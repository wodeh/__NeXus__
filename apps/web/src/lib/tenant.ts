export interface TenantConfig {
  id: string;
  name: string;
  property_type: string;
  license_tier: "core" | "operations" | "revenue" | "enterprise";
  license_status: "active" | "trial" | "expired" | "suspended";
  license_expires_at: string;
  max_rooms: number;
  max_users: number;
  capabilities: string[];
  settings: {
    timezone: string;
    currency_code: string;
    date_format: string;
    language: string;
  };
  created_at: string;
  updated_at: string;
}

export interface LicenseTier {
  id: string;
  name: string;
  price_monthly: number;
  max_rooms: number;
  max_users: number;
}

export interface CapabilityDef {
  id: string;
  category: string;
  name: string;
  description: string;
}

export const CAPABILITIES = {
  CORE: {
    RESERVATIONS: "core:reservations",
    GUESTS: "core:guests",
    PROPERTIES: "core:properties",
    ROOMS: "core:rooms",
    HOUSEKEEPING: "core:housekeeping",
    SETTINGS: "core:settings",
    AUDIT_LOGS: "core:audit_logs",
  },
  OPERATIONS: {
    FLOOR_DASHBOARD: "operations:floor_dashboard",
    ROOM_BLOCKS: "operations:room_blocks",
    GROUP_RESERVATIONS: "operations:group_reservations",
    MAINTENANCE: "operations:maintenance",
    FRONT_DESK: "operations:front_desk",
    IPTV_BASIC: "operations:iptv_basic",
    SMART_LOCKS: "operations:smart_locks",
    REMOTE_UNLOCK: "operations:remote_unlock",
  },
  REVENUE: {
    DYNAMIC_PRICING: "revenue:dynamic_pricing",
    OTA_INTEGRATION: "revenue:ota_integration",
    CHANNEL_MANAGER: "revenue:channel_manager",
    WHATSAPP_BOT: "revenue:whatsapp_bot",
    DIRECT_BOOKING: "revenue:direct_booking",
    GUEST_REVIEWS: "revenue:guest_reviews",
    COMMUNICATIONS: "revenue:communications",
    REVENUE_FORECAST: "revenue:revenue_forecasting",
    AGENT_MANAGEMENT: "revenue:agent_management",
    IPTV_PREMIUM: "revenue:iptv_premium",
    IPTV_WELCOME: "revenue:iptv_welcome",
    IPTV_CONTENT: "revenue:iptv_content",
    ACCESS_CODES: "revenue:access_codes",
  },
  ENTERPRISE: {
    MULTI_PROPERTY: "enterprise:multi_property",
    ADVANCED_CRM: "enterprise:advanced_crm",
    API_ACCESS: "enterprise:api_access",
    WHITE_LABEL: "enterprise:white_label",
    CUSTOM_REPORTS: "enterprise:custom_reports",
  },
} as const;

export function hasCapability(config: TenantConfig | null, cap: string): boolean {
  if (!config) return false;
  return config.capabilities.includes(cap);
}

export function hasAnyCapability(config: TenantConfig | null, caps: string[]): boolean {
  if (!config) return false;
  return caps.some((c) => config.capabilities.includes(c));
}

export function tierName(tier: string): string {
  const names: Record<string, string> = {
    core: "Core",
    operations: "Operations",
    revenue: "Revenue",
    enterprise: "Enterprise",
  };
  return names[tier] || tier;
}
