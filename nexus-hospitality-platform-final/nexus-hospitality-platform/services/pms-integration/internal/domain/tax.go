// Package domain provides tax-related aggregates.
package domain

import (
	"time"
)

// TaxRate represents a jurisdiction's tax configuration.
type TaxRate struct {
	ID          string
	TenantID    string
	Name        string
	Jurisdiction string
	CountryCode  string
	StateCode    string
	CityCode     string
	Rate         float64
	Type         string // vat, gst, sales_tax, tourism_tax, resort_fee
	AppliesTo    string // room, fb, service, all
	IsCompound   bool
	IsActive     bool
	EffectiveFrom time.Time
	EffectiveTo   *time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// TaxCalculation holds the result of applying tax rules to a charge.
type TaxCalculation struct {
	ChargeID        string
	BaseAmount      float64
	TaxBreakdown    []TaxLine
	TotalTax        float64
	TotalWithTax    float64
}

// TaxLine is a single tax application.
type TaxLine struct {
	TaxRateID       string
	Name            string
	Rate            float64
	TaxAmount       float64
}

// TaxExemption represents a guest or corporate tax exemption certificate.
type TaxExemption struct {
	ID             string
	TenantID       string
	GuestID        *string
	CorporateID    *string
	CertificateNum string
	Jurisdiction   string
	TaxType        string
	ExemptPercent  float64
	ValidFrom      time.Time
	ValidTo        time.Time
	IsActive       bool
	CreatedAt      time.Time
}
