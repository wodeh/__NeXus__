package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// ChannelManagerRepository handles OTA channel data.
type ChannelManagerRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

// NewChannelManagerRepository creates a new repository.
func NewChannelManagerRepository(pool *db.Pool, metrics *RepositoryMetrics) *ChannelManagerRepository {
	return &ChannelManagerRepository{pool: pool, metrics: metrics}
}

// ========== CONNECTIONS ==========

func (r *ChannelManagerRepository) CreateConnection(ctx context.Context, tenantID string, conn *domain.ChannelConnection) (*domain.ChannelConnection, error) {
	r.pool.SetTenant(ctx, tenantID)
	start := time.Now()
	defer func() { r.metrics.ObserveDuration("channel_manager", "CreateConnection", time.Since(start).Seconds()) }()

	query := `
		INSERT INTO channel_connections (tenant_id, source, display_name, api_key, api_secret, property_id, is_active, commission_pct)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query, tenantID, conn.Source, conn.DisplayName, conn.APIKey, conn.APISecret, conn.PropertyID, conn.IsActive, conn.CommissionPct)
	err := row.Scan(&conn.ID, &conn.CreatedAt, &conn.UpdatedAt)
	if err != nil {
		r.metrics.IncError("channel_manager", "CreateConnection", "db_error")
		return nil, fmt.Errorf("create connection: %w", err)
	}
	r.metrics.IncQuery("channel_manager", "CreateConnection")
	return conn, nil
}

func (r *ChannelManagerRepository) ListConnections(ctx context.Context, tenantID string) ([]*domain.ChannelConnection, error) {
	r.pool.SetTenant(ctx, tenantID)
	start := time.Now()
	defer func() { r.metrics.ObserveDuration("channel_manager", "ListConnections", time.Since(start).Seconds()) }()

	rows, err := r.pool.Query(ctx, `
		SELECT id, source, display_name, property_id, is_active, commission_pct, last_sync_at, last_sync_status, last_sync_error, created_at, updated_at
		FROM channel_connections WHERE tenant_id = $1 ORDER BY display_name`, tenantID)
	if err != nil {
		r.metrics.IncError("channel_manager", "ListConnections", "db_error")
		return nil, err
	}
	defer rows.Close()

	var out []*domain.ChannelConnection
	for rows.Next() {
		var c domain.ChannelConnection
		err := rows.Scan(&c.ID, &c.Source, &c.DisplayName, &c.PropertyID, &c.IsActive, &c.CommissionPct, &c.LastSyncAt, &c.LastSyncStatus, &c.LastSyncError, &c.CreatedAt, &c.UpdatedAt)
		if err != nil {
			continue
		}
		c.TenantID = tenantID
		out = append(out, &c)
	}
	r.metrics.IncQuery("channel_manager", "ListConnections")
	return out, nil
}

func (r *ChannelManagerRepository) UpdateConnection(ctx context.Context, tenantID string, id string, updates map[string]interface{}) error {
	r.pool.SetTenant(ctx, tenantID)
	return r.updateByID(ctx, tenantID, "channel_connections", id, updates)
}

func (r *ChannelManagerRepository) DeleteConnection(ctx context.Context, tenantID string, id string) error {
	r.pool.SetTenant(ctx, tenantID)
	_, err := r.pool.Exec(ctx, `DELETE FROM channel_connections WHERE id = $1 AND tenant_id = $2`, id, tenantID)
	return err
}

// ========== CHANNEL RESERVATIONS ==========

func (r *ChannelManagerRepository) CreateChannelReservation(ctx context.Context, tenantID string, res *domain.ChannelReservation) (*domain.ChannelReservation, error) {
	r.pool.SetTenant(ctx, tenantID)
	start := time.Now()
	defer func() { r.metrics.ObserveDuration("channel_manager", "CreateChannelReservation", time.Since(start).Seconds()) }()

	query := `
		INSERT INTO channel_reservations (tenant_id, channel_id, source, external_ref, guest_name, guest_email, guest_phone, room_type_id, room_id, check_in, check_out, nights, adults, children, total_amount, commission, net_amount, currency, status, special_requests, raw_payload)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)
		RETURNING id, created_at, updated_at`

	row := r.pool.QueryRow(ctx, query, tenantID, res.ChannelID, res.Source, res.ExternalRef, res.GuestName, res.GuestEmail, res.GuestPhone, res.RoomTypeID, res.RoomID, res.CheckIn, res.CheckOut, res.Nights, res.Adults, res.Children, res.TotalAmount, res.Commission, res.NetAmount, res.Currency, res.Status, res.SpecialRequests, res.RawPayload)
	err := row.Scan(&res.ID, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		r.metrics.IncError("channel_manager", "CreateChannelReservation", "db_error")
		return nil, fmt.Errorf("create channel reservation: %w", err)
	}
	r.metrics.IncQuery("channel_manager", "CreateChannelReservation")
	return res, nil
}

func (r *ChannelManagerRepository) ListChannelReservations(ctx context.Context, tenantID string, source *domain.ChannelManagerSource, status *string, fromDate, toDate *time.Time) ([]*domain.ChannelReservation, error) {
	r.pool.SetTenant(ctx, tenantID)
	start := time.Now()
	defer func() { r.metrics.ObserveDuration("channel_manager", "ListChannelReservations", time.Since(start).Seconds()) }()

	query := `SELECT id, channel_id, source, external_ref, guest_name, guest_email, guest_phone, room_type_id, room_id, check_in, check_out, nights, adults, children, total_amount, commission, net_amount, currency, status, special_requests, mapped_to_reservation_id, created_at, updated_at FROM channel_reservations WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argCount := 1

	if source != nil {
		argCount++
		query += fmt.Sprintf(" AND source = $%d", argCount)
		args = append(args, *source)
	}
	if status != nil {
		argCount++
		query += fmt.Sprintf(" AND status = $%d", argCount)
		args = append(args, *status)
	}
	if fromDate != nil {
		argCount++
		query += fmt.Sprintf(" AND check_out >= $%d", argCount)
		args = append(args, *fromDate)
	}
	if toDate != nil {
		argCount++
		query += fmt.Sprintf(" AND check_in <= $%d", argCount)
		args = append(args, *toDate)
	}

	query += " ORDER BY check_in DESC"

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		r.metrics.IncError("channel_manager", "ListChannelReservations", "db_error")
		return nil, err
	}
	defer rows.Close()

	var out []*domain.ChannelReservation
	for rows.Next() {
		var r domain.ChannelReservation
		rows.Scan(&r.ID, &r.ChannelID, &r.Source, &r.ExternalRef, &r.GuestName, &r.GuestEmail, &r.GuestPhone, &r.RoomTypeID, &r.RoomID, &r.CheckIn, &r.CheckOut, &r.Nights, &r.Adults, &r.Children, &r.TotalAmount, &r.Commission, &r.NetAmount, &r.Currency, &r.Status, &r.SpecialRequests, &r.MappedToReservationID, &r.CreatedAt, &r.UpdatedAt)
		r.TenantID = tenantID
		out = append(out, &r)
	}
	r.metrics.IncQuery("channel_manager", "ListChannelReservations")
	return out, nil
}

func (r *ChannelManagerRepository) MapToReservation(ctx context.Context, tenantID, channelResID, reservationID string) error {
	r.pool.SetTenant(ctx, tenantID)
	_, err := r.pool.Exec(ctx, `UPDATE channel_reservations SET mapped_to_reservation_id = $1, status = 'confirmed' WHERE id = $2 AND tenant_id = $3`, reservationID, channelResID, tenantID)
	return err
}

// ========== SYNC LOGS ==========

func (r *ChannelManagerRepository) CreateSyncLog(ctx context.Context, tenantID string, log *domain.ChannelSyncLog) (*domain.ChannelSyncLog, error) {
	r.pool.SetTenant(ctx, tenantID)
	query := `INSERT INTO channel_sync_logs (tenant_id, channel_id, direction, status, records, error_msg, started_at, completed_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`
	err := r.pool.QueryRow(ctx, query, tenantID, log.ChannelID, log.Direction, log.Status, log.Records, log.ErrorMsg, log.StartedAt, log.CompletedAt).Scan(&log.ID)
	return log, err
}

func (r *ChannelManagerRepository) ListSyncLogs(ctx context.Context, tenantID string, limit int) ([]*domain.ChannelSyncLog, error) {
	r.pool.SetTenant(ctx, tenantID)
	rows, err := r.pool.Query(ctx, `SELECT id, channel_id, direction, status, records, error_msg, started_at, completed_at FROM channel_sync_logs WHERE tenant_id = $1 ORDER BY started_at DESC LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.ChannelSyncLog
	for rows.Next() {
		var l domain.ChannelSyncLog
		rows.Scan(&l.ID, &l.ChannelID, &l.Direction, &l.Status, &l.Records, &l.ErrorMsg, &l.StartedAt, &l.CompletedAt)
		l.TenantID = tenantID
		out = append(out, &l)
	}
	return out, nil
}

// ========== HEALTH ==========

func (r *ChannelManagerRepository) GetChannelHealth(ctx context.Context, tenantID string, days int) ([]*domain.ChannelHealth, error) {
	r.pool.SetTenant(ctx, tenantID)
	query := `
		SELECT 
			c.id as channel_id,
			c.source,
			c.display_name,
			COUNT(cr.id) as bookings_30d,
			COALESCE(SUM(cr.total_amount), 0) as revenue_30d,
			COALESCE(SUM(cr.commission), 0) as commission_30d,
			COALESCE(SUM(cr.net_amount), 0) as net_revenue_30d,
			COALESCE(AVG(cr.total_amount / NULLIF(cr.nights, 0)), 0) as adr,
			COUNT(CASE WHEN cr.status = 'cancelled' THEN 1 END)::float / NULLIF(COUNT(cr.id), 0) * 100 as cancellation_rate
		FROM channel_connections c
		LEFT JOIN channel_reservations cr ON cr.channel_id = c.id AND cr.check_in >= NOW() - INTERVAL '1 day' * $2
		WHERE c.tenant_id = $1
		GROUP BY c.id, c.source, c.display_name
	`
	rows, err := r.pool.Query(ctx, query, tenantID, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*domain.ChannelHealth
	for rows.Next() {
		var h domain.ChannelHealth
		var cancelRate float64
		rows.Scan(&h.ChannelID, &h.Source, &h.DisplayName, &h.Bookings30d, &h.Revenue30d, &h.Commission30d, &h.NetRevenue30d, &h.ADR, &cancelRate)
		h.CancellationRate = cancelRate
		h.IsHealthy = cancelRate < 15
		out = append(out, &h)
	}
	return out, nil
}

// updateByID builds and executes a dynamic UPDATE query.
func (r *ChannelManagerRepository) updateByID(ctx context.Context, tenantID, table, id string, updates map[string]interface{}) error {
	query := "UPDATE " + table + " SET updated_at = NOW()"
	args := []interface{}{}
	argCount := 0
	for col, val := range updates {
		argCount++
		query += fmt.Sprintf(", %s = $%d", col, argCount)
		args = append(args, val)
	}
	argCount++
	query += fmt.Sprintf(" WHERE tenant_id = $%d AND id = $%d", argCount, argCount+1)
	args = append(args, tenantID, id)
	_, err := r.pool.Exec(ctx, query, args...)
	return err
}
