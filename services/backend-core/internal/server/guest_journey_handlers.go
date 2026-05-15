package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

func (s *Server) registerGuestJourneyHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/guest-journey/", s.withTenant(s.handleGuestJourney))
	mux.HandleFunc("/v1/guest-journey/events", s.withTenant(s.handleGuestJourneyEvents))
}

func (s *Server) handleGuestJourney(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	
	// Extract reservation ID from path: /v1/guest-journey/:id
	path := r.URL.Path[len("/v1/guest-journey/"):]
	if path == "" || path == "events" {
		writeJSONError(w, http.StatusBadRequest, "missing reservation id")
		return
	}
	
	reservationID, err := uuid.Parse(path)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid reservation id")
		return
	}

	switch r.Method {
	case http.MethodGet:
		timeline, err := s.repo.GuestJourney.GetTimeline(ctx, uuid.MustParse(tenantID), reservationID)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, timeline)

	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleGuestJourneyEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	
	var req struct {
		ReservationID string                 `json:"reservation_id"`
		EventType     string                 `json:"event_type"`
		EventData     map[string]interface{} `json:"event_data,omitempty"`
		CreatedBy     string                 `json:"created_by"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}
	
	reservationID, err := uuid.Parse(req.ReservationID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid reservation id")
		return
	}
	
	event, err := s.repo.GuestJourney.CreateEvent(ctx, uuid.MustParse(tenantID), reservationID, req.EventType, req.CreatedBy, req.EventData)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusCreated, event)
}
