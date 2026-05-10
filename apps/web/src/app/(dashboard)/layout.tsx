"use client";

import { useState, useEffect } from "react";
import Sidebar from "@/components/Sidebar";
import Topbar from "@/components/Topbar";
import { apiClient } from "@/lib/api";
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
  ],
  settings: { timezone: "UTC", currency_code: "USD", date_format: "YYYY-MM-DD", language: "en" },
  created_at: "",
  updated_at: "",
};

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const [tenantConfig, setTenantConfig] = useState<TenantConfig | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const tenant = localStorage.getItem("nexus-tenant") || "demo";
    apiClient
      .get(`/tenants/${tenant}/config`)
      .then((data) => {
        setTenantConfig(data as TenantConfig);
        setLoading(false);
      })
      .catch(() => {
        setTenantConfig(fallbackConfig);
        setLoading(false);
      });
  }, []);

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-slate-950">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-nexus-500 border-t-transparent" />
      </div>
    );
  }

  return (
    <div className="flex min-h-screen">
      <Sidebar tenantConfig={tenantConfig || fallbackConfig} />
      <div className="ml-60 flex flex-1 flex-col">
        <Topbar tenantConfig={tenantConfig || fallbackConfig} />
        <main className="flex-1 p-6">{children}</main>
      </div>
    </div>
  );
}
