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

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
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

	logger.Info("connected to database, starting comprehensive seed...")

	// ─── Data pools ───
	firstNames := []string{"Alice", "Bob", "Carol", "David", "Eva", "Frank", "Grace", "Henry", "Ivy", "Jack",
		"Kate", "Liam", "Mia", "Noah", "Olivia", "Paul", "Quinn", "Rachel", "Sam", "Tina",
		"Uma", "Victor", "Wendy", "Xavier", "Yara", "Zane", "Anna", "Ben", "Clara", "Derek",
		"Elena", "Finn", "Gina", "Hugo", "Isla", "James", "Kara", "Leo", "Maya", "Nate",
		"Opal", "Pete", "Rosa", "Sean", "Tara", "Ursula", "Vince", "Will", "Xena", "Yvonne",
		"Ahmad", "Fatima", "Hassan", "Layla", "Omar", "Nadia", "Khaled", "Rania", "Youssef", "Zainab",
		"Sofia", "Marco", "Ingrid", "Raj", "Priya", "Chen", "Wei", "Hiroshi", "Sakura", "Juan"}
	lastNames := []string{"Johnson", "Smith", "Davis", "Wilson", "Martinez", "Brown", "Lee", "Taylor", "Chen", "Anderson",
		"White", "Harris", "Clark", "Lewis", "Walker", "Hall", "Allen", "Young", "King", "Wright",
		"Scott", "Green", "Baker", "Adams", "Nelson", "Hill", "Ramirez", "Campbell", "Mitchell", "Roberts",
		"Carter", "Phillips", "Evans", "Turner", "Torres", "Parker", "Collins", "Edwards", "Stewart", "Flores",
		"Morris", "Nguyen", "Murphy", "Rivera", "Cook", "Rogers", "Morgan", "Peterson", "Cooper", "Reed",
		"Al-Rashid", "Hassan", "Khalil", "Farhat", "Nasser", "Suleiman", "Ibrahim", "Omar", "Haddad", "Nassar",
		"Mueller", "Schmidt", "Ferrari", "Patel", "Sharma", "Liu", "Wang", "Tanaka", "Suzuki", "Garcia"}

	roomTypes := []string{"Standard", "Deluxe", "Suite", "Premium", "Family"}
	bedTypes := []string{"Queen", "King", "Twin", "Double Queen", "King Sofa"}
	roomStatuses := []string{"vacant_clean", "occupied", "vacant_dirty", "blocked", "maintenance"}
	resStatuses := []string{"confirmed", "checked_in", "checked_out", "cancelled", "no_show"}
	sources := []string{"direct", "ota", "walk_in", "agent"}

	// ─── 10 Hotels (Tenants) ───
	hotels := []struct {
		externalID, name, city, country, region string
		resCount                               int
	}{
		{"hotel-01", "Grand Plaza Hotel", "New York", "USA", "us-east-1", 40},
		{"hotel-02", "Marina Bay Resort", "Dubai", "UAE", "me-east-1", 150},
		{"hotel-03", "Alpine Lodge", "Zermatt", "Switzerland", "eu-west-1", 65},
		{"hotel-04", "Ocean View Resort", "Maldives", "Maldives", "ap-south-1", 130},
		{"hotel-05", "City Center Inn", "London", "UK", "eu-west-1", 20},
		{"hotel-06", "Tokyo Gardens", "Tokyo", "Japan", "ap-northeast-1", 20},
		{"hotel-07", "Rio Palace", "Rio de Janeiro", "Brazil", "sa-east-1", 20},
		{"hotel-08", "Sydney Harbour", "Sydney", "Australia", "ap-southeast-1", 20},
		{"hotel-09", "Cape Grace", "Cape Town", "South Africa", "af-south-1", 20},
		{"hotel-10", "Ritz Carlton", "Paris", "France", "eu-central-1", 20},
	}

	totalRooms := 0
	totalReservations := 0

	for hi, h := range hotels {
		// Create tenant
		var tenantID string
		err := pool.QueryRow(ctx, `
			INSERT INTO tenants (id, external_id, name, region, tier, config)
			VALUES (gen_random_uuid(), $1, $2, $3, 'enterprise', '{}')
			ON CONFLICT (external_id) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`, h.externalID, h.name, h.region).Scan(&tenantID)
		if err != nil {
			logger.Error("hotel tenant insert failed", slog.String("name", h.name), slog.String("error", err.Error()))
			continue
		}
		logger.Info("hotel created", slog.Int("hotel", hi+1), slog.String("name", h.name), slog.String("id", tenantID))

		// ─── Seed Users for hotel ───
		hash := "$2a$10$wKvRKo1uJX0fC4p8K0OdHujT5jaMFeS9OqkZ/vEvX.s.yNXkgXGvK"
		hotelUsers := []struct{ email, name, role string }{
			{fmt.Sprintf("admin@%s.com", h.externalID), fmt.Sprintf("%s Admin", h.name), "admin"},
			{fmt.Sprintf("manager@%s.com", h.externalID), fmt.Sprintf("%s Manager", h.name), "manager"},
			{fmt.Sprintf("frontdesk@%s.com", h.name), fmt.Sprintf("%s Front Desk", h.name), "staff"},
			{fmt.Sprintf("hk@%s.com", h.externalID), fmt.Sprintf("%s Housekeeping", h.name), "cleaner"},
			{fmt.Sprintf("owner@%s.com", h.externalID), fmt.Sprintf("%s Owner", h.name), "manager"},
		}
		for _, u := range hotelUsers {
			_, err := pool.Exec(ctx, `
				INSERT INTO users (id, tenant_id, email, password_hash, name, role, is_active)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, true)
				ON CONFLICT (email) DO UPDATE SET
					name = EXCLUDED.name,
					role = EXCLUDED.role,
					is_active = true
			`, tenantID, u.email, hash, u.name, u.role)
			if err != nil {
				logger.Warn("user insert failed", slog.String("email", u.email), slog.String("error", err.Error()))
			}
		}

		// ─── Housekeeping Staff ───
		for si := 0; si < 5; si++ {
			name := fmt.Sprintf("%s %s", firstNames[(hi*5+si)%len(firstNames)], lastNames[(hi*5+si)%len(lastNames)])
			_, err := pool.Exec(ctx, `
				INSERT INTO housekeeping_staff (id, tenant_id, name, role, active_shift, max_rooms_per_day, phone, active)
				VALUES (gen_random_uuid(), $1, $2, 'cleaner', 'day', 15, $3, true)
				ON CONFLICT DO NOTHING
			`, tenantID, name, fmt.Sprintf("+%d-555-%04d", hi+1, 100+si))
			if err != nil {
				logger.Warn("hk staff insert failed", slog.String("name", name), slog.String("error", err.Error()))
			}
		}

		// ─── 10 Floors × 25 Rooms = 250 rooms ───
		roomRepo := repository.NewRoomRepository(pool)
		hotelRooms := 0
		for floor := 1; floor <= 10; floor++ {
			for roomIdx := 1; roomIdx <= 25; roomIdx++ {
				roomNum := fmt.Sprintf("%d%02d", floor, roomIdx)
				roomType := roomTypes[(roomIdx-1)%len(roomTypes)]
				bedType := bedTypes[(roomIdx-1)%len(bedTypes)]
				status := roomStatuses[(floor+roomIdx+hi)%len(roomStatuses)]
				rate := 89 + (floor * 10) + (roomIdx % 5 * 5)
				switch roomType {
				case "Deluxe":
					rate += 40
				case "Suite":
					rate += 140
				case "Premium":
					rate += 100
				case "Family":
					rate += 70
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
					logger.Warn("room insert failed", slog.String("hotel", h.name), slog.String("room", roomNum), slog.String("error", err.Error()))
				} else {
					hotelRooms++
				}
			}
		}
		totalRooms += hotelRooms
		logger.Info("rooms seeded", slog.String("hotel", h.name), slog.Int("count", hotelRooms))

		// ─── Reservations via direct SQL (bypass repository column mismatch) ───
		hotelReservations := 0
		baseDate := time.Date(2026, 5, 9, 0, 0, 0, 0, time.UTC)

		for i := 0; i < h.resCount; i++ {
			floor := (i % 10) + 1
			roomIdx := (i % 25) + 1
			roomNum := fmt.Sprintf("%d%02d", floor, roomIdx)
			roomType := roomTypes[i%len(roomTypes)]

			checkIn := baseDate.AddDate(0, 0, (i%20)-5).Format("2006-01-02")
			stayLength := 1 + (i % 7)
			checkOut := baseDate.AddDate(0, 0, (i%20)-5+stayLength).Format("2006-01-02")

			adults := 1 + (i % 3)
			children := i % 2
			source := sources[i%len(sources)]
			vip := i%10 == 0
			status := resStatuses[i%len(resStatuses)]
			if status == "no_show" || status == "cancelled" {
				status = "confirmed"
			}
			if i%3 == 0 { status = "checked_in" }
			if i%5 == 0 { status = "checked_out" }

			total := stayLength * (89 + (floor * 10))
			balance := 0
			if i%4 == 0 { balance = total / 2 }

			fn := firstNames[(hi*100+i)%len(firstNames)]
			ln := lastNames[(hi*100+i)%len(lastNames)]
			guestName := fmt.Sprintf("%s %s", fn, ln)

			_, err := pool.Exec(ctx, `
				INSERT INTO reservations (id, tenant_id, property_id, guest_name, email, phone, room_type, room_number, check_in, check_out, adults, children, status, source, total, balance, vip, pre_arrival_ready, deposit_paid, special_requests_acknowledged)
				VALUES (gen_random_uuid(), $1, 'main', $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, false, 0, false)
				ON CONFLICT DO NOTHING
			`, tenantID, guestName,
				fmt.Sprintf("guest.%d@%s.com", i+1, h.externalID),
				fmt.Sprintf("+%d-555-%04d", hi+1, 1000+i),
				roomType, roomNum, checkIn, checkOut, adults, children, status, source, total, balance, vip)
			if err != nil {
				logger.Warn("reservation insert failed", slog.String("hotel", h.name), slog.String("guest", guestName), slog.String("error", err.Error()))
			} else {
				hotelReservations++
			}
		}
		totalReservations += hotelReservations
		logger.Info("reservations seeded", slog.String("hotel", h.name), slog.Int("count", hotelReservations))
	}

	// ─── Create or get demo tenant ───
	var demoTenantID string
	_ = pool.QueryRow(ctx, `SELECT id FROM tenants WHERE external_id = 'demo'`).Scan(&demoTenantID)
	if demoTenantID == "" {
		err = pool.QueryRow(ctx, `
			INSERT INTO tenants (id, external_id, name, region, tier, config)
			VALUES ('a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 'demo', 'Demo Hotel', 'us-east-1', 'enterprise', '{"timezone":"America/New_York","locale":"en","currency":"USD","modules":["reservations","housekeeping","iptv","locks","communications","revenue","reviews","audit","whatsapp","channel_manager","booking_engine"]}')
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name
			RETURNING id
		`).Scan(&demoTenantID)
		if err != nil {
			logger.Error("demo tenant failed", slog.String("error", err.Error()))
		} else {
			logger.Info("demo tenant created", slog.String("id", demoTenantID))
		}
	} else {
		logger.Info("demo tenant exists", slog.String("id", demoTenantID))
	}

	if demoTenantID == "" {
		logger.Error("cannot seed villas without demo tenant")
		fmt.Println("\n⚠️  Seeding incomplete: demo tenant missing, villas skipped")
		os.Exit(1)
	}

	// ─── Seed Villas ───
	villaData := []struct {
		name, description, address, city, country string
		bedrooms, bathrooms, maxGuests            int
		pricePerNight, cleaningFee, securityDep   float64
		amenities                                 []string
	}{
		{"Villa Al-Mashta", "Luxury mountain villa with panoramic views", "Jabal Al-Mashta, Ramallah", "Ramallah", "Palestine", 3, 2, 6, 350, 50, 500, []string{"wifi", "parking", "pool", "garden", "bbq", "ac", "fireplace"}},
		{"Chalet Al-Balad", "Cozy chalet in the heart of Bethlehem", "Manger Street", "Bethlehem", "Palestine", 4, 3, 8, 450, 75, 750, []string{"wifi", "parking", "fireplace", "kitchen", "ac", "tv"}},
		{"Villa Al-Reef", "Modern villa with sea view terrace", "Gaza Beach Road", "Gaza", "Palestine", 5, 4, 10, 600, 100, 1000, []string{"wifi", "parking", "pool", "gym", "ac", "jacuzzi", "bbq"}},
		{"Chalet Al-Jabal", "Rustic chalet surrounded by olive trees", "Nablus Mountain Road", "Nablus", "Palestine", 2, 1, 4, 250, 40, 300, []string{"wifi", "parking", "garden", "fireplace", "kitchen"}},
		{"Villa Al-Bahar", "Beachfront villa with private access", "Gaza Beach North", "Gaza", "Palestine", 4, 3, 8, 500, 80, 600, []string{"wifi", "parking", "pool", "beach", "ac", "bbq"}},
		{"Chalet Al-Zaytoun", "Traditional stone chalet among olive groves", "Jenin Olive Grove", "Jenin", "Palestine", 3, 2, 6, 300, 45, 400, []string{"wifi", "parking", "garden", "fireplace", "bbq"}},
		{"Villa Al-Safa", "Modern family villa near city center", "Hebron Hills", "Hebron", "Palestine", 4, 2, 7, 400, 60, 500, []string{"wifi", "parking", "pool", "ac", "tv", "kitchen"}},
		{"Chalet Al-Tal", "Hilltop chalet with stunning sunset views", "Bethlehem Hills", "Bethlehem", "Palestine", 3, 2, 5, 320, 50, 350, []string{"wifi", "parking", "balcony", "fireplace", "kitchen"}},
	}

	villaIDs := make([]string, len(villaData))
	for i, v := range villaData {
		var id string
		err := pool.QueryRow(ctx, `
			INSERT INTO villa_properties (id, tenant_id, name, description, address, city, country, bedrooms, bathrooms, max_guests, amenities, price_per_night, currency, cleaning_fee, security_deposit, is_active, status)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'USD', $12, $13, true, 'available')
			ON CONFLICT (tenant_id, name) DO UPDATE SET
				description = EXCLUDED.description,
				address = EXCLUDED.address,
				city = EXCLUDED.city,
				country = EXCLUDED.country,
				bedrooms = EXCLUDED.bedrooms,
				bathrooms = EXCLUDED.bathrooms,
				max_guests = EXCLUDED.max_guests,
				amenities = EXCLUDED.amenities,
				price_per_night = EXCLUDED.price_per_night,
				cleaning_fee = EXCLUDED.cleaning_fee,
				security_deposit = EXCLUDED.security_deposit
			RETURNING id
		`, demoTenantID, v.name, v.description, v.address, v.city, v.country, v.bedrooms, v.bathrooms, v.maxGuests, v.amenities, v.pricePerNight, v.cleaningFee, v.securityDep).Scan(&id)
		if err != nil {
			logger.Warn("villa insert failed", slog.String("name", v.name), slog.String("error", err.Error()))
			continue
		}
		villaIDs[i] = id
		logger.Info("villa seeded", slog.String("name", v.name), slog.String("id", id))
	}

	// ─── Seed Villa Owners (as users) ───
	villaHash := "$2a$10$wKvRKo1uJX0fC4p8K0OdHujT5jaMFeS9OqkZ/vEvX.s.yNXkgXGvK"
	villaOwners := []struct {
		name  string
		email string
	}{
		{"Ahmad Khalil", "ahmad.khalil@villa.test"},
		{"Sarah Nassar", "sarah.nassar@villa.test"},
		{"Mohammed Ali", "mohammed.ali@villa.test"},
		{"Layla Farhat", "layla.farhat@villa.test"},
		{"Nadia Ibrahim", "nadia.ibrahim@villa.test"},
		{"Omar Khalil", "omar.khalil@villa.test"},
		{"Rania Suleiman", "rania.suleiman@villa.test"},
		{"Youssef Nasser", "youssef.nasser@villa.test"},
	}
	for _, o := range villaOwners {
		_, err := pool.Exec(ctx, `
			INSERT INTO users (id, tenant_id, email, password_hash, name, role, is_active)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, 'manager', true)
			ON CONFLICT (email) DO UPDATE SET
				name = EXCLUDED.name,
				role = EXCLUDED.role,
				is_active = true
		`, demoTenantID, o.email, villaHash, o.name)
		if err != nil {
			logger.Warn("villa owner insert failed", slog.String("email", o.email), slog.String("error", err.Error()))
		} else {
			logger.Info("villa owner seeded", slog.String("name", o.name))
		}
	}

	// ─── Seed Villa Reservations ───
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
		{4, "Hassan Farouk", "+970599444555", "hassan@email.com", 6, "2026-05-13", "2026-05-18", 5, 2500, "reserved", "phone", 0, true},
		{4, "Lina Mahmoud", "+970599666777", "lina@email.com", 4, "2026-05-20", "2026-05-24", 4, 2000, "pending", "website", 2000, false},
		{5, "Tariq Salem", "+970599888999", "tariq@email.com", 3, "2026-05-10", "2026-05-14", 4, 1200, "completed", "phone", 0, true},
		{5, "Dana Khatib", "+970599123789", "dana@email.com", 5, "2026-05-18", "2026-05-22", 4, 1200, "pending", "walkin", 1200, false},
		{6, "Samir Haddad", "+970599456123", "samir@email.com", 7, "2026-05-12", "2026-05-16", 4, 1600, "reserved", "whatsapp", 0, true},
		{6, "Nour Abdallah", "+970599789456", "nour@email.com", 4, "2026-05-22", "2026-05-27", 5, 2000, "pending", "website", 2000, false},
		{7, "Waleed Fares", "+970599147258", "waleed@email.com", 3, "2026-05-15", "2026-05-19", 4, 1280, "completed", "phone", 0, true},
		{7, "Hana Ibrahim", "+970599369852", "hana@email.com", 4, "2026-05-25", "2026-05-29", 4, 1280, "pending", "walkin", 1280, false},
	}

	for _, b := range bookings {
		villaID := villaIDs[b.villaIdx]
		if villaID == "" {
			continue
		}
		villaName := villaData[b.villaIdx].name
		_, err := pool.Exec(ctx, `
			INSERT INTO villa_reservations (id, tenant_id, villa_id, villa_name, guest_name, guest_phone, guest_email, guest_count, check_in_date, check_out_date, nights, total_amount, currency, status, source, balance_due, balance_paid)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'USD', $12, $13, $14, $15)
			ON CONFLICT DO NOTHING
		`, demoTenantID, villaID, villaName, b.guest, b.phone, b.email, b.guestCount, b.checkIn, b.checkOut, b.nights, b.totalAmount, b.status, b.source, b.balanceDue, b.balancePaid)
		if err != nil {
			logger.Warn("villa booking insert failed", slog.String("guest", b.guest), slog.String("error", err.Error()))
		}
	}

	// ─── Seed System Config ───
	_, _ = pool.Exec(ctx, `
		INSERT INTO system_configs (tenant_id, default_check_in_time, default_check_out_time, auto_confirm, deposit_percent, allow_walk_in, overbooking_enabled)
		VALUES ($1, '15:00', '11:00', true, 20, true, false)
		ON CONFLICT (tenant_id) DO NOTHING
	`, demoTenantID)

	// ─── Seed properties table entries ───
	for hi, h := range hotels {
		var hotelTenantID string
		_ = pool.QueryRow(ctx, `SELECT id FROM tenants WHERE external_id = $1`, h.externalID).Scan(&hotelTenantID)
		if hotelTenantID == "" {
			continue
		}
		_, _ = pool.Exec(ctx, `
			INSERT INTO properties (id, tenant_id, name, address, city, country, timezone, currency, star_rating, is_active)
			VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, 'UTC', 'USD', $6, true)
			ON CONFLICT (tenant_id, name) DO NOTHING
		`, hotelTenantID, h.name, fmt.Sprintf("%s Main Street", h.city), h.city, h.country, 3+(hi%3))
	}

	logger.Info("🎉 SEED COMPLETE",
		slog.Int("hotels", len(hotels)),
		slog.Int("total_rooms", totalRooms),
		slog.Int("total_reservations", totalReservations),
		slog.Int("villas", len(villaData)),
	)
	fmt.Printf("\n✅ Seeded %d hotels, %d rooms, %d reservations, %d villas\n", len(hotels), totalRooms, totalReservations, len(villaData))
}

func strPtr(s string) *string {
	return &s
}
