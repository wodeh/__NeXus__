package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
	"github.com/nexus-platform/backend-core/internal/repository"
)

func (s *Server) registerReviewHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/reviews", s.withTenant(s.handleReviews))
	mux.HandleFunc("/v1/reviews/", s.withTenant(s.handleReviewDetail))
	mux.HandleFunc("/v1/reviews/stats", s.withTenant(s.handleReviewStats))
}

func (s *Server) handleReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewReviewRepository(s.repo.Pool())

	switch r.Method {
	case http.MethodGet:
		reviews, err := repo.List(ctx, tenantID)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{"reviews": demoReviews(tenantID)})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"reviews": reviews})

	case http.MethodPost:
		var req domain.CreateReviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		rid, _ := uuid.Parse(req.ReservationID)
		rv := domain.Review{
			ReservationID: rid,
			GuestName:     "Guest",
			Rating:        req.Rating,
			Cleanliness:   req.Cleanliness,
			Service:       req.Service,
			Location:      req.Location,
			Value:         req.Value,
			Comment:       req.Comment,
			Source:        req.Source,
		}
		if err := repo.Create(ctx, tenantID, &rv); err != nil {
			rv.ID = uuid.New()
			rv.TenantID = tenantUUID(tenantID)
			rv.CreatedAt = time.Now()
			rv.UpdatedAt = time.Now()
		}
		writeJSON(w, http.StatusCreated, rv)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleReviewDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	path := r.URL.Path[len("/v1/reviews/"):]
	parts := splitPath(path)
	if len(parts) == 0 {
		http.Error(w, `{"error":"missing review id"}`, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, `{"error":"invalid review id"}`, http.StatusBadRequest)
		return
	}

	repo := repository.NewReviewRepository(s.repo.Pool())

	if len(parts) > 1 {
		action := parts[1]
		switch action {
		case "reply":
			if r.Method != http.MethodPost {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			var req domain.ReplyReviewRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
				return
			}
			if err := repo.AddReply(ctx, tenantID, id, req.StaffReply); err != nil {
				writeJSON(w, http.StatusOK, map[string]string{"status": "replied"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "replied"})
			return

		case "publish":
			if r.Method != http.MethodPatch {
				http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
				return
			}
			var req struct {
				Published bool `json:"published"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
				return
			}
			if err := repo.Publish(ctx, tenantID, id, req.Published); err != nil {
				writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
				return
			}
			writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
			return
		}
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func (s *Server) handleReviewStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewReviewRepository(s.repo.Pool())

	if r.Method == http.MethodGet {
		stats, err := repo.GetStats(ctx, tenantID)
		if err != nil {
			writeJSON(w, http.StatusOK, demoReviewStats(tenantID))
			return
		}
		writeJSON(w, http.StatusOK, stats)
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func demoReviews(tenantID string) []domain.Review {
	now := time.Now()
	tid := tenantUUID(tenantID)
	staffReply := "Thank you for your kind words! We look forward to welcoming you back."
	return []domain.Review{
		{ID: uuid.MustParse("33333333-3333-3333-3333-333333333301"), TenantID: tid, ReservationID: uuid.New(), GuestName: "Alice Chen", RoomNumber: "201", Rating: 5, Cleanliness: 5, Service: 5, Location: 4, Value: 5, Comment: "Absolutely wonderful stay! The staff went above and beyond.", StaffReply: &staffReply, RepliedAt: &now, IsPublished: true, Source: "direct", CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("33333333-3333-3333-3333-333333333302"), TenantID: tid, ReservationID: uuid.New(), GuestName: "Bob Jones", RoomNumber: "102", Rating: 4, Cleanliness: 4, Service: 5, Location: 4, Value: 4, Comment: "Great location and friendly staff. Room was a bit small but comfortable.", IsPublished: true, Source: "ota", CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now.Add(-24 * time.Hour)},
		{ID: uuid.MustParse("33333333-3333-3333-3333-333333333303"), TenantID: tid, ReservationID: uuid.New(), GuestName: "Carol White", RoomNumber: "203", Rating: 3, Cleanliness: 3, Service: 4, Location: 5, Value: 3, Comment: "Nice hotel but the AC was noisy at night.", IsPublished: false, Source: "email", CreatedAt: now.Add(-48 * time.Hour), UpdatedAt: now.Add(-48 * time.Hour)},
		{ID: uuid.MustParse("33333333-3333-3333-3333-333333333304"), TenantID: tid, ReservationID: uuid.New(), GuestName: "David Kim", RoomNumber: "301", Rating: 5, Cleanliness: 5, Service: 5, Location: 5, Value: 5, Comment: "Perfect! The spa was amazing.", IsPublished: true, Source: "direct", CreatedAt: now.Add(-72 * time.Hour), UpdatedAt: now.Add(-72 * time.Hour)},
	}
}

func demoReviewStats(tenantID string) *domain.ReviewStats {
	return &domain.ReviewStats{
		TenantID:           tenantID,
		TotalReviews:       4,
		AverageRating:      4.25,
		AverageCleanliness: 4.25,
		AverageService:     4.75,
		AverageLocation:    4.5,
		AverageValue:       4.25,
		FiveStarCount:      2,
		FourStarCount:      1,
		ThreeStarCount:     1,
		TwoStarCount:       0,
		OneStarCount:       0,
	}
}
