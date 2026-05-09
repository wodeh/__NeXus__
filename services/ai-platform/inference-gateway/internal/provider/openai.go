// Package provider implements the OpenAI inference backend.
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

// OpenAIProvider calls the OpenAI REST API.
type OpenAIProvider struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
	cb         *CircuitBreaker
}

// NewOpenAIProvider creates an OpenAI provider.
func NewOpenAIProvider(endpoint, apiKey string) *OpenAIProvider {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}
	return &OpenAIProvider{
		endpoint: endpoint,
		apiKey:   apiKey,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		cb: NewCircuitBreaker("openai", 5, 30*time.Second),
	}
}

// Complete performs a synchronous completion.
func (p *OpenAIProvider) Complete(ctx context.Context, req InferenceRequest) (*InferenceResponse, error) {
	if err := p.cb.Allow(); err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"model":       req.ModelID,
		"messages":    []map[string]string{{"role": "user", "content": req.Prompt}},
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"top_p":       req.TopP,
	}
	if req.SystemPrompt != "" {
		body["messages"] = append([]map[string]string{{"role": "system", "content": req.SystemPrompt}}, body["messages"].([]map[string]string)...)
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

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
		return nil, fmt.Errorf("openai error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
		Model string `json:"model"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in openai response")
	}

	return &InferenceResponse{
		Content:          result.Choices[0].Message.Content,
		FinishReason:     result.Choices[0].FinishReason,
		PromptTokens:     result.Usage.PromptTokens,
		CompletionTokens: result.Usage.CompletionTokens,
		TotalTokens:      result.Usage.TotalTokens,
		Model:            result.Model,
		Latency:          latency,
	}, nil
}

// Stream performs a streaming completion.
func (p *OpenAIProvider) Stream(ctx context.Context, req InferenceRequest, callback func(chunk string) error) (*InferenceResponse, error) {
	if err := p.cb.Allow(); err != nil {
		return nil, err
	}

	body := map[string]interface{}{
		"model":       req.ModelID,
		"messages":    []map[string]string{{"role": "user", "content": req.Prompt}},
		"max_tokens":  req.MaxTokens,
		"temperature": req.Temperature,
		"top_p":       req.TopP,
		"stream":      true,
	}
	if req.SystemPrompt != "" {
		body["messages"] = append([]map[string]string{{"role": "system", "content": req.SystemPrompt}}, body["messages"].([]map[string]string)...)
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")
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
		return nil, fmt.Errorf("openai stream error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var fullContent string
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !bytes.HasPrefix([]byte(line), []byte("data: ")) {
			continue
		}
		data := bytes.TrimPrefix([]byte(line), []byte("data: "))
		if string(data) == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal(data, &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 {
			content := chunk.Choices[0].Delta.Content
			fullContent += content
			if err := callback(content); err != nil {
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
func (p *OpenAIProvider) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint+"/models", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
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
