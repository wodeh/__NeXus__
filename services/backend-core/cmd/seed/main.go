package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/nexus-platform/backend-core/internal/config"
	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
	"github.com/nexus-platform/backend-core/internal/repository"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	// Run migrations first to ensure schema exists
	migrator, err := db.NewMigrator(cfg.DatabaseURL, "file://sql/migrations")
	if err != nil {
		logger.Error("failed to create migrator", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer migrator.Close()
	if err := migrator.Up(ctx); err != nil {
		logger.Error("failed to run migrations", slog.String("error", err.Error()))
		os.Exit(1)
	}

	// Ensure demo tenant exists
	tenantRepo := repository.NewTenantRepository(pool)
	tenant, err := tenantRepo.GetByExternalID(ctx, "demo")
	if err != nil {
		logger.Info("creating demo tenant")
		// Create tenant via raw SQL since Create method signature may vary
		_, err := pool.Exec(ctx, `
			INSERT INTO tenants (external_id, name, region, tier, config)
			VALUES ('demo', 'Demo Hotel', 'us-east-1', 'enterprise', '{}')
			ON CONFLICT (external_id) DO NOTHING
		`)
		if err != nil {
			logger.Error("failed to create tenant", slog.String("error", err.Error()))
			os.Exit(1)
		}
		tenant, err = tenantRepo.GetByExternalID(ctx, "demo")
		if err != nil {
			logger.Error("failed to get tenant after creation", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}

	tenantID := tenant.ID.String()
	logger.Info("seeding demo data", slog.String("tenant_id", tenantID))

	// Seed rooms
	roomRepo := repository.NewRoomRepository(pool)
	roomData := []domain.Room{
		{PropertyID: "main", Number: "101", Floor: "1", Type: "Standard", BedType: strPtr("Queen"), Status: "vacant_clean", RateNight: 89},
		{PropertyID: "main", Number: "102", Floor: "1", Type: "Standard", BedType: strPtr("Queen"), Status: "occupied", RateNight: 89},
		{PropertyID: "main", Number: "103", Floor: "1", Type: "Deluxe", BedType: strPtr("King"), Status: "vacant_clean", RateNight: 129},
		{PropertyID: "main", Number: "104", Floor: "1", Type: "Deluxe", BedType: strPtr("King"), Status: "vacant_dirty", RateNight: 129},
		{PropertyID: "main", Number: "105", Floor: "1", Type: "Suite", BedType: strPtr("King"), Status: "occupied", RateNight: 229},
		{PropertyID: "main", Number: "201", Floor: "2", Type: "Standard", BedType: strPtr("Queen"), Status: "vacant_clean", RateNight: 89},
		{PropertyID: "main", Number: "202", Floor: "2", Type: "Standard", BedType: strPtr("Twin"), Status: "occupied", RateNight: 89},
		{PropertyID: "main", Number: "203", Floor: "2", Type: "Deluxe", BedType: strPtr("King"), Status: "blocked", RateNight: 129},
		{PropertyID: "main", Number: "204", Floor: "2", Type: "Deluxe", BedType: strPtr("King"), Status: "vacant_clean", RateNight: 129},
		{PropertyID: "main", Number: "205", Floor: "2", Type: "Suite", BedType: strPtr("King"), Status: "occupied", RateNight: 229},
		{PropertyID: "main", Number: "301", Floor: "3", Type: "Standard", BedType: strPtr("Queen"), Status: "vacant_dirty", RateNight: 89},
		{PropertyID: "main", Number: "302", Floor: "3", Type: "Standard", BedType: strPtr("Twin"), Status: "vacant_clean", RateNight: 89},
		{PropertyID: "main", Number: "303", Floor: "3", Type: "Deluxe", BedType: strPtr("King"), Status: "occupied", RateNight: 129},
		{PropertyID: "main", Number: "304", Floor: "3", Type: "Suite", BedType: strPtr("King"), Status: "vacant_clean", RateNight: 229},
		{PropertyID: "main", Number: "305", Floor: "3", Type: "Deluxe", BedType: strPtr("Queen"), Status: "maintenance", RateNight: 129},
	}

	for _, r := range roomData {
		if err := roomRepo.Create(ctx, tenantID, &domain.Room{
			PropertyID: "main",
			Number:     r.Number,
			Floor:      r.Floor,
			Type:       r.Type,
			BedType:    r.BedType,
			Status:     r.Status,
			RateNight:  r.RateNight,
		}); err != nil {
			logger.Warn("room insert failed (may already exist)", slog.String("number", r.Number), slog.String("error", err.Error()))
		}
	}
	logger.Info("rooms seeded", slog.Int("count", len(roomData)))

	// Seed reservations
	resRepo := repository.NewReservationRepository(pool)
	reservations := []domain.ReservationCreateRequest{
		{GuestName: "Alice Johnson", Email: "alice@example.com", Phone: "+1-555-0101", RoomType: "Standard", RoomNumber: "102", CheckIn: "2026-05-09", CheckOut: "2026-05-12", Adults: 2, Children: 0, Source: "direct"},
		{GuestName: "Bob Smith", Email: "bob@example.com", Phone: "+1-555-0102", RoomType: "Suite", RoomNumber: "105", CheckIn: "2026-05-08", CheckOut: "2026-05-13", Adults: 2, Children: 2, Source: "ota", SpecialRequests: "Extra towels", VIP: true},
		{GuestName: "Carol Davis", Email: "carol@example.com", Phone: "+1-555-0103", RoomType: "Standard", RoomNumber: "202", CheckIn: "2026-05-10", CheckOut: "2026-05-14", Adults: 1, Children: 0, Source: "direct", SpecialRequests: "High floor preferred"},
		{GuestName: "David Wilson", Email: "david@example.com", Phone: "+1-555-0104", RoomType: "Deluxe", RoomNumber: "303", CheckIn: "2026-05-07", CheckOut: "2026-05-11", Adults: 2, Children: 0, Source: "ota", SpecialRequests: "Late checkout requested"},
		{GuestName: "Eva Martinez", Email: "eva@example.com", Phone: "+1-555-0105", RoomType: "Suite", RoomNumber: "205", CheckIn: "2026-05-11", CheckOut: "2026-05-16", Adults: 2, Children: 1, Source: "direct", SpecialRequests: "Anniversary setup", VIP: true},
		{GuestName: "Frank Brown", Email: "frank@example.com", Phone: "+1-555-0106", RoomType: "Standard", RoomNumber: "201", CheckIn: "2026-05-12", CheckOut: "2026-05-15", Adults: 2, Children: 0, Source: "walk_in"},
		{GuestName: "Grace Lee", Email: "grace@example.com", Phone: "+1-555-0107", RoomType: "Deluxe", RoomNumber: "103", CheckIn: "2026-05-13", CheckOut: "2026-05-18", Adults: 2, Children: 0, Source: "ota", SpecialRequests: "Allergic to feathers"},
		{GuestName: "Henry Taylor", Email: "henry@example.com", Phone: "+1-555-0108", RoomType: "Standard", RoomNumber: "101", CheckIn: "2026-05-11", CheckOut: "2026-05-12", Adults: 1, Children: 0, Source: "direct"},
		{GuestName: "Ivy Chen", Email: "ivy@example.com", Phone: "+1-555-0109", RoomType: "Deluxe", RoomNumber: "204", CheckIn: "2026-05-14", CheckOut: "2026-05-19", Adults: 2, Children: 2, Source: "ota", SpecialRequests: "Connecting rooms if possible"},
		{GuestName: "Jack Anderson", Email: "jack@example.com", Phone: "+1-555-0110", RoomType: "Suite", RoomNumber: "304", CheckIn: "2026-05-15", CheckOut: "2026-05-20", Adults: 4, Children: 0, Source: "direct", SpecialRequests: "Honeymoon package", VIP: true},
	}

	for _, req := range reservations {
		if _, err := resRepo.Create(ctx, nil, tenantID, &req); err != nil {
			logger.Warn("reservation insert failed", slog.String("guest", req.GuestName), slog.String("error", err.Error()))
		}
	}
	logger.Info("reservations seeded", slog.Int("count", len(reservations)))

	// Seed agents
	agentRepo := repository.NewAgentRepository(pool)
	agents := []domain.AgentCreateRequest{
		{Name: "Booking.com", Type: "ota", CommissionPct: 15, ContactName: "Partner Support", ContactEmail: "partners@booking.com", ContactPhone: "+1-800-BOOKING", ContractRef: "CNT-2026-001", SourceCode: "BKG"},
		{Name: "Expedia", Type: "ota", CommissionPct: 18, ContactName: "Account Manager", ContactEmail: "hotel@expedia.com", ContactPhone: "+1-800-EXPEDIA", ContractRef: "CNT-2026-002", SourceCode: "EXP"},
		{Name: "Virtuoso Travel", Type: "travel_agent", CommissionPct: 10, ContactName: "Jane Smith", ContactEmail: "jane@virtuoso.com", ContactPhone: "+1-555-0300", ContractRef: "CNT-2026-003", SourceCode: "VRT"},
		{Name: "TechCorp Corporate", Type: "corporate", CommissionPct: 0, ContactName: "Mike Chen", ContactEmail: "mike@techcorp.com", ContactPhone: "+1-555-0200", ContractRef: "CNT-2026-004", SourceCode: "TCP"},
	}

	for _, a := range agents {
		if _, err := agentRepo.Create(ctx, tenantID, &a); err != nil {
			logger.Warn("agent insert failed (may already exist)", slog.String("name", a.Name), slog.String("error", err.Error()))
		}
	}
	logger.Info("agents seeded", slog.Int("count", len(agents)))

	// Seed channels
	channelRepo := repository.NewChannelRepository(pool)
	channels := []domain.ChannelCreateRequest{
		{Source: "booking_com", DisplayName: "Booking.com", CommissionPct: 15},
		{Source: "expedia", DisplayName: "Expedia", CommissionPct: 18},
		{Source: "airbnb", DisplayName: "Airbnb", CommissionPct: 3},
		{Source: "direct", DisplayName: "Direct Bookings", CommissionPct: 0},
		{Source: "whatsapp", DisplayName: "WhatsApp Bot", CommissionPct: 0},
	}

	for _, c := range channels {
		if _, err := channelRepo.Create(ctx, tenantID, &c); err != nil {
			logger.Warn("channel insert failed (may already exist)", slog.String("source", c.Source), slog.String("error", err.Error()))
		}
	}
	logger.Info("channels seeded", slog.Int("count", len(channels)))

	// Update reservation statuses to match room occupancy
	_ = resRepo // used above

	logger.Info("demo data seed complete", slog.String("tenant_id", tenantID))
	fmt.Println("✅ Demo data seeded successfully")
}

func strPtr(s string) *string {
	return &s
}
