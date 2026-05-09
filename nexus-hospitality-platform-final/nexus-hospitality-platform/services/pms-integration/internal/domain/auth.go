// Package domain defines authentication and tenant aggregates.
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ==================== USER ====================

type UserID string

type UserRole string

const (
	UserRoleSuperAdmin UserRole = "super_admin"
	UserRoleAdmin      UserRole = "admin"
	UserRoleManager    UserRole = "manager"
	UserRoleFrontDesk  UserRole = "front_desk"
	UserRoleHousekeeping UserRole = "housekeeping"
	UserRoleMaintenance  UserRole = "maintenance"
	UserRoleReadOnly     UserRole = "read_only"
)

type UserStatus string

const (
	UserStatusActive    UserStatus = "active"
	UserStatusInactive  UserStatus = "inactive"
	UserStatusSuspended UserStatus = "suspended"
)

type User struct {
	ID             UserID
	TenantID       string
	Email          string
	PasswordHash   string
	FirstName      string
	LastName       string
	Phone          string
	Role           UserRole
	Status         UserStatus
	EmailVerified  bool
	LastLoginAt    *time.Time
	PasswordChangedAt time.Time
	FailedLoginAttempts int
	LockedUntil    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Version        int
}

func NewUser(tenantID, email, password, firstName, lastName string, role UserRole) (*User, error) {
	if email == "" || password == "" {
		return nil, fmt.Errorf("email and password required")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().UTC()
	return &User{
		ID:             UserID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:       tenantID,
		Email:          email,
		PasswordHash:   string(hash),
		FirstName:      firstName,
		LastName:       lastName,
		Role:           role,
		Status:         UserStatusActive,
		EmailVerified:  false,
		PasswordChangedAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
		Version:        1,
	}, nil
}

func (u *User) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) SetPassword(password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hash)
	u.PasswordChangedAt = time.Now().UTC()
	u.Version++
	u.UpdatedAt = time.Now().UTC()
	u.FailedLoginAttempts = 0
	u.LockedUntil = nil
	return nil
}

func (u *User) RecordLogin() {
	now := time.Now().UTC()
	u.LastLoginAt = &now
	u.FailedLoginAttempts = 0
	u.UpdatedAt = now
}

func (u *User) RecordFailedLogin() {
	u.FailedLoginAttempts++
	if u.FailedLoginAttempts >= 5 {
		lockUntil := time.Now().UTC().Add(30 * time.Minute)
		u.LockedUntil = &lockUntil
	}
	u.UpdatedAt = time.Now().UTC()
}

func (u *User) IsLocked() bool {
	if u.LockedUntil == nil {
		return false
	}
	return time.Now().UTC().Before(*u.LockedUntil)
}

func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// ==================== TENANT ====================

type TenantID string

type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusTrial     TenantStatus = "trial"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusCancelled TenantStatus = "cancelled"
)

type Tenant struct {
	ID             TenantID
	Slug           string
	Name           string
	Email          string
	Phone          string
	Address        string
	Country        string
	CurrencyCode   string
	Timezone       string
	Status         TenantStatus
	Plan           string
	MaxProperties  int
	MaxRooms       int
	MaxUsers       int
	BillingAddress string
	TaxID          string
	StripeCustomerID string
	TrialEndsAt    *time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Version        int
}

func NewTenant(slug, name, email, currencyCode string) (*Tenant, error) {
	if slug == "" || name == "" || email == "" {
		return nil, fmt.Errorf("slug, name, email required")
	}

	now := time.Now().UTC()
	trialEnd := now.Add(14 * 24 * time.Hour)
	return &Tenant{
		ID:            TenantID(uuid.Must(uuid.NewRandom()).String()),
		Slug:          slug,
		Name:          name,
		Email:         email,
		CurrencyCode:  currencyCode,
		Timezone:      "UTC",
		Status:        TenantStatusTrial,
		Plan:          "trial",
		MaxProperties: 3,
		MaxRooms:      100,
		MaxUsers:      10,
		TrialEndsAt:   &trialEnd,
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
	}, nil
}

func (t *Tenant) Activate(plan string) {
	t.Status = TenantStatusActive
	t.Plan = plan
	t.TrialEndsAt = nil
	t.UpdatedAt = time.Now().UTC()
}

func (t *Tenant) IsTrialExpired() bool {
	if t.Status != TenantStatusTrial || t.TrialEndsAt == nil {
		return false
	}
	return time.Now().UTC().After(*t.TrialEndsAt)
}

// ==================== REFRESH TOKEN ====================

type RefreshTokenID string

type RefreshToken struct {
	ID        RefreshTokenID
	UserID    string
	TenantID  string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	RevokedAt *time.Time
	IPAddress string
	UserAgent string
}

func NewRefreshToken(userID, tenantID, token string, expiresIn time.Duration, ip, ua string) *RefreshToken {
	now := time.Now().UTC()
	hash, _ := bcrypt.GenerateFromPassword([]byte(token), bcrypt.MinCost)
	return &RefreshToken{
		ID:        RefreshTokenID(uuid.Must(uuid.NewRandom()).String()),
		UserID:    userID,
		TenantID:  tenantID,
		TokenHash: string(hash),
		ExpiresAt: now.Add(expiresIn),
		CreatedAt: now,
		IPAddress: ip,
		UserAgent: ua,
	}
}

func (rt *RefreshToken) Verify(token string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(rt.TokenHash), []byte(token))
	return err == nil && rt.RevokedAt == nil && time.Now().UTC().Before(rt.ExpiresAt)
}

func (rt *RefreshToken) Revoke() {
	now := time.Now().UTC()
	rt.RevokedAt = &now
}
