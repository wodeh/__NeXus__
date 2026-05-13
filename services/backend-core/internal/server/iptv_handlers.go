package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
	"github.com/nexus-platform/backend-core/internal/repository"
)

func (s *Server) registerIPTVHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/iptv/channels", s.withTenant(s.handleGetChannels))
	mux.HandleFunc("/v1/iptv/content", s.withTenant(s.handleGetContent))
	mux.HandleFunc("/v1/iptv/rooms", s.withTenant(s.handleGetIPTVRooms))
	mux.HandleFunc("/v1/iptv/welcome/", s.withTenant(s.handleIPTVWelcome))
}

// handleGetChannels returns TV channels.
func (s *Server) handleGetChannels(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewIPTVRepository(s.repo.Pool())
	channels, err := repo.ListChannels(ctx, tenantID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"channels": demoChannels(tenantID)})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"channels": channels})
}

// handleGetContent returns on-demand content.
func (s *Server) handleGetContent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewIPTVRepository(s.repo.Pool())
	content, err := repo.ListContent(ctx, tenantID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"content": demoContent(tenantID)})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"content": content})
}

// handleGetIPTVRooms returns IPTV room status.
func (s *Server) handleGetIPTVRooms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewIPTVRepository(s.repo.Pool())
	statuses, err := repo.ListRoomStatus(ctx, tenantID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"rooms": []domain.IPTVRoomStatus{}})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"rooms": statuses})
}

// handleIPTVWelcome returns welcome screen data for a villa's smart TV.
func (s *Server) handleIPTVWelcome(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	villaID := extractID(r.URL.Path, "/v1/iptv/welcome/")
	if villaID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing villa id"})
		return
	}

	ctx := r.Context()
	today := time.Now().Format("2006-01-02")

	// Handle missing database gracefully
	if s.repo == nil || s.repo.Villa == nil {
		writeJSON(w, http.StatusOK, domain.IPTVWelcomeScreen{
			VillaID:        villaID,
			VillaName:      "Green Villa",
			GuestName:      "Guest",
			WelcomeMessage: "Welcome to your villa",
			CheckInDate:    today,
			CheckOutDate:   today,
			Nights:         0,
			HasReservation: false,
		})
		return
	}

	// Try to get active reservation for this villa today
	res, err := s.repo.Villa.GetVillaReservationByDate(ctx, villaID, today)
	if err != nil || res == nil {
		// No active reservation — return villa-only welcome
		villa, vErr := s.repo.Villa.GetVillaByID(ctx, villaID)
		if vErr != nil {
			writeJSON(w, http.StatusOK, domain.IPTVWelcomeScreen{
				VillaID:        villaID,
				VillaName:      "Villa",
				WelcomeMessage: "Welcome to your villa",
				HasReservation: false,
			})
			return
		}
		writeJSON(w, http.StatusOK, domain.IPTVWelcomeScreen{
			VillaID:        villaID,
			VillaName:      villa.Name,
			WelcomeMessage: "Welcome to " + villa.Name,
			HasReservation: false,
		})
		return
	}

	// Build welcome screen with guest info
	welcome := domain.IPTVWelcomeScreen{
		VillaID:        villaID,
		VillaName:      res.VillaName,
		GuestName:      res.GuestName,
		CheckInDate:    res.CheckInDate,
		CheckOutDate:   res.CheckOutDate,
		Nights:         res.Nights,
		WelcomeMessage: "Welcome, " + res.GuestName + "! Enjoy your stay at " + res.VillaName,
		HasReservation: true,
	}
	writeJSON(w, http.StatusOK, welcome)
}

func demoChannels(tenantID string) []domain.IPTVChannel {
	tid := tenantUUID(tenantID)
	now := time.Now()
	return []domain.IPTVChannel{
		{ID: uuid.MustParse("cccc0001-0000-0000-0000-000000000001"), TenantID: tid, Name: "CNN International", Number: 1, StreamURL: "https://cnn-intl/stream", Category: "news", Language: "en", IsActive: true, IsPremium: false, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0001-0000-0000-0000-000000000002"), TenantID: tid, Name: "BBC World", Number: 2, StreamURL: "https://bbc-world/stream", Category: "news", Language: "en", IsActive: true, IsPremium: false, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0001-0000-0000-0000-000000000003"), TenantID: tid, Name: "Al Jazeera", Number: 3, StreamURL: "https://aljazeera/stream", Category: "news", Language: "en", IsActive: true, IsPremium: false, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0001-0000-0000-0000-000000000010"), TenantID: tid, Name: "ESPN", Number: 10, StreamURL: "https://espn/stream", Category: "sports", Language: "en", IsActive: true, IsPremium: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0001-0000-0000-0000-000000000011"), TenantID: tid, Name: "BeIN Sports", Number: 11, StreamURL: "https://bein/stream", Category: "sports", Language: "en", IsActive: true, IsPremium: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0001-0000-0000-0000-000000000020"), TenantID: tid, Name: "HBO", Number: 20, StreamURL: "https://hbo/stream", Category: "movies", Language: "en", IsActive: true, IsPremium: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0001-0000-0000-0000-000000000021"), TenantID: tid, Name: "Netflix Channel", Number: 21, StreamURL: "https://netflix-channel/stream", Category: "movies", Language: "en", IsActive: true, IsPremium: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0001-0000-0000-0000-000000000050"), TenantID: tid, Name: "Kids TV", Number: 50, StreamURL: "https://kids-tv/stream", Category: "kids", Language: "en", IsActive: true, IsPremium: false, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0001-0000-0000-0000-000000000051"), TenantID: tid, Name: "Cartoon Network", Number: 51, StreamURL: "https://cn/stream", Category: "kids", Language: "en", IsActive: true, IsPremium: false, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0001-0000-0000-0000-000000000080"), TenantID: tid, Name: "Music FM", Number: 80, StreamURL: "https://music-fm/stream", Category: "music", Language: "en", IsActive: true, IsPremium: false, CreatedAt: now, UpdatedAt: now},
	}
}

func demoContent(tenantID string) []domain.IPTVContent {
	tid := tenantUUID(tenantID)
	now := time.Now()
	return []domain.IPTVContent{
		{ID: uuid.MustParse("cccc0002-0000-0000-0000-000000000001"), TenantID: tid, Title: "Welcome to Grand Plaza", Type: "info", Description: "Hotel overview, amenities guide, and local attractions", Category: "welcome", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0002-0000-0000-0000-000000000002"), TenantID: tid, Title: "Room Service Menu", Type: "info", Description: "Full dining menu with ordering instructions", Category: "dining", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0002-0000-0000-0000-000000000003"), TenantID: tid, Title: "Spa & Wellness Guide", Type: "info", Description: "Spa treatments, gym hours, and booking info", Category: "wellness", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0002-0000-0000-0000-000000000004"), TenantID: tid, Title: "The Dark Knight", Type: "movie", Description: "Batman faces the Joker in Gotham City", Duration: intPtr(152), Category: "action", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0002-0000-0000-0000-000000000005"), TenantID: tid, Title: "Inception", Type: "movie", Description: "A thief who steals corporate secrets through dream-sharing technology", Duration: intPtr(148), Category: "sci-fi", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0002-0000-0000-0000-000000000006"), TenantID: tid, Title: "Breaking Bad S1", Type: "series", Description: "A high school chemistry teacher turned methamphetamine manufacturer", Category: "drama", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0002-0000-0000-0000-000000000007"), TenantID: tid, Title: "Local Attractions", Type: "info", Description: "Top 10 places to visit within 5km of the hotel", Category: "tourism", IsActive: true, CreatedAt: now, UpdatedAt: now},
	}
}
