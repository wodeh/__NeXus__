package config

import (
	"context"
	"fmt"
)

// VaultClient abstracts HashiCorp Vault operations for secret retrieval and lease management.
type VaultClient interface {
	GetSecret(ctx context.Context, path string) (map[string]string, error)
	RenewLease(ctx context.Context, leaseID string) error
}

// MockVaultClient is a no-op implementation suitable for development and CI.
type MockVaultClient struct {
	Secrets map[string]map[string]string
}

// NewMockVaultClient creates a mock vault with pre-populated secrets.
func NewMockVaultClient() *MockVaultClient {
	return &MockVaultClient{Secrets: make(map[string]map[string]string)}
}

// GetSecret returns a secret from the mock store.
func (m *MockVaultClient) GetSecret(ctx context.Context, path string) (map[string]string, error) {
	_ = ctx
	if s, ok := m.Secrets[path]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("secret not found: %s", path)
}

// RenewLease is a no-op for the mock.
func (m *MockVaultClient) RenewLease(ctx context.Context, leaseID string) error {
	_ = ctx
	_ = leaseID
	return nil
}
