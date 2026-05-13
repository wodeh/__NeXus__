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

	// Seed demo users for login
	hash := "$2a$10$wKvRKo1uJX0fC4p8K0OdHujT5jaMFeS9OqkZ/vEvX.s.yNXkgXGvK"
	users := []struct {
		email string
		name  string
		role  string
	}{
		{"ramiz@villa.test", "Ramiz Haddad", "manager"},
		{"owner1@villa.test", "Ahmad Khalil", "manager"},
		{"owner2@villa.test", "Sarah Nassar", "manager"},
		{"owner4@villa.test", "Layla Farhat", "manager"},
	}
	for _, u := range users {
		_, err := pool.Exec(ctx, `
			INSERT INTO users (tenant_id, email, password_hash, name, role, is_active)
			VALUES ($1, $2, $3, $4, $5, true)
			ON CONFLICT (email) DO UPDATE SET
				password_hash = EXCLUDED.password_hash,
				name = EXCLUDED.name,
				role = EXCLUDED.role,
				is_active = true
		`, tenantID, u.email, hash, u.name, u.role)
		if err != nil {
			logger.Warn("user insert failed", slog.String("email", u.email), slog.String("error", err.Error()))
		}
	}
	logger.Info("demo users seeded", slog.Int("count", len(users)))

	// ─── Seed Villa Properties ───
	villas := []struct {
		name, description, address, city, country string
		bedrooms, bathrooms, maxGuests            int
		pricePerNight, cleaningFee, securityDep   float64
		amenities                                 []string
	}{
		{"Villa Al-Mashta", "Luxury mountain villa with panoramic views", "Jabal Al-Mashta, Ramallah", "Ramallah", "Palestine", 3, 2, 6, 350, 50, 500, []string{"wifi", "parking", "pool", "garden", "bbq"}},
		{"Chalet Al-Balad", "Cozy chalet in the heart of Bethlehem", "Manger Street", "Bethlehem", "Palestine", 4, 3, 8, 450, 75, 750, []string{"wifi", "parking", "fireplace", "kitchen"}},
		{"Villa Al-Reef", "Modern villa with sea view terrace", "Gaza Beach Road", "Gaza", "Palestine", 5, 4, 10, 600, 100, 1000, []string{"wifi", "parking", "pool", "gym", "ac"}},
		{"Chalet Al-Jabal", "Rustic chalet surrounded by olive trees", "Nablus Mountain Road", "Nablus", "Palestine", 2, 1, 4, 250, 40, 300, []string{"wifi", "parking", "garden", "fireplace"}},
	}
	villaIDs := make([]string, len(villas))
	for i, v := range villas {
		var id string
		err := pool.QueryRow(ctx, `
			INSERT INTO villa_properties (tenant_id, name, description, address, city, country, bedrooms, bathrooms, max_guests, amenities, price_per_night, currency, cleaning_fee, security_deposit, is_active, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'USD', $12, $13, true, 'available')
			ON CONFLICT DO NOTHING
			RETURNING id
		`, tenantID, v.name, v.description, v.address, v.city, v.country, v.bedrooms, v.bathrooms, v.maxGuests, v.amenities, v.pricePerNight, v.cleaningFee, v.securityDep).Scan(&id)
		if err != nil {
			logger.Warn("villa insert failed", slog.String("name", v.name), slog.String("error", err.Error()))
			continue
		}
		villaIDs[i] = id
		logger.Info("villa seeded", slog.String("name", v.name), slog.String("id", id))
	}

	// ─── Seed Villa Bookings (10 bookings) ───
	bookings := []struct {
		villaIdx    int
		guest       string
		phone       string
		email       string
		guestCount  int
		checkIn     string
		checkOut    string
		nights      int
		totalAmount float64
		status      string
		source      string
		balanceDue  float64
		balancePaid bool
	}{
		{0, "Ahmad Khalil", "+970599123456", "ahmad@email.com", 4, "2026-05-10", "2026-05-15", 5, 1750, "reserved", "phone", 0, true},
		{0, "Sarah Nassar", "+970599987654", "sarah@email.com", 3, "2026-05-16", "2026-05-20", 4, 1400, "pending", "whatsapp", 1400, false},
		{1, "Layla Farhat", "+970599456789", "layla@email.com", 6, "2026-05-12", "2026-05-17", 5, 2250, "reserved", "phone", 750, false},
		{1, "Mohammed Ali", "+970599111222", "mohammed@email.com", 4, "2026-05-18", "2026-05-23", 5, 2250, "pending", "walkin", 2250, false},
		{2, "Fatima Hassan", "+970599333444", "fatima@email.com", 8, "2026-05-14", "2026-05-19", 5, 3000, "reserved", "website", 0, true},
		{2, "Omar Khalil", "+970599555666", "omar@email.com", 5, "2026-05-20", "2026-05-26", 6, 3600, "pending", "whatsapp", 3600, false},
		{3, "Nadia Ibrahim", "+970599777888", "nadia@email.com", 3, "2026-05-11", "2026-05-14", 3, 750, "completed", "phone", 0, true},
		{3, "Khaled Omar", "+970599999000", "khaled@email.com", 2, "2026-05-15", "2026-05-18", 3, 750, "reserved", "walkin", 0, true},
		{0, "Rania Suleiman", "+970599000111", "rania@email.com", 5, "2026-05-22", "2026-05-28", 6, 2100, "pending", "website", 2100, false},
		{1, "Youssef Nasser", "+970599222333", "youssef@email.com", 4, "2026-05-25", "2026-05-30", 5, 2250, "reserved", "whatsapp", 500, false},
	}
	for _, b := range bookings {
		villaID := villaIDs[b.villaIdx]
		if villaID == "" {
			continue
		}
		villaName := villas[b.villaIdx].name
		_, err := pool.Exec(ctx, `
			INSERT INTO villa_reservations (tenant_id, villa_id, villa_name, guest_name, guest_phone, guest_email, guest_count, check_in_date, check_out_date, nights, total_amount, currency, status, source, balance_due, balance_paid)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'USD', $12, $13, $14, $15)
			ON CONFLICT DO NOTHING
		`, tenantID, villaID, villaName, b.guest, b.phone, b.email, b.guestCount, b.checkIn, b.checkOut, b.nights, b.totalAmount, b.status, b.source, b.balanceDue, b.balancePaid)
		if err != nil {
			logger.Warn("booking insert failed", slog.String("guest", b.guest), slog.String("error", err.Error()))
		} else {
			logger.Info("booking seeded", slog.String("guest", b.guest), slog.String("villa", villaName))
		}
	}

	// ─── Seed Super Admin ───
	superAdminEmail := "admin@nexus.com"
	_, err = pool.Exec(ctx, `
		INSERT INTO users (tenant_id, email, password_hash, name, role, is_active)
		VALUES ($1, $2, $3, 'System Administrator', 'super_admin', true)
		ON CONFLICT (email) DO UPDATE SET
			password_hash = EXCLUDED.password_hash,
			name = EXCLUDED.name,
			role = EXCLUDED.role,
			is_active = true
	`, tenantID, superAdminEmail, hash)
	if err != nil {
		logger.Warn("super admin insert failed", slog.String("error", err.Error()))
	} else {
		logger.Info("super admin seeded", slog.String("email", superAdminEmail))
	}

	// ─── Seed 5 Hotels with Owners, Rooms, Reservations ───
	hotels := []struct {
		externalID, name, city, country string
		floors, roomsPerFloor           int
		ownerEmail, ownerName           string
	}{
		{"hotel-grand-plaza", "Grand Plaza Hotel", "New York", "USA", 5, 20, "owner.grand@nexus.com", "John Grand"},
		{"hotel-marina-bay", "Marina Bay Resort", "Dubai", "UAE", 6, 18, "owner.marina@nexus.com", "Fatima Al-Rashid"},
		{"hotel-alpine-lodge", "Alpine Lodge", "Zermatt", "Switzerland", 4, 15, "owner.alpine@nexus.com", "Hans Mueller"},
		{"hotel-ocean-view", "Ocean View Resort", "Maldives", "Maldives", 3, 25, "owner.ocean@nexus.com", "Aisha Hassan"},
		{"hotel-city-center", "City Center Inn", "London", "UK", 4, 22, "owner.city@nexus.com", "James Sterling"},
	}

	roomTypesHotel := []string{"Standard", "Deluxe", "Suite", "Premium", "Family"}
	bedTypesHotel := []string{"Queen", "King", "Twin", "Double Queen", "King Sofa"}
	statusesHotel := []string{"vacant_clean", "occupied", "vacant_dirty", "blocked", "maintenance"}
	sourcesHotel := []string{"direct", "ota", "walk_in", "agent"}

	for hi, h := range hotels {
		// Create hotel tenant
		var hotelTenantID string
		err := pool.QueryRow(ctx, `
			INSERT INTO tenants (external_id, name, region, tier, config)
			VALUES ($1, $2, 'us-east-1', 'enterprise', '{}')
			ON CONFLICT (external_id) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, h.externalID, h.name).Scan(&hotelTenantID)
		if err != nil {
			logger.Warn("hotel tenant insert failed", slog.String("name", h.name), slog.String("error", err.Error()))
			continue
		}
		logger.Info("hotel tenant created", slog.String("name", h.name), slog.String("id", hotelTenantID))

		// Create hotel owner user
		_, err = pool.Exec(ctx, `
			INSERT INTO users (tenant_id, email, password_hash, name, role, is_active)
			VALUES ($1, $2, $3, $4, 'manager', true)
			ON CONFLICT (email) DO UPDATE SET
				password_hash = EXCLUDED.password_hash,
				name = EXCLUDED.name,
				role = EXCLUDED.role,
				is_active = true
		`, hotelTenantID, h.ownerEmail, hash, h.ownerName)
		if err != nil {
			logger.Warn("hotel owner insert failed", slog.String("hotel", h.name), slog.String("error", err.Error()))
		}

		// Seed rooms for this hotel
		roomRepoHotel := repository.NewRoomRepository(pool)
		roomCountHotel := 0
		for floor := 1; floor <= h.floors; floor++ {
			for roomIdx := 1; roomIdx <= h.roomsPerFloor; roomIdx++ {
				roomNum := fmt.Sprintf("%d%02d", floor, roomIdx)
				roomType := roomTypesHotel[(roomIdx-1)%len(roomTypesHotel)]
				bedType := bedTypesHotel[(roomIdx-1)%len(bedTypesHotel)]
				status := statusesHotel[(floor+roomIdx+hi)%len(statusesHotel)]
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

				if err := roomRepoHotel.Create(ctx, hotelTenantID, &domain.Room{
					PropertyID: "main",
					Number:     roomNum,
					Floor:      fmt.Sprintf("%d", floor),
					Type:       roomType,
					BedType:    strPtr(bedType),
					Status:     status,
					RateNight:  rate,
				}); err != nil {
					logger.Warn("hotel room insert failed", slog.String("hotel", h.name), slog.String("room", roomNum), slog.String("error", err.Error()))
				} else {
					roomCountHotel++
				}
			}
		}
		logger.Info("hotel rooms seeded", slog.String("hotel", h.name), slog.Int("count", roomCountHotel))

		// Seed reservations for this hotel
		resRepoHotel := repository.NewReservationRepository(pool)
		resCountHotel := 0
		baseDateHotel := time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC)
		for i := 0; i < 15; i++ {
			floor := (i % h.floors) + 1
			roomIdx := (i % h.roomsPerFloor) + 1
			roomNum := fmt.Sprintf("%d%02d", floor, roomIdx)
			roomType := roomTypesHotel[i%len(roomTypesHotel)]
			checkIn := baseDateHotel.AddDate(0, 0, (i % 14) - 3).Format("2006-01-02")
			checkOut := baseDateHotel.AddDate(0, 0, (i % 14) + 2 + (i % 5)).Format("2006-01-02")
			adults := 1 + (i % 3)
			children := i % 2
			source := sourcesHotel[i%len(sourcesHotel)]
			vip := i%7 == 0

			req := domain.ReservationCreateRequest{
				GuestName:       fmt.Sprintf("%s %s", firstNames[i%len(firstNames)], lastNames[i%len(lastNames)]),
				Email:           fmt.Sprintf("%s.guest%d@email.com", h.externalID, i+1),
				Phone:           fmt.Sprintf("+%d-555-%04d", hi+1, 100+i),
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
			if _, err := resRepoHotel.Create(ctx, nil, hotelTenantID, &req); err != nil {
				logger.Warn("hotel reservation insert failed", slog.String("hotel", h.name), slog.String("guest", req.GuestName), slog.String("error", err.Error()))
			} else {
				resCountHotel++
			}
		}
		logger.Info("hotel reservations seeded", slog.String("hotel", h.name), slog.Int("count", resCountHotel))
	}

	logger.Info("demo data seed complete", slog.String("tenant_id", tenantID))
	fmt.Println("✅ Demo data seeded successfully")
}

func strPtr(s string) *string {
	return &s
}
