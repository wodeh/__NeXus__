"use client";

import { useState, useEffect } from "react";
import { useRouter, usePathname } from "next/navigation";
import Sidebar from "@/components/Sidebar";
import Topbar from "@/components/Topbar";
import { apiClient } from "@/lib/api";
import { TenantConfig } from "@/lib/tenant";
import { useAuth } from "@/lib/auth";

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
    "revenue:iptv_premium", "revenue:iptv_welcome", "revenue:iptv_content", "revenue:access_codes",
    "enterprise:multi_property", "enterprise:advanced_crm",
    "enterprise:api_access", "enterprise:white_label", "enterprise:custom_reports",
    "villa:dashboard", "villa:properties", "villa:reservations", "villa:revenue", "villa:cleaner_tracking",
  ],
  settings: { timezone: "UTC", currency_code: "USD", date_format: "YYYY-MM-DD", language: "en" },
  created_at: "",
  updated_at: "",
};

const villaOwnerCapabilities = [
  "villa:dashboard", "villa:properties", "villa:reservations", "villa:revenue", "villa:cleaner_tracking",
  "core:reservations", "core:guests", "core:settings",
  "revenue:iptv_premium", "revenue:iptv_welcome", "revenue:iptv_content", "revenue:access_codes",
];

export default function DashboardLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  const { user, token, isLoading, isVillaOwner } = useAuth();
  const router = useRouter();
  const pathname = usePathname();
  const [tenantConfig, setTenantConfig] = useState<TenantConfig | null>(null);
  const [loading, setLoading] = useState(true);

  // Auth guard: redirect to login if not authenticated
  useEffect(() => {
    if (!isLoading && !user && pathname !== "/login") {
      router.push("/login");
    }
  }, [isLoading, user, pathname, router]);

  useEffect(() => {
    if (!user) {
      setLoading(false);
      return;
    }
    const tenant = isVillaOwner ? "villa-owners" : (localStorage.getItem("nexus-tenant") || "demo");
    apiClient
      .get(`/v1/tenant/${tenant}`)
      .then((data) => {
        const cfg = data as TenantConfig;
        // Override capabilities for villa owners
        if (isVillaOwner) {
          cfg.capabilities = villaOwnerCapabilities;
          cfg.property_type = "villa";
          cfg.name = user.name + " — Villa Owner";
        }
        setTenantConfig(cfg);
        setLoading(false);
      })
      .catch(() => {
        const cfg = { ...fallbackConfig };
        if (isVillaOwner) {
          cfg.capabilities = villaOwnerCapabilities;
          cfg.property_type = "villa";
          cfg.name = user?.name + " — Villa Owner" || "Villa Owner";
        }
        setTenantConfig(cfg);
        setLoading(false);
      });
  }, [user, isVillaOwner]);

  if (isLoading || loading) {
    return (
      <div className="flex h-screen items-center justify-center bg-slate-950">
        <div className="h-8 w-8 animate-spin rounded-full border-2 border-nexus-500 border-t-transparent" />
      </div>
    );
  }

  if (!user) {
    return null; // Will redirect
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
