// Package middleware provides HTTP middleware for auth, tenant, and request handling.
package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims for Nexus platform.
type Claims struct {
	TenantID   string   `json:"tenant_id"`
	UserID     string   `json:"sub"`
	Email      string   `json:"email"`
	Roles      []string `json:"roles"`
	Properties []string `json:"properties"`
	jwt.RegisteredClaims
}

// Context keys.
type contextKey string

const (
	ContextKeyTenantID contextKey = "tenant_id"
	ContextKeyUserID   contextKey = "user_id"
	ContextKeyClaims   contextKey = "claims"
)

// JWTAuth validates Bearer tokens and injects claims into request context.
func JWTAuth(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health endpoints
			if r.URL.Path == "/health" || r.URL.Path == "/ready" || r.URL.Path == "/live" {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				respondError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				respondError(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			tokenStr := parts[1]
			token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
				}
				return []byte(secretKey), nil
			})
			if err != nil || !token.Valid {
				respondError(w, http.StatusUnauthorized, "invalid token")
				return
			}

			claims, ok := token.Claims.(*Claims)
			if !ok {
				respondError(w, http.StatusUnauthorized, "invalid claims")
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyTenantID, claims.TenantID)
			ctx = context.WithValue(ctx, ContextKeyUserID, claims.UserID)
			ctx = context.WithValue(ctx, ContextKeyClaims, claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TenantFromContext extracts tenant ID from context.
func TenantFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ContextKeyTenantID).(string); ok {
		return v
	}
	return ""
}

// UserFromContext extracts user ID from context.
func UserFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ContextKeyUserID).(string); ok {
		return v
	}
	return ""
}

// ClaimsFromContext extracts JWT claims from context.
func ClaimsFromContext(ctx context.Context) *Claims {
	if v, ok := ctx.Value(ContextKeyClaims).(*Claims); ok {
		return v
	}
	return nil
}

// CORS enables cross-origin requests for web clients.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RateLimiter provides simple per-tenant rate limiting.
type RateLimiter struct {
	requests map[string][]time.Time
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.RemoteAddr
		if tenantID := TenantFromContext(r.Context()); tenantID != "" {
			key = tenantID
		}

		now := time.Now()
		cutoff := now.Add(-rl.window)

		var recent []time.Time
		for _, t := range rl.requests[key] {
			if t.After(cutoff) {
				recent = append(recent, t)
			}
		}

		if len(recent) >= rl.limit {
			respondError(w, http.StatusTooManyRequests, "rate limit exceeded")
			return
		}

		recent = append(recent, now)
		rl.requests[key] = recent

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", rl.limit-len(recent)))

		next.ServeHTTP(w, r)
	})
}

func respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}
