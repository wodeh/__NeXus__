package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
	"github.com/nexus-platform/backend-core/internal/repository"
)

func (s *Server) registerChannelHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/channels", s.withTenant(s.handleChannels))
	mux.HandleFunc("/v1/channels/", s.withTenant(s.handleChannelDetail))
}

func (s *Server) handleChannels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewChannelRepository(s.repo.Pool())

	switch r.Method {
	case http.MethodGet:
		channels, err := repo.List(ctx, tenantID)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"channels": channels})
	case http.MethodPost:
		var req domain.ChannelCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		ch, err := repo.Create(ctx, tenantID, &req)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, ch)
	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleChannelDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	path := r.URL.Path[len("/v1/channels/"):]
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		http.Error(w, `{"error":"missing channel id"}`, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, `{"error":"invalid channel id"}`, http.StatusBadRequest)
		return
	}

	repo := repository.NewChannelRepository(s.repo.Pool())

	// Sub-routes: /v1/channels/{id}/sync-logs
	if len(parts) > 1 && parts[1] == "sync-logs" {
		if r.Method != http.MethodGet {
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return
		}
		logs, err := repo.ListSyncLogs(ctx, tenantID, &id, 50)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"logs": logs})
		return
	}

	switch r.Method {
	case http.MethodGet:
		ch, err := repo.GetByID(ctx, tenantID, id)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, ch)
	case http.MethodPatch:
		var req domain.ChannelUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		ch, err := repo.Update(ctx, tenantID, id, &req)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, ch)
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
