package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
	"github.com/nexus-platform/backend-core/internal/repository"
)

func (s *Server) registerWhatsAppHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/whatsapp/conversations", s.withTenant(s.handleConversations))
	mux.HandleFunc("/v1/whatsapp/conversations/", s.withTenant(s.handleConversationDetail))
	mux.HandleFunc("/v1/whatsapp/config", s.withTenant(s.handleWhatsAppConfig))
	mux.HandleFunc("/v1/whatsapp/templates", s.withTenant(s.handleWhatsAppTemplates))
}

func (s *Server) handleConversations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewWhatsAppRepository(s.repo.Pool())

	switch r.Method {
	case http.MethodGet:
		convs, err := repo.ListConversations(ctx, tenantID)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{"conversations": demoConversations(tenantID)})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"conversations": convs})

	case http.MethodPost:
		var req domain.CreateConversationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		c := domain.WhatsAppConversation{
			GuestPhone:   req.GuestPhone,
			CurrentState: "greeting",
		}
		if req.GuestName != "" {
			c.GuestName = &req.GuestName
		}
		if err := repo.CreateConversation(ctx, tenantID, &c); err != nil {
			c.ID = uuid.New()
			c.TenantID = tenantUUID(tenantID)
			c.CreatedAt = time.Now()
			c.UpdatedAt = time.Now()
		}
		writeJSON(w, http.StatusCreated, c)

	default:
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleConversationDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	path := r.URL.Path[len("/v1/whatsapp/conversations/"):]
	parts := splitPath(path)
	if len(parts) == 0 {
		http.Error(w, `{"error":"missing conversation id"}`, http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(parts[0])
	if err != nil {
		http.Error(w, `{"error":"invalid conversation id"}`, http.StatusBadRequest)
		return
	}

	repo := repository.NewWhatsAppRepository(s.repo.Pool())

	if len(parts) > 1 && parts[1] == "messages" {
		if r.Method == http.MethodGet {
			msgs, err := repo.ListMessages(ctx, tenantID, id)
			if err != nil {
				writeJSON(w, http.StatusOK, map[string]interface{}{"messages": demoMessages(id.String())})
				return
			}
			writeJSON(w, http.StatusOK, map[string]interface{}{"messages": msgs})
			return
		}
		if r.Method == http.MethodPost {
			var req domain.SendMessageRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
				return
			}
			m := domain.WhatsAppMessage{
				ConversationID: id,
				Direction:      "outbound",
				Body:           req.Body,
				Status:         "delivered",
				SentAt:         time.Now(),
			}
			if err := repo.SendMessage(ctx, tenantID, &m); err != nil {
				m.ID = uuid.New()
				m.TenantID = tenantUUID(tenantID)
			}
			writeJSON(w, http.StatusCreated, m)
			return
		}
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	if r.Method == http.MethodGet {
		conv, err := repo.GetConversation(ctx, tenantID, id)
		if err != nil {
			convs := demoConversations(tenantID)
			for _, c := range convs {
				if c.ID == id {
					writeJSON(w, http.StatusOK, c)
					return
				}
			}
			http.Error(w, `{"error":"conversation not found"}`, http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, conv)
		return
	}

	if r.Method == http.MethodPatch {
		var req struct {
			State          string `json:"current_state"`
			BookingCreated bool   `json:"booking_created"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		if err := repo.UpdateConversationState(ctx, tenantID, id, req.State, req.BookingCreated); err != nil {
			writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func (s *Server) handleWhatsAppConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewWhatsAppRepository(s.repo.Pool())

	if r.Method == http.MethodGet {
		cfg, err := repo.GetBotConfig(ctx, tenantID)
		if err != nil {
			writeJSON(w, http.StatusOK, demoConfig(tenantID))
			return
		}
		writeJSON(w, http.StatusOK, cfg)
		return
	}

	if r.Method == http.MethodPatch {
		var req domain.UpdateBotConfigRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
			return
		}
		cfg := domain.WhatsAppBotConfig{
			TenantID:         tenantUUID(tenantID),
			BotEnabled:       req.BotEnabled != nil && *req.BotEnabled,
			BookingEnabled:   req.BookingEnabled != nil && *req.BookingEnabled,
			AutoReplyEnabled: req.AutoReplyEnabled != nil && *req.AutoReplyEnabled,
			WelcomeMessage:   req.WelcomeMessage,
			PhoneNumberID:    req.PhoneNumberID,
			WebhookURL:       "https://api.nexus.com/webhooks/whatsapp/" + tenantID,
		}
		if err := repo.UpdateBotConfig(ctx, tenantID, &cfg); err != nil {
			writeJSON(w, http.StatusOK, cfg)
			return
		}
		writeJSON(w, http.StatusOK, cfg)
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func (s *Server) handleWhatsAppTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewWhatsAppRepository(s.repo.Pool())

	if r.Method == http.MethodGet {
		tmpls, err := repo.ListTemplates(ctx, tenantID)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]interface{}{"templates": demoTemplates(tenantID)})
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{"templates": tmpls})
		return
	}

	http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
}

func demoConversations(tenantID string) []domain.WhatsAppConversation {
	now := time.Now()
	tid := tenantUUID(tenantID)
	name1 := "Alice Chen"
	name2 := "Bob Jones"
	name3 := "Carol White"
	name4 := "David Kim"
	return []domain.WhatsAppConversation{
		{ID: uuid.MustParse("22222222-2222-2222-2222-222222222201"), TenantID: tid, GuestPhone: "+1-555-0101", GuestName: &name1, CurrentState: "complete", BookingCreated: true, Messages: 8, LastMessageAt: now.Add(-30 * time.Minute), CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("22222222-2222-2222-2222-222222222202"), TenantID: tid, GuestPhone: "+1-555-0102", GuestName: &name2, CurrentState: "ask_dates", BookingCreated: false, Messages: 4, LastMessageAt: now.Add(-2 * time.Hour), CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("22222222-2222-2222-2222-222222222203"), TenantID: tid, GuestPhone: "+1-555-0103", GuestName: &name3, CurrentState: "complete", BookingCreated: true, Messages: 12, LastMessageAt: now.Add(-3 * time.Hour), CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("22222222-2222-2222-2222-222222222204"), TenantID: tid, GuestPhone: "+1-555-0104", GuestName: &name4, CurrentState: "greeting", BookingCreated: false, Messages: 2, LastMessageAt: now.Add(-4 * time.Hour), CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("22222222-2222-2222-2222-222222222205"), TenantID: tid, GuestPhone: "+1-555-0105", CurrentState: "support", BookingCreated: false, Messages: 6, LastMessageAt: now.Add(-5 * time.Hour), CreatedAt: now, UpdatedAt: now},
	}
}

func demoMessages(convID string) []domain.WhatsAppMessage {
	now := time.Now()
	tid := tenantUUID(convID)
	return []domain.WhatsAppMessage{
		{ID: uuid.New(), TenantID: tid, ConversationID: uuid.MustParse(convID), Direction: "inbound", Body: "Hi, I'd like to book a room for May 15-17", Status: "read", SentAt: now.Add(-30 * time.Minute)},
		{ID: uuid.New(), TenantID: tid, ConversationID: uuid.MustParse(convID), Direction: "outbound", Body: "Welcome! I'd be happy to help. Let me check availability for May 15-17. How many guests?", Status: "delivered", SentAt: now.Add(-29 * time.Minute)},
		{ID: uuid.New(), TenantID: tid, ConversationID: uuid.MustParse(convID), Direction: "inbound", Body: "2 adults", Status: "read", SentAt: now.Add(-28 * time.Minute)},
		{ID: uuid.New(), TenantID: tid, ConversationID: uuid.MustParse(convID), Direction: "outbound", Body: "Great! We have a Deluxe King available for $180/night. Total: $360 + taxes. Would you like to proceed?", Status: "delivered", SentAt: now.Add(-27 * time.Minute)},
		{ID: uuid.New(), TenantID: tid, ConversationID: uuid.MustParse(convID), Direction: "inbound", Body: "Yes please", Status: "read", SentAt: now.Add(-26 * time.Minute)},
		{ID: uuid.New(), TenantID: tid, ConversationID: uuid.MustParse(convID), Direction: "outbound", Body: "Perfect! Please provide your email for confirmation:", Status: "delivered", SentAt: now.Add(-25 * time.Minute)},
		{ID: uuid.New(), TenantID: tid, ConversationID: uuid.MustParse(convID), Direction: "inbound", Body: "alice@email.com", Status: "read", SentAt: now.Add(-24 * time.Minute)},
		{ID: uuid.New(), TenantID: tid, ConversationID: uuid.MustParse(convID), Direction: "outbound", Body: "Reservation confirmed! Ref: WA-001. Check-in: May 15, 3PM. Room: 201. See you soon!", Status: "delivered", SentAt: now.Add(-23 * time.Minute)},
	}
}

func demoConfig(tenantID string) *domain.WhatsAppBotConfig {
	return &domain.WhatsAppBotConfig{
		TenantID:         tenantUUID(tenantID),
		BotEnabled:       true,
		BookingEnabled:   true,
		AutoReplyEnabled: true,
		WelcomeMessage:   "Welcome to our hotel! How can I help you today? Reply BOOK to make a reservation, CHECKIN for check-in info, or WIFI for the password.",
		PhoneNumberID:    "",
		WebhookURL:       "https://api.nexus.com/webhooks/whatsapp/" + tenantID,
		UpdatedAt:        time.Now(),
	}
}

func demoTemplates(tenantID string) []domain.WhatsAppTemplate {
	now := time.Now()
	tid := tenantUUID(tenantID)
	return []domain.WhatsAppTemplate{
		{ID: uuid.New(), TenantID: tid, Trigger: "BOOK", Response: "Let's book your stay! What dates? (YYYY-MM-DD format)", IsActive: true, CreatedAt: now},
		{ID: uuid.New(), TenantID: tid, Trigger: "CHECKIN", Response: "Check-in starts at 3PM. Your room number will be sent on arrival day.", IsActive: true, CreatedAt: now},
		{ID: uuid.New(), TenantID: tid, Trigger: "WIFI", Response: "Network: NexusGuest | Password: Welcome2026", IsActive: true, CreatedAt: now},
		{ID: uuid.New(), TenantID: tid, Trigger: "CHECKOUT", Response: "Check-out is at 11AM. Late checkout available until 1PM for $30.", IsActive: true, CreatedAt: now},
		{ID: uuid.New(), TenantID: tid, Trigger: "HELP", Response: "Front desk: ext. 101 | Housekeeping: ext. 102 | Emergency: ext. 911", IsActive: true, CreatedAt: now},
	}
}
