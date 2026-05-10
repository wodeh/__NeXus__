// Package repository provides tenant-aware data access for PMS integration.
package repository

import (
	"context"
	"fmt"

	"github.com/nexus-platform/pms-integration/internal/db"
)

// Store is the root repository aggregating all PMS repositories.
type Store struct {
	pool                  *db.Pool
	Reservations          ReservationRepository
	Guests                GuestRepository
	Folios                FolioRepository
	Properties            PropertyRepository
	RoomTypes             RoomTypeRepository
	Rooms                 RoomRepository
	AuditLogs             AuditLogRepository
	RevenueForecasts      RevenueForecastRepository
	DynamicPricingRules   DynamicPricingRuleRepository
	PriceRecommendations  PriceRecommendationRepository
	MobileDevices         MobileDeviceRepository
	GuestSelfServices     GuestSelfServiceRepository
	RoomBlocks            *RoomBlockRepository
	GroupReservations     *GroupReservationRepository
	GuestProfiles         *GuestProfileRepository
	Agents                *AgentRepository
	RatePlans             *RatePlanRepository
	CheckIns              *CheckInRepository
	Invoices              *InvoiceRepository
	IPTV                  *IPTVRepository
	SmartLocks            *SmartLockRepository
	Search                SearchIndexer
	metrics               *RepositoryMetrics
}

// NewStore creates a repository store backed by a PostgreSQL pool.
// It runs migrations automatically on startup.
func NewStore(pool *db.Pool) *Store {
	m := NewRepositoryMetrics()
	s := &Store{
		pool:                 pool,
		Reservations:         NewReservationRepository(pool, m),
		Guests:               NewGuestRepository(pool, m),
		Folios:               NewFolioRepository(pool, m),
		Properties:           NewPropertyRepository(pool, m),
		RoomTypes:            NewRoomTypeRepository(pool, m),
		Rooms:                NewRoomRepository(pool, m),
		AuditLogs:            NewAuditLogRepository(pool, m),
		RevenueForecasts:     NewRevenueForecastRepository(pool, m),
		DynamicPricingRules:  NewDynamicPricingRuleRepository(pool, m),
		PriceRecommendations: NewPriceRecommendationRepository(pool, m),
		MobileDevices:        NewMobileDeviceRepository(pool, m),
		GuestSelfServices:    NewGuestSelfServiceRepository(pool, m),
		RoomBlocks:           NewRoomBlockRepository(),
		GroupReservations:    NewGroupReservationRepository(),
		GuestProfiles:        NewGuestProfileRepository(),
		Agents:               NewAgentRepository(),
		RatePlans:            NewRatePlanRepository(),
		CheckIns:             NewCheckInRepository(),
		Invoices:             NewInvoiceRepository(),
		IPTV:                 NewIPTVRepository(pool, m),
		SmartLocks:           NewSmartLockRepository(pool, m),
		Search:               NewInMemorySearchIndexer(),
		metrics:              m,
	}
	return s
}

// RunMigrations executes all pending database migrations.
func (s *Store) RunMigrations(ctx context.Context) error {
	migrator := db.NewMigrator(s.pool)
	migrations, err := db.LoadMigrationsFromDir("internal/db/migrations")
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}
	return migrator.Up(ctx, migrations)
}

// Ping verifies database connectivity.
func (s *Store) Ping(ctx context.Context) error {
	if s.pool == nil {
		return fmt.Errorf("pool not initialized")
	}
	return s.pool.Ping(ctx)
}

// Close releases repository resources.
func (s *Store) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}
