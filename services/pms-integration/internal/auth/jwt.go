package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Config holds JWT settings.
type Config struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
	Issuer        string
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		Secret:        "nexus-dev-secret-change-in-production",
		AccessExpiry:  24 * time.Hour,
		RefreshExpiry: 7 * 24 * time.Hour,
		Issuer:        "nexus-pms",
	}
}

// Claims represents JWT custom claims.
type Claims struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// TokenPair holds access and refresh tokens.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// Service handles JWT operations.
type Service struct {
	cfg Config
}

// NewService creates a JWT service.
func NewService(cfg Config) *Service {
	return &Service{cfg: cfg}
}

// GenerateTokenPair creates access and refresh tokens for a user.
func (s *Service) GenerateTokenPair(userID, tenantID, email, role string) (*TokenPair, error) {
	access, err := s.generateToken(userID, tenantID, email, role, s.cfg.AccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}
	refresh, err := s.generateToken(userID, tenantID, email, role, s.cfg.RefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}
	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(s.cfg.AccessExpiry.Seconds()),
		TokenType:    "Bearer",
	}, nil
}

func (s *Service) generateToken(userID, tenantID, email, role string, expiry time.Duration) (string, error) {
	claims := Claims{
		UserID:   userID,
		TenantID: tenantID,
		Email:    email,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    s.cfg.Issuer,
			Subject:   userID,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Secret))
}

// ValidateToken parses and validates a JWT string.
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.Secret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
