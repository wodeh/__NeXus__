package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AuditContext carries non-repudiable request metadata for logging and compliance.
type AuditContext struct {
	RequestID string    `json:"request_id"`
	UserID    string    `json:"user_id"`
	TenantID  string    `json:"tenant_id"`
	Timestamp time.Time `json:"timestamp"`
	ClientIP  string    `json:"client_ip"`
	UserAgent string    `json:"user_agent"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
}

// NewAuditContext builds an audit context from an HTTP request context.
func NewAuditContext(ctx context.Context, clientIP, userAgent, action, resource string) AuditContext {
	userID, _ := UserIDFromContext(ctx)
	tenantID, _ := TenantIDFromContext(ctx)
	return AuditContext{
		RequestID: uuid.Must(uuid.NewRandom()).String(),
		UserID:    userID,
		TenantID:  tenantID,
		Timestamp: time.Now().UTC(),
		ClientIP:  clientIP,
		UserAgent: userAgent,
		Action:    action,
		Resource:  resource,
	}
}
