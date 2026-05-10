package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/nexus-platform/pms-integration/internal/domain"
)

func (h *Handler) runNightAudit(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	propertyID := r.URL.Query().Get("property_id")

	// Simulate night audit run
	run := &domain.NightAuditRun{
		ID:       generateID(),
		TenantID: tenantID,
		RunDate:  time.Now(),
		Status:   "completed",
	}
	if propertyID != "" {
		run.PropertyID = &propertyID
	}

	now := time.Now()
	run.StartedAt = &now
	completed := now.Add(2 * time.Minute)
	run.CompletedAt = &completed

	// Fetch today's arrivals/departures
	reservations, err := h.store.Reservations.List(r.Context(), tenantID, "confirmed", propertyID, 1000, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	today := time.Now().Truncate(24 * time.Hour)
	for _, res := range reservations {
		if res.CheckIn.Truncate(24*time.Hour).Equal(today) {
			run.ArrivalsProcessed++
		}
		if res.CheckOut.Truncate(24*time.Hour).Equal(today) {
			run.DeparturesProcessed++
		}
		if res.CheckIn.Before(now) && res.Status == "confirmed" {
			// No-show detection
			run.NoShowsProcessed++
		}
		run.RevenueTotal += res.TotalAmount
	}

	respondJSON(w, http.StatusOK, run)
}

func (h *Handler) getNightAuditStatus(w http.ResponseWriter, r *http.Request) {
	// Return last night audit run status
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":        "idle",
		"last_run":      time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
		"next_scheduled": time.Now().Add(24 * time.Hour).Format(time.RFC3339),
	})
}

func generateID() string {
	// Simple ID generation; production uses UUID
	return time.Now().Format("20060102150405") + "-audit"
}
