#!/bin/bash
# create_auth_handlers.sh - Creates the missing auth_handlers.go file

cat > /c/Projects/NeXus/services/backend-core/internal/server/auth_handlers.go << 'EOF'
package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// registerAuthHandlers registers authentication routes (no tenant middleware required).
func (s *Server) registerAuthHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/auth/login", s.handleLogin)
	mux.HandleFunc("/v1/auth/me", s.withTenant(s.handleMe))
}

/* ─── Login ─── */

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

	// Lookup user by email (cross-tenant)
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

	// Validate password
	if err := bcrypt.CompareHashAndPassword([]byte(creds.PasswordHash), []byte(req.Password)); err != nil {
		writeJSONError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	// Update last login
	_ = s.repo.Auth.UpdateLastLogin(r.Context(), creds.ID)

	// Fetch tenant info
	tenant, _ := s.repo.Auth.GetTenantByID(r.Context(), creds.TenantID)
	tenantName := ""
	tenantExternalID := ""
	if tenant != nil {
		tenantName = tenant.Name
		tenantExternalID = tenant.ExternalID
	}

	// Generate a simple session token (UUID for now, can be upgraded to JWT later)
	token := uuid.New().String()

	// Build capabilities list based on role
	capabilities := defaultCapabilities()
	if creds.Role == "manager" && tenantExternalID == "villa-owners" {
		capabilities = villaOwnerCapabilities()
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user": map[string]interface{}{
			"id":        creds.ID,
			"email":     creds.Email,
			"name":      creds.Name,
			"role":      creds.Role,
			"tenant_id": creds.TenantID,
			"is_active": creds.IsActive,
		},
		"tenant": map[string]interface{}{
			"id":          tenantExternalID,
			"name":        tenantName,
			"external_id": tenantExternalID,
		},
		"capabilities": capabilities,
	})
}

/* ─── Me (current user) ─── */

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	// For now, return the tenant context. In a full JWT setup this would decode the token.
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tenant_id": tenantID,
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

func villaOwnerCapabilities() []string {
	return []string{
		"villa:dashboard",
		"villa:properties",
		"villa:reservations",
		"villa:revenue",
		"villa:cleaner_tracking",
		"core:reservations",
		"core:guests",
		"core:settings",
	}
}
EOF

echo "✅ auth_handlers.go created"
echo "🔧 Rebuilding server..."
cd /c/Projects/NeXus/services/backend-core
go build -o server cmd/server/main.go

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo "🚀 Start the server: ./server"
else
    echo "❌ Build failed — check errors above"
fi
