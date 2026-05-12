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

	// ─── 10 floors × 25 rooms = 250 rooms ───
	roomRepo := repository.NewRoomRepository(pool)
	roomTypes := []string{"Standard", "Deluxe", "Suite", "Premium", "Family"}
	bedTypes := []string{"Queen", "King", "Twin", "Double Queen", "King Sofa"}
	statuses := []string{"vacant_clean", "occupied", "vacant_dirty", "blocked", "maintenance"}

	roomCount := 0
	for floor := 1; floor <= 10; floor++ {
		for roomIdx := 1; roomIdx <= 25; roomIdx++ {
			roomNum := fmt.Sprintf("%d%02d", floor, roomIdx)
			roomType := roomTypes[(roomIdx-1)%len(roomTypes)]
			bedType := bedTypes[(roomIdx-1)%len(bedTypes)]
			status := statuses[(floor+roomIdx)%len(statuses)]
			rate := 89
			switch roomType {
			case "Deluxe":
				rate = 129
			case "Suite":
				rate = 229
			case "Premium":
				rate = 189
			case "Family":
				rate = 159
			}

			if err := roomRepo.Create(ctx, tenantID, &domain.Room{
				PropertyID: "main",
				Number:     roomNum,
				Floor:      fmt.Sprintf("%d", floor),
				Type:       roomType,
				BedType:    strPtr(bedType),
				Status:     status,
				RateNight:  rate,
			}); err != nil {
				logger.Warn("room insert failed", slog.String("number", roomNum), slog.String("error", err.Error()))
			} else {
				roomCount++
			}
		}
	}
	logger.Info("rooms seeded", slog.Int("count", roomCount))

	// ─── 50 reservations ───
	resRepo := repository.NewReservationRepository(pool)
	firstNames := []string{"Alice", "Bob", "Carol", "David", "Eva", "Frank", "Grace", "Henry", "Ivy", "Jack",
		"Kate", "Liam", "Mia", "Noah", "Olivia", "Paul", "Quinn", "Rachel", "Sam", "Tina",
		"Uma", "Victor", "Wendy", "Xavier", "Yara", "Zane", "Anna", "Ben", "Clara", "Derek",
		"Elena", "Finn", "Gina", "Hugo", "Isla", "James", "Kara", "Leo", "Maya", "Nate",
		"Opal", "Pete", "Rosa", "Sean", "Tara", "Ursula", "Vince", "Will", "Xena", "Yvonne"}
	lastNames := []string{"Johnson", "Smith", "Davis", "Wilson", "Martinez", "Brown", "Lee", "Taylor", "Chen", "Anderson",
		"White", "Harris", "Clark", "Lewis", "Walker", "Hall", "Allen", "Young", "King", "Wright",
		"Scott", "Green", "Baker", "Adams", "Nelson", "Hill", "Ramirez", "Campbell", "Mitchell", "Roberts",
		"Carter", "Phillips", "Evans", "Turner", "Torres", "Parker", "Collins", "Edwards", "Stewart", "Flores",
		"Morris", "Nguyen", "Murphy", "Rivera", "Cook", "Rogers", "Morgan", "Peterson", "Cooper", "Reed"}
	sources := []string{"direct", "ota", "walk_in", "agent"}
	roomTypesRes := []string{"Standard", "Deluxe", "Suite", "Premium", "Family"}

	resCount := 0
	baseDate := time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 50; i++ {
		floor := (i%10) + 1
		roomIdx := (i % 25) + 1
		roomNum := fmt.Sprintf("%d%02d", floor, roomIdx)
		roomType := roomTypesRes[i%len(roomTypesRes)]
		checkIn := baseDate.AddDate(0, 0, (i%14)-3).Format("2006-01-02")
		checkOut := baseDate.AddDate(0, 0, (i%14)+2+(i%5)).Format("2006-01-02")
		adults := 1 + (i % 3)
		children := i % 2
		source := sources[i%len(sources)]
		vip := i%7 == 0

		req := domain.ReservationCreateRequest{
			GuestName:       fmt.Sprintf("%s %s", firstNames[i%len(firstNames)], lastNames[i%len(lastNames)]),
			Email:           fmt.Sprintf("guest%d@example.com", i+1),
			Phone:           fmt.Sprintf("+1-555-%04d", 100+i),
			RoomType:        roomType,
			RoomNumber:      roomNum,
			CheckIn:         checkIn,
			CheckOut:        checkOut,
			Adults:          adults,
			Children:        children,
			Source:          source,
			SpecialRequests: "",
			VIP:             vip,
		}
		if _, err := resRepo.Create(ctx, nil, tenantID, &req); err != nil {
			logger.Warn("reservation insert failed", slog.String("guest", req.GuestName), slog.String("error", err.Error()))
		} else {
			resCount++
		}
	}
	logger.Info("reservations seeded", slog.Int("count", resCount))

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
			logger.Warn("agent insert failed", slog.String("name", a.Name))
		}
	}
	logger.Info("agents seeded", slog.Int("count", len(agents)))

	// Seed channels
	channelRepo := repository.NewChannelRepository(pool)
	channels := []domain.Channel{
		{Source: "booking_com", DisplayName: "Booking.com", CommissionPct: 15},
		{Source: "expedia", DisplayName: "Expedia", CommissionPct: 18},
		{Source: "airbnb", DisplayName: "Airbnb", CommissionPct: 3},
		{Source: "direct", DisplayName: "Direct Bookings", CommissionPct: 0},
		{Source: "whatsapp", DisplayName: "WhatsApp Bot", CommissionPct: 0},
	}
	for _, c := range channels {
		if err := channelRepo.CreateChannel(ctx, &c); err != nil {
			logger.Warn("channel insert failed", slog.String("source", c.Source))
		}
	}
	logger.Info("channels seeded", slog.Int("count", len(channels)))

	logger.Info("demo data seed complete", slog.String("tenant_id", tenantID))
	fmt.Println("✅ Demo data seeded successfully")
}

func strPtr(s string) *string {
	return &s
}
