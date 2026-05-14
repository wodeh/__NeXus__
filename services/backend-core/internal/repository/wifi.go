package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/backend-core/internal/db"
)

// WiFiAccessPoint represents an enterprise AP in the hotel.
type WiFiAccessPoint struct {
	ID             uuid.UUID `json:"id"`
	TenantID       uuid.UUID `json:"tenant_id"`
	Name           string    `json:"name"`
	Floor          string    `json:"floor"`
	Location       string    `json:"location"`
	MacAddress     string    `json:"mac_address"`
	IPAddress      *string   `json:"ip_address,omitempty"`
	Model          string    `json:"model"`
	Firmware       string    `json:"firmware"`
	Channels2GHz   []int32   `json:"channels_2ghz,omitempty"`
	Channels5GHz   []int32   `json:"channels_5ghz,omitempty"`
	MaxClients     int32     `json:"max_clients"`
	Status         string    `json:"status"`
	SNMPCommunity  *string   `json:"snmp_community,omitempty"`
	SNMPVersion    string    `json:"snmp_version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// WiFiMetric represents a single SNMP poll reading.
type WiFiMetric struct {
	ID            uuid.UUID `json:"id"`
	TenantID      uuid.UUID `json:"tenant_id"`
	APID          uuid.UUID `json:"ap_id"`
	Floor         string    `json:"floor"`
	Timestamp     time.Time `json:"timestamp"`
	RSSIDbm       *int32    `json:"rssi_dbm,omitempty"`
	NoiseDbm      *int32    `json:"noise_dbm,omitempty"`
	SNRDb         *int32    `json:"snr_db,omitempty"`
	QualityScore  *int32    `json:"quality_score,omitempty"`
	ClientCount   int32     `json:"client_count"`
	BandwidthMbps *float64  `json:"bandwidth_mbps,omitempty"`
	PacketLossPct *float64  `json:"packet_loss_pct,omitempty"`
	ChannelUtil   *float64  `json:"channel_util,omitempty"`
	TXRateMbps    *float64  `json:"tx_rate_mbps,omitempty"`
	RXRateMbps    *float64  `json:"rx_rate_mbps,omitempty"`
	RetriesPct    *float64  `json:"retries_pct,omitempty"`
}

// WiFiAlert represents a detected network issue.
type WiFiAlert struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	APID         *uuid.UUID `json:"ap_id,omitempty"`
	Floor        string     `json:"floor"`
	Type         string     `json:"type"`
	Severity     string     `json:"severity"`
	Message      string     `json:"message"`
	SuggestedFix *string    `json:"suggested_fix,omitempty"`
	IsResolved   bool       `json:"is_resolved"`
	ResolvedAt   *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy   *string    `json:"resolved_by,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// WiFiFloorSummary provides fast dashboard reads.
type WiFiFloorSummary struct {
	ID           uuid.UUID `json:"id"`
	TenantID     uuid.UUID `json:"tenant_id"`
	Floor        string    `json:"floor"`
	APCount      int32     `json:"ap_count"`
	OnlineAPs    int32     `json:"online_aps"`
	AvgSignalDbm *int32    `json:"avg_signal_dbm,omitempty"`
	TotalClients int32     `json:"total_clients"`
	ActiveAlerts int32     `json:"active_alerts"`
	OverallHealth string   `json:"overall_health"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// WiFiRepository handles WiFi monitoring database operations.
type WiFiRepository struct {
	pool *db.Pool
}

// NewWiFiRepository creates a new WiFi repository.
func NewWiFiRepository(pool *db.Pool) *WiFiRepository {
	return &WiFiRepository{pool: pool}
}

// ── Access Points ──

// ListAccessPoints returns all APs for a tenant.
func (r *WiFiRepository) ListAccessPoints(ctx context.Context, tenantID uuid.UUID) ([]WiFiAccessPoint, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, floor, location, mac_address, ip_address, model, firmware,
		       channels_2ghz, channels_5ghz, max_clients, status, snmp_community, snmp_version,
		       created_at, updated_at
		FROM wifi_access_points
		WHERE tenant_id = $1
		ORDER BY floor, name
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	aps := make([]WiFiAccessPoint, 0)
	for rows.Next() {
		var ap WiFiAccessPoint
		var ipAddr, snmpComm *string
		err := rows.Scan(
			&ap.ID, &ap.TenantID, &ap.Name, &ap.Floor, &ap.Location, &ap.MacAddress, &ipAddr,
			&ap.Model, &ap.Firmware, &ap.Channels2GHz, &ap.Channels5GHz, &ap.MaxClients, &ap.Status,
			&snmpComm, &ap.SNMPVersion, &ap.CreatedAt, &ap.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		ap.IPAddress = ipAddr
		ap.SNMPCommunity = snmpComm
		aps = append(aps, ap)
	}
	return aps, rows.Err()
}

// GetAccessPoint returns a single AP by ID.
func (r *WiFiRepository) GetAccessPoint(ctx context.Context, tenantID, apID uuid.UUID) (*WiFiAccessPoint, error) {
	var ap WiFiAccessPoint
	var ipAddr, snmpComm *string
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, floor, location, mac_address, ip_address, model, firmware,
		       channels_2ghz, channels_5ghz, max_clients, status, snmp_community, snmp_version,
		       created_at, updated_at
		FROM wifi_access_points
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, apID).Scan(
		&ap.ID, &ap.TenantID, &ap.Name, &ap.Floor, &ap.Location, &ap.MacAddress, &ipAddr,
		&ap.Model, &ap.Firmware, &ap.Channels2GHz, &ap.Channels5GHz, &ap.MaxClients, &ap.Status,
		&snmpComm, &ap.SNMPVersion, &ap.CreatedAt, &ap.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	ap.IPAddress = ipAddr
	ap.SNMPCommunity = snmpComm
	return &ap, nil
}

// CreateAccessPoint creates a new AP.
func (r *WiFiRepository) CreateAccessPoint(ctx context.Context, ap *WiFiAccessPoint) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO wifi_access_points (tenant_id, name, floor, location, mac_address, ip_address,
			model, firmware, channels_2ghz, channels_5ghz, max_clients, status, snmp_community, snmp_version)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`, ap.TenantID, ap.Name, ap.Floor, ap.Location, ap.MacAddress, ap.IPAddress,
		ap.Model, ap.Firmware, ap.Channels2GHz, ap.Channels5GHz, ap.MaxClients, ap.Status,
		ap.SNMPCommunity, ap.SNMPVersion,
	).Scan(&ap.ID, &ap.CreatedAt, &ap.UpdatedAt)
}

// UpdateAccessPoint updates an existing AP.
func (r *WiFiRepository) UpdateAccessPoint(ctx context.Context, ap *WiFiAccessPoint) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE wifi_access_points SET
			name = $1, floor = $2, location = $3, mac_address = $4, ip_address = $5,
			model = $6, firmware = $7, channels_2ghz = $8, channels_5ghz = $9,
			max_clients = $10, status = $11, snmp_community = $12, snmp_version = $13,
			updated_at = NOW()
		WHERE tenant_id = $14 AND id = $15
	`, ap.Name, ap.Floor, ap.Location, ap.MacAddress, ap.IPAddress,
		ap.Model, ap.Firmware, ap.Channels2GHz, ap.Channels5GHz, ap.MaxClients, ap.Status,
		ap.SNMPCommunity, ap.SNMPVersion, ap.TenantID, ap.ID,
	)
	return err
}

// DeleteAccessPoint removes an AP.
func (r *WiFiRepository) DeleteAccessPoint(ctx context.Context, tenantID, apID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM wifi_access_points WHERE tenant_id = $1 AND id = $2`, tenantID, apID)
	return err
}

// ── Metrics ──

// InsertMetric records a new polling reading.
func (r *WiFiRepository) InsertMetric(ctx context.Context, m *WiFiMetric) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO wifi_metrics (tenant_id, ap_id, floor, rssi_dbm, noise_dbm, snr_db,
			quality_score, client_count, bandwidth_mbps, packet_loss_pct, channel_util,
			tx_rate_mbps, rx_rate_mbps, retries_pct)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, timestamp
	`, m.TenantID, m.APID, m.Floor, m.RSSIDbm, m.NoiseDbm, m.SNRDb,
		m.QualityScore, m.ClientCount, m.BandwidthMbps, m.PacketLossPct, m.ChannelUtil,
		m.TXRateMbps, m.RXRateMbps, m.RetriesPct,
	).Scan(&m.ID, &m.Timestamp)
}

// GetFloorMetrics returns the latest metric per AP for a floor.
func (r *WiFiRepository) GetFloorMetrics(ctx context.Context, tenantID uuid.UUID, floor string) ([]WiFiMetric, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT ON (ap_id)
			m.id, m.tenant_id, m.ap_id, m.floor, m.timestamp, m.rssi_dbm, m.noise_dbm, m.snr_db,
			m.quality_score, m.client_count, m.bandwidth_mbps, m.packet_loss_pct, m.channel_util,
			m.tx_rate_mbps, m.rx_rate_mbps, m.retries_pct
		FROM wifi_metrics m
		JOIN wifi_access_points a ON a.id = m.ap_id
		WHERE m.tenant_id = $1 AND m.floor = $2
		ORDER BY ap_id, m.timestamp DESC
	`, tenantID, floor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMetrics(rows)
}

// GetAPMetricsHistory returns metrics for a specific AP over time.
func (r *WiFiRepository) GetAPMetricsHistory(ctx context.Context, tenantID, apID uuid.UUID, since time.Time) ([]WiFiMetric, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, ap_id, floor, timestamp, rssi_dbm, noise_dbm, snr_db,
		       quality_score, client_count, bandwidth_mbps, packet_loss_pct, channel_util,
		       tx_rate_mbps, rx_rate_mbps, retries_pct
		FROM wifi_metrics
		WHERE tenant_id = $1 AND ap_id = $2 AND timestamp >= $3
		ORDER BY timestamp DESC
		LIMIT 1000
	`, tenantID, apID, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanMetrics(rows)
}

func scanMetrics(rows pgx.Rows) ([]WiFiMetric, error) {
	metrics := make([]WiFiMetric, 0)
	for rows.Next() {
		var m WiFiMetric
		var rssi, noise, snr, quality *int32
		var bw, loss, util, tx, rx, retries *float64
		err := rows.Scan(
			&m.ID, &m.TenantID, &m.APID, &m.Floor, &m.Timestamp,
			&rssi, &noise, &snr, &quality, &m.ClientCount,
			&bw, &loss, &util, &tx, &rx, &retries,
		)
		if err != nil {
			return nil, err
		}
		m.RSSIDbm = rssi; m.NoiseDbm = noise; m.SNRDb = snr; m.QualityScore = quality
		m.BandwidthMbps = bw; m.PacketLossPct = loss; m.ChannelUtil = util
		m.TXRateMbps = tx; m.RXRateMbps = rx; m.RetriesPct = retries
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

// ── Alerts ──

// CreateAlert inserts a new WiFi alert.
func (r *WiFiRepository) CreateAlert(ctx context.Context, alert *WiFiAlert) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO wifi_alerts (tenant_id, ap_id, floor, type, severity, message, suggested_fix)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`, alert.TenantID, alert.APID, alert.Floor, alert.Type, alert.Severity, alert.Message, alert.SuggestedFix,
	).Scan(&alert.ID, &alert.CreatedAt)
}

// ListAlerts returns alerts with optional filtering.
func (r *WiFiRepository) ListAlerts(ctx context.Context, tenantID uuid.UUID, floor string, unresolvedOnly bool) ([]WiFiAlert, error) {
	q := `
		SELECT id, tenant_id, ap_id, floor, type, severity, message, suggested_fix,
		       is_resolved, resolved_at, resolved_by, created_at
		FROM wifi_alerts
		WHERE tenant_id = $1`
	args := []interface{}{tenantID}
	argCount := 1

	if floor != "" {
		argCount++
		q += fmt.Sprintf(" AND floor = $%d", argCount)
		args = append(args, floor)
	}
	if unresolvedOnly {
		q += ` AND is_resolved = FALSE`
	}
	q += ` ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanAlerts(rows)
}

// ResolveAlert marks an alert as resolved.
func (r *WiFiRepository) ResolveAlert(ctx context.Context, tenantID, alertID uuid.UUID, resolvedBy string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE wifi_alerts
		SET is_resolved = TRUE, resolved_at = NOW(), resolved_by = $1
		WHERE tenant_id = $2 AND id = $3
	`, resolvedBy, tenantID, alertID)
	return err
}

func scanAlerts(rows pgx.Rows) ([]WiFiAlert, error) {
	alerts := make([]WiFiAlert, 0)
	for rows.Next() {
		var a WiFiAlert
		var apID *uuid.UUID
		var resolvedAt, resolvedBy *time.Time
		var suggestedFix *string
		err := rows.Scan(
			&a.ID, &a.TenantID, &apID, &a.Floor, &a.Type, &a.Severity, &a.Message,
			&suggestedFix, &a.IsResolved, &resolvedAt, &resolvedBy, &a.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		a.APID = apID
		a.SuggestedFix = suggestedFix
		if resolvedAt != nil {
			a.ResolvedAt = resolvedAt
		}
		if resolvedBy != nil {
			s := resolvedBy.Format(time.RFC3339)
			a.ResolvedBy = &s
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// ── Floor Summary ──

// ListFloorSummaries returns all floor summaries for a tenant.
func (r *WiFiRepository) ListFloorSummaries(ctx context.Context, tenantID uuid.UUID) ([]WiFiFloorSummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, floor, ap_count, online_aps, avg_signal_dbm,
		       total_clients, active_alerts, overall_health, updated_at
		FROM wifi_floor_summary
		WHERE tenant_id = $1
		ORDER BY floor
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]WiFiFloorSummary, 0)
	for rows.Next() {
		var s WiFiFloorSummary
		var avgSig *int32
		err := rows.Scan(
			&s.ID, &s.TenantID, &s.Floor, &s.APCount, &s.OnlineAPs, &avgSig,
			&s.TotalClients, &s.ActiveAlerts, &s.OverallHealth, &s.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		s.AvgSignalDbm = avgSig
		summaries = append(summaries, s)
	}
	return summaries, rows.Err()
}

// GetFloorSummary returns a single floor summary.
func (r *WiFiRepository) GetFloorSummary(ctx context.Context, tenantID uuid.UUID, floor string) (*WiFiFloorSummary, error) {
	var s WiFiFloorSummary
	var avgSig *int32
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, floor, ap_count, online_aps, avg_signal_dbm,
		       total_clients, active_alerts, overall_health, updated_at
		FROM wifi_floor_summary
		WHERE tenant_id = $1 AND floor = $2
	`, tenantID, floor).Scan(
		&s.ID, &s.TenantID, &s.Floor, &s.APCount, &s.OnlineAPs, &avgSig,
		&s.TotalClients, &s.ActiveAlerts, &s.OverallHealth, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	s.AvgSignalDbm = avgSig
	return &s, nil
}

// PurgeOldMetrics deletes metrics older than the retention period.
func (r *WiFiRepository) PurgeOldMetrics(ctx context.Context, tenantID uuid.UUID, olderThan time.Time) (int64, error) {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM wifi_metrics
		WHERE tenant_id = $1 AND timestamp < $2
	`, tenantID, olderThan)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
