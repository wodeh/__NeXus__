# nhc-ai-pricing-service
# Reinforcement Learning-based dynamic pricing for hospitality

import numpy as np
import tensorflow as tf
from typing import Dict, List, Optional
from dataclasses import dataclass
from datetime import datetime, timedelta
import redis
import json
import os
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import uvicorn

@dataclass
class PricingContext:
    property_id: str
    room_type_id: str
    check_in_date: datetime
    check_out_date: datetime
    current_occupancy: float
    competitor_rates: Dict[str, float]
    historical_demand: List[float]
    events: List[str]
    day_of_week: int
    season: str
    lead_time_days: int
    cancellation_rate: float
    guest_segment: str
    length_of_stay: int
    booking_window: str
    market_segment: str

class ReinforcementLearningPricingAgent:
    def __init__(self, model_path: str, redis_url: str):
        self.model = tf.saved_model.load(model_path)
        self.redis_client = redis.from_url(redis_url, decode_responses=True)
        self.min_price = 50.0
        self.max_multiplier = 3.0
        self.min_multiplier = 0.5

    def get_features(self, context: PricingContext) -> np.ndarray:
        cache_key = f"pricing:features:{context.property_id}:{context.room_type_id}:{context.check_in_date.strftime('%Y-%m-%d')}"
        cached = self.redis_client.get(cache_key)
        if cached:
            return np.array(json.loads(cached))
        features = self._compute_features(context)
        self.redis_client.setex(cache_key, 300, json.dumps(features.tolist()))
        return features

    def _compute_features(self, context: PricingContext) -> np.ndarray:
        comp_avg = np.mean(list(context.competitor_rates.values())) if context.competitor_rates else 200.0
        comp_min = np.min(list(context.competitor_rates.values())) if context.competitor_rates else 150.0
        comp_max = np.max(list(context.competitor_rates.values())) if context.competitor_rates else 300.0

        feature_vector = np.array([
            context.current_occupancy,
            comp_avg / 200.0, comp_min / 200.0, comp_max / 200.0,
            len(context.events) * 0.1,
            context.lead_time_days / 365.0,
            context.day_of_week / 7.0,
            context.cancellation_rate,
            context.length_of_stay / 30.0,
            1.0 if context.market_segment == "corporate" else 0.0,
            1.0 if context.market_segment == "leisure" else 0.0,
            1.0 if context.guest_segment == "vip" else 0.0,
            1.0 if context.season == "peak" else 0.0,
            1.0 if context.season == "low" else 0.0,
            1.0 if context.booking_window == "last_minute" else 0.0,
            1.0 if context.booking_window == "advance" else 0.0,
            np.mean(context.historical_demand[-7:]) if context.historical_demand else 0.5,
            np.std(context.historical_demand[-7:]) if len(context.historical_demand) >= 7 else 0.1,
            max(context.historical_demand[-7:]) if context.historical_demand else 1.0,
            min(context.historical_demand[-7:]) if context.historical_demand else 0.0,
        ])
        if len(feature_vector) < 50:
            feature_vector = np.pad(feature_vector, (0, 50 - len(feature_vector)))
        return feature_vector.reshape(1, -1)

    def predict_optimal_rate(self, context: PricingContext, base_rate: float) -> Dict:
        features = self.get_features(context)
        action = self.model(features)
        price_multiplier = float(action[0][0])

        constraints = []
        if context.current_occupancy > 0.9 and price_multiplier > 2.0:
            price_multiplier = 2.0
            constraints.append("max_increase_high_demand")
        if context.lead_time_days < 2 and price_multiplier < 0.8:
            price_multiplier = 0.8
            constraints.append("min_price_last_minute")
        if context.competitor_rates:
            comp_avg = np.mean(list(context.competitor_rates.values()))
            max_parity = comp_avg * 1.05
            recommended = base_rate * price_multiplier
            if recommended > max_parity:
                price_multiplier = max_parity / base_rate
                constraints.append("rate_parity_cap")
        if context.length_of_stay >= 7:
            price_multiplier *= 0.9
            constraints.append("weekly_discount")
        if context.market_segment == "corporate" and price_multiplier > 1.5:
            price_multiplier = 1.5
            constraints.append("corporate_rate_ceiling")
        if context.lead_time_days > 30 and price_multiplier > 1.2:
            price_multiplier *= 0.95
            constraints.append("advance_purchase_discount")

        price_multiplier = max(self.min_multiplier, min(self.max_multiplier, price_multiplier))
        recommended_rate = round(base_rate * price_multiplier, 2)
        expected_occupancy = self._estimate_occupancy(recommended_rate, context)
        expected_revenue = recommended_rate * expected_occupancy * context.length_of_stay

        return {
            "recommended_rate": recommended_rate,
            "confidence": 0.92,
            "price_multiplier": round(price_multiplier, 2),
            "expected_revenue": round(expected_revenue, 2),
            "expected_occupancy": round(expected_occupancy, 3),
            "constraints_applied": constraints,
            "model_version": "rl_pricing_v3.2",
            "computed_at": datetime.utcnow().isoformat()
        }

    def _estimate_occupancy(self, rate: float, context: PricingContext) -> float:
        base_demand = 0.7
        price_sensitivity = -0.3
        rate_ratio = rate / 200.0
        segment_multiplier = 1.0
        if context.market_segment == "corporate":
            segment_multiplier = 0.9
        elif context.market_segment == "leisure":
            segment_multiplier = 1.1
        occupancy = base_demand + (price_sensitivity * (rate_ratio - 1)) * segment_multiplier
        return max(0.1, min(0.95, occupancy))

app = FastAPI(title="NHC AI Pricing Service", version="3.2.0")
agent = ReinforcementLearningPricingAgent(
    model_path=os.getenv("MODEL_PATH", "/models/pricing-v3.2"),
    redis_url=os.getenv("REDIS_URL", "redis://localhost:6379")
)

class PricingRequest(BaseModel):
    property_id: str
    room_type_id: str
    check_in: str
    check_out: str
    base_rate: float
    current_occupancy: float = 0.7
    competitor_rates: Optional[Dict[str, float]] = None
    market_segment: str = "leisure"
    guest_segment: str = "standard"
    events: Optional[List[str]] = None

class PricingResponse(BaseModel):
    recommended_rate: float
    confidence: float
    price_multiplier: float
    expected_revenue: float
    expected_occupancy: float
    constraints_applied: List[str]
    model_version: str
    computed_at: str

@app.post("/api/v1/pricing/optimize", response_model=PricingResponse)
async def optimize_pricing(request: PricingRequest):
    try:
        check_in = datetime.fromisoformat(request.check_in)
        check_out = datetime.fromisoformat(request.check_out)
        lead_time = (check_in - datetime.now()).days
        los = (check_out - check_in).days

        context = PricingContext(
            property_id=request.property_id,
            room_type_id=request.room_type_id,
            check_in_date=check_in,
            check_out_date=check_out,
            current_occupancy=request.current_occupancy,
            competitor_rates=request.competitor_rates or {},
            historical_demand=[0.6, 0.7, 0.8, 0.75, 0.9, 0.85, 0.7],
            events=request.events or [],
            day_of_week=check_in.weekday(),
            season="peak" if check_in.month in [6, 7, 8, 12] else "low",
            lead_time_days=max(0, lead_time),
            cancellation_rate=0.15,
            guest_segment=request.guest_segment,
            length_of_stay=los,
            booking_window="advance" if lead_time > 14 else "last_minute",
            market_segment=request.market_segment
        )

        result = agent.predict_optimal_rate(context, request.base_rate)
        return PricingResponse(**result)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
async def health_check():
    return {"status": "healthy", "model_version": "rl_pricing_v3.2"}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8090)
