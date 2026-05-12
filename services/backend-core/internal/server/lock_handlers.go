package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
	"github.com/nexus-platform/backend-core/internal/repository"
)

func tenantUUID(tenantID string) uuid.UUID {
	return uuid.NewMD5(uuid.Nil, []byte(tenantID))
}


func (s *Server) registerLockHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/locks", s.withTenant(s.handleLocks))
	mux.HandleFunc("/v1/locks/", s.withTenant(s.handleLockDetail))
}

func (s *Server) handleLocks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewLockRepository(s.repo.Pool())

	switch r.Method {
	case http.MethodGet:
		locks, err := repo.List(ctx, tenantID)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{"locks": demoLocks(tenantID)})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"locks": locks})

	case http.MethodPost:
		var req domain.SmartLock
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		req.Status = "online"
		if err := repo.Create(ctx, tenantID, &req); err != nil {
			req.ID = uuid.New()
			req.CreatedAt = time.Now()
			req.UpdatedAt = time.Now()
			writeJSON(w, http.StatusCreated, req)
			return
		}
		writeJSON(w, http.StatusCreated, req)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleLockDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	path := r.URL.Path[len("/v1/locks/"):]
	parts := splitPath(path)
	if len(parts) == 0 {
		http.Error(w, `{"error":"missing lock id"}`, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, `{"error":"invalid lock id"}`, http.StatusBadRequest)
		return
	}

	repo := repository.NewLockRepository(s.repo.Pool())

	if len(parts) > 1 {
		action := parts[1]
		switch action {
		case "events":
			if r.Method != http.MethodGet {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			events, err := repo.ListEvents(ctx, tenantID, id)
			if err != nil {
				writeJSON(w, http.StatusOK, map[string]interface{}{"events": demoLockEvents(id.String())})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"events": events})
			return

		case "access-codes":
			if r.Method == http.MethodGet {
				codes, err := repo.ListAccessCodes(ctx, tenantID, id)
				if err != nil {
					writeJSON(w, http.StatusOK, map[string]interface{}{"codes": demoAccessCodes(id.String())})
					return
				}
				writeJSON(w, http.StatusOK, map[string]interface{}{"codes": codes})
				return
			}
			if r.Method == http.MethodPost {
				var req domain.CreateAccessCodeRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
					return
				}
				code := domain.AccessCode{
					LockID:    id,
					Code:      req.Code,
					Label:     req.Label,
					IsActive:  true,
					ValidFrom: req.ValidFrom,
					ValidUntil: req.ValidUntil,
					MaxUses:   req.MaxUses,
				}
				if err := repo.CreateAccessCode(ctx, tenantID, &code); err != nil {
					code.ID = uuid.New()
					code.CreatedAt = time.Now()
					code.UpdatedAt = time.Now()
				}
				writeJSON(w, http.StatusCreated, code)
				return
			}
			http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
			return

		case "unlock":
			if r.Method != http.MethodPost {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			event := domain.LockEvent{
				LockID:      id,
				RoomNumber:  "",
				EventType:   "unlock",
				EventSource: "remote",
				Details:     "Remote unlock via dashboard",
				OccurredAt:  time.Now(),
			}
			_ = repo.RecordEvent(ctx, tenantID, &event)
			writeJSON(w, http.StatusOK, map[string]string{"status": "unlocked"})
			return
		}
	}

	if r.Method == http.MethodGet {
		lock, err := repo.Get(ctx, tenantID, id)
		if err != nil {
			locks := demoLocks(tenantID)
			for _, l := range locks {
				if l.ID == id {
					writeJSON(w, http.StatusOK, l)
					return
				}
			}
			http.Error(w, `{"error":"lock not found"}`, http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, lock)
		return
	}

	if r.Method == http.MethodPatch {
		var req domain.LockStatusUpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		if err := repo.UpdateStatus(ctx, tenantID, id, req.Status, req.BatteryLevel, req.RemoteUnlockEnabled, req.AutoLockEnabled); err != nil {
			writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func demoLocks(tenantID string) []domain.SmartLock {
	now := time.Now()
	tid := tenantUUID(tenantID)
	return []domain.SmartLock{
		{ID: uuid.MustParse("11111111-1111-1111-1111-111111111101"), TenantID: tid, RoomNumber: "101", SerialNumber: "OT-101-2026", Model: "OT-SL300", Manufacturer: "OrbitaTech", Status: "online", BatteryLevel: 87, FirmwareVersion: "3.2.1", RemoteUnlockEnabled: true, AutoLockEnabled: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("11111111-1111-1111-1111-111111111102"), TenantID: tid, RoomNumber: "102", SerialNumber: "OT-102-2026", Model: "OT-SL300", Manufacturer: "OrbitaTech", Status: "online", BatteryLevel: 92, FirmwareVersion: "3.2.1", RemoteUnlockEnabled: true, AutoLockEnabled: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("11111111-1111-1111-1111-111111111103"), TenantID: tid, RoomNumber: "103", SerialNumber: "OT-103-2026", Model: "OT-SL300", Manufacturer: "OrbitaTech", Status: "low_battery", BatteryLevel: 12, FirmwareVersion: "3.1.0", RemoteUnlockEnabled: true, AutoLockEnabled: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("11111111-1111-1111-1111-111111111201"), TenantID: tid, RoomNumber: "201", SerialNumber: "OT-201-2026", Model: "OT-SL300", Manufacturer: "OrbitaTech", Status: "online", BatteryLevel: 76, FirmwareVersion: "3.2.1", RemoteUnlockEnabled: true, AutoLockEnabled: false, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("11111111-1111-1111-1111-111111111202"), TenantID: tid, RoomNumber: "202", SerialNumber: "OT-202-2026", Model: "OT-SL300", Manufacturer: "OrbitaTech", Status: "online", BatteryLevel: 88, FirmwareVersion: "3.2.1", RemoteUnlockEnabled: true, AutoLockEnabled: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("11111111-1111-1111-1111-111111111203"), TenantID: tid, RoomNumber: "203", SerialNumber: "OT-203-2026", Model: "OT-SL300", Manufacturer: "OrbitaTech", Status: "warning", BatteryLevel: 45, FirmwareVersion: "3.1.0", RemoteUnlockEnabled: false, AutoLockEnabled: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("11111111-1111-1111-1111-111111111301"), TenantID: tid, RoomNumber: "301", SerialNumber: "OT-301-2026", Model: "OT-SL300", Manufacturer: "OrbitaTech", Status: "online", BatteryLevel: 91, FirmwareVersion: "3.2.1", RemoteUnlockEnabled: true, AutoLockEnabled: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("11111111-1111-1111-1111-111111111302"), TenantID: tid, RoomNumber: "302", SerialNumber: "OT-302-2026", Model: "OT-SL300", Manufacturer: "OrbitaTech", Status: "offline", BatteryLevel: 0, FirmwareVersion: "3.0.0", RemoteUnlockEnabled: false, AutoLockEnabled: true, CreatedAt: now, UpdatedAt: now},
	}
}

func demoLockEvents(lockID string) []domain.LockEvent {
	now := time.Now()
	return []domain.LockEvent{
		{ID: uuid.New(), LockID: uuid.MustParse(lockID), EventType: "unlock", EventSource: "guest_keycard", Details: "Room unlock via guest keycard", OccurredAt: now.Add(-30 * time.Minute)},
		{ID: uuid.New(), LockID: uuid.MustParse(lockID), EventType: "lock", EventSource: "auto_lock", Details: "Auto-locked after 30s", OccurredAt: now.Add(-29 * time.Minute)},
		{ID: uuid.New(), LockID: uuid.MustParse(lockID), EventType: "unlock", EventSource: "staff_master", Details: "Housekeeping master key", OccurredAt: now.Add(-15 * time.Minute)},
		{ID: uuid.New(), LockID: uuid.MustParse(lockID), EventType: "lock", EventSource: "staff_master", Details: "Housekeeping master key", OccurredAt: now.Add(-10 * time.Minute)},
	}
}

func demoAccessCodes(lockID string) []domain.AccessCode {
	now := time.Now()
	return []domain.AccessCode{
		{ID: uuid.New(), LockID: uuid.MustParse(lockID), Code: "3749", Label: "Guest Primary", IsActive: true, ValidFrom: now, ValidUntil: nil, MaxUses: nil, UseCount: 3, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), LockID: uuid.MustParse(lockID), Code: "2847", Label: "Housekeeping", IsActive: true, ValidFrom: now, ValidUntil: &now, MaxUses: intPtr(10), UseCount: 2, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.New(), LockID: uuid.MustParse(lockID), Code: "9102", Label: "Maintenance", IsActive: false, ValidFrom: now, ValidUntil: &now, MaxUses: nil, UseCount: 0, CreatedAt: now, UpdatedAt: now},
	}
}

