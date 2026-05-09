// Package repository provides tenant-aware data access for documents.
package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/nexus-platform/pms-integration/internal/db"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

// DocumentRepository persists document metadata.
type DocumentRepository interface {
	Create(ctx context.Context, doc *domain.Document) error
	GetByID(ctx context.Context, tenantID, id string) (*domain.Document, error)
	ListByEntity(ctx context.Context, tenantID, entityType, entityID string, limit, offset int) ([]*domain.Document, error)
	ListByType(ctx context.Context, tenantID string, docType domain.DocumentType, limit, offset int) ([]*domain.Document, error)
	Delete(ctx context.Context, tenantID, id string) error
	UpdateS3Info(ctx context.Context, tenantID, id, bucket, key, url string) error
}

type PostgresDocumentRepository struct {
	pool    *db.Pool
	metrics *RepositoryMetrics
}

func NewDocumentRepository(pool *db.Pool, metrics *RepositoryMetrics) *PostgresDocumentRepository {
	return &PostgresDocumentRepository{pool: pool, metrics: metrics}
}

func (r *PostgresDocumentRepository) Create(ctx context.Context, doc *domain.Document) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO documents (id, tenant_id, entity_type, entity_id, type, file_name, content_type, size_bytes, s3_bucket, s3_key, s3_url, uploaded_by, is_public, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`, doc.ID, doc.TenantID, doc.EntityType, doc.EntityID, doc.Type, doc.FileName, doc.ContentType, doc.SizeBytes,
		doc.S3Bucket, doc.S3Key, doc.S3URL, doc.UploadedBy, doc.IsPublic, doc.Metadata, doc.CreatedAt, doc.UpdatedAt)
	return err
}

func (r *PostgresDocumentRepository) GetByID(ctx context.Context, tenantID, id string) (*domain.Document, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, entity_type, entity_id, type, file_name, content_type, size_bytes, s3_bucket, s3_key, s3_url, uploaded_by, is_public, metadata, created_at, updated_at
		FROM documents
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, id)
	return scanDocument(row)
}

func (r *PostgresDocumentRepository) ListByEntity(ctx context.Context, tenantID, entityType, entityID string, limit, offset int) ([]*domain.Document, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, entity_type, entity_id, type, file_name, content_type, size_bytes, s3_bucket, s3_key, s3_url, uploaded_by, is_public, metadata, created_at, updated_at
		FROM documents
		WHERE tenant_id = $1 AND entity_type = $2 AND entity_id = $3
		ORDER BY created_at DESC
		LIMIT $4 OFFSET $5
	`, tenantID, entityType, entityID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDocuments(rows)
}

func (r *PostgresDocumentRepository) ListByType(ctx context.Context, tenantID string, docType domain.DocumentType, limit, offset int) ([]*domain.Document, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, tenant_id, entity_type, entity_id, type, file_name, content_type, size_bytes, s3_bucket, s3_key, s3_url, uploaded_by, is_public, metadata, created_at, updated_at
		FROM documents
		WHERE tenant_id = $1 AND type = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, tenantID, docType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDocuments(rows)
}

func (r *PostgresDocumentRepository) Delete(ctx context.Context, tenantID, id string) error {
	_, err := r.pool.Exec(ctx, `
		DELETE FROM documents
		WHERE tenant_id = $1 AND id = $2
	`, tenantID, id)
	return err
}

func (r *PostgresDocumentRepository) UpdateS3Info(ctx context.Context, tenantID, id, bucket, key, url string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE documents
		SET s3_bucket = $1, s3_key = $2, s3_url = $3, updated_at = NOW()
		WHERE tenant_id = $4 AND id = $5
	`, bucket, key, url, tenantID, id)
	return err
}

func scanDocument(row pgx.Row) (*domain.Document, error) {
	var d domain.Document
	var metadata map[string]string
	err := row.Scan(
		&d.ID, &d.TenantID, &d.EntityType, &d.EntityID, &d.Type, &d.FileName, &d.ContentType,
		&d.SizeBytes, &d.S3Bucket, &d.S3Key, &d.S3URL, &d.UploadedBy, &d.IsPublic,
		&metadata, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	d.Metadata = metadata
	return &d, nil
}

func scanDocuments(rows pgx.Rows) ([]*domain.Document, error) {
	var docs []*domain.Document
	for rows.Next() {
		var d domain.Document
		var metadata map[string]string
		err := rows.Scan(
			&d.ID, &d.TenantID, &d.EntityType, &d.EntityID, &d.Type, &d.FileName, &d.ContentType,
			&d.SizeBytes, &d.S3Bucket, &d.S3Key, &d.S3URL, &d.UploadedBy, &d.IsPublic,
			&metadata, &d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		d.Metadata = metadata
		docs = append(docs, &d)
	}
	return docs, rows.Err()
}
