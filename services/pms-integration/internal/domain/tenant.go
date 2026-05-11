package domain

import "time"

// PropertyType defines the kind of hospitality property.
type PropertyType string

const (
	PropertyTypeBoutique   PropertyType = "boutique"
	PropertyTypeMotel      PropertyType = "motel"
	PropertyTypeResort     PropertyType = "resort"
	PropertyTypeHostel     PropertyType = "hostel"
	PropertyTypeAparthotel PropertyType = "aparthotel"
	PropertyTypeBnB        PropertyType = "bnb"
)

// LicenseTier defines the subscription level.
type LicenseTier string

const (
	LicenseTierCore       LicenseTier = "core"
	LicenseTierOperations LicenseTier = "operations"
	LicenseTierRevenue    LicenseTier = "revenue"
	LicenseTierEnterprise LicenseTier = "enterprise"
)

// LicenseStatus defines the current state of a subscription.
type LicenseStatus string

const (
	LicenseStatusActive    LicenseStatus = "active"
	LicenseStatusTrial     LicenseStatus = "trial"
	LicenseStatusExpired   LicenseStatus = "expired"
	LicenseStatusSuspended LicenseStatus = "suspended"
)

// Capability is a granular feature flag.
type Capability string

// Core capabilities (available in all tiers).
const (
	CapReservations  Capability = "core:reservations"
	CapGuests        Capability = "core:guests"
	CapProperties    Capability = "core:properties"
	CapRooms         Capability = "core:rooms"
	CapHousekeeping  Capability = "core:housekeeping"
	CapSettings      Capability = "core:settings"
	CapAuditLogs     Capability = "core:audit_logs"
)

// Operations capabilities.
const (
	CapFloorDashboard     Capability = "operations:floor_dashboard"
	CapRoomBlocks         Capability = "operations:room_blocks"
	CapGroupReservations  Capability = "operations:group_reservations"
	CapMaintenance        Capability = "operations:maintenance"
	CapFrontDeskWorkflow  Capability = "operations:front_desk"
	CapIPTVBasic          Capability = "operations:iptv_basic"
	CapSmartLocks         Capability = "operations:smart_locks"
	CapRemoteUnlock       Capability = "operations:remote_unlock"
)

// Revenue capabilities.
const (
	CapDynamicPricing    Capability = "revenue:dynamic_pricing"
	CapOTAIntegration    Capability = "revenue:ota_integration"
	CapRevenueForecast   Capability = "revenue:revenue_forecasting"
	CapAgentManagement   Capability = "revenue:agent_management"
	CapIPTVPremium       Capability = "revenue:iptv_premium"
	CapIPTVWelcome       Capability = "revenue:iptv_welcome"
	CapIPTVContent       Capability = "revenue:iptv_content"
	CapChannelManager   Capability = "revenue:channel_manager"
	CapWhatsAppBot      Capability = "revenue:whatsapp_bot"
	CapDirectBooking    Capability = "revenue:direct_booking"
	CapGuestReviews     Capability = "revenue:guest_reviews"
	CapCommunications   Capability = "revenue:communications"

)

// Enterprise capabilities.
const (
	CapMultiProperty    Capability = "enterprise:multi_property"
	CapAdvancedCRM      Capability = "enterprise:advanced_crm"
	CapAPIAccess        Capability = "enterprise:api_access"
	CapWhiteLabel       Capability = "enterprise:white_label"
	CapCustomReports    Capability = "enterprise:custom_reports"
)

// TenantConfig holds license and property configuration.
type TenantConfig struct {
	ID             string        `json:"id"`
	Name           string        `json:"name"`
	PropertyType   PropertyType  `json:"property_type"`
	LicenseTier    LicenseTier   `json:"license_tier"`
	LicenseStatus  LicenseStatus `json:"license_status"`
	LicenseExpires time.Time     `json:"license_expires_at"`
	MaxRooms       int           `json:"max_rooms"`
	MaxUsers       int           `json:"max_users"`
	Capabilities   []Capability  `json:"capabilities"`
	Settings       TenantSettings `json:"settings"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// TenantSettings holds per-tenant preferences.
type TenantSettings struct {
	Timezone     string `json:"timezone"`
	CurrencyCode string `json:"currency_code"`
	DateFormat   string `json:"date_format"`
	Language     string `json:"language"`
}

// NewTenantConfig creates a tenant with default capabilities derived from tier.
func NewTenantConfig(id, name string, pt PropertyType, tier LicenseTier) *TenantConfig {
	now := time.Now()
	tc := &TenantConfig{
		ID:            id,
		Name:          name,
		PropertyType:  pt,
		LicenseTier:   tier,
		LicenseStatus: LicenseStatusTrial,
		LicenseExpires: now.AddDate(0, 1, 0), // 30-day trial
		MaxRooms:      tierMaxRooms(tier),
		MaxUsers:      tierMaxUsers(tier),
		Settings: TenantSettings{
			Timezone:     "UTC",
			CurrencyCode: "USD",
			DateFormat:   "YYYY-MM-DD",
			Language:     "en",
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
	tc.Capabilities = ComputeCapabilities(pt, tier)
	return tc
}

// HasCapability checks if a tenant has a specific capability.
func (tc *TenantConfig) HasCapability(cap Capability) bool {
	for _, c := range tc.Capabilities {
		if c == cap {
			return true
		}
	}
	return false
}

// ComputeCapabilities returns the full capability set for a property type + tier.
func ComputeCapabilities(pt PropertyType, tier LicenseTier) []Capability {
	caps := []Capability{
		CapReservations, CapGuests, CapProperties, CapRooms,
		CapHousekeeping, CapSettings, CapAuditLogs,
	}

	// Property-type-specific capabilities
	switch pt {
	case PropertyTypeHostel:
		caps = append(caps, Capability("hostel:bunk_management"))
	case PropertyTypeResort:
		caps = append(caps, Capability("resort:spa_management"), Capability("resort:activities"))
	case PropertyTypeAparthotel:
		caps = append(caps, Capability("aparthotel:extended_stay_rates"))
	}

	if tier == LicenseTierCore {
		return caps
	}

	// Operations tier +
	caps = append(caps,
		CapFloorDashboard, CapRoomBlocks, CapGroupReservations,
		CapMaintenance, CapFrontDeskWorkflow,
		CapIPTVBasic, CapSmartLocks, CapRemoteUnlock,
	)

	if tier == LicenseTierOperations {
		return caps
	}

	// Revenue tier +
	caps = append(caps,
		CapDynamicPricing, CapOTAIntegration,
		CapRevenueForecast, CapAgentManagement,
		CapIPTVPremium, CapIPTVWelcome, CapIPTVContent,
		CapAccessCodes,
	)

	if tier == LicenseTierRevenue {
		return caps
	}

	// Enterprise tier — everything
	caps = append(caps,
		CapMultiProperty, CapAdvancedCRM, CapAPIAccess,
		CapWhiteLabel, CapCustomReports,
	)

	return caps
}

func tierMaxRooms(tier LicenseTier) int {
	switch tier {
	case LicenseTierCore:
		return 50
	case LicenseTierOperations:
		return 150
	case LicenseTierRevenue:
		return 500
	case LicenseTierEnterprise:
		return 5000
	}
	return 50
}

func tierMaxUsers(tier LicenseTier) int {
	switch tier {
	case LicenseTierCore:
		return 5
	case LicenseTierOperations:
		return 20
	case LicenseTierRevenue:
		return 50
	case LicenseTierEnterprise:
		return 200
	}
	return 5
}
