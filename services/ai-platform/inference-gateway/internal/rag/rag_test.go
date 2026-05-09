package rag

import (
	"context"
	"testing"
)

func TestNew(t *testing.T) {
	e, err := New(Config{VectorStoreURL: "http://localhost:8080"})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestRetrieve_NotConfigured(t *testing.T) {
	e, _ := New(Config{VectorStoreURL: ""})
	_, err := e.Retrieve(context.Background(), "prompt", nil)
	if err == nil {
		t.Error("expected error when vector store not configured")
	}
}
