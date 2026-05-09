package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/ai-platform/inference-gateway/internal/config"
	"github.com/nexus-platform/ai-platform/inference-gateway/internal/guardrails"
	"github.com/nexus-platform/ai-platform/inference-gateway/internal/metrics"
	"github.com/nexus-platform/ai-platform/inference-gateway/internal/modelrouter"
	"github.com/nexus-platform/ai-platform/inference-gateway/internal/provider"
	"github.com/nexus-platform/ai-platform/inference-gateway/internal/rag"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("ai-inference-gateway")

type InferenceRequest struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	ModelID      string                 `json:"model_id"`
	Prompt       string                 `json:"prompt"`
	SystemPrompt string                 `json:"system_prompt,omitempty"`
	MaxTokens    int                    `json:"max_tokens,omitempty"`
	Temperature  float64                `json:"temperature,omitempty"`
	TopP         float64                `json:"top_p,omitempty"`
	Stream       bool                   `json:"stream,omitempty"`
	Context      map[string]interface{} `json:"context,omitempty"`
	UseRAG       bool                   `json:"use_rag,omitempty"`
	RAGConfig    *rag.Config            `json:"rag_config,omitempty"`
	GuestID      string                 `json:"guest_id,omitempty"`
	SessionID    string                 `json:"session_id,omitempty"`
}

type InferenceResponse struct {
	ID           string          `json:"id"`
	RequestID    string          `json:"request_id"`
	Content      string          `json:"content"`
	FinishReason string          `json:"finish_reason"`
	Usage        TokenUsage      `json:"usage"`
	Model        string          `json:"model"`
	Guardrails   GuardrailResult `json:"guardrails"`
	RAGSources   []rag.Source    `json:"rag_sources,omitempty"`
	LatencyMs    int64           `json:"latency_ms"`
}

type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type GuardrailResult struct {
	Passed     bool     `json:"passed"`
	Violations []string `json:"violations,omitempty"`
	Checks     []string `json:"checks"`
}

type InferenceGateway struct {
	config         *config.Config
	router         *modelrouter.Router
	guardrails     *guardrails.Engine
	ragEngine      *rag.Engine
	metrics        *metrics.Collector
	rateLimiter    *RateLimiter
	cache          *ResponseCache
	providerFactory *provider.ProviderFactory
	mu             sync.RWMutex
	activeRequests map[string]*RequestContext
}

type RequestContext struct {
	Request   InferenceRequest
	StartTime time.Time
	Cancel    context.CancelFunc
}

type RateLimiter struct {
	limits map[string]*TenantLimit
	mu     sync.RWMutex
}

type TenantLimit struct {
	TenantID       string
	RequestsPerMin int
	BurstSize      int
	LastRequest    time.Time
	Tokens         float64
}

type ResponseCache struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
}

type CacheEntry struct {
	Response  *InferenceResponse
	Timestamp time.Time
}

func NewInferenceGateway(cfg *config.Config) (*InferenceGateway, error) {
	router, err := modelrouter.New(cfg.Models)
	if err != nil {
		return nil, fmt.Errorf("model router init failed: %w", err)
	}

	guardrailEngine, err := guardrails.New(cfg.Guardrails)
	if err != nil {
		return nil, fmt.Errorf("guardrails init failed: %w", err)
	}

	ragEngine, err := rag.New(cfg.RAG)
	if err != nil {
		return nil, fmt.Errorf("RAG engine init failed: %w", err)
	}

	factory := provider.NewProviderFactory()
	for _, bc := range cfg.Models.Backends {
		switch bc.Type {
		case "openai":
			factory.Register(bc.ID, provider.NewOpenAIProvider(bc.Endpoint, bc.APIKey))
		case "anthropic":
			factory.Register(bc.ID, provider.NewAnthropicProvider(bc.Endpoint, bc.APIKey))
		}
	}

	return &InferenceGateway{
		config:          cfg,
		router:          router,
		guardrails:      guardrailEngine,
		ragEngine:       ragEngine,
		metrics:         metrics.NewCollector(metrics.Config{Enabled: cfg.Metrics.Enabled, Port: cfg.Metrics.Port}),
		rateLimiter:     NewRateLimiter(),
		cache:           NewResponseCache(cfg.CacheTTL),
		providerFactory: factory,
		activeRequests:  make(map[string]*RequestContext),
	}, nil
}

func (g *InferenceGateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "inference-request",
		trace.WithAttributes(attribute.String("http.method", r.Method)))
	defer span.End()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InferenceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		span.RecordError(err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	span.SetAttributes(
		attribute.String("request.id", req.ID),
		attribute.String("tenant.id", req.TenantID),
		attribute.String("model.id", req.ModelID),
	)

	if !g.rateLimiter.Allow(req.TenantID) {
		g.metrics.IncrementCounter("inference_rate_limited_total", req.TenantID)
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	if cached := g.cache.Get(req.TenantID, req.Prompt, req.ModelID); cached != nil {
		g.metrics.IncrementCounter("inference_cache_hit_total", req.TenantID)
		respondJSON(w, cached)
		return
	}

	guardResult, err := g.guardrails.CheckInput(ctx, req.Prompt, req.TenantID)
	if err != nil {
		span.RecordError(err)
		log.Printf("Guardrail check failed: %v", err)
	}
	if !guardResult.Passed {
		g.metrics.IncrementCounter("inference_guardrail_blocked_total", req.TenantID)
		http.Error(w, fmt.Sprintf("Content blocked: %s", strings.Join(guardResult.Violations, ", ")), http.StatusBadRequest)
		return
	}

	var ragSources []rag.Source
	if req.UseRAG && req.RAGConfig != nil {
		sources, err := g.ragEngine.Retrieve(ctx, req.Prompt, req.RAGConfig)
		if err != nil {
			log.Printf("RAG retrieval failed: %v", err)
		} else {
			ragSources = sources
			req.Prompt = g.augmentPromptWithRAG(req.Prompt, sources)
		}
	}

	backend, err := g.router.SelectBackend(req.ModelID, req.TenantID)
	if err != nil {
		span.RecordError(err)
		http.Error(w, fmt.Sprintf("Model routing failed: %v", err), http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	reqCtx := &RequestContext{
		Request:   req,
		StartTime: time.Now(),
		Cancel:    cancel,
	}

	g.mu.Lock()
	g.activeRequests[req.ID] = reqCtx
	g.mu.Unlock()

	response, err := g.executeInference(ctx, backend, req)
	if err != nil {
		g.mu.Lock()
		delete(g.activeRequests, req.ID)
		g.mu.Unlock()

		span.RecordError(err)
		g.metrics.IncrementCounter("inference_requests_total", req.TenantID, "POST", "error")
		http.Error(w, fmt.Sprintf("Inference failed: %v", err), http.StatusInternalServerError)
		return
	}

	g.mu.Lock()
	delete(g.activeRequests, req.ID)
	g.mu.Unlock()

	postGuardResult, err := g.guardrails.CheckOutput(ctx, response.Content, req.Prompt, ragSources)
	if err != nil {
		log.Printf("Post-inference guardrail check failed: %v", err)
	}
	response.Guardrails = GuardrailResult{
		Passed:     postGuardResult.Passed,
		Violations: postGuardResult.Violations,
		Checks:     postGuardResult.Checks,
	}

	if !postGuardResult.Passed {
		g.metrics.IncrementCounter("inference_hallucination_detected_total", req.TenantID)
		log.Printf("Hallucination detected for request %s: %v", req.ID, postGuardResult.Violations)
	}

	response.RAGSources = ragSources
	response.LatencyMs = time.Since(reqCtx.StartTime).Milliseconds()

	if response.FinishReason == "stop" && len(response.Content) > 0 {
		g.cache.Set(req.TenantID, req.Prompt, req.ModelID, response)
	}

	g.metrics.IncrementCounter("inference_requests_total", req.TenantID, "POST", "success")
	g.metrics.HistogramObserve("inference_latency_seconds", float64(response.LatencyMs)/1000.0, req.TenantID)
	g.metrics.HistogramObserve("inference_tokens_total", float64(response.Usage.TotalTokens), req.TenantID)

	span.SetAttributes(
		attribute.Int64("inference.latency_ms", response.LatencyMs),
		attribute.Int("inference.tokens", response.Usage.TotalTokens),
		attribute.Bool("inference.hallucination_detected", !postGuardResult.Passed),
	)

	respondJSON(w, response)
}

func (g *InferenceGateway) executeInference(ctx context.Context, backend *modelrouter.Backend, req InferenceRequest) (*InferenceResponse, error) {
	p, err := g.providerFactory.Get(backend.ID)
	if err != nil {
		return nil, fmt.Errorf("provider not found: %w", err)
	}

	pReq := provider.InferenceRequest{
		ModelID:      req.ModelID,
		Prompt:       req.Prompt,
		SystemPrompt: req.SystemPrompt,
		MaxTokens:    req.MaxTokens,
		Temperature:  req.Temperature,
		TopP:         req.TopP,
		Stream:       req.Stream,
	}

	var result *provider.InferenceResponse
	if req.Stream {
		result, err = p.Stream(ctx, pReq, func(chunk string) error {
			// Streaming chunks are collected; in production, this would write to an SSE stream.
			return nil
		})
	} else {
		result, err = p.Complete(ctx, pReq)
	}
	if err != nil {
		return nil, err
	}

	return &InferenceResponse{
		ID:           req.ID,
		RequestID:    req.ID,
		Content:      result.Content,
		FinishReason: result.FinishReason,
		Usage: TokenUsage{
			PromptTokens:     result.PromptTokens,
			CompletionTokens: result.CompletionTokens,
			TotalTokens:      result.TotalTokens,
		},
		Model:     result.Model,
		LatencyMs: result.Latency.Milliseconds(),
	}, nil
}

func (g *InferenceGateway) augmentPromptWithRAG(prompt string, sources []rag.Source) string {
	var context strings.Builder
	context.WriteString("Use the following information to answer the question. If the answer is not in the provided information, say you don't know.\n\n")
	for i, source := range sources {
		context.WriteString(fmt.Sprintf("[%d] %s\n", i+1, source.Content))
	}
	context.WriteString(fmt.Sprintf("\nQuestion: %s\nAnswer:", prompt))
	return context.String()
}

func respondJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limits: make(map[string]*TenantLimit),
	}
}

func (rl *RateLimiter) Allow(tenantID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limit, exists := rl.limits[tenantID]
	if !exists {
		limit = &TenantLimit{
			TenantID:       tenantID,
			RequestsPerMin: 1000,
			BurstSize:      100,
			Tokens:         100,
		}
		rl.limits[tenantID] = limit
	}

	now := time.Now()
	elapsed := now.Sub(limit.LastRequest).Minutes()
	limit.LastRequest = now

	limit.Tokens += elapsed * float64(limit.RequestsPerMin)
	if limit.Tokens > float64(limit.BurstSize) {
		limit.Tokens = float64(limit.BurstSize)
	}

	if limit.Tokens >= 1 {
		limit.Tokens--
		return true
	}
	return false
}

func NewResponseCache(ttl time.Duration) *ResponseCache {
	c := &ResponseCache{
		entries: make(map[string]*CacheEntry),
		ttl:     ttl,
	}
	go c.cleanup()
	return c
}

func (c *ResponseCache) Get(tenantID, prompt, modelID string) *InferenceResponse {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := c.key(tenantID, prompt, modelID)
	entry, exists := c.entries[key]
	if !exists || time.Since(entry.Timestamp) > c.ttl {
		return nil
	}
	return entry.Response
}

func (c *ResponseCache) Set(tenantID, prompt, modelID string, response *InferenceResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := c.key(tenantID, prompt, modelID)
	c.entries[key] = &CacheEntry{
		Response:  response,
		Timestamp: time.Now(),
	}
}

func (c *ResponseCache) key(tenantID, prompt, modelID string) string {
	return fmt.Sprintf("%s:%s:%s", tenantID, modelID, prompt[:min(len(prompt), 100)])
}

func (c *ResponseCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.Sub(entry.Timestamp) > c.ttl {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	gateway, err := NewInferenceGateway(cfg)
	if err != nil {
		log.Fatalf("Failed to create gateway: %v", err)
	}

	http.Handle("/v1/inference", gateway)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "healthy", "service": "ai-platform"})
	})
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	http.HandleFunc("/live", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "alive"})
	})

	log.Printf("AI Inference Gateway starting on :%s", cfg.HTTPPort)
	log.Fatal(http.ListenAndServe(":"+cfg.HTTPPort, nil))
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
