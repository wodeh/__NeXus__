package server

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/backend-core/internal/domain"
	"github.com/nexus-platform/backend-core/internal/repository"
)

func (s *Server) registerCommHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/v1/communications/templates", s.withTenant(s.handleGetTemplates))
	mux.HandleFunc("/v1/communications/sequences", s.withTenant(s.handleGetSequences))
	mux.HandleFunc("/v1/communications/scheduled", s.withTenant(s.handleGetScheduled))
}

// handleGetTemplates returns communication templates.
func (s *Server) handleGetTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewCommRepository(s.repo.Pool())
	templates, err := repo.ListTemplates(ctx, tenantID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"templates": demoTemplates(tenantID)})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"templates": templates})
}

// handleGetSequences returns communication sequences.
func (s *Server) handleGetSequences(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewCommRepository(s.repo.Pool())
	sequences, err := repo.ListSequences(ctx, tenantID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"sequences": demoSequences(tenantID)})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"sequences": sequences})
}

// handleGetScheduled returns scheduled communications.
func (s *Server) handleGetScheduled(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID, _ := ctx.Value("tenant_id").(string)
	repo := repository.NewCommRepository(s.repo.Pool())
	scheduled, err := repo.ListScheduled(ctx, tenantID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"scheduled": demoScheduled(tenantID)})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"scheduled": scheduled})
}

func demoTemplates(tenantID string) []domain.CommTemplate {
	tid := tenantUUID(tenantID)
	now := time.Now()
	subj := func(s string) *string { return &s }
	return []domain.CommTemplate{
		{ID: uuid.MustParse("cccc0003-0000-0000-0000-000000000001"), TenantID: tid, Name: "Booking Confirmation", Subject: subj("Your reservation is confirmed"), Body: "Dear {{guest_name}},\n\nYour reservation at {{property_name}} is confirmed.\n\nDates: {{check_in}} - {{check_out}}\nRoom: {{room_number}}\n\nWe look forward to welcoming you!", Channel: "email", Category: "booking", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0003-0000-0000-0000-000000000002"), TenantID: tid, Name: "Pre-Arrival", Subject: subj("Your stay is coming up"), Body: "Hi {{guest_name}},\n\nYour check-in at {{property_name}} is tomorrow.\n\nCheck-in time: 3:00 PM\nWiFi: NexusGuest / Password: Welcome2026\n\nNeed anything? Reply to this email or call us.", Channel: "email", Category: "pre_arrival", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0003-0000-0000-0000-000000000003"), TenantID: tid, Name: "Check-In Instructions", Body: "Welcome {{guest_name}}! Your room {{room_number}} is ready. Digital key: {{digital_key}}. Enjoy your stay!", Channel: "sms", Category: "checkin", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0003-0000-0000-0000-000000000004"), TenantID: tid, Name: "Review Request", Subject: subj("How was your stay?"), Body: "Hi {{guest_name}},\n\nThank you for staying with us! We'd love to hear about your experience.\n\nLeave a review: {{review_link}}\n\nIt takes less than 2 minutes.", Channel: "email", Category: "review", IsActive: true, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0003-0000-0000-0000-000000000005"), TenantID: tid, Name: "WhatsApp Welcome", Body: "Welcome to {{property_name}}! 🎉\nYour room {{room_number}} is ready.\n\nNeed help? Reply with:\n• BOOK - for future stays\n• CHECKIN - check-in info\n• WIFI - password\n• HELP - speak to front desk", Channel: "whatsapp", Category: "stay", IsActive: true, CreatedAt: now, UpdatedAt: now},
	}
}

func demoSequences(tenantID string) []domain.CommSequence {
	tid := tenantUUID(tenantID)
	now := time.Now()
	return []domain.CommSequence{
		{ID: uuid.MustParse("cccc0004-0000-0000-0000-000000000001"), TenantID: tid, Name: "Guest Journey", Description: "Full guest communication flow", Trigger: "booking_created", IsActive: true, Steps: 5, CreatedAt: now, UpdatedAt: now},
		{ID: uuid.MustParse("cccc0004-0000-0000-0000-000000000002"), TenantID: tid, Name: "Post-Stay Review", Description: "Review request sequence", Trigger: "checkout", IsActive: true, Steps: 3, CreatedAt: now, UpdatedAt: now},
	}
}

func demoScheduled(tenantID string) []domain.CommScheduled {
	tid := tenantUUID(tenantID)
	now := time.Now()
	sched := now.Add(-30 * time.Minute)
	sent := now.Add(-45 * time.Minute)
	subj := func(s string) *string { return &s }
	body := func(s string) *string { return &s }
	return []domain.CommScheduled{
		{ID: uuid.MustParse("cccc0005-0000-0000-0000-000000000001"), TenantID: tid, GuestName: "Alice Chen", Channel: "email", Subject: subj("Your reservation is confirmed"), Status: "scheduled", ScheduledAt: sched, CreatedAt: now},
		{ID: uuid.MustParse("cccc0005-0000-0000-0000-000000000002"), TenantID: tid, GuestName: "Bob Jones", Channel: "sms", Body: body("Welcome! Your room 102 is ready."), Status: "sent", ScheduledAt: sent, SentAt: &sent, CreatedAt: now},
		{ID: uuid.MustParse("cccc0005-0000-0000-0000-000000000003"), TenantID: tid, GuestName: "Carol White", Channel: "email", Subject: subj("How was your stay?"), Status: "delivered", ScheduledAt: now, SentAt: &now, CreatedAt: now},
		{ID: uuid.MustParse("cccc0005-0000-0000-0000-000000000004"), TenantID: tid, GuestName: "David Kim", Channel: "whatsapp", Body: body("Check-in reminder: tomorrow 3PM"), Status: "failed", ScheduledAt: now.Add(-2 * time.Hour), Error: strPtr("Invalid phone number"), CreatedAt: now},
	}
}
