package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nexus-platform/pms-integration/internal/auth"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== AUTH ====================

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)

	var req struct {
		Email     string         `json:"email"`
		Password  string         `json:"password"`
		FirstName string         `json:"first_name"`
		LastName  string         `json:"last_name"`
		Role      domain.Role    `json:"role"`
		PropertyID string        `json:"property_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" {
		respondError(w, http.StatusBadRequest, "email, password, first_name, last_name required")
		return
	}

	// Check if user already exists
	if _, err := h.store.Users.GetByEmail(ctx, tenant, req.Email); err == nil {
		respondError(w, http.StatusConflict, "user with this email already exists")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	u, err := domain.NewUser(tenant, req.Email, req.FirstName, req.LastName, req.Role)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	u.PropertyID = req.PropertyID
	u.PasswordHash = hash

	if err := h.store.Users.Create(ctx, u); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Don't return password hash
	u.PasswordHash = ""
	respondJSON(w, http.StatusCreated, u)
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)

	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Email == "" || req.Password == "" {
		respondError(w, http.StatusBadRequest, "email and password required")
		return
	}

	u, err := h.store.Users.GetByEmail(ctx, tenant, req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if !u.IsActive {
		respondError(w, http.StatusUnauthorized, "account deactivated")
		return
	}

	if err := auth.CheckPassword(req.Password, u.PasswordHash); err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Update last login
	_ = h.store.Users.UpdateLastLogin(ctx, tenant, u.ID)

	jwtSvc := auth.NewService(auth.DefaultConfig())
	tokens, err := jwtSvc.GenerateTokenPair(u.ID, u.TenantID, u.Email, string(u.Role))
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"user": map[string]interface{}{
			"id":         u.ID,
			"email":      u.Email,
			"first_name": u.FirstName,
			"last_name":  u.LastName,
			"role":       u.Role,
			"tenant_id":  u.TenantID,
		},
		"tokens": tokens,
	})
}

func (h *Handler) me(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user").(*auth.Claims)
	if !ok {
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	ctx := r.Context()
	u, err := h.store.Users.GetByID(ctx, claims.TenantID, claims.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	u.PasswordHash = ""
	respondJSON(w, http.StatusOK, u)
}

func (h *Handler) listUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenant := tenantID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 { limit = 50 }
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 { page = 1 }
	offset := (page - 1) * limit

	users, err := h.store.Users.ListByTenant(ctx, tenant, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	for _, u := range users {
		u.PasswordHash = ""
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"data": users, "page": page, "limit": limit})
}
