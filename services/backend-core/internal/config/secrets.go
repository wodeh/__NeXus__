package config

import (
	"context"
	"fmt"
	"os"
	"sync"
)

// SecretResolver resolves configuration values with strict precedence:
//   1. Runtime override
//   2. Vault secret (if vaultPath provided)
//   3. Environment variable
//   4. Secure default (fallback)
type SecretResolver struct {
	vault     VaultClient
	overrides map[string]string
	mu        sync.RWMutex
	metrics   *ConfigMetrics
}

// NewSecretResolver creates a resolver backed by Vault.
// If vault is nil, resolution skips the Vault layer.
func NewSecretResolver(vault VaultClient) *SecretResolver {
	return &SecretResolver{
		vault:     vault,
		overrides: make(map[string]string),
		metrics:   NewConfigMetrics(nil),
	}
}

// SetOverride injects a runtime configuration value.
func (r *SecretResolver) SetOverride(key, value string) {
	r.mu.Lock()
	r.overrides[key] = value
	r.mu.Unlock()
}

// Resolve returns the configuration value for key using the precedence chain.
func (r *SecretResolver) Resolve(ctx context.Context, key string, vaultPath string, fallback string) (string, error) {
	// 1. Runtime override
	r.mu.RLock()
	if v, ok := r.overrides[key]; ok {
		r.mu.RUnlock()
		r.metrics.IncResolved(key, "override")
		return v, nil
	}
	r.mu.RUnlock()

	// 2. Vault secret
	if vaultPath != "" && r.vault != nil {
		if secrets, err := r.vault.GetSecret(ctx, vaultPath); err == nil {
			if v, ok := secrets[key]; ok {
				r.metrics.IncResolved(key, "vault")
				return v, nil
			}
		}
	}

	// 3. Environment variable
	if v := os.Getenv(key); v != "" {
		r.metrics.IncResolved(key, "env")
		return v, nil
	}

	// 4. Secure default
	if fallback != "" {
		r.metrics.IncResolved(key, "default")
		return fallback, nil
	}

	return "", fmt.Errorf("required config %s not found in any source", key)
}

// ResolveRequired is like Resolve but returns an error if no value is found.
func (r *SecretResolver) ResolveRequired(ctx context.Context, key string, vaultPath string) (string, error) {
	return r.Resolve(ctx, key, vaultPath, "")
}
