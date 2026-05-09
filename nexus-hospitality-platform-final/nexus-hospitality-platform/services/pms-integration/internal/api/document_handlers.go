// Package api provides document REST handlers.
package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nexus-platform/pms-integration/internal/domain"
	"github.com/nexus-platform/pms-integration/internal/middleware"
)

func (h *Handler) createDocument(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	var req struct {
		EntityType  string `json:"entity_type"`
		EntityID    string `json:"entity_id"`
		Type        string `json:"type"`
		FileName    string `json:"file_name"`
		ContentType string `json:"content_type"`
		SizeBytes   int64  `json:"size_bytes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var docType domain.DocumentType
	switch req.Type {
	case "guest_id":
		docType = domain.DocumentTypeGuestID
	case "contract":
		docType = domain.DocumentTypeContract
	case "invoice":
		docType = domain.DocumentTypeInvoice
	case "receipt":
		docType = domain.DocumentTypeReceipt
	case "housekeeping_photo":
		docType = domain.DocumentTypeHousekeeping
	case "maintenance_photo":
		docType = domain.DocumentTypeMaintenance
	case "rate_sheet":
		docType = domain.DocumentTypeRateSheet
	default:
		docType = domain.DocumentTypeGeneral
	}

	claims := middleware.ClaimsFromContext(r.Context())
	uploadedBy := ""
	if claims != nil {
		uploadedBy = claims.UserID
	}

	doc := domain.NewDocument(tenantID, req.EntityType, req.EntityID, docType, req.FileName, req.ContentType, req.SizeBytes, uploadedBy)
	doc.S3Key = doc.GenerateS3Key()

	if err := h.store.Documents.Create(r.Context(), doc); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, doc)
}

func (h *Handler) getDocument(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	docID := chi.URLParam(r, "documentId")

	doc, err := h.store.Documents.GetByID(r.Context(), tenantID, docID)
	if err != nil {
		respondError(w, http.StatusNotFound, "document not found")
		return
	}

	respondJSON(w, http.StatusOK, doc)
}

func (h *Handler) listDocumentsByEntity(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	entityType := chi.URLParam(r, "entityType")
	entityID := chi.URLParam(r, "entityId")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 { limit = 50 }
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	docs, err := h.store.Documents.ListByEntity(r.Context(), tenantID, entityType, entityID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, docs)
}

func (h *Handler) deleteDocument(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	docID := chi.URLParam(r, "documentId")

	if err := h.store.Documents.Delete(r.Context(), tenantID, docID); err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}
