package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
	"github.com/nexus-platform/backend-core/internal/repository"
)

func (s *Server) registerAdminHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/admin/tenants", s.withTenant(s.handleAdminTenants))
	mux.HandleFunc("/v1/admin/tenants/", s.withTenant(s.handleAdminTenantDetail))
	mux.HandleFunc("/v1/admin/users", s.withTenant(s.handleAdminUsers))
	mux.HandleFunc("/v1/admin/users/", s.withTenant(s.handleAdminUserDetail))
	mux.HandleFunc("/v1/admin/roles", s.withTenant(s.handleAdminRoles))
	mux.HandleFunc("/v1/admin/roles/", s.withTenant(s.handleAdminRoleDetail))
}

// ─── Tenants (Hotel List) ───

func (s *Server) handleAdminTenants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Only super_admin can list all tenants; others get their own tenant only
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewTenantRepository(s.repo.Pool())

	if r.Method == http.MethodGet {
		// For now, return single tenant info. Full multi-tenant listing requires super_admin check.
		tenant, err := repo.GetByExternalID(ctx, tenantID)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"tenants": []interface{}{tenant}})
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func (s *Server) handleAdminTenantDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	path := r.URL.Path[len("/v1/admin/tenants/"):]
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		http.Error(w, `{"error":"missing tenant id"}`, http.StatusBadRequest)
		return
	}

	// Summary endpoint: /v1/admin/tenants/{id}/summary
	if len(parts) > 1 && parts[1] == "summary" {
		repo := repository.NewTenantRepository(s.repo.Pool())
		tid, err := uuid.Parse(tenantID)
	if err != nil {
		http.Error(w, `{"error":"invalid tenant id"}`, http.StatusBadRequest)
		return
	}
	props, err := repo.ListProperties(ctx, tid)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"property_count": len(props),
		})
		return
	}

	http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
}

// ─── Users ───

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewUserRepository(s.repo.Pool())

	switch r.Method {
	case http.MethodGet:
		users, err := repo.List(ctx, tenantID)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"users": users})

	case http.MethodPost:
		var req domain.User
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		if req.Email == "" || req.Name == "" {
			http.Error(w, `{"error":"email and name are required"}`, http.StatusBadRequest)
			return
		}
		if req.Role == "" {
			req.Role = "front_desk"
		}
		if req.Status == "" {
			req.Status = "active"
		}
		if err := repo.Create(ctx, tenantID, &req); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, req)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAdminUserDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	path := r.URL.Path[len("/v1/admin/users/"):]
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		http.Error(w, `{"error":"missing user id"}`, http.StatusBadRequest)
		return
	}
	id := parts[0]
	repo := repository.NewUserRepository(s.repo.Pool())

	switch r.Method {
	case http.MethodGet:
		user, err := repo.Get(ctx, tenantID, id)
		if err != nil {
			http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, user)

	case http.MethodPatch:
		var req domain.User
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		if err := repo.Update(ctx, tenantID, id, &req); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, req)

	case http.MethodDelete:
		if err := repo.Delete(ctx, tenantID, id); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

// ─── Roles ───

func (s *Server) handleAdminRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewRoleRepository(s.repo.Pool())

	switch r.Method {
	case http.MethodGet:
		roles, err := repo.List(ctx, tenantID)
		if err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"roles": roles})

	case http.MethodPost:
		var req domain.Role
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, `{"error":"role name is required"}`, http.StatusBadRequest)
			return
		}
		if err := repo.Create(ctx, tenantID, &req); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, req)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleAdminRoleDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	path := r.URL.Path[len("/v1/admin/roles/"):]
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		http.Error(w, `{"error":"missing role id"}`, http.StatusBadRequest)
		return
	}
	id := parts[0]
	repo := repository.NewRoleRepository(s.repo.Pool())

	switch r.Method {
	case http.MethodPatch:
		var req domain.Role
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		if err := repo.Update(ctx, tenantID, id, &req); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, req)

	case http.MethodDelete:
		if err := repo.Delete(ctx, tenantID, id); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
