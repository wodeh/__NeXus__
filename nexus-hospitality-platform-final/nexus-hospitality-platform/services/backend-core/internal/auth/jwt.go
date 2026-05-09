package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// JWTValidator validates RS256-signed JWT access tokens.
type JWTValidator struct {
	publicKey interface{}
	issuer    string
	audience  string
}

// NewJWTValidator creates a validator with the given RSA public key, issuer, and audience.
func NewJWTValidator(publicKey interface{}, issuer, audience string) *JWTValidator {
	return &JWTValidator{publicKey: publicKey, issuer: issuer, audience: audience}
}

// Validate parses and verifies a JWT string.
func (v *JWTValidator) Validate(tokenString string) (*jwt.Token, jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.publicKey, nil
	},
		jwt.WithIssuer(v.issuer),
		jwt.WithAudience(v.audience),
		jwt.WithValidMethods([]string{"RS256"}),
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
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}
			_, claims, err := validator.Validate(parts[1])
			if err != nil {
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusUnauthorized)
				return
			}
			ctx := r.Context()
			if sub, ok := claims["sub"].(string); ok {
				ctx = WithUserID(ctx, sub)
			}
			if tenantID, ok := claims["tenant_id"].(string); ok {
				ctx = WithTenantID(ctx, tenantID)
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
