// Package api provides Stripe webhook and billing REST handlers.
package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/nexus-platform/pms-integration/internal/domain"
)

var stripeWebhookSecret = os.Getenv("STRIPE_WEBHOOK_SECRET")

func (h *Handler) stripeWebhook(w http.ResponseWriter, r *http.Request) {
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	// In production, verify Stripe signature using stripe-webhook-secret
	// For now, we parse the event directly
	var event struct {
		Type    string `json:"type"`
		Data    struct {
			Object json.RawMessage `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		respondError(w, http.StatusBadRequest, "invalid json")
		return
	}

	switch event.Type {
	case "invoice.payment_succeeded":
		var invoice struct {
			ID           string `json:"id"`
			Customer     string `json:"customer"`
			Subscription string `json:"subscription"`
			AmountDue    int    `json:"amount_due"`
			AmountPaid   int    `json:"amount_paid"`
			Currency     string `json:"currency"`
			PeriodStart  int64  `json:"period_start"`
			PeriodEnd    int64  `json:"period_end"`
			Pdf          string `json:"invoice_pdf"`
			Status       string `json:"status"`
		}
		_ = json.Unmarshal(event.Data.Object, &invoice)
		_ = h.handleInvoicePaymentSucceeded(r.Context(), invoice)

	case "invoice.payment_failed":
		var invoice struct {
			ID           string `json:"id"`
			Customer     string `json:"customer"`
			Subscription string `json:"subscription"`
		}
		_ = json.Unmarshal(event.Data.Object, &invoice)
		_ = h.handleInvoicePaymentFailed(r.Context(), invoice)

	case "customer.subscription.updated":
		var sub struct {
			ID               string `json:"id"`
			Customer         string `json:"customer"`
			Status           string `json:"status"`
			CurrentPeriodEnd int64  `json:"current_period_end"`
			CancelAtPeriodEnd bool  `json:"cancel_at_period_end"`
		}
		_ = json.Unmarshal(event.Data.Object, &sub)
		_ = h.handleSubscriptionUpdated(r.Context(), sub)

	case "customer.subscription.deleted":
		var sub struct {
			ID       string `json:"id"`
			Customer string `json:"customer"`
		}
		_ = json.Unmarshal(event.Data.Object, &sub)
		_ = h.handleSubscriptionDeleted(r.Context(), sub)
	}

	respondJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleInvoicePaymentSucceeded(ctx context.Context, invoice struct {
	ID           string
	Customer     string
	Subscription string
	AmountDue    int
	AmountPaid   int
	Currency     string
	PeriodStart  int64
	PeriodEnd    int64
	Pdf          string
	Status       string
}) error {
	// Find subscription by stripe subscription ID
	sub, err := h.store.Subscriptions.GetByStripeSubscriptionID(ctx, invoice.Subscription)
	if err != nil {
		return err
	}

	i := &domain.Invoice{
		ID:             domain.InvoiceID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:       sub.TenantID,
		SubscriptionID: string(sub.ID),
		StripeInvoiceID: invoice.ID,
		Status:         domain.InvoiceStatusPaid,
		AmountDueCents:  invoice.AmountDue,
		AmountPaidCents: invoice.AmountPaid,
		CurrencyCode:    invoice.Currency,
		PeriodStart:     time.Unix(invoice.PeriodStart, 0).UTC(),
		PeriodEnd:       time.Unix(invoice.PeriodEnd, 0).UTC(),
		PaidAt:          func() *time.Time { t := time.Now().UTC(); return &t }(),
		PdfUrl:          invoice.Pdf,
		CreatedAt:       time.Now().UTC(),
	}
	return h.store.Invoices.Create(ctx, i)
}

func (h *Handler) handleInvoicePaymentFailed(ctx context.Context, invoice struct {
	ID           string
	Customer     string
	Subscription string
}) error {
	sub, err := h.store.Subscriptions.GetByStripeSubscriptionID(ctx, invoice.Subscription)
	if err != nil {
		return err
	}
	sub.Status = domain.SubscriptionStatusPastDue
	return h.store.Subscriptions.Update(ctx, sub)
}

func (h *Handler) handleSubscriptionUpdated(ctx context.Context, sub struct {
	ID               string
	Customer         string
	Status           string
	CurrentPeriodEnd int64
	CancelAtPeriodEnd bool
}) error {
	subscription, err := h.store.Subscriptions.GetByStripeSubscriptionID(ctx, sub.ID)
	if err != nil {
		return err
	}

	switch sub.Status {
	case "active":
		subscription.Status = domain.SubscriptionStatusActive
	case "trialing":
		subscription.Status = domain.SubscriptionStatusTrialing
	case "past_due":
		subscription.Status = domain.SubscriptionStatusPastDue
	case "canceled", "unpaid":
		subscription.Status = domain.SubscriptionStatusCanceled
	}

	subscription.CurrentPeriodEnd = time.Unix(sub.CurrentPeriodEnd, 0).UTC()
	subscription.CancelAtPeriodEnd = sub.CancelAtPeriodEnd
	subscription.UpdatedAt = time.Now().UTC()

	// Update tenant limits based on plan
	if subscription.Status == domain.SubscriptionStatusActive {
		plan, _ := h.store.SubscriptionPlans.GetByID(ctx, subscription.PlanID)
		if plan != nil {
			tenant, _ := h.store.Tenants.GetByID(ctx, subscription.TenantID)
			if tenant != nil {
				tenant.MaxProperties = plan.MaxProperties
				tenant.MaxRooms = plan.MaxRooms
				tenant.MaxUsers = plan.MaxUsers
				tenant.Status = domain.TenantStatusActive
				_ = h.store.Tenants.Update(ctx, tenant)
			}
		}
	}

	return h.store.Subscriptions.Update(ctx, subscription)
}

func (h *Handler) handleSubscriptionDeleted(ctx context.Context, sub struct {
	ID       string
	Customer string
}) error {
	subscription, err := h.store.Subscriptions.GetByStripeSubscriptionID(ctx, sub.ID)
	if err != nil {
		return err
	}
	subscription.Status = domain.SubscriptionStatusCanceled
	now := time.Now().UTC()
	subscription.CanceledAt = &now
	subscription.UpdatedAt = now

	// Suspend tenant
	tenant, _ := h.store.Tenants.GetByID(ctx, subscription.TenantID)
	if tenant != nil {
		tenant.Status = domain.TenantStatusSuspended
		_ = h.store.Tenants.Update(ctx, tenant)
	}

	return h.store.Subscriptions.Update(ctx, subscription)
}

// Billing REST handlers

func (h *Handler) listPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.store.SubscriptionPlans.ListActive(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, plans)
}

func (h *Handler) getSubscription(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	sub, err := h.store.Subscriptions.GetByTenant(r.Context(), tenantID)
	if err != nil {
		respondError(w, http.StatusNotFound, "no subscription found")
		return
	}
	respondJSON(w, http.StatusOK, sub)
}

func (h *Handler) listInvoices(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 { limit = 50 }
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	invoices, err := h.store.Invoices.ListByTenant(r.Context(), tenantID, limit, offset)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, invoices)
}

// Helper — needed for Stripe subscription lookup
type stripeSubscriptionLookup interface {
	GetByStripeSubscriptionID(ctx context.Context, stripeID string) (*domain.Subscription, error)
}

// Add this method to PostgresSubscriptionRepository
func (r *PostgresSubscriptionRepository) GetByStripeSubscriptionID(ctx context.Context, stripeID string) (*domain.Subscription, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, tenant_id, plan_id, status, stripe_subscription_id, stripe_customer_id,
		       current_period_start, current_period_end, trial_start, trial_end,
		       canceled_at, cancel_at_period_end, created_at, updated_at
		FROM subscriptions WHERE stripe_subscription_id = $1
	`, stripeID)
	return scanSubscription(row)
}
