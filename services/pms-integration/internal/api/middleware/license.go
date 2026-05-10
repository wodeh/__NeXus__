package middleware

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
	"github.com/nexus-platform/pms-integration/internal/license"
)

type contextKey string

const tenantConfigKey contextKey = "tenant_config"

// License creates middleware that checks tenant capabilities.
func License(svc *license.Service, required ...domain.Capability) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := chi.URLParam(r, "tenantId")
			if tenantID == "" {
				http.Error(w, `{"error":"tenant required"}`, http.StatusBadRequest)
				return
			}

			cfg, err := svc.GetConfig(r.Context(), tenantID)
			if err != nil {
				http.Error(w, `{"error":"tenant not found"}`, http.StatusNotFound)
				return
			}

			// Inject config into context
			ctx := context.WithValue(r.Context(), tenantConfigKey, cfg)
			r = r.WithContext(ctx)

			// If no specific capability required, just inject and proceed
			if len(required) == 0 {
				next.ServeHTTP(w, r)
				return
			}

			ok, err := svc.CheckAnyCapability(r.Context(), tenantID, required...)
			if err != nil {
				http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusForbidden)
				return
			}
			if !ok {
				http.Error(w, `{"error":"license tier does not include this feature"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TenantConfigFromContext retrieves tenant config from request context.
func TenantConfigFromContext(ctx context.Context) *domain.TenantConfig {
	if cfg, ok := ctx.Value(tenantConfigKey).(*domain.TenantConfig); ok {
		return cfg
	}
	return nil
}
