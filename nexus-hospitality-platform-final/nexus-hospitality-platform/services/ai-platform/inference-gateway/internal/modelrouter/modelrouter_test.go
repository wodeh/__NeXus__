package modelrouter

import (
	"testing"
)

func TestNew_EmptyConfig(t *testing.T) {
	r, err := New(Config{Backends: []BackendConfig{}})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	_, err = r.SelectBackend("any-model", "tenant-1")
	if err == nil {
		t.Error("expected error when no backends available")
	}
}

func TestSelectBackend_Fallback(t *testing.T) {
	r, err := New(Config{
		Backends: []BackendConfig{
			{ID: "b1", Type: "openai", Models: []string{"gpt-4"}},
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	b, err := r.SelectBackend("unknown-model", "tenant-1")
	if err != nil {
		t.Fatalf("SelectBackend error = %v", err)
	}
	if b.ID != "b1" {
		t.Errorf("backend ID = %q, want %q", b.ID, "b1")
	}
}

func TestSelectBackend_ByModelID(t *testing.T) {
	r, err := New(Config{
		Backends: []BackendConfig{
			{ID: "b1", Type: "openai", Models: []string{"gpt-4"}},
			{ID: "b2", Type: "anthropic", Models: []string{"claude-3"}},
		},
	})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	b, err := r.SelectBackend("claude-3", "tenant-1")
	if err != nil {
		t.Fatalf("SelectBackend error = %v", err)
	}
	if b.ID != "b2" {
		t.Errorf("backend ID = %q, want %q", b.ID, "b2")
	}
}

func TestRegisterBackend_Dynamic(t *testing.T) {
	r, _ := New(Config{Backends: []BackendConfig{}})
	r.RegisterBackend(&Backend{ID: "dyn", Type: "local", Models: []string{"llama-2"}})
	b, err := r.SelectBackend("llama-2", "tenant-1")
	if err != nil {
		t.Fatalf("SelectBackend error = %v", err)
	}
	if b.ID != "dyn" {
		t.Errorf("backend ID = %q, want %q", b.ID, "dyn")
	}
}
