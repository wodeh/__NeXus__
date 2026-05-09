package domain

import (
	"testing"
	"time"
)

func TestFolio_AddChargeAndPayment(t *testing.T) {
	f := NewFolio("tenant-1", "prop-1", "res-1", "guest-1", "USD")

	if err := f.AddCharge(Charge{ID: "ch-1", Description: "Room rate", Amount: 200.0, PostedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("add charge: %v", err)
	}
	if f.Balance != 200.0 {
		t.Fatalf("expected balance 200, got %f", f.Balance)
	}

	if err := f.AddPayment(Payment{ID: "pay-1", Method: "cc", Amount: 100.0, PostedAt: time.Now().UTC()}); err != nil {
		t.Fatalf("add payment: %v", err)
	}
	if f.Balance != 100.0 {
		t.Fatalf("expected balance 100, got %f", f.Balance)
	}
}

func TestFolio_Close(t *testing.T) {
	f := NewFolio("tenant-1", "prop-1", "res-1", "guest-1", "USD")
	if err := f.Close(); err != nil {
		t.Fatalf("close folio: %v", err)
	}
	if f.Status != FolioStatusClosed {
		t.Fatalf("expected closed, got %s", f.Status)
	}
	if err := f.AddCharge(Charge{Amount: 10}); err == nil {
		t.Fatal("expected error adding charge to closed folio")
	}
}
