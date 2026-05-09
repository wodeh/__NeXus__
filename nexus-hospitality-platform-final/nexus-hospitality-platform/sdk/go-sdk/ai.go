// sdk/go-sdk/ai.go
package nexus

import (
	"context"
)

type AIService struct {
	client *Client
}

func (s *AIService) ConciergeQuery(ctx context.Context, req *ConciergeRequest) (*ConciergeResponse, *Response, error) {
	r, err := s.client.newRequest(ctx, "POST", "/ai/concierge", req)
	if err != nil {
		return nil, nil, err
	}

	var resp ConciergeResponse
	response, err := s.client.do(r, &resp)
	if err != nil {
		return nil, response, err
	}

	return &resp, response, nil
}

func (s *AIService) AnalyzeSentiment(ctx context.Context, req *SentimentRequest) (*SentimentResponse, *Response, error) {
	r, err := s.client.newRequest(ctx, "POST", "/ai/sentiment", req)
	if err != nil {
		return nil, nil, err
	}

	var resp SentimentResponse
	response, err := s.client.do(r, &resp)
	if err != nil {
		return nil, response, err
	}

	return &resp, response, nil
}

func (s *AIService) ForecastOccupancy(ctx context.Context, req *ForecastRequest) (*ForecastResponse, *Response, error) {
	r, err := s.client.newRequest(ctx, "POST", "/ai/forecast/occupancy", req)
	if err != nil {
		return nil, nil, err
	}

	var resp ForecastResponse
	response, err := s.client.do(r, &resp)
	if err != nil {
		return nil, response, err
	}

	return &resp, response, nil
}

func (s *AIService) GetRecommendations(ctx context.Context, req *RecommendationRequest) (*RecommendationResponse, *Response, error) {
	r, err := s.client.newRequest(ctx, "POST", "/ai/recommendations", req)
	if err != nil {
		return nil, nil, err
	}

	var resp RecommendationResponse
	response, err := s.client.do(r, &resp)
	if err != nil {
		return nil, response, err
	}

	return &resp, response, nil
}

type ConciergeRequest struct {
	Prompt      string                 `json:"prompt"`
	GuestID     string                 `json:"guest_id,omitempty"`
	RoomID      string                 `json:"room_id,omitempty"`
	Language    string                 `json:"language,omitempty"`
	UseRAG      bool                   `json:"use_rag,omitempty"`
	MaxTokens   int                    `json:"max_tokens,omitempty"`
	Temperature float64                `json:"temperature,omitempty"`
}

type ConciergeResponse struct {
	ID           string `json:"id"`
	Content      string `json:"content"`
	FinishReason string `json:"finish_reason"`
	Sources      []struct {
		DocumentID string  `json:"document_id"`
		Content    string  `json:"content"`
		Score      float64 `json:"score"`
	} `json:"sources,omitempty"`
	LatencyMs int64 `json:"latency_ms"`
}

type SentimentRequest struct {
	Text     string `json:"text"`
	TenantID string `json:"tenant_id,omitempty"`
}

type SentimentResponse struct {
	Score       float64 `json:"score"`       // -1 to 1
	Confidence  float64 `json:"confidence"`
	Label       string  `json:"label"`       // positive, negative, neutral
	Aspects     []struct {
		Aspect    string  `json:"aspect"`
		Sentiment float64 `json:"sentiment"`
	} `json:"aspects,omitempty"`
}

type ForecastRequest struct {
	PropertyID   string `json:"property_id"`
	HorizonDays  int    `json:"horizon_days"`
	TenantID     string `json:"tenant_id,omitempty"`
}

type ForecastResponse struct {
	PropertyID       string    `json:"property_id"`
	Forecast         []struct {
		Date              string  `json:"date"`
		PredictedOccupancy float64 `json:"predicted_occupancy"`
		PredictedADR       float64 `json:"predicted_adr"`
		PredictedRevPAR    float64 `json:"predicted_revpar"`
		Confidence         float64 `json:"confidence"`
	} `json:"forecast"`
	Accuracy float64 `json:"accuracy"`
}

type RecommendationRequest struct {
	GuestID string                 `json:"guest_id"`
	Context map[string]interface{} `json:"context,omitempty"`
}

type RecommendationResponse struct {
	GuestID         string `json:"guest_id"`
	Recommendations []struct {
		Type        string  `json:"type"`        // dining, activity, service, upgrade
		Title       string  `json:"title"`
		Description string  `json:"description"`
		Score       float64 `json:"score"`
		ImageURL    string  `json:"image_url,omitempty"`
	} `json:"recommendations"`
}
