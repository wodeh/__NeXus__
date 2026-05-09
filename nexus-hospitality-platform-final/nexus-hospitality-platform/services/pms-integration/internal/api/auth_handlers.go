// Package api provides HTTP REST handlers for authentication.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
	"github.com/nexus-platform/pms-integration/internal/middleware"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

// JWT secret — in production, load from env
var jwtSecret = []byte("nexus-platform-jwt-secret-change-in-production")

func SetJWTSecret(secret string) {
	jwtSecret = []byte(secret)
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	User         userResponse `json:"user"`
}

type userResponse struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	FirstName string   `json:"first_name"`
	LastName  string   `json:"last_name"`
	Role      string   `json:"role"`
	TenantID  string   `json:"tenant_id"`
}

func toUserResponse(u *domain.User) userResponse {
	return userResponse{
		ID:        string(u.ID),
		Email:     u.Email,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Role:      string(u.Role),
		TenantID:  u.TenantID,
	}
}

func (h *Handler) login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Extract tenant from header or use default
	tenantID := r.Header.Get("X-Tenant-ID")
	if tenantID == "" {
		tenantID = chi.URLParam(r, "tenantId")
	}
	if tenantID == "" {
		respondError(w, http.StatusBadRequest, "tenant_id required")
		return
	}

	user, err := h.store.Users.GetByEmail(r.Context(), tenantID, req.Email)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if user.Status != domain.UserStatusActive {
		respondError(w, http.StatusForbidden, "account is not active")
		return
	}

	if user.IsLocked() {
		respondError(w, http.StatusForbidden, "account is temporarily locked")
		return
	}

	if !user.VerifyPassword(req.Password) {
		user.RecordFailedLogin()
		_ = h.store.Users.Update(r.Context(), user)
		respondError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	user.RecordLogin()
	_ = h.store.Users.Update(r.Context(), user)

	accessToken, err := generateAccessToken(user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	refreshTokenStr := generateRefreshToken()
	rt := domain.NewRefreshToken(string(user.ID), tenantID, refreshTokenStr, refreshTokenTTL, r.RemoteAddr, r.UserAgent())
	_ = h.store.RefreshTokens.Create(r.Context(), rt)

	respondJSON(w, http.StatusOK, loginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		TokenType:    "Bearer",
		ExpiresIn:    int(accessTokenTTL.Seconds()),
		User:         toUserResponse(user),
	})
}

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TenantSlug   string           `json:"tenant_slug"`
		TenantName   string           `json:"tenant_name"`
		TenantEmail  string           `json:"tenant_email"`
		CurrencyCode string           `json:"currency_code"`
		Email        string           `json:"email"`
		Password     string           `json:"password"`
		FirstName    string           `json:"first_name"`
		LastName     string           `json:"last_name"`
		Phone        string           `json:"phone"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Create tenant
	tenant, err := domain.NewTenant(req.TenantSlug, req.TenantName, req.TenantEmail, req.CurrencyCode)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.store.Tenants.Create(r.Context(), tenant); err != nil {
		respondError(w, http.StatusConflict, "tenant already exists")
		return
	}

	// Create first user as admin
	user, err := domain.NewUser(string(tenant.ID), req.Email, req.Password, req.FirstName, req.LastName, domain.UserRoleAdmin)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	user.Phone = req.Phone

	if err := h.store.Users.Create(r.Context(), user); err != nil {
		respondError(w, http.StatusConflict, "user already exists")
		return
	}

	accessToken, _ := generateAccessToken(user)
	refreshTokenStr := generateRefreshToken()
	rt := domain.NewRefreshToken(string(user.ID), string(tenant.ID), refreshTokenStr, refreshTokenTTL, r.RemoteAddr, r.UserAgent())
	_ = h.store.RefreshTokens.Create(r.Context(), rt)

	respondJSON(w, http.StatusCreated, loginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshTokenStr,
		TokenType:    "Bearer",
		ExpiresIn:    int(accessTokenTTL.Seconds()),
		User:         toUserResponse(user),
	})
}

func (h *Handler) refreshToken(w http.ResponseWriter, r *http.Request) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	rt, err := h.store.RefreshTokens.GetByToken(r.Context(), req.RefreshToken)
	if err != nil || !rt.Verify(req.RefreshToken) {
		respondError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}

	user, err := h.store.Users.GetByID(r.Context(), rt.TenantID, rt.UserID)
	if err != nil {
		respondError(w, http.StatusUnauthorized, "user not found")
		return
	}

	accessToken, err := generateAccessToken(user)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"access_token": accessToken,
		"token_type":   "Bearer",
		"expires_in":   int(accessTokenTTL.Seconds()),
	})
}

func (h *Handler) getMe(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims == nil {
		respondError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	user, err := h.store.Users.GetByID(r.Context(), claims.TenantID, claims.UserID)
	if err != nil {
		respondError(w, http.StatusNotFound, "user not found")
		return
	}

	respondJSON(w, http.StatusOK, toUserResponse(user))
}

func (h *Handler) logout(w http.ResponseWriter, r *http.Request) {
	claims := middleware.ClaimsFromContext(r.Context())
	if claims != nil {
		_ = h.store.RefreshTokens.RevokeAllForUser(r.Context(), claims.UserID)
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// ==================== DASHBOARD STATS ====================

func (h *Handler) getDashboardStats(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)

	// Count properties
	var totalProperties int
	_ = h.store.pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM properties WHERE tenant_id = $1`, tenantID).Scan(&totalProperties)

	// Count rooms
	var totalRooms, occupiedRooms int
	_ = h.store.pool.QueryRow(r.Context(), `SELECT COUNT(*), COUNT(*) FILTER (WHERE status = 'occupied') FROM rooms WHERE tenant_id = $1`, tenantID).Scan(&totalRooms, &occupiedRooms)

	// Today's check-ins/check-outs
	today := time.Now().Format("2006-01-02")
	var todayCheckIns, todayCheckOuts int
	_ = h.store.pool.QueryRow(r.Context(), `
		SELECT 
			COUNT(*) FILTER (WHERE check_in_date = $2),
			COUNT(*) FILTER (WHERE check_out_date = $2)
		FROM reservations WHERE tenant_id = $1
	`, tenantID, today).Scan(&todayCheckIns, &todayCheckOuts)

	// Pending housekeeping
	var pendingHK int
	_ = h.store.pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM housekeeping_tasks WHERE tenant_id = $1 AND status IN ('pending', 'assigned')`, tenantID).Scan(&pendingHK)

	// Open work orders
	var openWO int
	_ = h.store.pool.QueryRow(r.Context(), `SELECT COUNT(*) FROM work_orders WHERE tenant_id = $1 AND status IN ('open', 'assigned', 'in_progress')`, tenantID).Scan(&openWO)

	// Revenue today
	var revenueToday float64
	_ = h.store.pool.QueryRow(r.Context(), `
		SELECT COALESCE(SUM(total_amount), 0) FROM charges 
		WHERE tenant_id = $1 AND charge_date = $2 AND is_voided = FALSE
	`, tenantID, today).Scan(&revenueToday)

	occupancyRate := 0.0
	if totalRooms > 0 {
		occupancyRate = float64(occupiedRooms) / float64(totalRooms) * 100
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_properties":     totalProperties,
		"total_rooms":          totalRooms,
		"occupied_rooms":       occupiedRooms,
		"available_rooms":      totalRooms - occupiedRooms,
		"today_check_ins":      todayCheckIns,
		"today_check_outs":     todayCheckOuts,
		"pending_housekeeping": pendingHK,
		"open_work_orders":     openWO,
		"revenue_today":        revenueToday,
		"occupancy_rate":       occupancyRate,
	})
}

// ==================== HELPERS ====================

func generateAccessToken(user *domain.User) (string, error) {
	now := time.Now().UTC()
	claims := middleware.Claims{
		TenantID:   user.TenantID,
		UserID:     string(user.ID),
		Email:      user.Email,
		Roles:      []string{string(user.Role)},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   string(user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(accessTokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func generateRefreshToken() string {
	b := make([]byte, 32)
	for i := range b {
		b[i] = byte(time.Now().UnixNano() % 256)
	}
	return string(b)
}
