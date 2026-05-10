package api

import (
	"net/http"

	"github.com/nexus-platform/pms-integration/internal/domain"
	"github.com/nexus-platform/pms-integration/internal/license"
)

// TenantHandler holds tenant and license endpoints.
type TenantHandler struct {
	licenseSvc *license.Service
}

// NewTenantHandler creates a new tenant handler.
func NewTenantHandler(svc *license.Service) *TenantHandler {
	return &TenantHandler{licenseSvc: svc}
}

// getTenantConfig returns the full tenant configuration with capabilities.
func (h *TenantHandler) getTenantConfig(w http.ResponseWriter, r *http.Request) {
	tenant := tenantID(r)
	cfg, err := h.licenseSvc.GetConfig(r.Context(), tenant)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, cfg)
}

// listPropertyTypes returns available property types.
func (h *TenantHandler) listPropertyTypes(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"property_types": h.licenseSvc.PropertyTypes(),
	})
}

// listLicenseTiers returns available license tiers with pricing.
func (h *TenantHandler) listLicenseTiers(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"license_tiers": h.licenseSvc.LicenseTiers(),
	})
}

// listCapabilities returns all capability definitions.
func (h *TenantHandler) listCapabilities(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"capabilities": []map[string]string{
			{"id": string(domain.CapReservations), "category": "Core", "name": "Reservations", "description": "Individual booking management"},
			{"id": string(domain.CapGuests), "category": "Core", "name": "Guest Profiles", "description": "Guest CRM and preferences"},
			{"id": string(domain.CapProperties), "category": "Core", "name": "Properties", "description": "Property and room type management"},
			{"id": string(domain.CapRooms), "category": "Core", "name": "Room Inventory", "description": "Room status and housekeeping"},
			{"id": string(domain.CapHousekeeping), "category": "Core", "name": "Housekeeping", "description": "Task management for housekeeping"},
			{"id": string(domain.CapSettings), "category": "Core", "name": "Settings", "description": "Property configuration"},
			{"id": string(domain.CapAuditLogs), "category": "Core", "name": "Audit Logs", "description": "Activity tracking and compliance"},
			{"id": string(domain.CapFloorDashboard), "category": "Operations", "name": "Floor Dashboard", "description": "Interactive floor plan with room status"},
			{"id": string(domain.CapRoomBlocks), "category": "Operations", "name": "Room Blocks", "description": "Block rooms for groups, VIP, maintenance"},
			{"id": string(domain.CapGroupReservations), "category": "Operations", "name": "Group Reservations", "description": "Group bookings with master folio"},
			{"id": string(domain.CapMaintenance), "category": "Operations", "name": "Maintenance", "description": "Engineering and maintenance management"},
			{"id": string(domain.CapFrontDeskWorkflow), "category": "Operations", "name": "Front Desk", "description": "Check-in/out workflow tools"},
			{"id": string(domain.CapDynamicPricing), "category": "Revenue", "name": "Dynamic Pricing", "description": "Automated pricing rules and recommendations"},
			{"id": string(domain.CapOTAIntegration), "category": "Revenue", "name": "OTA Integration", "description": "Channel manager for online travel agencies"},
			{"id": string(domain.CapRevenueForecast), "category": "Revenue", "name": "Revenue Forecasting", "description": "Occupancy and revenue predictions"},
			{"id": string(domain.CapAgentManagement), "category": "Revenue", "name": "Agent Management", "description": "OTA and travel agent relationships"},
			{"id": string(domain.CapMultiProperty), "category": "Enterprise", "name": "Multi-Property", "description": "Manage multiple properties in one account"},
			{"id": string(domain.CapAdvancedCRM), "category": "Enterprise", "name": "Advanced CRM", "description": "Guest history, VIP management, notes"},
			{"id": string(domain.CapAPIAccess), "category": "Enterprise", "name": "API Access", "description": "Webhooks and full API integration"},
			{"id": string(domain.CapWhiteLabel), "category": "Enterprise", "name": "White Label", "description": "Custom branding options"},
			{"id": string(domain.CapCustomReports), "category": "Enterprise", "name": "Custom Reports", "description": "Advanced analytics and reporting"},
		},
	})
}
