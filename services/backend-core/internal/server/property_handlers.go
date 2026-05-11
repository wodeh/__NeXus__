package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/nexus-platform/backend-core/internal/repository"
)

func (s *Server) registerPropertyHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/properties", s.withTenant(s.handleProperties))
	mux.HandleFunc("/v1/properties/", s.withTenant(s.handlePropertyDetail))
}

func (s *Server) handleProperties(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewTenantRepository(s.repo.Pool())

	if r.Method == http.MethodGet {
		props, err := repo.ListProperties(ctx, tenantID)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"properties": props})
		return
	}

	if r.Method == http.MethodPost {
		var req repository.TenantProperty
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		// For now, properties are read-only via API since they require DB schema alignment
		http.Error(w, `{"error":"create not yet implemented"}`, http.StatusNotImplemented)
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func (s *Server) handlePropertyDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	path := r.URL.Path[len("/v1/properties/"):]
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		http.Error(w, `{"error":"missing property id"}`, http.StatusBadRequest)
		return
	}
	propertyID := parts[0]

	repo := repository.NewTenantRepository(s.repo.Pool())
	props, err := repo.ListProperties(ctx, tenantID)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	var found *repository.TenantProperty
	for i := range props {
		if props[i].PropertyID == propertyID {
			found = &props[i]
			break
		}
	}
	if found == nil {
		http.Error(w, `{"error":"property not found"}`, http.StatusNotFound)
		return
	}

	if r.Method == http.MethodGet {
		writeJSON(w, http.StatusOK, found)
		return
	}

	if r.Method == http.MethodPatch {
		var req repository.TenantProperty
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		// Update logic would go here
		writeJSON(w, http.StatusOK, found)
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}
