// Package api provides HTTP REST handlers for financial, housekeeping, and maintenance operations.
package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ==================== FOLIO ====================

type createFolioRequest struct {
	ReservationID string `json:"reservation_id"`
	GuestID       string `json:"guest_id"`
	CurrencyCode  string `json:"currency_code"`
	IsMaster      bool   `json:"is_master"`
	MasterFolioID string `json:"master_folio_id,omitempty"`
	Notes         string `json:"notes"`
}

type folioResponse struct {
	ID               string    `json:"id"`
	ReservationID    string    `json:"reservation_id"`
	GuestID          string    `json:"guest_id"`
	Status           string    `json:"status"`
	Balance          float64   `json:"balance"`
	TotalCharges     float64   `json:"total_charges"`
	TotalPayments    float64   `json:"total_payments"`
	TotalAdjustments float64   `json:"total_adjustments"`
	CurrencyCode     string    `json:"currency_code"`
	IsMaster         bool      `json:"is_master"`
	MasterFolioID    *string   `json:"master_folio_id,omitempty"`
	Notes            string    `json:"notes"`
	Version          int       `json:"version"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func toFolioResponse(f *domain.Folio) folioResponse {
	return folioResponse{
		ID:               string(f.ID),
		ReservationID:    f.ReservationID,
		GuestID:          f.GuestID,
		Status:           string(f.Status),
		Balance:          f.Balance,
		TotalCharges:     f.TotalCharges,
		TotalPayments:    f.TotalPayments,
		TotalAdjustments: f.TotalAdjustments,
		CurrencyCode:     f.CurrencyCode,
		IsMaster:         f.IsMaster,
		MasterFolioID:    f.MasterFolioID,
		Notes:            f.Notes,
		Version:          f.Version,
		CreatedAt:        f.CreatedAt,
		UpdatedAt:        f.UpdatedAt,
	}
}

func (h *Handler) createFolio(w http.ResponseWriter, r *http.Request) {
	var req createFolioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := tenantID(r)
	folio := domain.NewFolio(tenantID, req.ReservationID, req.GuestID, req.CurrencyCode)
	folio.IsMaster = req.IsMaster
	if req.MasterFolioID != "" {
		folio.MasterFolioID = &req.MasterFolioID
	}
	folio.Notes = req.Notes

	if err := h.store.Folios.Create(r.Context(), folio); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, toFolioResponse(folio))
}

func (h *Handler) getFolio(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "folioId")

	folio, err := h.store.Folios.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "folio not found")
		return
	}

	respondJSON(w, http.StatusOK, toFolioResponse(folio))
}

func (h *Handler) getFolioByReservation(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	reservationID := chi.URLParam(r, "reservationId")

	folio, err := h.store.Folios.GetByReservation(r.Context(), tenantID, reservationID)
	if err != nil {
		respondError(w, http.StatusNotFound, "folio not found for reservation")
		return
	}

	respondJSON(w, http.StatusOK, toFolioResponse(folio))
}

func (h *Handler) listFolios(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	folios, err := h.store.Folios.ListByTenant(r.Context(), tenantID, status, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]folioResponse, len(folios))
	for i, f := range folios {
		resp[i] = toFolioResponse(f)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"folios": resp,
		"total":  len(resp),
	})
}

func (h *Handler) closeFolio(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "folioId")

	if err := h.store.Folios.CloseFolio(r.Context(), tenantID, id); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "closed"})
}

// ==================== CHARGES ====================

type createChargeRequest struct {
	FolioID       string  `json:"folio_id"`
	ReservationID string  `json:"reservation_id,omitempty"`
	RoomID        string  `json:"room_id,omitempty"`
	ChargeType    string  `json:"charge_type"`
	Category      string  `json:"category"`
	Description   string  `json:"description"`
	Quantity      float64 `json:"quantity"`
	UnitPrice     float64 `json:"unit_price"`
	TaxRate       float64 `json:"tax_rate"`
	ChargeDate    string  `json:"charge_date"`
	PostedBy      string  `json:"posted_by"`
	RevenueCenter string  `json:"revenue_center"`
	Source        string  `json:"source"`
}

type chargeResponse struct {
	ID            string                 `json:"id"`
	FolioID       string                 `json:"folio_id"`
	ReservationID *string                `json:"reservation_id,omitempty"`
	RoomID        *string                `json:"room_id,omitempty"`
	ChargeType    string                 `json:"charge_type"`
	Category      string                 `json:"category"`
	Description   string                 `json:"description"`
	Quantity      float64                `json:"quantity"`
	UnitPrice     float64                `json:"unit_price"`
	TotalAmount   float64                `json:"total_amount"`
	TaxAmount     float64                `json:"tax_amount"`
	TaxRate       float64                `json:"tax_rate"`
	CurrencyCode  string                 `json:"currency_code"`
	ChargeDate    time.Time              `json:"charge_date"`
	PostedBy      string                 `json:"posted_by"`
	IsRevenue     bool                   `json:"is_revenue"`
	RevenueCenter string                 `json:"revenue_center"`
	Source        string                 `json:"source"`
	IsVoided      bool                   `json:"is_voided"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

func toChargeResponse(c *domain.Charge) chargeResponse {
	return chargeResponse{
		ID:            string(c.ID),
		FolioID:       c.FolioID,
		ReservationID: c.ReservationID,
		RoomID:        c.RoomID,
		ChargeType:    c.ChargeType,
		Category:      c.Category,
		Description:   c.Description,
		Quantity:      c.Quantity,
		UnitPrice:     c.UnitPrice,
		TotalAmount:   c.TotalAmount,
		TaxAmount:     c.TaxAmount,
		TaxRate:       c.TaxRate,
		CurrencyCode:  c.CurrencyCode,
		ChargeDate:    c.ChargeDate,
		PostedBy:      c.PostedBy,
		IsRevenue:     c.IsRevenue,
		RevenueCenter: c.RevenueCenter,
		Source:        c.Source,
		IsVoided:      c.IsVoided,
		CreatedAt:     c.CreatedAt,
		UpdatedAt:     c.UpdatedAt,
	}
}

func (h *Handler) createCharge(w http.ResponseWriter, r *http.Request) {
	var req createChargeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := tenantID(r)
	charge := domain.NewCharge(tenantID, req.FolioID, req.ChargeType, req.Description, req.Quantity, req.UnitPrice)
	if req.Category != "" {
		charge.Category = req.Category
	}
	if req.ReservationID != "" {
		charge.ReservationID = &req.ReservationID
	}
	if req.RoomID != "" {
		charge.RoomID = &req.RoomID
	}
	charge.SetTax(req.TaxRate)
	charge.PostedBy = req.PostedBy
	charge.RevenueCenter = req.RevenueCenter
	if req.Source != "" {
		charge.Source = req.Source
	}
	if req.ChargeDate != "" {
		date, err := time.Parse("2006-01-02", req.ChargeDate)
		if err == nil {
			charge.ChargeDate = date
		}
	}

	if err := h.store.Charges.Create(r.Context(), charge); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Auto-recalculate folio balance
	_ = h.store.Folios.RecalculateBalance(r.Context(), req.FolioID)

	respondJSON(w, http.StatusCreated, toChargeResponse(charge))
}

func (h *Handler) listCharges(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	folioID := chi.URLParam(r, "folioId")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	charges, err := h.store.Charges.ListByFolio(r.Context(), tenantID, folioID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]chargeResponse, len(charges))
	for i, c := range charges {
		resp[i] = toChargeResponse(c)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"charges": resp,
		"total":   len(resp),
	})
}

func (h *Handler) voidCharge(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	chargeID := chi.URLParam(r, "chargeId")

	var req struct {
		VoidedBy string `json:"voided_by"`
		Reason   string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.store.Charges.VoidCharge(r.Context(), tenantID, chargeID, req.VoidedBy, req.Reason); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "voided"})
}

// ==================== PAYMENTS ====================

type createPaymentRequest struct {
	FolioID       string  `json:"folio_id"`
	ReservationID string  `json:"reservation_id,omitempty"`
	GuestID       string  `json:"guest_id,omitempty"`
	PaymentMethod string  `json:"payment_method"`
	Amount        float64 `json:"amount"`
	CurrencyCode  string  `json:"currency_code"`
	ReferenceNum  string  `json:"reference_number"`
	TransactionID string  `json:"transaction_id"`
	ProcessedBy   string  `json:"processed_by"`
}

type paymentResponse struct {
	ID            string    `json:"id"`
	FolioID       string    `json:"folio_id"`
	PaymentMethod string    `json:"payment_method"`
	Amount        float64   `json:"amount"`
	CurrencyCode  string    `json:"currency_code"`
	Status        string    `json:"status"`
	ReferenceNum  string    `json:"reference_number"`
	TransactionID string    `json:"transaction_id"`
	CreatedAt     time.Time `json:"created_at"`
}

func toPaymentResponse(p *domain.Payment) paymentResponse {
	return paymentResponse{
		ID:            string(p.ID),
		FolioID:       p.FolioID,
		PaymentMethod: p.PaymentMethod,
		Amount:        p.Amount,
		CurrencyCode:  p.CurrencyCode,
		Status:        string(p.Status),
		ReferenceNum:  p.ReferenceNumber,
		TransactionID: p.TransactionID,
		CreatedAt:     p.CreatedAt,
	}
}

func (h *Handler) createPayment(w http.ResponseWriter, r *http.Request) {
	var req createPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := tenantID(r)
	payment := domain.NewPayment(tenantID, req.FolioID, req.Amount, req.PaymentMethod)
	if req.CurrencyCode != "" {
		payment.CurrencyCode = req.CurrencyCode
	}
	payment.ReferenceNumber = req.ReferenceNum
	payment.TransactionID = req.TransactionID
	payment.ProcessedBy = req.ProcessedBy

	if err := h.store.Payments.Create(r.Context(), payment); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Auto-recalculate folio balance
	_ = h.store.Folios.RecalculateBalance(r.Context(), req.FolioID)

	respondJSON(w, http.StatusCreated, toPaymentResponse(payment))
}

func (h *Handler) listPayments(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	folioID := chi.URLParam(r, "folioId")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	payments, err := h.store.Payments.ListByFolio(r.Context(), tenantID, folioID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]paymentResponse, len(payments))
	for i, p := range payments {
		resp[i] = toPaymentResponse(p)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"payments": resp,
		"total":    len(resp),
	})
}

func (h *Handler) refundPayment(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	paymentID := chi.URLParam(r, "paymentId")

	var req struct {
		Amount float64 `json:"amount"`
		Reason string  `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.store.Payments.Refund(r.Context(), tenantID, paymentID, req.Amount, req.Reason); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "refunded"})
}

// ==================== HOUSEKEEPING ====================

type createHousekeepingTaskRequest struct {
	PropertyID        string `json:"property_id"`
	RoomID            string `json:"room_id"`
	TaskType          string `json:"task_type"`
	Priority          string `json:"priority"`
	AssignedTo        string `json:"assigned_to"`
	ScheduledAt       string `json:"scheduled_at"`
	EstimatedDuration int    `json:"estimated_duration_minutes"`
	Notes             string `json:"notes"`
}

type housekeepingTaskResponse struct {
	ID                string    `json:"id"`
	PropertyID        string    `json:"property_id"`
	RoomID            string    `json:"room_id"`
	TaskType          string    `json:"task_type"`
	Priority          string    `json:"priority"`
	Status            string    `json:"status"`
	AssignedTo        string    `json:"assigned_to"`
	ScheduledAt       *time.Time `json:"scheduled_at,omitempty"`
	StartedAt         *time.Time `json:"started_at,omitempty"`
	CompletedAt       *time.Time `json:"completed_at,omitempty"`
	EstimatedDuration int       `json:"estimated_duration_minutes"`
	ActualDuration    int       `json:"actual_duration_minutes"`
	Notes             string    `json:"notes"`
	InspectionScore   int       `json:"inspection_score"`
	Version           int       `json:"version"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func toHousekeepingTaskResponse(t *domain.HousekeepingTask) housekeepingTaskResponse {
	return housekeepingTaskResponse{
		ID:                string(t.ID),
		PropertyID:        t.PropertyID,
		RoomID:            t.RoomID,
		TaskType:          t.TaskType,
		Priority:          string(t.Priority),
		Status:            string(t.Status),
		AssignedTo:        t.AssignedTo,
		ScheduledAt:       t.ScheduledAt,
		StartedAt:         t.StartedAt,
		CompletedAt:       t.CompletedAt,
		EstimatedDuration: t.EstimatedDuration,
		ActualDuration:    t.ActualDuration,
		Notes:             t.Notes,
		InspectionScore:   t.InspectionScore,
		Version:           t.Version,
		CreatedAt:         t.CreatedAt,
		UpdatedAt:         t.UpdatedAt,
	}
}

func (h *Handler) createHousekeepingTask(w http.ResponseWriter, r *http.Request) {
	var req createHousekeepingTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := tenantID(r)
	task := domain.NewHousekeepingTask(tenantID, req.PropertyID, req.RoomID, req.TaskType)
	if req.Priority != "" {
		task.Priority = domain.TaskPriority(req.Priority)
	}
	if req.AssignedTo != "" {
		task.Assign(req.AssignedTo)
	}
	if req.ScheduledAt != "" {
		t, err := time.Parse(time.RFC3339, req.ScheduledAt)
		if err == nil {
			task.ScheduledAt = &t
		}
	}
	task.EstimatedDuration = req.EstimatedDuration
	task.Notes = req.Notes

	if err := h.store.HousekeepingTasks.Create(r.Context(), task); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, toHousekeepingTaskResponse(task))
}

func (h *Handler) getHousekeepingTask(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "taskId")

	task, err := h.store.HousekeepingTasks.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "task not found")
		return
	}

	respondJSON(w, http.StatusOK, toHousekeepingTaskResponse(task))
}

func (h *Handler) listHousekeepingTasks(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	propertyID := r.URL.Query().Get("property_id")
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	var tasks []*domain.HousekeepingTask
	var err error

	if propertyID != "" {
		tasks, err = h.store.HousekeepingTasks.ListByProperty(r.Context(), tenantID, propertyID, status, limit, offset)
	} else {
		assignedTo := r.URL.Query().Get("assigned_to")
		tasks, err = h.store.HousekeepingTasks.ListByAssignedTo(r.Context(), tenantID, assignedTo, status, limit, offset)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]housekeepingTaskResponse, len(tasks))
	for i, t := range tasks {
		resp[i] = toHousekeepingTaskResponse(t)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"tasks": resp,
		"total": len(resp),
	})
}

func (h *Handler) getHousekeepingBoard(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	propertyID := chi.URLParam(r, "propertyId")

	tasks, err := h.store.HousekeepingTasks.ListBoard(r.Context(), tenantID, propertyID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]housekeepingTaskResponse, len(tasks))
	for i, t := range tasks {
		resp[i] = toHousekeepingTaskResponse(t)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"board": resp,
		"total": len(resp),
	})
}

func (h *Handler) startHousekeepingTask(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "taskId")

	task, err := h.store.HousekeepingTasks.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "task not found")
		return
	}

	task.Start()
	if err := h.store.HousekeepingTasks.Update(r.Context(), task); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toHousekeepingTaskResponse(task))
}

func (h *Handler) completeHousekeepingTask(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "taskId")

	task, err := h.store.HousekeepingTasks.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "task not found")
		return
	}

	task.Complete()
	if err := h.store.HousekeepingTasks.Update(r.Context(), task); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Update room housekeeping status to clean
	if task.RoomID != "" {
		room, _ := h.store.Rooms.GetByID(r.Context(), tenantID, task.RoomID)
		if room != nil {
			room.HousekeepingStatus = domain.HousekeepingClean
			_ = h.store.Rooms.UpdateHousekeeping(r.Context(), room)
		}
	}

	respondJSON(w, http.StatusOK, toHousekeepingTaskResponse(task))
}

func (h *Handler) inspectHousekeepingTask(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "taskId")

	var req struct {
		Score int    `json:"score"`
		By    string `json:"by"`
		Notes string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	task, err := h.store.HousekeepingTasks.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "task not found")
		return
	}

	task.Inspect(req.Score, req.By, req.Notes)
	if err := h.store.HousekeepingTasks.Update(r.Context(), task); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toHousekeepingTaskResponse(task))
}

// ==================== WORK ORDERS ====================

type createWorkOrderRequest struct {
	PropertyID    string  `json:"property_id"`
	RoomID        string  `json:"room_id,omitempty"`
	AssetID       string  `json:"asset_id,omitempty"`
	Category      string  `json:"category"`
	Priority      string  `json:"priority"`
	Title         string  `json:"title"`
	Description   string  `json:"description"`
	ReportedBy    string  `json:"reported_by"`
	EstimatedCost float64 `json:"estimated_cost"`
}

type workOrderResponse struct {
	ID              string    `json:"id"`
	WorkOrderNumber string    `json:"work_order_number"`
	PropertyID      string    `json:"property_id"`
	RoomID          *string   `json:"room_id,omitempty"`
	Category        string    `json:"category"`
	Priority        string    `json:"priority"`
	Status          string    `json:"status"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	ReportedBy      string    `json:"reported_by"`
	AssignedTo      string    `json:"assigned_to"`
	EstimatedCost   float64   `json:"estimated_cost"`
	ActualCost      float64   `json:"actual_cost"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func toWorkOrderResponse(wo *domain.WorkOrder) workOrderResponse {
	return workOrderResponse{
		ID:              string(wo.ID),
		WorkOrderNumber: wo.WorkOrderNumber,
		PropertyID:      wo.PropertyID,
		RoomID:          wo.RoomID,
		Category:        wo.Category,
		Priority:        string(wo.Priority),
		Status:          string(wo.Status),
		Title:           wo.Title,
		Description:     wo.Description,
		ReportedBy:      wo.ReportedBy,
		AssignedTo:      wo.AssignedTo,
		EstimatedCost:   wo.EstimatedCost,
		ActualCost:      wo.ActualCost,
		CreatedAt:       wo.CreatedAt,
		UpdatedAt:       wo.UpdatedAt,
	}
}

func (h *Handler) createWorkOrder(w http.ResponseWriter, r *http.Request) {
	var req createWorkOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tenantID := tenantID(r)
	woNumber := fmt.Sprintf("WO-%d-%s", time.Now().Unix(), tenantID[:8])
	wo := domain.NewWorkOrder(tenantID, req.PropertyID, woNumber, req.Category, req.Title)
	wo.Description = req.Description
	wo.ReportedBy = req.ReportedBy
	wo.EstimatedCost = req.EstimatedCost
	if req.Priority != "" {
		wo.Priority = domain.WorkOrderPriority(req.Priority)
	}
	if req.RoomID != "" {
		wo.RoomID = &req.RoomID
	}
	if req.AssetID != "" {
		wo.AssetID = &req.AssetID
	}

	if err := h.store.WorkOrders.Create(r.Context(), wo); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, toWorkOrderResponse(wo))
}

func (h *Handler) getWorkOrder(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "workOrderId")

	wo, err := h.store.WorkOrders.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "work order not found")
		return
	}

	respondJSON(w, http.StatusOK, toWorkOrderResponse(wo))
}

func (h *Handler) listWorkOrders(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	propertyID := r.URL.Query().Get("property_id")
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	var orders []*domain.WorkOrder
	var err error

	if propertyID != "" {
		orders, err = h.store.WorkOrders.ListByProperty(r.Context(), tenantID, propertyID, status, limit, offset)
	} else if status != "" {
		orders, err = h.store.WorkOrders.ListByStatus(r.Context(), tenantID, status, limit, offset)
	} else {
		orders, err = h.store.WorkOrders.ListByProperty(r.Context(), tenantID, "", status, limit, offset)
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := make([]workOrderResponse, len(orders))
	for i, wo := range orders {
		resp[i] = toWorkOrderResponse(wo)
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"work_orders": resp,
		"total":       len(resp),
	})
}

func (h *Handler) assignWorkOrder(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "workOrderId")

	var req struct {
		AssignedTo string `json:"assigned_to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	wo, err := h.store.WorkOrders.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "work order not found")
		return
	}

	wo.Assign(req.AssignedTo)
	if err := h.store.WorkOrders.Update(r.Context(), wo); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toWorkOrderResponse(wo))
}

func (h *Handler) startWorkOrder(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "workOrderId")

	wo, err := h.store.WorkOrders.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "work order not found")
		return
	}

	wo.Start()
	if err := h.store.WorkOrders.Update(r.Context(), wo); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toWorkOrderResponse(wo))
}

func (h *Handler) completeWorkOrder(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "workOrderId")

	var req struct {
		Resolution string  `json:"resolution"`
		ActualCost float64 `json:"actual_cost"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	wo, err := h.store.WorkOrders.GetByID(r.Context(), tenantID, id)
	if err != nil {
		respondError(w, http.StatusNotFound, "work order not found")
		return
	}

	wo.Complete(req.Resolution, req.ActualCost)
	if err := h.store.WorkOrders.Update(r.Context(), wo); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, toWorkOrderResponse(wo))
}
