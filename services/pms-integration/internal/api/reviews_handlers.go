package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ============== REVIEWS HANDLERS ==============

func (h *Handler) listReviews(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var channel *domain.ReviewChannel
	if c := r.URL.Query().Get("channel"); c != "" {
		ch := domain.ReviewChannel(c)
		channel = &ch
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}
	revs, err := h.store.Reviews.ListReviews(r.Context(), tenantID, channel, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, revs)
}

func (h *Handler) createReview(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var rev domain.GuestReview
	if err := json.NewDecoder(r.Body).Decode(&rev); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := h.store.Reviews.CreateReview(r.Context(), tenantID, &rev)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, res)
}

func (h *Handler) respondToReview(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	id := chi.URLParam(r, "reviewId")
	var body struct {
		Response   string `json:"response"`
		RespondedBy string `json:"responded_by"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	if err := h.store.Reviews.RespondToReview(r.Context(), tenantID, id, body.Response, body.RespondedBy); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"status": "responded"})
}

func (h *Handler) getRatingSnapshot(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "30d"
	}
	snap, err := h.store.Reviews.GetRatingSnapshot(r.Context(), tenantID, period)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, snap)
}

// ============== REVIEW REQUESTS ==============

func (h *Handler) listReviewRequests(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit == 0 {
		limit = 50
	}
	reqs, err := h.store.Reviews.ListReviewRequests(r.Context(), tenantID, status, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, reqs)
}

func (h *Handler) createReviewRequest(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var req domain.ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := h.store.Reviews.CreateReviewRequest(r.Context(), tenantID, &req)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, res)
}

// ============== PUBLIC REVIEW SUBMISSION (No Auth) ==============

func (h *Handler) publicSubmitReview(w http.ResponseWriter, r *http.Request) {
	tenantID := chi.URLParam(r, "tenantId")
	var rev domain.GuestReview
	if err := json.NewDecoder(r.Body).Decode(&rev); err != nil {
		respondError(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := h.store.Reviews.CreateReview(r.Context(), tenantID, &rev)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, res)
}
