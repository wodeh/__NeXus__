// Package domain defines billing and subscription aggregates for Stripe integration.
package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ==================== SUBSCRIPTION PLAN ====================

type PlanID string

type BillingInterval string

const (
	BillingMonthly  BillingInterval = "month"
	BillingYearly   BillingInterval = "year"
	BillingQuarterly BillingInterval = "quarter"
)

type SubscriptionPlan struct {
	ID                PlanID
	Slug              string
	Name              string
	Description       string
	PriceCents        int
	CurrencyCode      string
	Interval          BillingInterval
	MaxProperties     int
	MaxRooms          int
	MaxUsers          int
	Features          []string
	StripePriceID     string
	StripeProductID   string
	IsActive          bool
	IsDefault         bool
	TrialDays         int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewSubscriptionPlan(slug, name string, priceCents int, interval BillingInterval) *SubscriptionPlan {
	now := time.Now().UTC()
	return &SubscriptionPlan{
		ID:           PlanID(uuid.Must(uuid.NewRandom()).String()),
		Slug:         slug,
		Name:         name,
		PriceCents:   priceCents,
		CurrencyCode: "USD",
		Interval:     interval,
		MaxProperties: 1,
		MaxRooms:      50,
		MaxUsers:      5,
		IsActive:     true,
		TrialDays:    14,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func (p *SubscriptionPlan) PriceDollars() float64 {
	return float64(p.PriceCents) / 100.0
}

// ==================== SUBSCRIPTION ====================

type SubscriptionID string

type SubscriptionStatus string

const (
	SubscriptionStatusActive      SubscriptionStatus = "active"
	SubscriptionStatusTrialing    SubscriptionStatus = "trialing"
	SubscriptionStatusPastDue     SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled    SubscriptionStatus = "canceled"
	SubscriptionStatusUnpaid      SubscriptionStatus = "unpaid"
	SubscriptionStatusIncomplete  SubscriptionStatus = "incomplete"
)

type Subscription struct {
	ID                   SubscriptionID
	TenantID             string
	PlanID               string
	Status               SubscriptionStatus
	StripeSubscriptionID string
	StripeCustomerID     string
	CurrentPeriodStart   time.Time
	CurrentPeriodEnd     time.Time
	TrialStart           *time.Time
	TrialEnd             *time.Time
	CanceledAt           *time.Time
	CancelAtPeriodEnd    bool
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func NewSubscription(tenantID, planID, stripeSubID, stripeCustID string) *Subscription {
	now := time.Now().UTC()
	return &Subscription{
		ID:                   SubscriptionID(uuid.Must(uuid.NewRandom()).String()),
		TenantID:             tenantID,
		PlanID:               planID,
		Status:               SubscriptionStatusTrialing,
		StripeSubscriptionID: stripeSubID,
		StripeCustomerID:     stripeCustID,
		CurrentPeriodStart:   now,
		CurrentPeriodEnd:     now.AddDate(0, 1, 0),
		TrialStart:           &now,
		TrialEnd:             func() *time.Time { t := now.AddDate(0, 0, 14); return &t }(),
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

func (s *Subscription) IsActive() bool {
	return s.Status == SubscriptionStatusActive || s.Status == SubscriptionStatusTrialing
}

func (s *Subscription) IsTrialing() bool {
	return s.Status == SubscriptionStatusTrialing
}

func (s *Subscription) IsCanceled() bool {
	return s.CanceledAt != nil || s.CancelAtPeriodEnd
}

// ==================== INVOICE ====================

type InvoiceID string

type InvoiceStatus string

const (
	InvoiceStatusDraft      InvoiceStatus = "draft"
	InvoiceStatusOpen       InvoiceStatus = "open"
	InvoiceStatusPaid       InvoiceStatus = "paid"
	InvoiceStatusUncollectible InvoiceStatus = "uncollectible"
	InvoiceStatusVoid       InvoiceStatus = "void"
)

type Invoice struct {
	ID              InvoiceID
	TenantID        string
	SubscriptionID  string
	StripeInvoiceID string
	Status          InvoiceStatus
	AmountDueCents  int
	AmountPaidCents int
	CurrencyCode    string
	PeriodStart     time.Time
	PeriodEnd       time.Time
	DueDate         *time.Time
	PaidAt          *time.Time
	PdfUrl          string
	CreatedAt       time.Time
}

func (i *Invoice) AmountDueDollars() float64 {
	return float64(i.AmountDueCents) / 100.0
}

// ==================== PAYMENT METHOD ====================

type PaymentMethodID string

type PaymentMethodType string

const (
	PaymentMethodCard      PaymentMethodType = "card"
	PaymentMethodBankTransfer PaymentMethodType = "bank_transfer"
)

type PaymentMethod struct {
	ID             PaymentMethodID
	TenantID       string
	StripePaymentMethodID string
	Type           PaymentMethodType
	Brand          string
	Last4          string
	ExpMonth       int
	ExpYear        int
	IsDefault      bool
	CreatedAt      time.Time
}
