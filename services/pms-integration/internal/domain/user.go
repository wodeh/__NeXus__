package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Role defines user permission levels.
type Role string

const (
	RoleOwner       Role = "owner"
	RoleManager     Role = "manager"
	RoleReceptionist Role = "receptionist"
	RoleHousekeeper Role = "housekeeper"
	RoleReadOnly    Role = "readonly"
)

// User represents a staff member or admin in the PMS.
type User struct {
	ID           string
	TenantID     string
	PropertyID   string
	Email        string
	PasswordHash string
	FirstName    string
	LastName     string
	Role         Role
	IsActive     bool
	LastLoginAt  *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Version      int
}

// NewUser creates a new user with validation.
func NewUser(tenantID, email, firstName, lastName string, role Role) (*User, error) {
	if tenantID == "" {
		return nil, errors.New("tenant_id is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if firstName == "" || lastName == "" {
		return nil, errors.New("first and last name are required")
	}
	if role == "" {
		role = RoleReceptionist
	}
	now := time.Now().UTC()
	return &User{
		ID:        uuid.Must(uuid.NewRandom()).String(),
		TenantID:  tenantID,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Role:      role,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
		Version:   1,
	}, nil
}

// FullName returns the user's full name.
func (u *User) FullName() string {
	return u.FirstName + " " + u.LastName
}

// CanAccess checks if user role meets minimum required role.
func (u *User) CanAccess(required Role) bool {
	roleHierarchy := map[Role]int{
		RoleReadOnly:    0,
		RoleHousekeeper: 1,
		RoleReceptionist: 2,
		RoleManager:     3,
		RoleOwner:       4,
	}
	return roleHierarchy[u.Role] >= roleHierarchy[required]
}

// Session represents an active JWT session.
type Session struct {
	ID        string
	UserID    string
	TenantID  string
	Token     string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// IsExpired returns true if session is past expiry.
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}
