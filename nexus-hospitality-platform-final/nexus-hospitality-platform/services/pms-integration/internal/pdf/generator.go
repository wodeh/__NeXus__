// Package pdf provides PDF generation for hospitality documents.
// Uses gofpdf-style approach (text-based layout for invoices/reports).
package pdf

import (
	"fmt"
	"time"
)

// InvoiceData holds all fields needed for an invoice PDF.
type InvoiceData struct {
	InvoiceNumber   string
	InvoiceDate     time.Time
	DueDate         time.Time
	BilledTo        string
	BilledToAddress string
	PropertyName    string
	PropertyAddress string
	Items           []InvoiceItem
	Subtotal        float64
	TaxRate         float64
	TaxAmount       float64
	Total           float64
	AmountPaid      float64
	BalanceDue      float64
	Notes           string
}

// InvoiceItem is a line item on an invoice.
type InvoiceItem struct {
	Description string
	Quantity    int
	UnitPrice   float64
	Amount      float64
}

// FolioStatementData holds data for a guest folio PDF.
type FolioStatementData struct {
	GuestName       string
	RoomNumber      string
	CheckIn         time.Time
	CheckOut        time.Time
	ConfirmationNum string
	Charges         []FolioCharge
	Payments        []FolioPayment
	Balance         float64
	PropertyName    string
	PropertyAddress string
	PrintedAt       time.Time
}

// FolioCharge is a charge line on a folio.
type FolioCharge struct {
	Date        time.Time
	Description string
	Amount      float64
}

// FolioPayment is a payment line on a folio.
type FolioPayment struct {
	Date   time.Time
	Method string
	Amount float64
}

// NightlyReportData holds data for the nightly operations report.
type NightlyReportData struct {
	ReportDate       time.Time
	PropertyName     string
	OccupancyRate    float64
	ADR              float64 // Average Daily Rate
	RevPAR           float64 // Revenue Per Available Room
	RoomsSold        int
	RoomsAvailable   int
	TotalRevenue     float64
	Arrivals         int
	Departures       int
	WalkIns          int
	NoShows          int
	Cancellations    int
	HousekeepingDone int
	MaintenanceOpen  int
}

// PDFGenerator defines the interface for PDF document generation.
type PDFGenerator interface {
	GenerateInvoice(data InvoiceData) ([]byte, error)
	GenerateFolioStatement(data FolioStatementData) ([]byte, error)
	GenerateNightlyReport(data NightlyReportData) ([]byte, error)
}

// SimplePDFGenerator implements PDFGenerator using basic text layout.
// In production, swap in a proper PDF library (gofpdf, unidoc, or wkhtmltopdf).
type SimplePDFGenerator struct{}

// NewSimplePDFGenerator creates a new simple PDF generator.
func NewSimplePDFGenerator() *SimplePDFGenerator {
	return &SimplePDFGenerator{}
}

// GenerateInvoice creates a simple text-based invoice.
func (g *SimplePDFGenerator) GenerateInvoice(data InvoiceData) ([]byte, error) {
	// Placeholder: returns a structured text representation.
	// Production: integrate gofpdf or call wkhtmltopdf with HTML template.
	var buf []byte
	buf = append(buf, []byte(fmt.Sprintf("INVOICE %s\n", data.InvoiceNumber))...)
	buf = append(buf, []byte(fmt.Sprintf("Date: %s\n", data.InvoiceDate.Format("2006-01-02"))...)
	buf = append(buf, []byte(fmt.Sprintf("Due: %s\n\n", data.DueDate.Format("2006-01-02"))...))
	buf = append(buf, []byte(fmt.Sprintf("Billed To:\n%s\n%s\n\n", data.BilledTo, data.BilledToAddress))...)
	buf = append(buf, []byte(fmt.Sprintf("From:\n%s\n%s\n\n", data.PropertyName, data.PropertyAddress))...)
	buf = append(buf, []byte("Items:\n")...)
	for _, item := range data.Items {
		buf = append(buf, []byte(fmt.Sprintf("  %s x%d @ %.2f = %.2f\n", item.Description, item.Quantity, item.UnitPrice, item.Amount))...)
	}
	buf = append(buf, []byte(fmt.Sprintf("\nSubtotal: %.2f\n", data.Subtotal))...)
	buf = append(buf, []byte(fmt.Sprintf("Tax (%.1f%%): %.2f\n", data.TaxRate*100, data.TaxAmount))...)
	buf = append(buf, []byte(fmt.Sprintf("Total: %.2f\n", data.Total))...)
	buf = append(buf, []byte(fmt.Sprintf("Paid: %.2f\n", data.AmountPaid))...)
	buf = append(buf, []byte(fmt.Sprintf("Balance Due: %.2f\n", data.BalanceDue))...)
	if data.Notes != "" {
		buf = append(buf, []byte(fmt.Sprintf("\nNotes: %s\n", data.Notes))...)
	}
	return buf, nil
}

// GenerateFolioStatement creates a guest folio statement.
func (g *SimplePDFGenerator) GenerateFolioStatement(data FolioStatementData) ([]byte, error) {
	var buf []byte
	buf = append(buf, []byte(fmt.Sprintf("FOLIO STATEMENT\n"))...)
	buf = append(buf, []byte(fmt.Sprintf("Guest: %s\n", data.GuestName))...)
	buf = append(buf, []byte(fmt.Sprintf("Room: %s\n", data.RoomNumber))...)
	buf = append(buf, []byte(fmt.Sprintf("Check-in: %s\n", data.CheckIn.Format("2006-01-02")))...)
	buf = append(buf, []byte(fmt.Sprintf("Check-out: %s\n", data.CheckOut.Format("2006-01-02")))...)
	buf = append(buf, []byte(fmt.Sprintf("Confirmation: %s\n\n", data.ConfirmationNum))...)
	buf = append(buf, []byte(fmt.Sprintf("Property: %s\n", data.PropertyName))...)
	buf = append(buf, []byte("\n--- CHARGES ---\n")...)
	for _, c := range data.Charges {
		buf = append(buf, []byte(fmt.Sprintf("%s  %-30s  %.2f\n", c.Date.Format("2006-01-02"), c.Description, c.Amount))...)
	}
	buf = append(buf, []byte("\n--- PAYMENTS ---\n")...)
	for _, p := range data.Payments {
		buf = append(buf, []byte(fmt.Sprintf("%s  %-20s  %.2f\n", p.Date.Format("2006-01-02"), p.Method, p.Amount))...)
	}
	buf = append(buf, []byte(fmt.Sprintf("\nBALANCE: %.2f\n", data.Balance))...)
	buf = append(buf, []byte(fmt.Sprintf("Printed: %s\n", data.PrintedAt.Format("2006-01-02 15:04")))...)
	return buf, nil
}

// GenerateNightlyReport creates a nightly operations report.
func (g *SimplePDFGenerator) GenerateNightlyReport(data NightlyReportData) ([]byte, error) {
	var buf []byte
	buf = append(buf, []byte(fmt.Sprintf("NIGHTLY OPERATIONS REPORT\n"))...)
	buf = append(buf, []byte(fmt.Sprintf("Property: %s\n", data.PropertyName))...)
	buf = append(buf, []byte(fmt.Sprintf("Date: %s\n\n", data.ReportDate.Format("2006-01-02")))...)
	buf = append(buf, []byte("--- OCCUPANCY ---\n")...)
	buf = append(buf, []byte(fmt.Sprintf("Occupancy Rate: %.1f%%\n", data.OccupancyRate*100))...)
	buf = append(buf, []byte(fmt.Sprintf("ADR: $%.2f\n", data.ADR))...)
	buf = append(buf, []byte(fmt.Sprintf("RevPAR: $%.2f\n", data.RevPAR))...)
	buf = append(buf, []byte(fmt.Sprintf("Rooms Sold: %d / %d\n\n", data.RoomsSold, data.RoomsAvailable))...)
	buf = append(buf, []byte("--- REVENUE ---\n")...)
	buf = append(buf, []byte(fmt.Sprintf("Total Revenue: $%.2f\n\n", data.TotalRevenue))...)
	buf = append(buf, []byte("--- ACTIVITY ---\n")...)
	buf = append(buf, []byte(fmt.Sprintf("Arrivals: %d\n", data.Arrivals))...)
	buf = append(buf, []byte(fmt.Sprintf("Departures: %d\n", data.Departures))...)
	buf = append(buf, []byte(fmt.Sprintf("Walk-ins: %d\n", data.WalkIns))...)
	buf = append(buf, []byte(fmt.Sprintf("No-shows: %d\n", data.NoShows))...)
	buf = append(buf, []byte(fmt.Sprintf("Cancellations: %d\n", data.Cancellations))...)
	buf = append(buf, []byte(fmt.Sprintf("\nHousekeeping Completed: %d\n", data.HousekeepingDone))...)
	buf = append(buf, []byte(fmt.Sprintf("Open Maintenance: %d\n", data.MaintenanceOpen))...)
	return buf, nil
}
