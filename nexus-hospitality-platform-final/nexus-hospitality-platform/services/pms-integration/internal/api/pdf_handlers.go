package api

import (
	"net/http"
	"time"

	"github.com/nexus-platform/pms-integration/internal/middleware"
	"github.com/nexus-platform/pms-integration/internal/pdf"
)

func (h *Handler) generateInvoicePDF(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	invoiceID := chi.URLParam(r, "invoiceId")

	// Fetch invoice from store
	inv, err := h.store.Invoices.GetByID(r.Context(), tenantID, invoiceID)
	if err != nil {
		respondError(w, http.StatusNotFound, "invoice not found")
		return
	}

	// Build PDF data
	data := pdf.InvoiceData{
		InvoiceNumber:   inv.InvoiceNumber,
		InvoiceDate:     inv.InvoiceDate,
		DueDate:         inv.DueDate,
		BilledTo:        inv.CustomerName,
		BilledToAddress: inv.CustomerEmail,
		PropertyName:    "Nexus Property",
		Subtotal:        inv.Subtotal,
		TaxRate:         inv.TaxRate,
		TaxAmount:       inv.TaxAmount,
		Total:           inv.Total,
		AmountPaid:      inv.AmountPaid,
		BalanceDue:      inv.BalanceDue,
		Notes:           inv.Notes,
	}

	// Add line items
	for _, item := range inv.LineItems {
		data.Items = append(data.Items, pdf.InvoiceItem{
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			Amount:      item.Amount,
		})
	}

	gen := pdf.NewSimplePDFGenerator()
	buf, err := gen.GenerateInvoice(data)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"invoice_"+inv.InvoiceNumber+".pdf\"")
	w.Write(buf)
}

func (h *Handler) generateFolioPDF(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)
	folioID := chi.URLParam(r, "folioId")

	folio, err := h.store.Folios.GetByID(r.Context(), tenantID, folioID)
	if err != nil {
		respondError(w, http.StatusNotFound, "folio not found")
		return
	}

	// Fetch guest for folio
	guest, err := h.store.Guests.GetByID(r.Context(), tenantID, folio.GuestID)
	guestName := "Guest"
	if err == nil && guest != nil {
		guestName = guest.FirstName + " " + guest.LastName
	}

	// Fetch charges
	charges, err := h.store.Charges.ListByFolio(r.Context(), tenantID, folioID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Fetch payments
	payments, err := h.store.Payments.ListByFolio(r.Context(), tenantID, folioID)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	data := pdf.FolioStatementData{
		GuestName:  guestName,
		Balance:    folio.Balance,
		PrintedAt:  time.Now(),
		PropertyName: "Nexus Property",
	}

	for _, c := range charges {
		data.Charges = append(data.Charges, pdf.FolioCharge{
			Date:        c.CreatedAt,
			Description: c.Description,
			Amount:      c.Amount,
		})
	}

	for _, p := range payments {
		data.Payments = append(data.Payments, pdf.FolioPayment{
			Date:   p.PaymentDate,
			Method: p.PaymentMethod,
			Amount: p.Amount,
		})
	}

	gen := pdf.NewSimplePDFGenerator()
	buf, err := gen.GenerateFolioStatement(data)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"folio_"+folioID+".pdf\"")
	w.Write(buf)
}

func (h *Handler) generateNightlyReport(w http.ResponseWriter, r *http.Request) {
	tenantID := tenantID(r)

	// Calculate yesterday's date
	reportDate := time.Now().AddDate(0, 0, -1)

	// Fetch property stats
	properties, err := h.store.Properties.List(r.Context(), tenantID, 100, 0)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Calculate stats
	var totalRevenue float64
	var roomsSold, roomsAvailable int

	for _, prop := range properties {
		roomsAvailable += prop.TotalRooms
		// Fetch reservations for this property
		reservations, _ := h.store.Reservations.List(r.Context(), tenantID, "confirmed", prop.ID, 1000, 0)
		for _, res := range reservations {
			if res.CheckIn.Before(reportDate.AddDate(0, 0, 1)) && res.CheckOut.After(reportDate) {
				roomsSold++
				totalRevenue += res.TotalAmount
			}
		}
	}

	occupancyRate := 0.0
	if roomsAvailable > 0 {
		occupancyRate = float64(roomsSold) / float64(roomsAvailable)
	}

	adr := 0.0
	if roomsSold > 0 {
		adr = totalRevenue / float64(roomsSold)
	}

	revPAR := 0.0
	if roomsAvailable > 0 {
		revPAR = totalRevenue / float64(roomsAvailable)
	}

	data := pdf.NightlyReportData{
		ReportDate:     reportDate,
		PropertyName:   "Nexus Properties",
		OccupancyRate:  occupancyRate,
		ADR:            adr,
		RevPAR:         revPAR,
		RoomsSold:      roomsSold,
		RoomsAvailable: roomsAvailable,
		TotalRevenue:   totalRevenue,
	}

	gen := pdf.NewSimplePDFGenerator()
	buf, err := gen.GenerateNightlyReport(data)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"nightly_report_"+reportDate.Format("2006-01-02")+".pdf\"")
	w.Write(buf)
}
