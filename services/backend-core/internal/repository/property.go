package repository

import (
	"context"
	"fmt"

	"github.com/nexus-platform/backend-core/internal/db"
	"github.com/nexus-platform/backend-core/internal/domain"
)

type PropertyRepository struct {
	pool *db.Pool
}

func NewPropertyRepository(pool *db.Pool) *PropertyRepository {
	return &PropertyRepository{pool: pool}
}

func (r *PropertyRepository) List(ctx context.Context, tenantID string) ([]domain.Property, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, name, code, address, city, country, timezone, status, settings, created_at, updated_at
		FROM properties
		WHERE tenant_id = $1
		ORDER BY name
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var props []domain.Property
	for rows.Next() {
		var p domain.Property
		var settings []byte
		if err := rows.Scan(&p.ID, &p.TenantID, &p.Name, &p.Code, &p.Address, &p.City, &p.Country, &p.Timezone, &p.Status, &settings, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		if len(settings) > 0 {
			p.Settings = mustUnmarshalJSON(settings)
		}
		props = append(props, p)
	}
	return props, rows.Err()
}

func (r *PropertyRepository) Get(ctx context.Context, tenantID, id string) (*domain.Property, error) {
	var p domain.Property
	var settings []byte
	err := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, name, code, address, city, country, timezone, status, settings, created_at, updated_at
		FROM properties
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, id).Scan(&p.ID, &p.TenantID, &p.Name, &p.Code, &p.Address, &p.City, &p.Country, &p.Timezone, &p.Status, &settings, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if len(settings) > 0 {
		p.Settings = mustUnmarshalJSON(settings)
	}
	return &p, nil
}

func (r *PropertyRepository) Create(ctx context.Context, tenantID string, p *domain.Property) error {
	return r.pool.QueryRow(ctx, `
		INSERT INTO properties (tenant_id, name, code, address, city, country, timezone, status, settings)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`, tenantID, p.Name, p.Code, p.Address, p.City, p.Country, p.Timezone, p.Status, mustMarshalJSON(p.Settings)).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PropertyRepository) Update(ctx context.Context, tenantID, id string, p *domain.Property) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE properties SET name=$1, code=$2, address=$3, city=$4, country=$5, timezone=$6, status=$7, settings=$8
		WHERE tenant_id=$9 AND id=$10
	`, p.Name, p.Code, p.Address, p.City, p.Country, p.Timezone, p.Status, mustMarshalJSON(p.Settings), tenantID, id)
	return err
}

func (r *PropertyRepository) Delete(ctx context.Context, tenantID, id string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM properties WHERE tenant_id=$1 AND id=$2`, tenantID, id)
	return err
}

func (r *PropertyRepository) CountByTenant(ctx context.Context, tenantID string) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM properties WHERE tenant_id=$1`, tenantID).Scan(&count)
	return count, err
}

func (r *PropertyRepository) Summary(ctx context.Context, tenantID string) ([]domain.PropertySummary, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT p.id, p.name, p.code, p.status, p.city, COUNT(r.id) as room_count
		FROM properties p
		LEFT JOIN rooms r ON r.property_id = p.id
		WHERE p.tenant_id = $1
		GROUP BY p.id, p.name, p.code, p.status, p.city
		ORDER BY p.name
	`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []domain.PropertySummary
	for rows.Next() {
		var s domain.PropertySummary
		if err := rows.Scan(&s.ID, &s.Name, &s.Code, &s.Status, &s.City, &s.RoomCount); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, rows.Err()
}
