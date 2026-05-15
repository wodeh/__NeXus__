package server

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

func (s *Server) registerOverbookingHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/overbooking/confidence", s.withTenant(s.handleOverbookingConfidence))
}

func (s *Server) handleOverbookingConfidence(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx := r.Context()
	tenantID := tenantIDFromContext(ctx)
	
	date := r.URL.Query().Get("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}
	
	scores, err := s.repo.Overbooking.GetConfidenceScores(ctx, uuid.MustParse(tenantID), date)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	
	writeJSON(w, http.StatusOK, scores)
}
