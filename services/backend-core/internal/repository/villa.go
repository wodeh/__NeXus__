package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

// VillaRepository handles villa rental data access.
type VillaRepository struct {
	pool *db.Pool
}

// NewVillaRepository creates a new instance.
func NewVillaRepository(pool *db.Pool) *VillaRepository {
	return &VillaRepository{pool: pool}
}

// --- Villa Properties ---

func (r *VillaRepository) ListVillas(ctx context.Context) ([]domain.VillaProperty, error) {
	query := `
		SELECT id, tenant_id, name, description, address, city, country, 
		       latitude, longitude, elevation, bedrooms, bathrooms, max_guests,
		       amenities, images, price_per_night, currency, cleaning_fee, 
		       security_deposit, is_active, status, created_at, updated_at
		FROM villa_properties
		ORDER BY name
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list villas: %w", err)
	}
	defer rows.Close()

	var villas []domain.VillaProperty
	for rows.Next() {
		var v domain.VillaProperty
		err := rows.Scan(
			&v.ID, &v.TenantID, &v.Name, &v.Description, &v.Address, &v.City, &v.Country,
			&v.Latitude, &v.Longitude, &v.Elevation, &v.Bedrooms, &v.Bathrooms, &v.MaxGuests,
			&v.Amenities, &v.Images, &v.PricePerNight, &v.Currency, &v.CleaningFee,
			&v.SecurityDeposit, &v.IsActive, &v.Status, &v.CreatedAt, &v.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		villas = append(villas, v)
	}
	return villas, nil
}

func (r *VillaRepository) GetVillaByID(ctx context.Context, id string) (*domain.VillaProperty, error) {
	query := `
		SELECT id, tenant_id, name, description, address, city, country,
		       latitude, longitude, elevation, bedrooms, bathrooms, max_guests,
		       amenities, images, price_per_night, currency, cleaning_fee,
		       security_deposit, is_active, status, created_at, updated_at
		FROM villa_properties WHERE id = $1
	`
	var v domain.VillaProperty
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&v.ID, &v.TenantID, &v.Name, &v.Description, &v.Address, &v.City, &v.Country,
		&v.Latitude, &v.Longitude, &v.Elevation, &v.Bedrooms, &v.Bathrooms, &v.MaxGuests,
		&v.Amenities, &v.Images, &v.PricePerNight, &v.Currency, &v.CleaningFee,
		&v.SecurityDeposit, &v.IsActive, &v.Status, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("get villa: %w", err)
	}
	return &v, nil
}

func (r *VillaRepository) CreateVilla(ctx context.Context, v *domain.VillaProperty) error {
	query := `
		INSERT INTO villa_properties (
			id, tenant_id, name, description, address, city, country,
			latitude, longitude, elevation, bedrooms, bathrooms, max_guests,
			amenities, images, price_per_night, currency, cleaning_fee,
			security_deposit, is_active, status, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18, $19, $20, NOW(), NOW()
		) RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		v.TenantID, v.Name, v.Description, v.Address, v.City, v.Country,
		v.Latitude, v.Longitude, v.Elevation, v.Bedrooms, v.Bathrooms, v.MaxGuests,
		v.Amenities, v.Images, v.PricePerNight, v.Currency, v.CleaningFee,
		v.SecurityDeposit, v.IsActive, v.Status,
	).Scan(&v.ID, &v.CreatedAt, &v.UpdatedAt)
}

func (r *VillaRepository) UpdateVilla(ctx context.Context, v *domain.VillaProperty) error {
	query := `
		UPDATE villa_properties SET
			name = $2, description = $3, address = $4, city = $5, country = $6,
			latitude = $7, longitude = $8, elevation = $9, bedrooms = $10, bathrooms = $11,
			max_guests = $12, amenities = $13, images = $14, price_per_night = $15,
			currency = $16, cleaning_fee = $17, security_deposit = $18,
			is_active = $19, status = $20, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		v.ID, v.Name, v.Description, v.Address, v.City, v.Country,
		v.Latitude, v.Longitude, v.Elevation, v.Bedrooms, v.Bathrooms, v.MaxGuests,
		v.Amenities, v.Images, v.PricePerNight, v.Currency, v.CleaningFee,
		v.SecurityDeposit, v.IsActive, v.Status,
	)
	return err
}

func (r *VillaRepository) DeleteVilla(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM villa_properties WHERE id = $1`, id)
	return err
}

// --- Villa Reservations ---

func (r *VillaRepository) ListVillaReservations(ctx context.Context) ([]domain.VillaReservation, error) {
	query := `
		SELECT id, tenant_id, villa_id, villa_name, guest_name, guest_phone, guest_email,
		       guest_count, check_in_date, check_out_date, nights, total_amount, currency,
		       status, source, internal_notes, down_payment, balance_due, balance_paid,
		       created_at, updated_at, completed_at
		FROM villa_reservations
		ORDER BY check_in_date DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []domain.VillaReservation
	for rows.Next() {
		var res domain.VillaReservation
		err := rows.Scan(
			&res.ID, &res.TenantID, &res.VillaID, &res.VillaName, &res.GuestName, &res.GuestPhone,
			&res.GuestEmail, &res.GuestCount, &res.CheckInDate, &res.CheckOutDate, &res.Nights,
			&res.TotalAmount, &res.Currency, &res.Status, &res.Source, &res.InternalNotes,
			&res.DownPayment, &res.BalanceDue, &res.BalancePaid,
			&res.CreatedAt, &res.UpdatedAt, &res.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, res)
	}
	return reservations, nil
}

func (r *VillaRepository) GetVillaReservationByID(ctx context.Context, id string) (*domain.VillaReservation, error) {
	query := `
		SELECT id, tenant_id, villa_id, villa_name, guest_name, guest_phone, guest_email,
		       guest_count, check_in_date, check_out_date, nights, total_amount, currency,
		       status, source, internal_notes, down_payment, balance_due, balance_paid,
		       created_at, updated_at, completed_at
		FROM villa_reservations WHERE id = $1
	`
	var res domain.VillaReservation
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&res.ID, &res.TenantID, &res.VillaID, &res.VillaName, &res.GuestName, &res.GuestPhone,
		&res.GuestEmail, &res.GuestCount, &res.CheckInDate, &res.CheckOutDate, &res.Nights,
		&res.TotalAmount, &res.Currency, &res.Status, &res.Source, &res.InternalNotes,
		&res.DownPayment, &res.BalanceDue, &res.BalancePaid,
		&res.CreatedAt, &res.UpdatedAt, &res.CompletedAt,
	)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *VillaRepository) CreateVillaReservation(ctx context.Context, res *domain.VillaReservation) error {
	query := `
		INSERT INTO villa_reservations (
			id, tenant_id, villa_id, villa_name, guest_name, guest_phone, guest_email,
			guest_count, check_in_date, check_out_date, nights, total_amount, currency,
			status, source, internal_notes, down_payment, balance_due, balance_paid,
			created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12,
			$13, $14, $15, $16, $17, $18, NOW(), NOW()
		) RETURNING id, created_at, updated_at
	`
	return r.pool.QueryRow(ctx, query,
		res.TenantID, res.VillaID, res.VillaName, res.GuestName, res.GuestPhone,
		res.GuestEmail, res.GuestCount, res.CheckInDate, res.CheckOutDate, res.Nights,
		res.TotalAmount, res.Currency, res.Status, res.Source, res.InternalNotes,
		res.DownPayment, res.BalanceDue, res.BalancePaid,
	).Scan(&res.ID, &res.CreatedAt, &res.UpdatedAt)
}

func (r *VillaRepository) UpdateVillaReservation(ctx context.Context, res *domain.VillaReservation) error {
	query := `
		UPDATE villa_reservations SET
			villa_id = $2, villa_name = $3, guest_name = $4, guest_phone = $5, guest_email = $6,
			guest_count = $7, check_in_date = $8, check_out_date = $9, nights = $10,
			total_amount = $11, currency = $12, status = $13, source = $14,
			internal_notes = $15, down_payment = $16, balance_due = $17, balance_paid = $18,
			completed_at = $19, updated_at = NOW()
		WHERE id = $1
	`
	_, err := r.pool.Exec(ctx, query,
		res.ID, res.VillaID, res.VillaName, res.GuestName, res.GuestPhone,
		res.GuestEmail, res.GuestCount, res.CheckInDate, res.CheckOutDate, res.Nights,
		res.TotalAmount, res.Currency, res.Status, res.Source, res.InternalNotes,
		res.DownPayment, res.BalanceDue, res.BalancePaid, res.CompletedAt,
	)
	return err
}

func (r *VillaRepository) DeleteVillaReservation(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM villa_reservations WHERE id = $1`, id)
	return err
}

func (r *VillaRepository) GetVillaReservationByDate(ctx context.Context, villaID string, date string) (*domain.VillaReservation, error) {
	query := `
		SELECT id, tenant_id, villa_id, villa_name, guest_name, guest_phone, guest_email,
		       guest_count, check_in_date::text, check_out_date::text, nights, total_amount, currency,
		       status, source, internal_notes, balance_due, balance_paid,
		       created_at, updated_at
		FROM villa_reservations
		WHERE villa_id = $1 AND check_in_date <= $2::date AND check_out_date > $2::date
		  AND status IN ('reserved', 'completed')
		ORDER BY check_in_date DESC
		LIMIT 1
	`
	var res domain.VillaReservation
	err := r.pool.QueryRow(ctx, query, villaID, date).Scan(
		&res.ID, &res.TenantID, &res.VillaID, &res.VillaName, &res.GuestName, &res.GuestPhone,
		&res.GuestEmail, &res.GuestCount, &res.CheckInDate, &res.CheckOutDate, &res.Nights,
		&res.TotalAmount, &res.Currency, &res.Status, &res.Source, &res.InternalNotes,
		&res.BalanceDue, &res.BalancePaid,
		&res.CreatedAt, &res.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *VillaRepository) RecordSensorLog(ctx context.Context, log *domain.CleanerSensorLog) error {
	query := `
		INSERT INTO cleaner_sensor_logs (
			id, tenant_id, villa_id, villa_name, cleaner_id, cleaner_name,
			temperature, latitude, longitude, altitude, floor, location_type,
			battery_level, recorded_at, created_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW()
		) RETURNING id, created_at
	`
	return r.pool.QueryRow(ctx, query,
		log.TenantID, log.VillaID, log.VillaName, log.CleanerID, log.CleanerName,
		log.Temperature, log.Latitude, log.Longitude, log.Altitude, log.Floor,
		log.LocationType, log.BatteryLevel, log.RecordedAt,
	).Scan(&log.ID, &log.CreatedAt)
}

func (r *VillaRepository) GetSensorLogsByVilla(ctx context.Context, villaID string) ([]domain.CleanerSensorLog, error) {
	query := `
		SELECT id, tenant_id, villa_id, villa_name, cleaner_id, cleaner_name,
		       temperature, latitude, longitude, altitude, floor, location_type,
		       battery_level, recorded_at, created_at
		FROM cleaner_sensor_logs
		WHERE villa_id = $1
		ORDER BY recorded_at DESC
	`
	rows, err := r.pool.Query(ctx, query, villaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.CleanerSensorLog
	for rows.Next() {
		var log domain.CleanerSensorLog
		err := rows.Scan(
			&log.ID, &log.TenantID, &log.VillaID, &log.VillaName, &log.CleanerID, &log.CleanerName,
			&log.Temperature, &log.Latitude, &log.Longitude, &log.Altitude, &log.Floor,
			&log.LocationType, &log.BatteryLevel, &log.RecordedAt, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	return logs, nil
}

func (r *VillaRepository) GetLatestSensorLog(ctx context.Context, villaID string, cleanerID string) (*domain.CleanerSensorLog, error) {
	query := `
		SELECT id, tenant_id, villa_id, villa_name, cleaner_id, cleaner_name,
		       temperature, latitude, longitude, altitude, floor, location_type,
		       battery_level, recorded_at, created_at
		FROM cleaner_sensor_logs
		WHERE villa_id = $1 AND cleaner_id = $2
		ORDER BY recorded_at DESC
		LIMIT 1
	`
	var log domain.CleanerSensorLog
	err := r.pool.QueryRow(ctx, query, villaID, cleanerID).Scan(
		&log.ID, &log.TenantID, &log.VillaID, &log.VillaName, &log.CleanerID, &log.CleanerName,
		&log.Temperature, &log.Latitude, &log.Longitude, &log.Altitude, &log.Floor,
		&log.LocationType, &log.BatteryLevel, &log.RecordedAt, &log.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// --- Availability ---

func (r *VillaRepository) GetVillaAvailability(ctx context.Context, villaID string, start, end string) ([]domain.VillaAvailability, error) {
	query := `
		WITH date_range AS (
			SELECT generate_series($2::date, $3::date, '1 day')::date AS date
		)
		SELECT dr.date, COALESCE(vr.status, 'available') AS status, vr.reservation_id
		FROM date_range dr
		LEFT JOIN LATERAL (
			SELECT 'reserved' AS status, id AS reservation_id
			FROM villa_reservations
			WHERE villa_id = $1
			  AND dr.date >= check_in_date
			  AND dr.date < check_out_date
			  AND status IN ('reserved', 'completed')
			LIMIT 1
		) vr ON true
		ORDER BY dr.date
	`
	rows, err := r.pool.Query(ctx, query, villaID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var avail []domain.VillaAvailability
	for rows.Next() {
		var a domain.VillaAvailability
		a.VillaID = villaID
		err := rows.Scan(&a.Date, &a.Status, &a.ReservationID)
		if err != nil {
			return nil, err
		}
		avail = append(avail, a)
	}
	return avail, nil
}

// --- Revenue Stats ---

func (r *VillaRepository) GetVillaRevenueStats(ctx context.Context, start, end string) (*domain.VillaRevenueStats, error) {
	query := `
		SELECT 
			COUNT(*) AS total_reservations,
			COALESCE(SUM(total_amount), 0) AS total_revenue,
			COALESCE(AVG(total_amount), 0) AS avg_booking_value,
			COALESCE(SUM(CASE WHEN down_payment->>'status' = 'received' THEN (down_payment->>'amount')::numeric ELSE 0 END), 0) AS down_payment_total,
			COALESCE(SUM(balance_due), 0) AS pending_balance
		FROM villa_reservations
		WHERE status IN ('reserved', 'completed')
		  AND check_in_date >= $1
		  AND check_in_date <= $2
	`
	var stats domain.VillaRevenueStats
	stats.PeriodStart = start
	stats.PeriodEnd = end
	err := r.pool.QueryRow(ctx, query, start, end).Scan(
		&stats.TotalReservations, &stats.TotalRevenue, &stats.AvgBookingValue,
		&stats.DownPaymentTotal, &stats.PendingBalance,
	)
	if err != nil {
		return nil, err
	}

	// Calculate occupancy
	villaQuery := `
		SELECT vp.id, vp.name,
			COALESCE(SUM(vr.nights), 0) AS nights_booked,
			COUNT(CASE WHEN vr.status IN ('reserved', 'completed') THEN 1 END) AS booking_count
		FROM villa_properties vp
		LEFT JOIN villa_reservations vr ON vp.id = vr.villa_id
			AND vr.check_in_date >= $1
			AND vr.check_in_date <= $2
			AND vr.status IN ('reserved', 'completed')
		GROUP BY vp.id, vp.name
	`
	rows, err := r.pool.Query(ctx, villaQuery, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	totalVillaNights := 0
	for rows.Next() {
		var vb domain.VillaRevenueBreakdown
		var nightsBooked int
		err := rows.Scan(&vb.VillaID, &vb.VillaName, &nightsBooked, &vb.NightsBooked)
		if err != nil {
			return nil, err
		}
		vb.NightsBooked = int(nightsBooked)
		vb.Revenue = 0 // Will be filled by separate query if needed
		totalVillaNights += int(nightsBooked)
		stats.VillaBreakdown = append(stats.VillaBreakdown, vb)
	}

	startT, _ := time.Parse("2006-01-02", start)
	endT, _ := time.Parse("2006-01-02", end)
	daysInPeriod := int(endT.Sub(startT).Hours()/24) + 1
	if daysInPeriod > 0 {
		stats.OccupancyRate = float64(totalVillaNights) / float64(len(stats.VillaBreakdown)*daysInPeriod) * 100
	}

	return &stats, nil
}
