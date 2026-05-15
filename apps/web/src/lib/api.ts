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

/* ─── Reservation API ─── */
export interface Reservation {
  id: string;
  guest_name: string;
  email: string;
  phone: string;
  room_number: string;
  room_type: string;
  check_in: string;
  check_out: string;
  adults: number;
  children: number;
  status: string;
  source: string;
  total: number;
  balance: number;
  special_requests: string;
  vip: boolean;
  property_id?: string;
  color?: string;
}

export interface Room {
  id: string;
  number: string;
  type: string;
  floor: string;
  bed_type: string;
  status: string;
  rate_night: number;
  config: Record<string, any>;
  property_id?: string;
}

export async function getReservations(): Promise<Reservation[]> {
  const data = await api<{ reservations: Reservation[] }>("/v1/reservations");
  return data.reservations || [];
}

export async function getRooms(): Promise<Room[]> {
  const data = await api<{ rooms: Room[] }>("/v1/rooms");
  return data.rooms || [];
}

export async function getReservationDetail(id: string): Promise<Reservation> {
  return api<Reservation>(`/v1/reservations/${id}`);
}

export async function checkInReservation(id: string): Promise<void> {
  return api<void>(`/v1/reservations/${id}`, { method: "PATCH", body: JSON.stringify({ action: "check-in" }) });
}

export async function checkOutReservation(id: string): Promise<void> {
  return api<void>(`/v1/reservations/${id}`, { method: "PATCH", body: JSON.stringify({ action: "check-out" }) });
}

export async function cancelReservation(id: string): Promise<void> {
  return api<void>(`/v1/reservations/${id}`, { method: "PATCH", body: JSON.stringify({ action: "cancel" }) });
}

export async function restoreReservation(id: string): Promise<void> {
  return api<void>(`/v1/reservations/${id}`, { method: "PATCH", body: JSON.stringify({ action: "restore" }) });
}

export async function assignRoom(id: string, roomNumber: string): Promise<void> {
  return api<void>(`/v1/reservations/${id}`, { method: "PATCH", body: JSON.stringify({ action: "assign-room", room_number: roomNumber }) });
}

export async function moveReservation(id: string, payload: { room_number: string; check_in: string; check_out: string }): Promise<Reservation> {
  return api<Reservation>(`/v1/reservations/${id}`, { method: "PATCH", body: JSON.stringify({ action: "move-room", ...payload }) });
}

export async function updateRoomStatus(roomId: string, status: string): Promise<void> {
  return api<void>(`/v1/rooms/${roomId}`, { method: "PATCH", body: JSON.stringify({ status }) });
}

export async function createReservation(data: Partial<Reservation>): Promise<Reservation> {
  return api<Reservation>("/v1/reservations", { method: "POST", body: JSON.stringify(data) });
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

/* ─── Admin API ─── */
export interface AdminUser {
  id: string;
  email: string;
  name: string;
  role: string;
  is_active: boolean;
  last_login_at?: string;
  created_at: string;
}

export interface AdminProperty {
  id: string;
  name: string;
  address: string;
  city: string;
  country: string;
  phone: string;
  email: string;
  timezone: string;
  currency: string;
  star_rating: number;
  is_active: boolean;
  room_count: number;
  user_count: number;
}

export interface SystemConfig {
  default_check_in_time: string;
  default_check_out_time: string;
  auto_confirm: boolean;
  require_deposit: boolean;
  deposit_percent: number;
  allow_walk_in: boolean;
  overbooking_enabled: boolean;
}

export async function getAdminUsers(): Promise<AdminUser[]> {
  const data = await api<{ users: AdminUser[] }>("/v1/admin/users");
  return data.users || [];
}

export async function getAdminProperties(): Promise<AdminProperty[]> {
  const data = await api<{ properties: AdminProperty[] }>("/v1/admin/properties");
  return data.properties || [];
}

export async function getAdminConfig(): Promise<SystemConfig> {
  return api<SystemConfig>("/v1/admin/config");
}

/* ─── WiFi API ─── */
export interface WiFiAlert {
  id: string;
  floor: string;
  type: string;
  severity: string;
  message: string;
  suggested_fix: string;
  resolved_at?: string;
  resolved_by?: string;
}

export interface WiFiFloorSummary {
  floor: string;
  ap_count: number;
  online_aps: number;
  avg_signal_dbm: number;
  total_clients: number;
  active_alerts: number;
  overall_health: string;
}

export interface WiFiDashboard {
  overview: {
    total_aps: number;
    online_aps: number;
    total_clients: number;
    active_alerts: number;
    overall_health: string;
    avg_signal_dbm: number;
  };
  floors: WiFiFloorSummary[];
  alerts: WiFiAlert[];
}

export async function getWiFiDashboard(): Promise<WiFiDashboard> {
  return api<WiFiDashboard>("/v1/wifi/dashboard");
}

export async function getWiFiHeatmap(floor: string): Promise<{ floor: string; grid: number[][] }> {
  return api<{ floor: string; grid: number[][] }>(`/v1/wifi/heatmap?floor=${encodeURIComponent(floor)}`);
}

export async function resolveWiFiAlert(alertId: string): Promise<void> {
  return api<void>(`/v1/wifi/alerts/${alertId}`, { method: "PATCH", body: JSON.stringify({ resolved_by: "system" }) });
}

export async function triggerWiFiScan(floor?: string): Promise<{ status: string; scan_id: string }> {
  const qs = floor ? `?floor=${encodeURIComponent(floor)}` : "";
  return api<{ status: string; scan_id: string }>(`/v1/wifi/scan${qs}`, { method: "POST" });
}

/* ─── Villa API ─── */
export interface VillaProperty {
  id: string;
  name: string;
  city: string;
  country: string;
  bedrooms: number;
  bathrooms: number;
  max_guests: number;
  price_per_night: number;
  currency: string;
  cleaning_fee: number;
  security_deposit: number;
  is_active: boolean;
  status: string;
  description?: string;
  amenities?: string[];
  address?: string;
  latitude?: number;
  longitude?: number;
  images?: string[];
  owner_id?: string;
  owner_name?: string;
  owner_phone?: string;
  owner_email?: string;
}

export async function getVillas(): Promise<VillaProperty[]> {
  const data = await api<{ villas: VillaProperty[] }>("/v1/villas");
  return data.villas || [];
}

export async function createVilla(data: Partial<VillaProperty>): Promise<VillaProperty> {
  return api<VillaProperty>("/v1/villas", { method: "POST", body: JSON.stringify(data) });
}

export async function updateVilla(id: string, data: Partial<VillaProperty>): Promise<VillaProperty> {
  return api<VillaProperty>(`/v1/villas/${id}`, { method: "PATCH", body: JSON.stringify(data) });
}

export async function deleteVilla(id: string): Promise<void> {
  return api<void>(`/v1/villas/${id}`, { method: "DELETE" });
}
