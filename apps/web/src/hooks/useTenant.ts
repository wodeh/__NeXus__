import { useState, useEffect } from "react";
import { apiClient, getTenantConfig } from "@/lib/api";
import { TenantConfig } from "@/lib/tenant";

const fallbackConfig: TenantConfig = {
  id: "demo",
  name: "Demo Hotel",
  property_type: "boutique",
  license_tier: "enterprise",
  license_status: "active",
  license_expires_at: "",
  max_rooms: 5000,
  max_users: 200,
  capabilities: [
    "core:reservations", "core:guests", "core:properties", "core:rooms",
    "core:housekeeping", "core:settings", "core:audit_logs",
    "operations:floor_dashboard", "operations:room_blocks",
    "operations:group_reservations", "operations:maintenance", "operations:front_desk",
    "revenue:dynamic_pricing", "revenue:ota_integration",
    "revenue:revenue_forecasting", "revenue:agent_management",
    "enterprise:multi_property", "enterprise:advanced_crm",
    "enterprise:api_access", "enterprise:white_label", "enterprise:custom_reports",
    "villa:dashboard", "villa:properties", "villa:reservations", "villa:revenue", "villa:cleaner_tracking",
  ],
  settings: { timezone: "UTC", currency_code: "USD", date_format: "YYYY-MM-DD", language: "en" },
  created_at: "",
  updated_at: "",
};

export function useTenant() {
  const [config, setConfig] = useState<TenantConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const tenant = localStorage.getItem("nexus-tenant") || "hotel-01";
    getTenantConfig(tenant)
      .then((data) => {
        // Normalize: ensure capabilities array exists
        if (!data || !Array.isArray(data.capabilities)) {
          console.warn("Tenant config missing capabilities, using fallback");
          setConfig(fallbackConfig);
        } else {
          setConfig(data);
        }
        setLoading(false);
      })
      .catch((err) => {
        console.warn("Tenant config fetch failed, using fallback:", err.message);
        setConfig(fallbackConfig);
        setError(err.message);
        setLoading(false);
      });
  }, []);

  return { config: config || fallbackConfig, loading, error };
}
