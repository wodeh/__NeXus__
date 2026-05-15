package server

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/nexus-platform/backend-core/internal/auth"
)

// jwtSecret returns the HMAC secret for signing JWTs.
// In production this MUST be set via JWT_SECRET environment variable.
// The fallback is only safe for local development.
func jwtSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret != "" {
		return []byte(secret)
	}
	// Fallback development key — NOT for production
	return []byte("nexus-dev-jwt-secret-do-not-use-in-production-2024")
}

func (s *Server) registerAuthHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/auth/login", s.withRateLimit(s.handleLogin))
	// /v1/auth/me is a session-check endpoint — it must work without X-Tenant-ID
	// so the frontend can call restoreSession() before login. handleMe already
	// returns {authenticated: false} gracefully when no JWT context exists.
	mux.HandleFunc("/v1/auth/me", s.handleMe)
}

// loginRequest is the request body for the login endpoint.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if req.Email == "" || req.Password == "" {
		writeJSONError(w, http.StatusBadRequest, "email and password required")
		return
	}
	if s.repo == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "database not connected — check postgres and DATABASE_URL")
		return
	}
	if s.repo.Auth == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "auth repository not configured")
		return
	}
	creds, err := s.repo.Auth.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if !creds.IsActive {
		writeJSONError(w, http.StatusUnauthorized, "account disabled")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(req.Password)); err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	_ = s.repo.Auth.UpdateLastLogin(r.Context(), creds.ID)
	tenant, _ := s.repo.Auth.GetTenantByID(r.Context(), creds.TenantID)
	tenantName := ""
	tenantExternalID := ""
	if tenant != nil {
		tenantName = tenant.Name
		tenantExternalID = tenant.ExternalID
	}

	// Build capability list based on role
	capabilities := defaultCapabilities()
	if creds.Role == "super_admin" || creds.Role == "admin" {
		capabilities = adminCapabilities()
	} else if creds.Role == "manager" && tenantExternalID == "villa-owners" {
		capabilities = villaOwnerCapabilities()
	}

	// Generate proper JWT with claims
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":          creds.ID.String(),
		"email":        creds.Email,
		"role":         creds.Role,
		"tenant_id":    creds.TenantID.String(),
		"capabilities": capabilities,
		"iat":          now.Unix(),
		"exp":          now.Add(24 * time.Hour).Unix(),
		"iss":          "nexus-backend",
		"aud":          "nexus-api",
	})

	tokenString, err := token.SignedString(jwtSecret())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": tokenString,
		"user": map[string]interface{}{
			"id": creds.ID, "email": creds.Email, "name": creds.Name,
			"role": creds.Role, "tenant_id": creds.TenantID, "is_active": creds.IsActive,
		},
		"tenant": map[string]interface{}{
			"id": tenantExternalID, "name": tenantName, "external_id": tenantExternalID,
		},
		"capabilities": capabilities,
	})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	// Read JWT directly from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	// Validate the token
	validator := auth.NewHMACValidator(jwtSecretFromEnv(), "nexus-backend", jwtAudience)
	_, claims, err := validator.Validate(parts[1])
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	// Extract user_id and tenant_id from claims
	userID, _ := claims["sub"].(string)
	tenantID, _ := claims["tenant_id"].(string)
	if userID == "" {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"authenticated": false,
		})
		return
	}

	// Look up full user record
	if s.repo == nil || s.repo.Auth == nil {
		role, _ := claims["role"].(string)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"authenticated": true,
			"user": map[string]interface{}{
				"id": userID, "role": role,
			},
		})
		return
	}

	creds, err := s.repo.Auth.GetUserByID(r.Context(), uuid.MustParse(userID))
	if err != nil {
		role, _ := claims["role"].(string)
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"authenticated": true,
			"user": map[string]interface{}{
				"id": userID, "role": role,
			},
		})
		return
	}

	// Build capability list based on role
	capabilities := defaultCapabilities()
	if creds.Role == "super_admin" || creds.Role == "admin" {
		capabilities = adminCapabilities()
	} else if creds.Role == "manager" {
		tenant, _ := s.repo.Auth.GetTenantByID(r.Context(), creds.TenantID)
		if tenant != nil && tenant.ExternalID == "villa-owners" {
			capabilities = villaOwnerCapabilities()
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"authenticated": true,
		"user": map[string]interface{}{
			"id":         creds.ID,
			"email":      creds.Email,
			"name":       creds.Name,
			"role":       creds.Role,
			"tenant_id":  creds.TenantID,
			"is_active":  creds.IsActive,
		},
		"tenant": map[string]interface{}{
			"id":          tenantID,
			"name":        creds.Name + "'s Hotel",
			"external_id": tenantID,
		},
		"capabilities": capabilities,
	})
}

func defaultCapabilities() []string {
	return []string{
		"core:reservations", "core:guests", "core:properties", "core:rooms",
		"core:housekeeping", "core:settings", "core:audit_logs",
		"operations:floor_dashboard", "operations:room_blocks",
		"operations:group_reservations", "operations:maintenance", "operations:front_desk",
		"revenue:dynamic_pricing", "revenue:ota_integration",
		"revenue:revenue_forecasting", "revenue:agent_management",
		"enterprise:multi_property", "enterprise:advanced_crm",
		"enterprise:api_access", "enterprise:white_label", "enterprise:custom_reports",
		"villa:dashboard", "villa:properties", "villa:reservations", "villa:revenue", "villa:cleaner_tracking",
	}
}

func adminCapabilities() []string {
	return []string{
		"admin:full",
		"core:reservations", "core:guests", "core:properties", "core:rooms",
		"core:housekeeping", "core:settings", "core:audit_logs",
		"operations:floor_dashboard", "operations:room_blocks",
		"operations:group_reservations", "operations:maintenance", "operations:front_desk",
		"operations:iptv_basic", "operations:smart_locks",
		"revenue:dynamic_pricing", "revenue:ota_integration",
		"revenue:revenue_forecasting", "revenue:agent_management",
		"revenue:channel_manager", "revenue:whatsapp_bot", "revenue:direct_booking",
		"revenue:guest_reviews", "revenue:communications",
		"enterprise:multi_property", "enterprise:advanced_crm",
		"enterprise:api_access", "enterprise:white_label", "enterprise:custom_reports",
		"villa:dashboard", "villa:properties", "villa:reservations", "villa:revenue", "villa:cleaner_tracking",
	}
}

func villaOwnerCapabilities() []string {
	return []string{
		"villa:dashboard", "villa:properties", "villa:reservations",
		"villa:revenue", "villa:cleaner_tracking",
		"core:reservations", "core:guests", "core:settings",
	}
}
