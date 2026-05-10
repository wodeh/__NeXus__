package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/nexus-platform/pms-integration/internal/domain"
)

// SearchIndexer defines the interface for search indexing.
type SearchIndexer interface {
	IndexReservation(ctx context.Context, res *domain.Reservation) error
	IndexGuest(ctx context.Context, guest *domain.Guest) error
	Search(ctx context.Context, tenantID, query string, types []string, limit int) ([]SearchResult, error)
	Delete(ctx context.Context, tenantID, docType, id string) error
}

// SearchResult represents a single search hit.
type SearchResult struct {
	ID        string
	Type      string // reservation, guest, room, folio
	Title     string
	Subtitle  string
	Score     float64
	Data      map[string]interface{}
}

// ElasticsearchIndexer is a placeholder for Elasticsearch integration.
// In production, this would use github.com/elastic/go-elasticsearch/v8
type ElasticsearchIndexer struct {
	address string
	index   string
}

// NewElasticsearchIndexer creates a new indexer.
func NewElasticsearchIndexer(address, index string) SearchIndexer {
	return &ElasticsearchIndexer{address: address, index: index}
}

func (e *ElasticsearchIndexer) IndexReservation(ctx context.Context, res *domain.Reservation) error {
	// Placeholder: In production, serialize and POST to ES
	fmt.Printf("[ES] Index reservation %s for tenant %s\n", res.ID, res.TenantID)
	return nil
}

func (e *ElasticsearchIndexer) IndexGuest(ctx context.Context, guest *domain.Guest) error {
	fmt.Printf("[ES] Index guest %s for tenant %s\n", guest.ID, guest.TenantID)
	return nil
}

func (e *ElasticsearchIndexer) Search(ctx context.Context, tenantID, query string, types []string, limit int) ([]SearchResult, error) {
	// Placeholder: Return mock results
	fmt.Printf("[ES] Search '%s' for tenant %s\n", query, tenantID)
	return []SearchResult{
		{ID: "1", Type: "guest", Title: "John Smith", Subtitle: "john@example.com", Score: 1.0, Data: map[string]interface{}{"email": "john@example.com"}},
		{ID: "2", Type: "reservation", Title: "RES-2024-0001", Subtitle: "John Smith - Room 101", Score: 0.9, Data: map[string]interface{}{"status": "checked_in"}},
	}, nil
}

func (e *ElasticsearchIndexer) Delete(ctx context.Context, tenantID, docType, id string) error {
	fmt.Printf("[ES] Delete %s/%s for tenant %s\n", docType, id, tenantID)
	return nil
}

// InMemorySearchIndexer is a development fallback.
type InMemorySearchIndexer struct {
	reservations map[string]*domain.Reservation
	guests       map[string]*domain.Guest
}

// NewInMemorySearchIndexer creates a new in-memory indexer.
func NewInMemorySearchIndexer() SearchIndexer {
	return &InMemorySearchIndexer{
		reservations: make(map[string]*domain.Reservation),
		guests:       make(map[string]*domain.Guest),
	}
}

func (m *InMemorySearchIndexer) IndexReservation(ctx context.Context, res *domain.Reservation) error {
	m.reservations[res.ID] = res
	return nil
}

func (m *InMemorySearchIndexer) IndexGuest(ctx context.Context, guest *domain.Guest) error {
	m.guests[guest.ID] = guest
	return nil
}

func (m *InMemorySearchIndexer) Search(ctx context.Context, tenantID, query string, types []string, limit int) ([]SearchResult, error) {
	var results []SearchResult
	for _, g := range m.guests {
		if g.TenantID == tenantID && (query == "" || contains(g.FirstName, query) || contains(g.LastName, query) || contains(g.Email, query)) {
			results = append(results, SearchResult{
				ID:       g.ID,
				Type:     "guest",
				Title:    fmt.Sprintf("%s %s", g.FirstName, g.LastName),
				Subtitle: g.Email,
				Data:     map[string]interface{}{"phone": g.Phone},
			})
		}
	}
	for _, res := range m.reservations {
		if res.TenantID == tenantID {
			results = append(results, SearchResult{
				ID:       res.ID,
				Type:     "reservation",
				Title:    fmt.Sprintf("Reservation %s", res.ID[:8]),
				Subtitle: fmt.Sprintf("%s - %s", res.CheckIn.Format("2006-01-02"), res.CheckOut.Format("2006-01-02")),
				Data:     map[string]interface{}{"status": res.Status, "total": res.TotalAmount},
			})
		}
	}
	return results, nil
}

func (m *InMemorySearchIndexer) Delete(ctx context.Context, tenantID, docType, id string) error {
	if docType == "guest" {
		delete(m.guests, id)
	} else if docType == "reservation" {
		delete(m.reservations, id)
	}
	return nil
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr))
}
