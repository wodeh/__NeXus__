package domain

import "time"

type User struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	Email        string                 `json:"email"`
	Name         string                 `json:"name"`
	Phone        string                 `json:"phone"`
	Role         string                 `json:"role"`
	Status       string                 `json:"status"`
	LastLogin    *time.Time             `json:"last_login,omitempty"`
	PasswordHash string                 `json:"-"`
	Permissions  []string               `json:"permissions"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

type Role struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	Name        string                 `json:"name"`
	Permissions []string               `json:"permissions"`
	Description string                 `json:"description"`
	IsSystem    bool                   `json:"is_system"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

type TenantSummary struct {
	ID         string    `json:"id"`
	ExternalID string    `json:"external_id"`
	Name       string    `json:"name"`
	Region     string    `json:"region"`
	Tier       string    `json:"tier"`
	Status     string    `json:"status"`
	UserCount  int       `json:"user_count"`
	RoomCount  int       `json:"room_count"`
	CreatedAt  time.Time `json:"created_at"`
}
