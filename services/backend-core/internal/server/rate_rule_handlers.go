package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
)

func (s *Server) registerRateRuleHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/rate-rules", s.withTenant(s.handleRateRules))
	mux.HandleFunc("/v1/rate-rules/", s.withTenant(s.handleRateRuleDetail))
	mux.HandleFunc("/v1/rate-rules/calculate", s.withTenant(s.handleRateCalculate))
}

func (s *Server) handleRateRules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)

	switch r.Method {
	case http.MethodGet:
		rules, err := s.repo.RateRules.List(ctx, uuid.MustParse(tenantID))
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, rules)

	case http.MethodPost:
		var req domain.RateRuleCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSONError(w, http.StatusBadRequest, "invalid body")
			return
		}
		rule, err := s.repo.RateRules.Create(ctx, uuid.MustParse(tenantID), req)
		if err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, rule)

	default:
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (s *Server) handleRateRuleDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	idStr := r.URL.Path[len("/v1/rate-rules/"):]
	if idStr == "" {
		writeJSONError(w, http.StatusBadRequest, "missing id")
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if r.Method == http.MethodDelete {
		if err := s.repo.RateRules.Delete(ctx, uuid.MustParse(tenantID), id); err != nil {
			writeJSONError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
		return
	}

	writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func (s *Server) handleRateCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)

	var req domain.RateCalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid body")
		return
	}

	result, err := s.repo.RateRules.CalculateRate(ctx, uuid.MustParse(tenantID), req)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, result)
}
