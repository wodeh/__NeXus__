// Package provider implements the Anthropic inference backend.
package provider

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AnthropicProvider calls the Anthropic REST API.
type AnthropicProvider struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
	cb         *CircuitBreaker
}

// NewAnthropicProvider creates an Anthropic provider.
func NewAnthropicProvider(endpoint, apiKey string) *AnthropicProvider {
	if endpoint == "" {
		endpoint = "https://api.anthropic.com/v1"
	}
	return &AnthropicProvider{
		endpoint: endpoint,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		cb: NewCircuitBreaker("anthropic", 5, 30*time.Second),
	}
}

// Complete performs a synchronous completion.
func (p *AnthropicProvider) Complete(ctx context.Context, req InferenceRequest) (*InferenceResponse, error) {
	if err := p.cb.Allow(); err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"model":      req.ModelID,
		"max_tokens": req.MaxTokens,
		"messages":   []map[string]string{{"role": "user", "content": req.Prompt}},
	}
	if req.SystemPrompt != "" {
		body["system"] = req.SystemPrompt
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}
	if req.TopP > 0 {
		body["top_p"] = req.TopP
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/messages", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	start := time.Now()
	resp, err := p.httpClient.Do(httpReq)
	latency := time.Since(start)
	if err != nil {
		p.cb.RecordFailure()
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		p.cb.RecordFailure()
	} else {
		p.cb.RecordSuccess()
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		StopReason string `json:"stop_reason"`
		Usage      struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var content string
	for _, c := range result.Content {
		if c.Type == "text" {
			content += c.Text
		}
	}

	return &InferenceResponse{
		Content:          content,
		FinishReason:     result.StopReason,
		PromptTokens:     result.Usage.InputTokens,
		CompletionTokens: result.Usage.OutputTokens,
		TotalTokens:      result.Usage.InputTokens + result.Usage.OutputTokens,
		Model:            result.Model,
		Latency:          latency,
	}, nil
}

// Stream performs a streaming completion.
func (p *AnthropicProvider) Stream(ctx context.Context, req InferenceRequest, callback func(chunk string) error) (*InferenceResponse, error) {
	if err := p.cb.Allow(); err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"model":      req.ModelID,
		"max_tokens": req.MaxTokens,
		"messages":   []map[string]string{{"role": "user", "content": req.Prompt}},
		"stream":     true,
	}
	if req.SystemPrompt != "" {
		body["system"] = req.SystemPrompt
	}
	if req.Temperature > 0 {
		body["temperature"] = req.Temperature
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/messages", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")

	start := time.Now()
	resp, err := p.httpClient.Do(httpReq)
	latency := time.Since(start)
	if err != nil {
		p.cb.RecordFailure()
		return nil, fmt.Errorf("http do: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 500 {
		p.cb.RecordFailure()
	} else {
		p.cb.RecordSuccess()
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("anthropic stream error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var fullContent string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !bytes.HasPrefix([]byte(line), []byte("data: ")) {
			continue
		}
		data := bytes.TrimPrefix([]byte(line), []byte("data: "))

		var event struct {
			Type string `json:"type"`
			Delta struct {
				Text string `json:"text"`
			} `json:"delta"`
		}
		if err := json.Unmarshal(data, &event); err != nil {
			continue
		}
		if event.Type == "content_block_delta" {
			fullContent += event.Delta.Text
			if err := callback(event.Delta.Text); err != nil {
				return nil, fmt.Errorf("stream callback: %w", err)
			}
		}
	}

	return &InferenceResponse{
		Content:      fullContent,
		FinishReason: "stop",
		Model:        req.ModelID,
		Latency:      latency,
	}, nil
}

// HealthCheck verifies provider connectivity.
func (p *AnthropicProvider) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint+"/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %d", resp.StatusCode)
	}
	return nil
}
