# nhc-fraud-detection-service
import numpy as np
from typing import Dict, List, Optional
from datetime import datetime
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import redis
import joblib
import os
import uvicorn

app = FastAPI(title="NHC Fraud Detection", version="1.0.0")

class FraudDetectionEngine:
    def __init__(self, model_path: str, redis_url: str):
        self.model = joblib.load(model_path)
        self.redis = redis.from_url(redis_url, decode_responses=True)
        self.risk_threshold = 0.7
        self.block_threshold = 0.9

    def analyze_transaction(self, transaction: Dict) -> Dict:
        features = self._extract_features(transaction)
        ml_score = self._ml_score(features)
        rule_score = self._rule_based_score(transaction)
        combined_score = 0.6 * ml_score + 0.4 * rule_score

        if combined_score >= self.block_threshold:
            risk_level, action = "block", "DECLINE"
        elif combined_score >= self.risk_threshold:
            risk_level, action = "high", "REVIEW"
        elif combined_score >= 0.4:
            risk_level, action = "medium", "MONITOR"
        else:
            risk_level, action = "low", "APPROVE"

        self._store_transaction(transaction, combined_score)

        return {
            "transaction_id": transaction.get("transaction_id"),
            "fraud_score": round(combined_score, 3),
            "risk_level": risk_level,
            "action": action,
            "ml_score": round(ml_score, 3),
            "rule_score": round(rule_score, 3),
            "flags": self._generate_flags(transaction, combined_score),
            "analyzed_at": datetime.utcnow().isoformat()
        }

    def _extract_features(self, transaction: Dict) -> np.ndarray:
        card_fp = transaction.get("card_fingerprint", "")
        email = transaction.get("email", "")
        ip = transaction.get("ip_address", "")

        card_txns_1h = int(self.redis.get(f"fraud:card:{card_fp}:1h") or 0)
        card_txns_24h = int(self.redis.get(f"fraud:card:{card_fp}:24h") or 0)
        email_txns_24h = int(self.redis.get(f"fraud:email:{email}:24h") or 0)
        ip_txns_1h = int(self.redis.get(f"fraud:ip:{ip}:1h") or 0)

        features = np.array([
            transaction.get("amount", 0) / 1000.0,
            card_txns_1h, card_txns_24h, email_txns_24h, ip_txns_1h,
            1.0 if transaction.get("is_mobile") else 0.0,
            1.0 if transaction.get("guest_id") is None else 0.0,
            len(transaction.get("billing_address", "")) / 100.0,
            1.0 if transaction.get("is_vpn") else 0.0,
            1.0 if transaction.get("is_tor") else 0.0,
            transaction.get("typing_speed", 200) / 500.0,
            transaction.get("time_on_page", 60) / 300.0,
        ])
        return features.reshape(1, -1)

    def _ml_score(self, features: np.ndarray) -> float:
        try:
            return float(self.model.predict_proba(features)[0][1])
        except:
            return 0.5

    def _rule_based_score(self, transaction: Dict) -> float:
        score = 0.0
        if transaction.get("amount", 0) > 5000:
            score += 0.3
        card_fp = transaction.get("card_fingerprint", "")
        card_bookings = int(self.redis.get(f"fraud:card:{card_fp}:bookings") or 0)
        if card_bookings > 3:
            score += 0.2
        email = transaction.get("email", "")
        suspicious = ["tempmail", "10minutemail", "guerrillamail"]
        if any(d in email for d in suspicious):
            score += 0.4
        if transaction.get("is_vpn"):
            score += 0.15
        if transaction.get("is_tor"):
            score += 0.5
        if transaction.get("billing_country") != transaction.get("booking_country"):
            score += 0.2
        if transaction.get("time_on_page", 60) < 10:
            score += 0.3
        return min(1.0, score)

    def _generate_flags(self, transaction: Dict, score: float) -> List[str]:
        flags = []
        if transaction.get("is_vpn"):
            flags.append("VPN_DETECTED")
        if transaction.get("is_tor"):
            flags.append("TOR_EXIT_NODE")
        if transaction.get("amount", 0) > 5000:
            flags.append("HIGH_VALUE_TRANSACTION")
        if transaction.get("time_on_page", 60) < 10:
            flags.append("RAPID_BOOKING")
        if transaction.get("billing_country") != transaction.get("booking_country"):
            flags.append("GEO_MISMATCH")
        if score > 0.7:
            flags.append("HIGH_RISK_SCORE")
        return flags

    def _store_transaction(self, transaction: Dict, score: float):
        card_fp = transaction.get("card_fingerprint", "")
        email = transaction.get("email", "")
        ip = transaction.get("ip_address", "")
        pipe = self.redis.pipeline()
        pipe.incr(f"fraud:card:{card_fp}:1h")
        pipe.expire(f"fraud:card:{card_fp}:1h", 3600)
        pipe.incr(f"fraud:card:{card_fp}:24h")
        pipe.expire(f"fraud:card:{card_fp}:24h", 86400)
        pipe.incr(f"fraud:email:{email}:24h")
        pipe.expire(f"fraud:email:{email}:24h", 86400)
        pipe.incr(f"fraud:ip:{ip}:1h")
        pipe.expire(f"fraud:ip:{ip}:1h", 3600)
        pipe.execute()

engine = FraudDetectionEngine(
    model_path="/models/fraud-xgboost-v1.pkl",
    redis_url=os.getenv("REDIS_URL", "redis://localhost:6379")
)

class FraudCheckRequest(BaseModel):
    transaction_id: str
    guest_id: Optional[str] = None
    amount: float
    currency: str = "USD"
    card_fingerprint: Optional[str] = None
    email: Optional[str] = None
    ip_address: Optional[str] = None
    billing_country: Optional[str] = None
    booking_country: Optional[str] = None
    is_mobile: bool = False
    is_vpn: bool = False
    is_tor: bool = False
    time_on_page: int = 60

class FraudCheckResponse(BaseModel):
    transaction_id: str
    fraud_score: float
    risk_level: str
    action: str
    flags: List[str]
    analyzed_at: str

@app.post("/api/v1/fraud/check", response_model=FraudCheckResponse)
async def check_fraud(request: FraudCheckRequest):
    try:
        transaction = request.dict()
        result = engine.analyze_transaction(transaction)
        return FraudCheckResponse(**result)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
async def health_check():
    return {"status": "healthy", "version": "1.0.0"}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8093)
