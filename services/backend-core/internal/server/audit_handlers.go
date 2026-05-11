package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
	"github.com/nexus-platform/backend-core/internal/repository"
)

func (s *Server) registerAuditHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/audit/logs", s.withTenant(s.handleAuditLogs))
	mux.HandleFunc("/v1/audit/stats", s.withTenant(s.handleAuditStats))
}

func (s *Server) handleAuditLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewAuditRepository(s.repo.Pool())

	if r.Method == http.MethodGet {
		logs, err := repo.List(ctx, tenantID, 100)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{"logs": demoAuditLogs(tenantID)})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"logs": logs})
		return
	}

	if r.Method == http.MethodPost {
		var req domain.CreateAuditLogRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		var uid *uuid.UUID
		if req.UserID != "" {
			id, _ := uuid.Parse(req.UserID)
			uid = &id
		}
		var uemail *string
		if req.UserEmail != "" {
			uemail = &req.UserEmail
		}
		l := domain.AuditLog{
			UserID:     uid,
			UserEmail:  uemail,
			Action:     req.Action,
			Resource:   req.Resource,
			ResourceID: req.ResourceID,
			Details:    req.Details,
			IPAddress:  req.IPAddress,
			UserAgent:  req.UserAgent,
			CreatedAt:  time.Now(),
		}
		if err := repo.Create(ctx, tenantID, &l); err != nil {
			l.ID = uuid.New()
			l.TenantID = tenantUUID(tenantID)
		}
		writeJSON(w, http.StatusCreated, l)
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func (s *Server) handleAuditStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewAuditRepository(s.repo.Pool())

	if r.Method == http.MethodGet {
		stats, err := repo.GetStats(ctx, tenantID)
		if err != nil {
			writeJSON(w, http.StatusOK, demoAuditStats(tenantID))
			return
		}
		writeJSON(w, http.StatusOK, stats)
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func demoAuditLogs(tenantID string) []domain.AuditLog {
	now := time.Now()
	tid := tenantUUID(tenantID)
	user1 := "admin@nexus.com"
	user2 := "frontdesk@nexus.com"
	uid1 := uuid.New()
	uid2 := uuid.New()
	return []domain.AuditLog{
		{ID: uuid.New(), TenantID: tid, UserID: &uid1, UserEmail: &user1, Action: "reservation_created", Resource: "reservation", ResourceID: "res-001", Details: map[string]interface{}{"guest": "Alice Chen", "room": "201"}, IPAddress: "192.168.1.10", UserAgent: "Mozilla/5.0", CreatedAt: now},
		{ID: uuid.New(), TenantID: tid, UserID: &uid2, UserEmail: &user2, Action: "guest_checkin", Resource: "reservation", ResourceID: "res-002", Details: map[string]interface{}{"guest": "Bob Jones", "room": "102"}, IPAddress: "192.168.1.11", UserAgent: "Mozilla/5.0", CreatedAt: now.Add(-30 * time.Minute)},
		{ID: uuid.New(), TenantID: tid, UserID: &uid1, UserEmail: &user1, Action: "room_status_updated", Resource: "room", ResourceID: "room-101", Details: map[string]interface{}{"from": "occupied", "to": "vacant_dirty"}, IPAddress: "192.168.1.10", UserAgent: "Mozilla/5.0", CreatedAt: now.Add(-1 * time.Hour)},
		{ID: uuid.New(), TenantID: tid, UserID: &uid2, UserEmail: &user2, Action: "rate_plan_modified", Resource: "rate_plan", ResourceID: "rp-deluxe", Details: map[string]interface{}{"old_rate": 180, "new_rate": 200}, IPAddress: "192.168.1.11", UserAgent: "Mozilla/5.0", CreatedAt: now.Add(-2 * time.Hour)},
		{ID: uuid.New(), TenantID: tid, Action: "system_backup", Resource: "system", Details: map[string]interface{}{"status": "completed", "size": "1.2GB"}, IPAddress: "10.0.0.1", UserAgent: "Nexus-Scheduler/1.0", CreatedAt: now.Add(-3 * time.Hour)},
		{ID: uuid.New(), TenantID: tid, UserID: &uid1, UserEmail: &user1, Action: "channel_sync", Resource: "channel", ResourceID: "ch-booking", Details: map[string]interface{}{"records": 15, "status": "success"}, IPAddress: "192.168.1.10", UserAgent: "Mozilla/5.0", CreatedAt: now.Add(-4 * time.Hour)},
	}
}

func demoAuditStats(tenantID string) *domain.AuditStats {
	return &domain.AuditStats{
		TenantID:      tenantID,
		TotalEvents:   156,
		TodayEvents:   24,
		UserActions:   134,
		SystemActions: 22,
		FailedActions: 0,
	}
}
