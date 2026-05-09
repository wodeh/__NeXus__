// Package domain defines document storage aggregates.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// DocumentID is a unique document identifier.
type DocumentID string

// DocumentType categorizes stored documents.
type DocumentType string

const (
	DocumentTypeGuestID       DocumentType = "guest_id"
	DocumentTypeContract      DocumentType = "contract"
	DocumentTypeInvoice       DocumentType = "invoice"
	DocumentTypeReceipt       DocumentType = "receipt"
	DocumentTypeHousekeeping  DocumentType = "housekeeping_photo"
	DocumentTypeMaintenance   DocumentType = "maintenance_photo"
	DocumentTypeRateSheet     DocumentType = "rate_sheet"
	DocumentTypeGeneral       DocumentType = "general"
)

// Document represents a file stored in S3 with metadata in Postgres.
type Document struct {
	ID           DocumentID
	TenantID     string
	EntityType   string      // "guest", "reservation", "folio", "property", "work_order"
	EntityID     string
	Type         DocumentType
	FileName     string
	ContentType  string
	SizeBytes    int64
	S3Bucket     string
	S3Key        string
	S3URL        string
	UploadedBy   string
	IsPublic     bool
	Metadata     map[string]string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewDocument(tenantID, entityType, entityID string, docType DocumentType, fileName, contentType string, sizeBytes int64, uploadedBy string) *Document {
	now := time.Now().UTC()
	return &Document{
		ID:         DocumentID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:   tenantID,
		EntityType: entityType,
		EntityID:   entityID,
		Type:       docType,
		FileName:   fileName,
		ContentType: contentType,
		SizeBytes:  sizeBytes,
		UploadedBy: uploadedBy,
		IsPublic:   false,
		Metadata:   make(map[string]string),
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// GenerateS3Key creates an S3 object key from document metadata.
func (d *Document) GenerateS3Key() string {
	return "tenants/" + d.TenantID + "/" + string(d.Type) + "/" + string(d.ID) + "/" + d.FileName
}
