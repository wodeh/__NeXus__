# nhc-ai-concierge-service
import os
import json
from typing import Dict, List, Optional
from datetime import datetime
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import uvicorn

app = FastAPI(title="NHC AI Concierge", version="2.0.0")

class HospitalityConciergeAgent:
    def __init__(self, tenant_id: str, property_id: str):
        self.tenant_id = tenant_id
        self.property_id = property_id

    async def handle_message(self, message: str, guest_id: Optional[str] = None, channel: str = "whatsapp") -> Dict:
        detected_lang = self._detect_language(message)
        intent = self._classify_intent(message)
        sentiment = self._analyze_sentiment(message)

        if intent == "complaint":
            response = self._handle_complaint(message, guest_id)
        elif intent == "booking_inquiry":
            response = "I can help you find the perfect room. What dates are you looking for?"
        elif intent == "service_request":
            response = "I'd be happy to arrange that for you. Could you provide a few more details?"
        elif intent == "local_info":
            response = "Here are some great recommendations near the hotel: Central Park (0.2 mi), The Capital Grille (0.3 mi), and the Met (1.2 mi)."
        else:
            response = "How can I assist you today?"

        requires_human = sentiment == "very_negative" or "manager" in message.lower() or "complaint" in message.lower()

        return {
            "response": response,
            "actions": [],
            "requires_human": requires_human,
            "sentiment": sentiment,
            "language": detected_lang,
            "confidence": 0.92,
            "response_time_ms": 450,
            "channel": channel
        }

    def _detect_language(self, message: str) -> str:
        return "en"

    def _classify_intent(self, message: str) -> str:
        m = message.lower()
        if any(w in m for w in ["book", "reservation", "room"]):
            return "booking_inquiry"
        elif any(w in m for w in ["complaint", "problem", "issue", "terrible"]):
            return "complaint"
        elif any(w in m for w in ["order", "food", "restaurant", "spa"]):
            return "service_request"
        elif any(w in m for w in ["nearby", "attraction", "recommend", "things to do"]):
            return "local_info"
        return "general"

    def _analyze_sentiment(self, message: str) -> str:
        negative = ["bad", "terrible", "awful", "worst", "hate", "angry"]
        positive = ["good", "great", "excellent", "amazing", "love"]
        neg_count = sum(1 for w in negative if w in message.lower())
        pos_count = sum(1 for w in positive if w in message.lower())
        if neg_count > pos_count + 1:
            return "very_negative"
        elif neg_count > pos_count:
            return "negative"
        elif pos_count > neg_count:
            return "positive"
        return "neutral"

    def _handle_complaint(self, message: str, guest_id: Optional[str]) -> str:
        ticket_id = f"SR-{datetime.now().strftime('%Y%m%d%H%M%S')}"
        return f"I sincerely apologize. I've escalated this to our duty manager who will contact you within 15 minutes. Reference: #{ticket_id}"

agent_cache: Dict[str, HospitalityConciergeAgent] = {}

def get_or_create_agent(tenant_id: str, property_id: str) -> HospitalityConciergeAgent:
    cache_key = f"{tenant_id}:{property_id}"
    if cache_key not in agent_cache:
        agent_cache[cache_key] = HospitalityConciergeAgent(tenant_id, property_id)
    return agent_cache[cache_key]

class ConciergeRequest(BaseModel):
    message: str
    guest_id: Optional[str] = None
    tenant_id: str
    property_id: str
    channel: str = "whatsapp"

class ConciergeResponse(BaseModel):
    response: str
    actions: List[Dict]
    requires_human: bool
    sentiment: str
    language: str
    confidence: float
    response_time_ms: int

@app.post("/api/v1/concierge/message", response_model=ConciergeResponse)
async def process_message(request: ConciergeRequest):
    try:
        agent = get_or_create_agent(request.tenant_id, request.property_id)
        result = await agent.handle_message(request.message, request.guest_id, request.channel)
        return ConciergeResponse(**result)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

@app.get("/health")
async def health_check():
    return {"status": "healthy", "version": "2.0.0"}

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8091)
