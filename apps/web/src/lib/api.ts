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

async function api<T>(path: string, opts?: RequestInit): Promise<T> {
  const isGet = !opts || !opts.method || opts.method === "GET";
  const tenantId = getTenantId();

  const res = await fetch(`${API_BASE}${path}`, {
    credentials: "include",
    headers: {
      "Content-Type": "application/json",
      "X-Tenant-ID": tenantId,
      ...(opts?.headers || {}),
    },
    cache: isGet ? "no-store" : undefined,
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
  config?: Record<string, unknown>;
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
  return data.reservations || [];
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

export async function restoreReservation(id: string): Promise<Reservation> {
  return api<Reservation>(`/v1/reservations/${id}/restore`, { method: "PATCH" });
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
  return data.rooms || [];
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
  return data.properties || [];
}

export async function getProperty(id: string): Promise<Property> {
  return api<Property>(`/v1/properties/${id}`);
}

/* ─── User API ─── */

export async function getUsers(): Promise<User[]> {
  const data = await api<{ users: User[] }>("/v1/admin/users");
  return data.users || [];
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
  return data.roles || [];
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
  const data = await api<{ tasks: HousekeepingTask[] }>("/v1/housekeeping/tasks");
  return data.tasks || [];
}

export async function createHousekeepingTask(payload: Omit<HousekeepingTask, "id" | "tenant_id" | "status" | "created_at" | "updated_at">): Promise<HousekeepingTask> {
  return api<HousekeepingTask>("/v1/housekeeping/tasks", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateHousekeepingTask(id: string, payload: Partial<HousekeepingTask>): Promise<HousekeepingTask> {
  return api<HousekeepingTask>(`/v1/housekeeping/tasks/${id}`, {
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
  return data.staff || [];
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
  return data.agents || [];
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
  return data.channels || [];
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
  return data.logs || [];
}

/* ─── Smart Lock API ─── */

export interface SmartLock {
  id: string;
  tenant_id: string;
  room_id: string;
  room_number: string;
  serial_number: string;
  model: string;
  manufacturer: string;
  status: "online" | "offline" | "low_battery" | "warning";
  battery_level: number;
  last_communication_at?: string;
  last_unlock_at?: string;
  last_lock_at?: string;
  firmware_version: string;
  remote_unlock_enabled: boolean;
  auto_lock_enabled: boolean;
}

export interface LockEvent {
  id: string;
  lock_id: string;
  room_number: string;
  event_type: string;
  event_source: string;
  details: string;
  occurred_at: string;
}

export interface AccessCode {
  id: string;
  lock_id: string;
  code: string;
  label: string;
  is_active: boolean;
  valid_from: string;
  valid_until?: string;
  max_uses?: number;
  use_count: number;
}

export async function getLocks(): Promise<SmartLock[]> {
  const data = await api<{ locks: SmartLock[] }>("/v1/locks");
  return data.locks || [];
}

export async function createLock(payload: Partial<SmartLock>): Promise<SmartLock> {
  return api<SmartLock>("/v1/locks", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateLock(id: string, payload: Partial<SmartLock>): Promise<SmartLock> {
  return api<SmartLock>(`/v1/locks/${id}`, {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export async function getLockEvents(lockId: string): Promise<LockEvent[]> {
  const data = await api<{ events: LockEvent[] }>(`/v1/locks/${lockId}/events`);
  return data.events || [];
}

export async function getLockAccessCodes(lockId: string): Promise<AccessCode[]> {
  const data = await api<{ codes: AccessCode[] }>(`/v1/locks/${lockId}/access-codes`);
  return data.codes || [];
}

export async function createAccessCode(lockId: string, payload: { code: string; label: string; valid_from: string; valid_until?: string; max_uses?: number }): Promise<AccessCode> {
  return api<AccessCode>(`/v1/locks/${lockId}/access-codes`, {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function remoteUnlock(lockId: string): Promise<{ status: string }> {
  return api<{ status: string }>(`/v1/locks/${lockId}/unlock`, {
    method: "POST",
  });
}

/* ─── WhatsApp API ─── */

export interface WhatsAppConversation {
  id: string;
  tenant_id: string;
  guest_phone: string;
  guest_name?: string;
  current_state: string;
  booking_created: boolean;
  messages: number;
  last_message_at: string;
}

export interface WhatsAppMessage {
  id: string;
  conversation_id: string;
  direction: "inbound" | "outbound";
  body: string;
  status: string;
  sent_at: string;
}

export interface WhatsAppBotConfig {
  tenant_id: string;
  bot_enabled: boolean;
  booking_enabled: boolean;
  auto_reply_enabled: boolean;
  welcome_message: string;
  phone_number_id: string;
  webhook_url: string;
}

export interface WhatsAppTemplate {
  id: string;
  trigger: string;
  response: string;
}

export async function getWhatsAppConversations(): Promise<WhatsAppConversation[]> {
  const data = await api<{ conversations: WhatsAppConversation[] }>("/v1/whatsapp/conversations");
  return data.conversations || [];
}

export async function getWhatsAppMessages(conversationId: string): Promise<WhatsAppMessage[]> {
  const data = await api<{ messages: WhatsAppMessage[] }>(`/v1/whatsapp/conversations/${conversationId}/messages`);
  return data.messages || [];
}

export async function sendWhatsAppMessage(conversationId: string, body: string): Promise<WhatsAppMessage> {
  return api<WhatsAppMessage>(`/v1/whatsapp/conversations/${conversationId}/messages`, {
    method: "POST",
    body: JSON.stringify({ body }),
  });
}

export async function getWhatsAppConfig(): Promise<WhatsAppBotConfig> {
  return api<WhatsAppBotConfig>("/v1/whatsapp/config");
}

export async function updateWhatsAppConfig(payload: Partial<WhatsAppBotConfig>): Promise<WhatsAppBotConfig> {
  return api<WhatsAppBotConfig>("/v1/whatsapp/config", {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export async function getWhatsAppTemplates(): Promise<WhatsAppTemplate[]> {
  const data = await api<{ templates: WhatsAppTemplate[] }>("/v1/whatsapp/templates");
  return data.templates || [];
}

/* ─── Review API ─── */

export interface Review {
  id: string;
  tenant_id: string;
  reservation_id: string;
  guest_name: string;
  room_number: string;
  rating: number;
  cleanliness: number;
  service: number;
  location: number;
  value: number;
  comment: string;
  staff_reply?: string;
  replied_at?: string;
  is_published: boolean;
  source: string;
}

export interface ReviewStats {
  tenant_id: string;
  total_reviews: number;
  average_rating: number;
  average_cleanliness: number;
  average_service: number;
  average_location: number;
  average_value: number;
  five_star_count: number;
  four_star_count: number;
  three_star_count: number;
  two_star_count: number;
  one_star_count: number;
}

export async function getReviews(): Promise<Review[]> {
  const data = await api<{ reviews: Review[] }>("/v1/reviews");
  return data.reviews || [];
}

export async function getReviewStats(): Promise<ReviewStats> {
  return api<ReviewStats>("/v1/reviews/stats");
}

export async function replyToReview(id: string, staffReply: string): Promise<{ status: string }> {
  return api<{ status: string }>(`/v1/reviews/${id}/reply`, {
    method: "POST",
    body: JSON.stringify({ staff_reply: staffReply }),
  });
}

export async function publishReview(id: string, published: boolean): Promise<{ status: string }> {
  return api<{ status: string }>(`/v1/reviews/${id}/publish`, {
    method: "PATCH",
    body: JSON.stringify({ published }),
  });
}

/* ─── Audit API ─── */

export interface AuditLog {
  id: string;
  tenant_id: string;
  user_id?: string;
  user_email?: string;
  action: string;
  resource: string;
  resource_id?: string;
  details?: Record<string, unknown>;
  ip_address?: string;
  user_agent?: string;
  created_at: string;
}

export interface AuditStats {
  tenant_id: string;
  total_events: number;
  today_events: number;
  user_actions: number;
  system_actions: number;
  failed_actions: number;
}

export async function getAuditLogs(): Promise<AuditLog[]> {
  const data = await api<{ logs: AuditLog[] }>("/v1/audit/logs");
  return data.logs || [];
}

export async function getAuditStats(): Promise<AuditStats> {
  return api<AuditStats>("/v1/audit/stats");
}

/* ─── IPTV API ─── */

export interface IPTVChannel {
  id: string;
  tenant_id: string;
  name: string;
  number: number;
  stream_url: string;
  logo_url?: string;
  category: string;
  language: string;
  is_active: boolean;
  is_premium: boolean;
}

export interface IPTVContent {
  id: string;
  tenant_id: string;
  title: string;
  type: "movie" | "series" | "music" | "info";
  description: string;
  duration?: number;
  thumbnail_url?: string;
  category: string;
  is_active: boolean;
}

export interface IPTVRoomStatus {
  id: string;
  tenant_id: string;
  room_id: string;
  room_number: string;
  is_online: boolean;
  current_channel?: number;
  last_activity_at?: string;
}

export async function getIPTVChannels(): Promise<IPTVChannel[]> {
  const data = await api<{ channels: IPTVChannel[] }>("/v1/iptv/channels");
  return data.channels || [];
}

export async function getIPTVContent(): Promise<IPTVContent[]> {
  const data = await api<{ content: IPTVContent[] }>("/v1/iptv/content");
  return data.content || [];
}

export async function getIPTVRooms(): Promise<IPTVRoomStatus[]> {
  const data = await api<{ rooms: IPTVRoomStatus[] }>("/v1/iptv/rooms");
  return data.rooms || [];
}

export interface IPTVWelcomeScreen {
  villa_id: string;
  villa_name: string;
  guest_name: string;
  check_in_date: string;
  check_out_date: string;
  nights: number;
  welcome_message: string;
  has_reservation: boolean;
}

export async function getIPTVWelcome(villaId: string): Promise<IPTVWelcomeScreen> {
  return api<IPTVWelcomeScreen>(`/v1/iptv/welcome/${villaId}`);
}

/* ─── Communications API ─── */

export interface CommTemplate {
  id: string;
  tenant_id: string;
  name: string;
  subject?: string;
  body: string;
  channel: "email" | "sms" | "whatsapp";
  category: string;
  is_active: boolean;
}

export interface CommSequence {
  id: string;
  tenant_id: string;
  name: string;
  description: string;
  trigger: string;
  is_active: boolean;
  steps: number;
}

export interface CommScheduled {
  id: string;
  tenant_id: string;
  guest_name: string;
  channel: string;
  subject?: string;
  body?: string;
  status: "scheduled" | "sent" | "delivered" | "failed";
  scheduled_at: string;
  sent_at?: string;
  error?: string;
}

export async function getCommTemplates(): Promise<CommTemplate[]> {
  const data = await api<{ templates: CommTemplate[] }>("/v1/communications/templates");
  return data.templates || [];
}

export async function getCommSequences(): Promise<CommSequence[]> {
  const data = await api<{ sequences: CommSequence[] }>("/v1/communications/sequences");
  return data.sequences || [];
}

export async function getCommScheduled(): Promise<CommScheduled[]> {
  const data = await api<{ scheduled: CommScheduled[] }>("/v1/communications/scheduled");
  return data.scheduled || [];
}


/* ─── Villa API ─── */

export interface VillaProperty {
  id: string;
  tenant_id: string;
  name: string;
  address?: string;
  city?: string;
  country?: string;
  elevation?: number;
  status: "available" | "reserved" | "occupied" | "maintenance";
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface VillaReservation {
  id: string;
  tenant_id: string;
  villa_id: string;
  villa_name: string;
  guest_name: string;
  check_in_date: string;
  check_out_date: string;
  nights: number;
  status: "pending" | "confirmed" | "reserved" | "checked_in" | "checked_out" | "cancelled";
  total_amount: number;
  source: string;
  created_at: string;
  updated_at: string;
}

export interface SystemConfig {
  tenant_id: string;
  default_check_in_time: string;
  default_check_out_time: string;
  auto_confirm: boolean;
  require_deposit: boolean;
  deposit_percent: number;
  allow_walk_in: boolean;
}

export async function getSystemConfig(): Promise<SystemConfig> {
  return api<SystemConfig>("/v1/admin/config");
}

export async function updateSystemConfig(payload: Partial<SystemConfig>): Promise<SystemConfig> {
  return api<SystemConfig>("/v1/admin/config", {
    method: "PATCH",
    body: JSON.stringify(payload),
  });
}

export async function getAdminProperties(): Promise<Property[]> {
  const data = await api<{ properties: Property[] }>("/v1/admin/properties");
  return data.properties;
}

export async function createAdminProperty(payload: { name: string; address?: string; city?: string; country?: string; phone?: string; email?: string; timezone?: string; currency?: string; star_rating?: number; config?: Record<string, unknown> }): Promise<Property> {
  return api<Property>("/v1/admin/properties", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function getAdminUsers(): Promise<User[]> {
  const data = await api<{ users: User[] }>("/v1/admin/users");
  return data.users;
}

export async function createAdminUser(payload: { email: string; name: string; role: string; password?: string }): Promise<User> {
  return api<User>("/v1/admin/users", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateAdminUser(id: string, payload: { name?: string; role?: string; is_active?: boolean }): Promise<{ status: string }> {
  return api<{ status: string }>("/v1/admin/users", {
    method: "PATCH",
    body: JSON.stringify({ id, ...payload }),
  });
}

export async function deleteAdminUser(id: string): Promise<{ status: string }> {
  return api<{ status: string }>(`/v1/admin/users?id=${id}`, { method: "DELETE" });
}

/* ─── Revenue Dashboard API ─── */

export interface RevenueDashboard {
  tenant_id: string;
  date: string;
  total_revenue: number;
  room_revenue: number;
  extra_revenue: number;
  tax_revenue: number;
  total_rooms: number;
  occupied_rooms: number;
  available_rooms: number;
  occupancy_rate: number;
  adr: number;
  revpar: number;
  arrivals: number;
  departures: number;
  stayovers: number;
  walk_ins: number;
  no_shows: number;
  cancellations: number;
}

export interface RevenueForecast {
  date: string;
  projected_revenue: number;
  projected_occupancy: number;
  confidence: number;
  booked_rooms: number;
  total_rooms: number;
}

export interface PricingRule {
  id: string;
  tenant_id: string;
  name: string;
  room_type: string;
  condition: string;
  trigger_value: number;
  adjustment_type: string;
  adjustment_value: number;
  min_rate: number;
  max_rate: number;
  is_active: boolean;
  priority: number;
  config?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface ChannelRevenueBreakdown {
  channel: string;
  bookings: number;
  revenue: number;
  commission: number;
  net_revenue: number;
  avg_rate: number;
}

export interface RoomTypeRevenue {
  room_type: string;
  nights_sold: number;
  revenue: number;
  avg_rate: number;
  occupancy_pct: number;
}

export async function getRevenueDashboard(from?: string, to?: string): Promise<{ stats: RevenueDashboard[]; period: { from: string; to: string } }> {
  const params = new URLSearchParams();
  if (from) params.append("from", from);
  if (to) params.append("to", to);
  return api(`/v1/revenue/dashboard?${params.toString()}`);
}

export async function getRevenueForecast(): Promise<{ forecast: RevenueForecast[] }> {
  return api("/v1/revenue/forecast");
}

export async function getChannelRevenue(): Promise<{ breakdown: ChannelRevenueBreakdown[] }> {
  return api("/v1/revenue/channels");
}

export async function getRoomTypeRevenue(): Promise<{ room_types: RoomTypeRevenue[] }> {
  return api("/v1/revenue/room-types");
}

export async function getPricingRules(): Promise<{ rules: PricingRule[] }> {
  return api("/v1/revenue/pricing-rules");
}

export async function createPricingRule(payload: Omit<PricingRule, "id" | "tenant_id" | "created_at" | "updated_at">): Promise<PricingRule> {
  return api<PricingRule>("/v1/revenue/pricing-rules", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

/* ─── Reservation Detail API ─── */

export interface ReservationDetail {
  id: string;
  tenant_id: string;
  guest: {
    name: string;
    email: string;
    phone: string;
    address?: string;
    city?: string;
    country?: string;
    id_type?: string;
    id_number?: string;
    birth_date?: string;
    nationality?: string;
    vip: boolean;
  };
  room: {
    room_number: string;
    room_type: string;
    floor?: string;
    bed_type?: string;
    rate_night: number;
  };
  stay: {
    check_in: string;
    check_out: string;
    nights: number;
    adults: number;
    children: number;
    status: string;
    source: string;
  };
  charges: {
    subtotal: number;
    tax: number;
    total: number;
    paid: number;
    balance: number;
  };
  extras: Array<{
    description: string;
    amount: number;
    quantity: number;
  }>;
  audit_trail: Array<{
    action: string;
    user: string;
    timestamp: string;
  }>;
}

/* ─── Villa API ─── */

export interface VillaProperty {
  id: string;
  tenant_id: string;
  name: string;
  address?: string;
  city?: string;
  country?: string;
  elevation?: number;
  status: "available" | "reserved" | "occupied" | "maintenance";
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface VillaReservation {
  id: string;
  tenant_id: string;
  villa_id: string;
  villa_name: string;
  guest_name: string;
  check_in_date: string;
  check_out_date: string;
  nights: number;
  status: "pending" | "confirmed" | "reserved" | "checked_in" | "checked_out" | "cancelled";
  total_amount: number;
  source: string;
  created_at: string;
  updated_at: string;
}

export async function getVillas(): Promise<VillaProperty[]> {
  const data = await api<{ villas: VillaProperty[] }>("/v1/villas");
  return data.villas || [];
}

export async function getVillaReservations(): Promise<VillaReservation[]> {
  const data = await api<{ reservations: VillaReservation[] }>("/v1/villa-reservations");
  return data.reservations || [];
}

export async function getVillaRevenue(from: string, to: string): Promise<{ total_revenue: number; villa_breakdown: Array<{ villa_id: string; villa_name: string; revenue: number; nights_booked: number; occupancy_pct: number }> }> {
  const params = new URLSearchParams();
  params.append("from", from);
  params.append("to", to);
  return api(`/v1/villa-revenue?${params.toString()}`);
}

export async function createVilla(payload: Omit<VillaProperty, "id" | "tenant_id" | "created_at" | "updated_at">): Promise<VillaProperty> {
  return api("/v1/villas", { method: "POST", body: JSON.stringify(payload) });
}

export async function updateVilla(id: string, payload: Partial<VillaProperty>): Promise<VillaProperty> {
  return api(`/v1/villas/${id}`, { method: "PATCH", body: JSON.stringify(payload) });
}

export async function deleteVilla(id: string): Promise<void> {
  return api(`/v1/villas/${id}`, { method: "DELETE" });
}

export async function updateVillaReservation(id: string, payload: Partial<VillaReservation>): Promise<VillaReservation> {
  return api(`/v1/villa-reservations/${id}`, { method: "PATCH", body: JSON.stringify(payload) });
}

export async function createVillaReservation(payload: Omit<VillaReservation, "id" | "tenant_id" | "created_at" | "updated_at">): Promise<VillaReservation> {
  return api("/v1/villa-reservations", { method: "POST", body: JSON.stringify(payload) });
}

export interface CleanerSensorLog {
  id: string;
  villa_id: string;
  cleaner_name: string;
  action: "check_in" | "check_out";
  temperature?: number;
  latitude?: number;
  longitude?: number;
  altitude?: number;
  floor?: number;
  location_type?: string;
  battery_level?: number;
  recorded_at?: string;
  timestamp: string;
  location?: string;
}

export async function getSensorLogs(villaId: string): Promise<CleanerSensorLog[]> {
  const data = await api<{ logs: CleanerSensorLog[] }>(`/v1/villas/${villaId}/sensor-logs`);
  return data.logs || [];
}

export async function recordSensorLog(villaId: string, payload: { villa_id?: string; villa_name?: string; cleaner_id?: string; cleaner_name: string; action?: string; temperature?: number; latitude?: number; longitude?: number; altitude?: number; floor?: number; location_type?: string; battery_level?: number; recorded_at?: string; location?: string }): Promise<CleanerSensorLog> {
  return api(`/v1/villas/${villaId}/sensor-logs`, { method: "POST", body: JSON.stringify(payload) });
}



/* Re-exports from api-missing.ts */
export * from "./api-missing";

