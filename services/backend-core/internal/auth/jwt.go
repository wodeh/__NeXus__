package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// writeJSONError writes a JSON error response. Duplicated here to avoid circular imports.
func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// JWTValidator validates JWT access tokens (RS256 or HS256).
type JWTValidator struct {
	publicKey  interface{}
	hmacSecret []byte
	issuer     string
	audience   string
}

// NewJWTValidator creates a validator with the given RSA public key, issuer, and audience.
func NewJWTValidator(publicKey interface{}, issuer, audience string) *JWTValidator {
	return &JWTValidator{publicKey: publicKey, issuer: issuer, audience: audience}
}

// NewHMACValidator creates a validator that verifies HMAC-SHA256 (HS256) signed tokens.
func NewHMACValidator(secret []byte, issuer, audience string) *JWTValidator {
	return &JWTValidator{hmacSecret: secret, issuer: issuer, audience: audience}
}

// Validate parses and verifies a JWT string.
func (v *JWTValidator) Validate(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		switch token.Method.(type) {
		case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS:
			if v.publicKey == nil {
				return nil, fmt.Errorf("RSA public key not configured")
			}
			return v.publicKey, nil
		case *jwt.SigningMethodHMAC:
			if len(v.hmacSecret) == 0 {
				return nil, fmt.Errorf("HMAC secret not configured")
			}
			return v.hmacSecret, nil
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
	},
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithValidMethods([]string{"RS256", "HS256"}),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("parse jwt: %w", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, nil, fmt.Errorf("invalid claims type")
	}
	return token, claims, nil
}

// JWTMiddleware returns an HTTP middleware that validates Bearer tokens.
func JWTMiddleware(validator *JWTValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				writeJSONError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				writeJSONError(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}
			_, claims, err := validator.Validate(parts[1])
			if err != nil {
				writeJSONError(w, http.StatusUnauthorized, err.Error())
				return
			}
			ctx := r.Context()
			if sub, ok := claims["sub"].(string); ok {
				ctx = WithUserID(ctx, sub)
				// Also set raw string key for backward compatibility with handlers
				ctx = context.WithValue(ctx, "user_id", sub)
			}
			if tenantID, ok := claims["tenant_id"].(string); ok {
				ctx = WithTenantID(ctx, tenantID)
				// Also set raw string key for backward compatibility with handlers
				ctx = context.WithValue(ctx, "tenant_id", tenantID)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
