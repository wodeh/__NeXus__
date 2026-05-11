package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
	"github.com/nexus-platform/backend-core/internal/repository"
)

func (s *Server) registerAgentHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/agents", s.withTenant(s.handleAgents))
	mux.HandleFunc("/v1/agents/", s.withTenant(s.handleAgentDetail))
}

func (s *Server) handleAgents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewAgentRepository(s.repo.Pool())

	switch r.Method {
	case http.MethodGet:
		agents, err := repo.List(ctx, tenantID)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"agents": agents})
	case http.MethodPost:
		var req domain.AgentCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		agent, err := repo.Create(ctx, tenantID, &req)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, agent)
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAgentDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	path := r.URL.Path[len("/v1/agents/"):]
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		http.Error(w, `{"error":"missing agent id"}`, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, `{"error":"invalid agent id"}`, http.StatusBadRequest)
		return
	}

	repo := repository.NewAgentRepository(s.repo.Pool())

	switch r.Method {
	case http.MethodGet:
		agent, err := repo.GetByID(ctx, tenantID, id)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, agent)
	case http.MethodPatch:
		var req domain.AgentUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		agent, err := repo.Update(ctx, tenantID, id, &req)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, agent)
	case http.MethodDelete:
		if err := repo.Delete(ctx, tenantID, id); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusNoContent, nil)
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
