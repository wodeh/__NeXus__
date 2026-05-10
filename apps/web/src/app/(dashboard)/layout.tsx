"use client";

import { useState, useEffect } from "react";
import Sidebar from "@/components/Sidebar";
import Topbar from "@/components/Topbar";
import { apiClient } from "@/lib/api";
import { TenantConfig } from "@/lib/tenant";

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
      <Sidebar tenantConfig={tenantConfig} />
      <div className="ml-60 flex flex-1 flex-col">
        <Topbar tenantConfig={tenantConfig} />
        <main className="flex-1 p-6">{children}</main>
      </div>
    </div>
  );
}
