package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ===================== HOUSEKEEPING TASKS =====================

func (h *Handler) listHousekeepingTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	status := r.URL.Query().Get("status")
	floor := r.URL.Query().Get("floor")
	staffID := r.URL.Query().Get("staff_id")
	shift := r.URL.Query().Get("shift")

	tasks, err := h.store.Housekeeping.ListTasks(ctx, tenantID, status, floor, staffID, shift, 100, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list tasks")
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"tasks": tasks})
}

func (h *Handler) createHousekeepingTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	var task domain.HousekeepingTask
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	task.TenantID = tenantID
	task.Status = "pending"
	if err := h.store.Housekeeping.CreateTask(ctx, &task); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create task")
		return
	}
	respondJSON(w, http.StatusCreated, task)
}

func (h *Handler) getHousekeepingTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	taskID := chi.URLParam(r, "taskId")
	task, err := h.store.Housekeeping.GetTask(ctx, tenantID, taskID)
	if err != nil {
		respondError(w, http.StatusNotFound, "task not found")
		return
	}
	respondJSON(w, http.StatusOK, task)
}

func (h *Handler) updateHousekeepingTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	taskID := chi.URLParam(r, "taskId")
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.store.Housekeeping.UpdateTask(ctx, tenantID, taskID, updates); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update task")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) deleteHousekeepingTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	taskID := chi.URLParam(r, "taskId")
	if err := h.store.Housekeeping.DeleteTask(ctx, tenantID, taskID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete task")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (h *Handler) assignHousekeepingTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	taskID := chi.URLParam(r, "taskId")
	var body struct {
		StaffID    string `json:"staff_id"`
		AssignedBy string `json:"assigned_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.store.Housekeeping.AssignTask(ctx, tenantID, taskID, body.StaffID, body.AssignedBy); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to assign task")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "assigned"})
}

func (h *Handler) startHousekeepingTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	taskID := chi.URLParam(r, "taskId")
	var body struct {
		CardMode string `json:"card_mode,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if err := h.store.Housekeeping.StartTask(ctx, tenantID, taskID, body.CardMode); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to start task")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "in_progress"})
}

func (h *Handler) completeHousekeepingTask(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	taskID := chi.URLParam(r, "taskId")
	var body struct {
		ActualMin     int                     `json:"actual_minutes"`
		ChecklistDone []string                `json:"checklist_done"`
		SuppliesUsed  []domain.SupplyUsage    `json:"supplies_used"`
		Notes         string                  `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.store.Housekeeping.CompleteTask(ctx, tenantID, taskID, body.ActualMin, body.ChecklistDone, body.SuppliesUsed, body.Notes); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to complete task")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

// ===================== BULK ASSIGN =====================

func (h *Handler) bulkAssignTasks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	var body struct {
		TaskIDs []string `json:"task_ids"`
		StaffID string   `json:"staff_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	assigned := 0
	for _, taskID := range body.TaskIDs {
		if err := h.store.Housekeeping.AssignTask(ctx, tenantID, taskID, body.StaffID, ""); err == nil {
			assigned++
		}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"assigned": assigned, "total": len(body.TaskIDs)})
}

// ===================== STAFF =====================

func (h *Handler) listHousekeepingStaff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	activeOnly := r.URL.Query().Get("active_only") == "true"
	shift := r.URL.Query().Get("shift")
	staff, err := h.store.Housekeeping.ListStaff(ctx, tenantID, activeOnly, shift)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list staff")
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"staff": staff})
}

func (h *Handler) createHousekeepingStaff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	var staff domain.HousekeepingStaff
	if err := json.NewDecoder(r.Body).Decode(&staff); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	staff.TenantID = tenantID
	if err := h.store.Housekeeping.CreateStaff(ctx, &staff); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create staff")
		return
	}
	respondJSON(w, http.StatusCreated, staff)
}

func (h *Handler) updateHousekeepingStaff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	staffID := chi.URLParam(r, "staffId")
	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.store.Housekeeping.UpdateStaff(ctx, tenantID, staffID, updates); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update staff")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (h *Handler) deleteHousekeepingStaff(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	staffID := chi.URLParam(r, "staffId")
	if err := h.store.Housekeeping.DeleteStaff(ctx, tenantID, staffID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to delete staff")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// ===================== CHECKLISTS =====================

func (h *Handler) getChecklistTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	roomType := r.URL.Query().Get("room_type")
	taskType := r.URL.Query().Get("task_type")
	if roomType == "" {
		respondError(w, http.StatusBadRequest, "room_type required")
		return
	}
	if taskType == "" {
		taskType = "clean"
	}
	tmpl, err := h.store.Housekeeping.GetChecklistTemplate(ctx, tenantID, roomType, taskType)
	if err != nil {
		respondJSON(w, http.StatusOK, map[string]interface{}{"items": []domain.ChecklistItem{}})
		return
	}
	respondJSON(w, http.StatusOK, tmpl)
}

func (h *Handler) saveChecklistTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	var tmpl domain.CleaningChecklistTemplate
	if err := json.NewDecoder(r.Body).Decode(&tmpl); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	tmpl.TenantID = tenantID
	if err := h.store.Housekeeping.SaveChecklistTemplate(ctx, &tmpl); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to save checklist")
		return
	}
	respondJSON(w, http.StatusOK, tmpl)
}

// ===================== CLEANER CARDS =====================

func (h *Handler) listCleanerCards(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	cards, err := h.store.Housekeeping.ListCleanerCards(ctx, tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list cards")
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"cards": cards})
}

func (h *Handler) createCleanerCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	var card domain.LockAccessCodeWithMode
	if err := json.NewDecoder(r.Body).Decode(&card); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.store.Housekeeping.CreateCleanerCard(ctx, tenantID, &card); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to create card")
		return
	}
	respondJSON(w, http.StatusCreated, card)
}

func (h *Handler) revokeCleanerCard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	cardID := chi.URLParam(r, "cardId")
	if err := h.store.Housekeeping.RevokeCleanerCard(ctx, tenantID, cardID); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to revoke card")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

// ===================== SUPPLIES =====================

func (h *Handler) listSupplies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	category := r.URL.Query().Get("category")
	supplies, err := h.store.Housekeeping.ListSupplies(ctx, tenantID, category)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to list supplies")
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"supplies": supplies})
}

func (h *Handler) updateSupplyStock(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	var body struct {
		ItemID string `json:"item_id"`
		Delta  int    `json:"delta"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.store.Housekeeping.UpdateSupplyStock(ctx, tenantID, body.ItemID, body.Delta); err != nil {
		respondError(w, http.StatusInternalServerError, "failed to update stock")
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// ===================== DASHBOARD =====================

func (h *Handler) getHousekeepingDashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	stats, err := h.store.Housekeeping.GetDashboardStats(ctx, tenantID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get dashboard")
		return
	}
	respondJSON(w, http.StatusOK, stats)
}

func (h *Handler) getStaffPerformance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := tenantID(r)
	staffID := chi.URLParam(r, "staffId")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" {
		from = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if to == "" {
		to = time.Now().Format("2006-01-02")
	}
	fromTime, _ := time.Parse("2006-01-02", from)
	toTime, _ := time.Parse("2006-01-02", to)
	perf, err := h.store.Housekeeping.GetStaffPerformance(ctx, tenantID, staffID, fromTime, toTime)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "failed to get performance")
		return
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"performance": perf})
}
