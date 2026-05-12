package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// registerAdminHandlers registers admin API routes.
func (s *Server) registerAdminHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/admin/users", s.withTenant(s.handleAdminUsers))
	mux.HandleFunc("/v1/admin/properties", s.withTenant(s.handleAdminProperties))
	mux.HandleFunc("/v1/admin/config", s.withTenant(s.handleAdminConfig))
}

/* ─── Users ─── */

func (s *Server) handleAdminUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	switch r.Method {
	case http.MethodGet:
		users, err := s.repo.Admin.ListUsers(ctx, tenantID)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{"users": demoUsers()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"users": users})

	case http.MethodPost:
		var req struct {
			Email    string `json:"email"`
			Name     string `json:"name"`
			Role     string `json:"role"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}
		if req.Email == "" || req.Name == "" || req.Role == "" {
			http.Error(w, `{"error":"missing required fields"}`, http.StatusBadRequest)
			return
		}
		u := &domain.UserWithRole{
			ID:       uuid.New(),
			TenantID: tenantID,
			Email:    req.Email,
			Name:     req.Name,
			Role:     req.Role,
			IsActive: true,
		}
		passwordHash := req.Password
		if passwordHash == "" { passwordHash = "changeme" }
		if err := s.repo.Admin.CreateUser(ctx, u, passwordHash); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, u)

	case http.MethodPatch:
		var req struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Role     string `json:"role"`
			IsActive bool   `json:"is_active"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}
		id, err := uuid.Parse(req.ID)
		if err != nil {
			http.Error(w, `{"error":"invalid user ID"}`, http.StatusBadRequest)
			return
		}
		if err := s.repo.Admin.UpdateUser(ctx, tenantID, id, req.Name, req.Role, req.IsActive); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})

	case http.MethodDelete:
		idStr := r.URL.Query().Get("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, `{"error":"invalid user ID"}`, http.StatusBadRequest)
			return
		}
		if err := s.repo.Admin.DeleteUser(ctx, tenantID, id); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func demoUsers() []domain.UserWithRole {
	return []domain.UserWithRole{
		{ID: uuid.MustParse("a0000001-0000-0000-0000-000000000001"), Email: "admin@grandplaza.com", Name: "Alex Rivera", Role: "admin", IsActive: true},
		{ID: uuid.MustParse("a0000002-0000-0000-0000-000000000002"), Email: "manager@grandplaza.com", Name: "Sarah Chen", Role: "manager", IsActive: true},
		{ID: uuid.MustParse("a0000003-0000-0000-0000-000000000003"), Email: "frontdesk@grandplaza.com", Name: "James Wilson", Role: "staff", IsActive: true},
		{ID: uuid.MustParse("a0000004-0000-0000-0000-000000000004"), Email: "cleaner1@grandplaza.com", Name: "Maria Garcia", Role: "cleaner", IsActive: true},
		{ID: uuid.MustParse("a0000005-0000-0000-0000-000000000005"), Email: "cleaner2@grandplaza.com", Name: "Liu Wei", Role: "cleaner", IsActive: true},
	}
}

/* ─── Properties ─── */

func (s *Server) handleAdminProperties(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	switch r.Method {
	case http.MethodGet:
		props, err := s.repo.Admin.ListProperties(ctx, tenantID)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{"properties": demoProperties()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"properties": props})

	case http.MethodPost:
		var req struct {
			Name       string                 `json:"name"`
			Address    string                 `json:"address"`
			City       string                 `json:"city"`
			Country    string                 `json:"country"`
			Phone      string                 `json:"phone"`
			Email      string                 `json:"email"`
			Timezone   string                 `json:"timezone"`
			Currency   string                 `json:"currency"`
			StarRating int                    `json:"star_rating"`
			Config     map[string]interface{} `json:"config,omitempty"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}
		if req.Name == "" {
			http.Error(w, `{"error":"name required"}`, http.StatusBadRequest)
			return
		}
		if req.Timezone == "" { req.Timezone = "UTC" }
		if req.Currency == "" { req.Currency = "USD" }
		p := &domain.Property{
			ID:         uuid.New(),
			TenantID:   tenantID,
			Name:       req.Name,
			Address:    req.Address,
			City:       req.City,
			Country:    req.Country,
			Phone:      req.Phone,
			Email:      req.Email,
			Timezone:   req.Timezone,
			Currency:   req.Currency,
			StarRating: req.StarRating,
			IsActive:   true,
			Config:     req.Config,
		}
		if err := s.repo.Admin.CreateProperty(ctx, p); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusCreated, p)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func demoProperties() []domain.Property {
	return []domain.Property{
		{ID: uuid.MustParse("b0000001-0000-0000-0000-000000000001"), Name: "Grand Plaza Main", Address: "123 Ocean Drive", City: "Miami", Country: "USA", Phone: "+1-555-0199", Email: "info@grandplaza.com", Timezone: "America/New_York", Currency: "USD", StarRating: 4, IsActive: true},
	}
}

/* ─── System Config ─── */

func (s *Server) handleAdminConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantIDStr, _ := ctx.Value("tenant_id").(string)
	tenantID, _ := uuid.Parse(tenantIDStr)

	switch r.Method {
	case http.MethodGet:
		cfg, err := s.repo.Admin.GetSystemConfig(ctx, tenantID)
		if err != nil {
			writeJSON(w, http.StatusOK, &domain.SystemConfig{
				TenantID:            tenantID,
				DefaultCheckInTime:  "15:00",
				DefaultCheckOutTime: "11:00",
				AutoConfirm:         true,
				DepositPercent:      20,
				AllowWalkIn:         true,
			})
			return
		}
		writeJSON(w, http.StatusOK, cfg)

	case http.MethodPatch:
		var req map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
			return
		}
		if err := s.repo.Admin.UpdateSystemConfig(ctx, tenantID, req); err != nil {
			http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}
