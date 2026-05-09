// Package provider implements provider-agnostic inference backends.
package provider

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// InferenceRequest carries normalized parameters for any LLM provider.
type InferenceRequest struct {
	ModelID      string
	Prompt       string
	SystemPrompt string
	MaxTokens    int
	Temperature  float64
	TopP         float64
	Stream       bool
}

// InferenceResponse carries normalized output from any LLM provider.
type InferenceResponse struct {
	Content      string
	FinishReason string
	PromptTokens int
	CompletionTokens int
	TotalTokens  int
	Model        string
	Latency      time.Duration
}

// Provider is the abstraction over LLM inference backends.
type Provider interface {
	Complete(ctx context.Context, req InferenceRequest) (*InferenceResponse, error)
	Stream(ctx context.Context, req InferenceRequest, callback func(chunk string) error) (*InferenceResponse, error)
	HealthCheck(ctx context.Context) error
}

// ProviderFactory creates Provider instances from configuration.
type ProviderFactory struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewProviderFactory creates a factory with registered providers.
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{providers: make(map[string]Provider)}
}

// Register adds a provider under a service name.
func (f *ProviderFactory) Register(name string, p Provider) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.providers[name] = p
}

// Get returns the provider for a given name.
func (f *ProviderFactory) Get(name string) (Provider, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	p, ok := f.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider %s not found", name)
	}
	return p, nil
}
