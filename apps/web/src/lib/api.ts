

"use client";

import { TenantConfig } from "@/lib/tenant";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";
const TENANT_ID = process.env.NEXT_PUBLIC_TENANT_ID || "demo";

function getTenantId(): string {
  if (typeof window !== "undefined") {
    const saved = localStorage.getItem("nexus-tenant");
    if (saved) return saved;
  }
  return TENANT_ID;
}

function getAuthToken(): string | null {
  if (typeof window !== "undefined") {
    return localStorage.getItem("nexus-token");
  }
  return null;
}

export async function api<T>(path: string, opts?: RequestInit): Promise<T> {
  const isGet = !opts || !opts.method || opts.method === "GET";
  const tenantId = getTenantId();
  const token = getAuthToken();

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    "X-Tenant-ID": tenantId,
  };

  if (token) {
    headers["Authorization"] = `Bearer ${token}`;
  }

  if (opts?.headers) {
    const extra = opts.headers as Record<string, string>;
    Object.assign(headers, extra);
  }

  const url = `${API_BASE}${path}`;

  const res = await fetch(url, {
    ...opts,
    headers,
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(text || `HTTP ${res.status}`);
  }

  // Handle empty responses (e.g., 204 No Content)
  if (res.status === 204) {
    return undefined as T;
  }

  return res.json();
}

// Convenience wrapper with .get / .post / .patch / .delete
export const apiClient = {
  get: <T>(path: string) => api<T>(path),
  post: <T>(path: string, body: unknown) => api<T>(path, { method: "POST", body: JSON.stringify(body) }),
  patch: <T>(path: string, body: unknown) => api<T>(path, { method: "PATCH", body: JSON.stringify(body) }),
  put: <T>(path: string, body: unknown) => api<T>(path, { method: "PUT", body: JSON.stringify(body) }),
  delete: <T>(path: string) => api<T>(path, { method: "DELETE" }),
};

/* ─── Tenant API ─── */
export async function getTenantConfig(externalId: string): Promise<TenantConfig> {
  return api<TenantConfig>(`/v1/tenant/${externalId}`);
}

/* ─── Overbooking API ─── */
export interface OverbookingConfidence {
  room_type: string;
  confidence: number;
  no_show_rate: number;
  cancellation_rate: number;
  current_occupancy: number;
  day_of_week_risk: number;
  suggested_overbook: number;
}

export async function getOverbookingConfidence(date?: string): Promise<OverbookingConfidence[]> {
  const qs = date ? `?date=${date}` : "";
  return api<OverbookingConfidence[]>(`/v1/overbooking/confidence${qs}`);
}
