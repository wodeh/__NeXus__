import { TenantConfig } from "@/lib/tenant";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function api<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(opts?.headers || {}),
    },
    ...opts,
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`API ${res.status}: ${body}`);
  }
  return res.json();
}

/* ─── Legacy axios-compatible client for existing hooks ─── */
export const apiClient = {
  async get<T>(path: string): Promise<T> {
    return api<T>(path);
  },
  async post<T>(path: string, payload: unknown): Promise<T> {
    return api<T>(path, { method: "POST", body: JSON.stringify(payload) });
  },
  async patch<T>(path: string, payload: unknown): Promise<T> {
    return api<T>(path, { method: "PATCH", body: JSON.stringify(payload) });
  },
  async delete<T>(path: string): Promise<T> {
    return api<T>(path, { method: "DELETE" });
  },
};

export interface Reservation {
  id: string;
  guest_name: string;
  email?: string;
  phone?: string;
  room_number?: string;
  room_type: string;
  check_in: string;
  check_out: string;
  adults: number;
  children: number;
  status: "confirmed" | "checked_in" | "checked_out" | "cancelled" | "no_show";
  source: "walk_in" | "ota" | "direct" | "agent";
  total: number;
  balance: number;
  special_requests?: string;
  vip: boolean;
  color?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateReservationPayload {
  guest_name: string;
  email?: string;
  phone?: string;
  room_type: string;
  room_number?: string;
  check_in: string;
  check_out: string;
  adults?: number;
  children?: number;
  source?: string;
  special_requests?: string;
  vip?: boolean;
}

export interface Room {
  id: string;
  number: string;
  type: string;
  floor: string;
  bed_type?: string;
  status: "occupied" | "vacant_clean" | "vacant_dirty" | "blocked" | "maintenance";
  rate_night: number;
}

export interface Property {
  id: string;
  tenant_id: string;
  name: string;
  code: string;
  address?: string;
  city?: string;
  country?: string;
  timezone: string;
  locale?: string;
  currency?: string;
  status: string;
  settings?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: string;
  tenant_id: string;
  email: string;
  name: string;
  phone?: string;
  role: string;
  status: string;
  last_login?: string;
  permissions?: string[];
  created_at: string;
  updated_at: string;
}

export interface Role {
  id: string;
  tenant_id: string;
  name: string;
  permissions: string[];
  description?: string;
  is_system: boolean;
  created_at: string;
  updated_at: string;
}

export interface HousekeepingTask {
  id: string;
  tenant_id: string;
  room_number: string;
  room_type?: string;
  task_type: string;
  priority?: string;
  status: string;
  assigned_to?: string;
  notes?: string;
  due_date?: string;
  completed_at?: string;
  completed_by?: string;
  created_at: string;
  updated_at: string;
}

export interface HousekeepingStaff {
  id: string;
  tenant_id: string;
  name: string;
  role: "cleaner" | "inspector" | "manager";
  active: boolean;
  active_shift: string;
  floors: string[];
  max_rooms_per_day: number;
  current_load: number;
  rating: number;
  completed_today: number;
  created_at: string;
  updated_at: string;
}

export interface Agent {
  id: string;
  tenant_id: string;
  name: string;
  type: "ota" | "travel_agent" | "corporate" | "direct";
  commission_pct: number;
  contact_name: string;
  contact_email: string;
  contact_phone: string;
  contract_ref: string;
  is_active: boolean;
  source_code: string;
  config?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface Channel {
  id: string;
  tenant_id: string;
  source: string;
  display_name: string;
  is_active: boolean;
  commission_pct: number;
  last_sync_at?: string;
  last_sync_status: "success" | "warning" | "error" | "n/a";
  config?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface ChannelSyncLog {
  id: string;
  tenant_id: string;
  channel_id: string;
  channel_source: string;
  direction: "pull" | "push";
  status: "success" | "partial" | "error";
  records: number;
  duration: string;
  error?: string;
  created_at: string;
}

/* ─── Reservation API ─── */

export async function getReservations(): Promise<Reservation[]> {
  const data = await api<{ reservations: Reservation[] }>("/v1/reservations");
  return data.reservations;
}

export async function getReservation(id: string): Promise<Reservation> {
  return api<Reservation>(`/v1/reservations/${id}`);
}

export async function createReservation(payload: CreateReservationPayload): Promise<Reservation> {
  return api<Reservation>("/v1/reservations", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function checkInReservation(id: string): Promise<Reservation> {
  return api<Reservation>(`/v1/reservations/${id}/checkin`, { method: "PATCH" });
}

export async function checkOutReservation(id: string): Promise<Reservation> {
  return api<Reservation>(`/v1/reservations/${id}/checkout`, { method: "PATCH" });
}

export async function cancelReservation(id: string): Promise<Reservation> {
  return api<Reservation>(`/v1/reservations/${id}/cancel`, { method: "PATCH" });
}

export async function moveReservation(id: string, payload: { room_number?: string; check_in?: string; check_out?: string }): Promise<Reservation> {
  return api<Reservation>(`/v1/reservations/${id}/move`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

/* ─── Room API ─── */

export async function getRooms(): Promise<Room[]> {
  const data = await api<{ rooms: Room[] }>("/v1/rooms");
  return data.rooms;
}

export async function updateRoomStatus(number: string, status: string): Promise<void> {
  await api(`/v1/rooms/${number}/status`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}

/* ─── Property API ─── */

export async function getProperties(): Promise<Property[]> {
  const data = await api<{ properties: Property[] }>("/v1/properties");
  return data.properties;
}

export async function getProperty(id: string): Promise<Property> {
  return api<Property>(`/v1/properties/${id}`);
}

/* ─── User API ─── */

export async function getUsers(): Promise<User[]> {
  const data = await api<{ users: User[] }>("/v1/admin/users");
  return data.users;
}

export async function createUser(payload: Omit<User, "id" | "created_at" | "updated_at">): Promise<User> {
  return api<User>("/v1/admin/users", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateUser(id: string, payload: Partial<User>): Promise<User> {
  return api<User>(`/v1/admin/users/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export async function deleteUser(id: string): Promise<void> {
  await api(`/v1/admin/users/${id}`, { method: "DELETE" });
}

/* ─── Role API ─── */

export async function getRoles(): Promise<Role[]> {
  const data = await api<{ roles: Role[] }>("/v1/admin/roles");
  return data.roles;
}

export async function createRole(payload: Omit<Role, "id" | "is_system" | "created_at" | "updated_at">): Promise<Role> {
  return api<Role>("/v1/admin/roles", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateRole(id: string, payload: Partial<Role>): Promise<Role> {
  return api<Role>(`/v1/admin/roles/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export async function deleteRole(id: string): Promise<void> {
  await api(`/v1/admin/roles/${id}`, { method: "DELETE" });
}

/* ─── Housekeeping API ─── */

export async function getHousekeepingTasks(): Promise<HousekeepingTask[]> {
  const data = await api<{ tasks: HousekeepingTask[] }>("/v1/housekeeping");
  return data.tasks;
}

export async function createHousekeepingTask(payload: Omit<HousekeepingTask, "id" | "tenant_id" | "status" | "created_at" | "updated_at">): Promise<HousekeepingTask> {
  return api<HousekeepingTask>("/v1/housekeeping", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateHousekeepingTask(id: string, payload: Partial<HousekeepingTask>): Promise<HousekeepingTask> {
  return api<HousekeepingTask>(`/v1/housekeeping/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

/* ─── Tenant Config API ─── */

export async function getTenantConfig(tenantId: string): Promise<TenantConfig> {
  return api<TenantConfig>(`/v1/tenant/${tenantId}`);
}

export async function updateTenantConfig(tenantId: string, payload: Partial<TenantConfig>): Promise<TenantConfig> {
  return api<TenantConfig>(`/v1/tenant/${tenantId}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

/* ─── Housekeeping Staff API ─── */

export async function getHousekeepingStaff(): Promise<HousekeepingStaff[]> {
  const data = await api<{ staff: HousekeepingStaff[] }>("/v1/housekeeping/staff");
  return data.staff;
}

export async function assignRoom(reservationId: string, roomNumber: string): Promise<Reservation> {
  return api<Reservation>(`/v1/reservations/${reservationId}/assign`, {
    method: "PATCH",
    body: JSON.stringify({ room_number: roomNumber }),
  });
}

export async function createProperty(payload: { name: string; timezone?: string; locale?: string; currency?: string; code?: string; status?: string; address?: string; city?: string; country?: string }): Promise<Property> {
  return api<Property>("/v1/properties", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function getAgents(): Promise<Agent[]> {
  const data = await api<{ agents: Agent[] }>("/v1/agents");
  return data.agents;
}

export async function createAgent(payload: Omit<Agent, "id" | "tenant_id" | "is_active" | "created_at" | "updated_at">): Promise<Agent> {
  return api<Agent>("/v1/agents", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateAgent(id: string, payload: Partial<Agent>): Promise<Agent> {
  return api<Agent>(`/v1/agents/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export async function deleteAgent(id: string): Promise<void> {
  await api(`/v1/agents/${id}`, { method: "DELETE" });
}

/* ─── Channel API ─── */

export async function getChannels(): Promise<Channel[]> {
  const data = await api<{ channels: Channel[] }>("/v1/channels");
  return data.channels;
}

export async function createChannel(payload: Omit<Channel, "id" | "tenant_id" | "is_active" | "last_sync_at" | "last_sync_status" | "created_at" | "updated_at">): Promise<Channel> {
  return api<Channel>("/v1/channels", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateChannel(id: string, payload: Partial<Channel>): Promise<Channel> {
  return api<Channel>(`/v1/channels/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export async function deleteChannel(id: string): Promise<void> {
  await api(`/v1/channels/${id}`, { method: "DELETE" });
}

export async function getChannelSyncLogs(channelId: string): Promise<ChannelSyncLog[]> {
  const data = await api<{ logs: ChannelSyncLog[] }>(`/v1/channels/${channelId}/sync-logs`);
  return data.logs;
}
