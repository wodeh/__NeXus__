import { useState, useEffect } from "react";
import { apiClient } from "@/lib/api";
import { TenantConfig } from "@/lib/tenant";

export function useTenant() {
  const [config, setConfig] = useState<TenantConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const tenant = localStorage.getItem("nexus-tenant") || "demo";
    apiClient
      .get(`/tenants/${tenant}/config`)
      .then((data) => {
        setConfig(data as TenantConfig);
        setLoading(false);
      })
      .catch((err) => {
        setError(err.message);
        setLoading(false);
      });
  }, []);

  return { config, loading, error };
}
