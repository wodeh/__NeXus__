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
  elevation?: number;
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



/* ─── Tenant Config Update ─── */
export async function updateTenantConfig(data: Partial<TenantConfig>): Promise<TenantConfig> {
  return api<TenantConfig>("/v1/tenant/config", { method: "PATCH", body: JSON.stringify(data) });
}

/* ─── Housekeeping API ─── */
export interface HousekeepingTask {
  id: string;
  room_id: string;
  room_number: string;
  floor: string;
  type: string;
  status: string;
  assigned_to?: string;
  assigned_name?: string;
  priority: string;
  estimated_minutes?: number;
  started_at?: string;
  completed_at?: string;
  notes?: string;
  created_at?: string;
}

export interface HousekeepingStaff {
  id: string;
  name: string;
  role: string;
  phone?: string;
  email?: string;
  is_active: boolean;
  current_task_count?: number;
}

export async function getHousekeepingTasks(): Promise<HousekeepingTask[]> {
  const data = await api<{ tasks: HousekeepingTask[] }>("/v1/housekeeping/tasks");
  return data.tasks || [];
}

export async function getHousekeepingStaff(): Promise<HousekeepingStaff[]> {
  const data = await api<{ staff: HousekeepingStaff[] }>("/v1/housekeeping/staff");
  return data.staff || [];
}

export async function createHousekeepingTask(data: Partial<HousekeepingTask>): Promise<HousekeepingTask> {
  return api<HousekeepingTask>("/v1/housekeeping/tasks", { method: "POST", body: JSON.stringify(data) });
}

export async function updateHousekeepingTask(id: string, data: Partial<HousekeepingTask>): Promise<HousekeepingTask> {
  return api<HousekeepingTask>(`/v1/housekeeping/tasks/${id}`, { method: "PATCH", body: JSON.stringify(data) });
}

/* ─── Villa Reservations API ─── */
export interface VillaReservation {
  id: string;
  villa_id: string;
  guest_name: string;
  email: string;
  phone?: string;
  check_in: string;
  check_out: string;
  adults: number;
  children: number;
  status: string;
  total: number;
  balance: number;
  special_requests?: string;
  source?: string;
  created_at?: string;
  updated_at?: string;
}

export async function getVillaReservations(): Promise<VillaReservation[]> {
  const data = await api<{ reservations: VillaReservation[] }>("/v1/villa-reservations");
  return data.reservations || [];
}

export async function createVillaReservation(data: Partial<VillaReservation>): Promise<VillaReservation> {
  return api<VillaReservation>("/v1/villa-reservations", { method: "POST", body: JSON.stringify(data) });
}

export async function updateVillaReservation(id: string, data: Partial<VillaReservation>): Promise<VillaReservation> {
  return api<VillaReservation>(`/v1/villa-reservations/${id}`, { method: "PATCH", body: JSON.stringify(data) });
}

export async function deleteVillaReservation(id: string): Promise<void> {
  return api<void>(`/v1/villa-reservations/${id}`, { method: "DELETE" });
}

/* ─── Smart Locks API ─── */
export interface SmartLock {
  id: string;
  room_id: string;
  room_number: string;
  device_id: string;
  brand: string;
  status: string;
  battery_level?: number;
  signal_strength?: number;
  firmware_version?: string;
  last_synced_at?: string;
  is_online: boolean;
}

export interface LockEvent {
  id: string;
  lock_id: string;
  event_type: string;
  details?: Record<string, any>;
  created_at: string;
}

export interface AccessCode {
  id: string;
  lock_id: string;
  code: string;
  label?: string;
  type: string;
  valid_from?: string;
  valid_until?: string;
  is_active: boolean;
}

export async function getLocks(): Promise<SmartLock[]> {
  const data = await api<{ locks: SmartLock[] }>("/v1/locks");
  return data.locks || [];
}

export async function getLockEvents(lockId: string): Promise<LockEvent[]> {
  const data = await api<{ events: LockEvent[] }>(`/v1/locks/${lockId}/events`);
  return data.events || [];
}

export async function getLockAccessCodes(lockId: string): Promise<AccessCode[]> {
  const data = await api<{ codes: AccessCode[] }>(`/v1/locks/${lockId}/codes`);
  return data.codes || [];
}

export async function remoteUnlock(lockId: string): Promise<void> {
  return api<void>(`/v1/locks/${lockId}/unlock`, { method: "POST" });
}

export interface GuestJourneyEvent {
  id: string;
  reservation_id?: string;
  guest_id?: string;
  type: string;
  description?: string;
  metadata?: Record<string, any>;
  created_at: string;
  icon?: string;
  color?: string;
}

export async function getGuestTimeline(guestId: string): Promise<GuestJourneyEvent[]> {
  const data = await api<{ events: GuestJourneyEvent[] }>(`/v1/guests/${guestId}/timeline`);
  return data.events || [];
}

export async function addGuestJourneyEvent(guestId: string, event: Partial<GuestJourneyEvent>): Promise<GuestJourneyEvent> {
  return api<GuestJourneyEvent>(`/v1/guests/${guestId}/events`, { method: "POST", body: JSON.stringify(event) });
}
export interface GuestJourney {
  id: string;
  name: string;
  description?: string;
  trigger: string;
  trigger_config?: Record<string, any>;
  steps: GuestJourneyStep[];
  is_active: boolean;
  created_at?: string;
}

export interface GuestJourneyStep {
  id: string;
  type: string;
  config?: Record<string, any>;
  delay_minutes?: number;
  order_index: number;
}

export interface GuestJourneyExecution {
  id: string;
  journey_id: string;
  journey_name?: string;
  reservation_id: string;
  guest_name?: string;
  status: string;
  current_step?: number;
  total_steps?: number;
  started_at?: string;
  completed_at?: string;
}

export async function getJourneys(): Promise<GuestJourney[]> {
  const data = await api<{ journeys: GuestJourney[] }>("/v1/journeys");
  return data.journeys || [];
}

export async function getJourneyExecutions(): Promise<GuestJourneyExecution[]> {
  const data = await api<{ executions: GuestJourneyExecution[] }>("/v1/journeys/executions");
  return data.executions || [];
}

export async function createJourney(data: Partial<GuestJourney>): Promise<GuestJourney> {
  return api<GuestJourney>("/v1/journeys", { method: "POST", body: JSON.stringify(data) });
}

export async function updateJourney(id: string, data: Partial<GuestJourney>): Promise<GuestJourney> {
  return api<GuestJourney>(`/v1/journeys/${id}`, { method: "PATCH", body: JSON.stringify(data) });
}

export async function deleteJourney(id: string): Promise<void> {
  return api<void>(`/v1/journeys/${id}`, { method: "DELETE" });
}

/* ─── Upsells API ─── */
export interface UpsellOffer {
  id: string;
  name: string;
  description?: string;
  category: string;
  price: number;
  currency: string;
  image_url?: string;
  is_active: boolean;
  auto_offer: boolean;
  created_at?: string;
}

export interface UpsellPurchase {
  id: string;
  offer_id: string;
  offer_name?: string;
  reservation_id: string;
  guest_name?: string;
  quantity: number;
  total: number;
  status: string;
  created_at?: string;
}

export async function getUpsellOffers(): Promise<UpsellOffer[]> {
  const data = await api<{ offers: UpsellOffer[] }>("/v1/upsells/offers");
  return data.offers || [];
}

export async function getUpsellPurchases(): Promise<UpsellPurchase[]> {
  const data = await api<{ purchases: UpsellPurchase[] }>("/v1/upsells/purchases");
  return data.purchases || [];
}

export async function createUpsellOffer(data: Partial<UpsellOffer>): Promise<UpsellOffer> {
  return api<UpsellOffer>("/v1/upsells/offers", { method: "POST", body: JSON.stringify(data) });
}

export async function updateUpsellOffer(id: string, data: Partial<UpsellOffer>): Promise<UpsellOffer> {
  return api<UpsellOffer>(`/v1/upsells/offers/${id}`, { method: "PATCH", body: JSON.stringify(data) });
}

export async function deleteUpsellOffer(id: string): Promise<void> {
  return api<void>(`/v1/upsells/offers/${id}`, { method: "DELETE" });
}

/* ─── Revenue API ─── */
export interface RevenueDashboard {
  total_revenue: number;
  total_bookings: number;
  occupancy_rate: number;
  adr: number;
  revpar: number;
  comparison_period?: { revenue_change: number; bookings_change: number };
}

export interface RevenueForecast {
  dates: string[];
  revenue: number[];
  occupancy: number[];
  bookings: number[];
}

export interface ChannelRevenueBreakdown {
  channel: string;
  revenue: number;
  bookings: number;
  percentage: number;
}

export interface RoomTypeRevenue {
  room_type: string;
  revenue: number;
  nights: number;
  occupancy: number;
}

export interface PricingRule {
  id: string;
  name: string;
  room_type: string;
  condition: string;
  adjustment_type: string;
  adjustment_value: number;
  is_active: boolean;
}

export async function getRevenueDashboard(period?: string): Promise<RevenueDashboard> {
  const qs = period ? `?period=${period}` : "";
  return api<RevenueDashboard>(`/v1/revenue/dashboard${qs}`);
}

export async function getRevenueForecast(days?: number): Promise<RevenueForecast> {
  const qs = days ? `?days=${days}` : "";
  return api<RevenueForecast>(`/v1/revenue/forecast${qs}`);
}

export async function getChannelRevenue(): Promise<ChannelRevenueBreakdown[]> {
  const data = await api<{ channels: ChannelRevenueBreakdown[] }>("/v1/revenue/channels");
  return data.channels || [];
}

export async function getRoomTypeRevenue(): Promise<RoomTypeRevenue[]> {
  const data = await api<{ room_types: RoomTypeRevenue[] }>("/v1/revenue/room-types");
  return data.room_types || [];
}

export async function getPricingRules(): Promise<PricingRule[]> {
  const data = await api<{ rules: PricingRule[] }>("/v1/revenue/pricing-rules");
  return data.rules || [];
}

/* ─── Reviews API ─── */
export interface Review {
  id: string;
  platform: string;
  guest_name: string;
  rating: number;
  title?: string;
  content?: string;
  reply?: string;
  is_published: boolean;
  stay_date?: string;
  created_at: string;
}

export interface ReviewStats {
  average_rating: number;
  total_reviews: number;
  five_star: number;
  four_star: number;
  three_star: number;
  two_star: number;
  one_star: number;
}

export async function getReviews(): Promise<Review[]> {
  const data = await api<{ reviews: Review[] }>("/v1/reviews");
  return data.reviews || [];
}

export async function getReviewStats(): Promise<ReviewStats> {
  return api<ReviewStats>("/v1/reviews/stats");
}

export async function replyToReview(id: string, reply: string): Promise<Review> {
  return api<Review>(`/v1/reviews/${id}`, { method: "PATCH", body: JSON.stringify({ reply }) });
}

export async function publishReview(id: string): Promise<Review> {
  return api<Review>(`/v1/reviews/${id}/publish`, { method: "POST" });
}

/* ─── Cleaner Tracking / Sensor Logs API ─── */
export interface CleanerSensorLog {
  id: string;
  villa_id: string;
  sensor_type: string;
  value: number;
  temperature?: number;
  unit?: string;
  recorded_at: string;
  location?: string;
}

export async function getSensorLogs(villaId?: string, limit?: number): Promise<CleanerSensorLog[]> {
  const params = new URLSearchParams();
  if (villaId) params.set("villa_id", villaId);
  if (limit) params.set("limit", String(limit));
  const qs = params.toString();
  const data = await api<{ logs: CleanerSensorLog[] }>(`/v1/sensor-logs${qs ? "?" + qs : ""}`);
  return data.logs || [];
}

export async function recordSensorLog(data: Partial<CleanerSensorLog>): Promise<CleanerSensorLog> {
  return api<CleanerSensorLog>("/v1/sensor-logs", { method: "POST", body: JSON.stringify(data) });
}

/* ─── Villa Revenue API ─── */
export interface VillaRevenueStats {
  total_revenue: number;
  total_bookings: number;
  total_nights: number;
  average_nightly_rate: number;
  occupancy_rate: number;
  by_villa?: { villa_id: string; villa_name: string; revenue: number; nights: number }[];
}

export async function getVillaRevenue(period?: string): Promise<VillaRevenueStats> {
  const qs = period ? `?period=${period}` : "";
  return api<VillaRevenueStats>(`/v1/villa-revenue${qs}`);
}

/* ─── WhatsApp API ─── */
export interface WhatsAppConversation {
  id: string;
  guest_phone: string;
  guest_name?: string;
  last_message?: string;
  last_message_at?: string;
  unread_count: number;
  status: string;
}

export interface WhatsAppMessage {
  id: string;
  conversation_id: string;
  direction: "inbound" | "outbound";
  content: string;
  sent_at: string;
  status: string;
}

export interface WhatsAppBotConfig {
  id: string;
  is_active: boolean;
  welcome_message?: string;
  auto_reply_enabled: boolean;
  fallback_phone?: string;
}

export interface WhatsAppTemplate {
  id: string;
  name: string;
  language: string;
  category: string;
  content: string;
  variables?: string[];
  status: string;
}

export async function getWhatsAppConversations(): Promise<WhatsAppConversation[]> {
  const data = await api<{ conversations: WhatsAppConversation[] }>("/v1/whatsapp/conversations");
  return data.conversations || [];
}

export async function getWhatsAppMessages(convId: string): Promise<WhatsAppMessage[]> {
  const data = await api<{ messages: WhatsAppMessage[] }>(`/v1/whatsapp/conversations/${convId}/messages`);
  return data.messages || [];
}

export async function sendWhatsAppMessage(convId: string, content: string): Promise<WhatsAppMessage> {
  return api<WhatsAppMessage>(`/v1/whatsapp/conversations/${convId}/messages`, { method: "POST", body: JSON.stringify({ content }) });
}

export async function getWhatsAppConfig(): Promise<WhatsAppBotConfig> {
  return api<WhatsAppBotConfig>("/v1/whatsapp/config");
}

export async function updateWhatsAppConfig(data: Partial<WhatsAppBotConfig>): Promise<WhatsAppBotConfig> {
  return api<WhatsAppBotConfig>("/v1/whatsapp/config", { method: "PATCH", body: JSON.stringify(data) });
}

export async function getWhatsAppTemplates(): Promise<WhatsAppTemplate[]> {
  const data = await api<{ templates: WhatsAppTemplate[] }>("/v1/whatsapp/templates");
  return data.templates || [];
}

/* ─── Communications API ─── */
export interface CommTemplate {
  id: string;
  name: string;
  channel: string;
  subject?: string;
  body: string;
  variables?: string[];
  is_active: boolean;
}

export interface CommSequence {
  id: string;
  name: string;
  description?: string;
  trigger: string;
  steps: CommSequenceStep[];
  is_active: boolean;
}

export interface CommSequenceStep {
  id: string;
  template_id: string;
  delay_minutes: number;
  order_index: number;
}

export interface CommScheduled {
  id: string;
  sequence_id?: string;
  template_id?: string;
  recipient: string;
  scheduled_at: string;
  status: string;
  sent_at?: string;
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

/* ─── Waitlist API ─── */
export interface WaitlistEntry {
  id: string;
  guest_name: string;
  email?: string;
  phone?: string;
  room_type?: string;
  check_in?: string;
  check_out?: string;
  adults?: number;
  children?: number;
  priority: number;
  status: string;
  notes?: string;
  created_at?: string;
}

export async function createGroupReservation(data: Partial<Reservation>): Promise<Reservation> {
  return api<Reservation>("/v1/reservations/group", { method: "POST", body: JSON.stringify(data) });
}

export async function getWaitlist(): Promise<WaitlistEntry[]> {
  const data = await api<{ entries: WaitlistEntry[] }>("/v1/waitlist");
  return data.entries || [];
}

export async function addToWaitlist(data: Partial<WaitlistEntry>): Promise<WaitlistEntry> {
  return api<WaitlistEntry>("/v1/waitlist", { method: "POST", body: JSON.stringify(data) });
}

export async function assignWaitlistRoom(id: string, roomNumber: string): Promise<WaitlistEntry> {
  return api<WaitlistEntry>(`/v1/waitlist/${id}/assign`, { method: "POST", body: JSON.stringify({ room_number: roomNumber }) });
}

export async function cancelWaitlistEntry(id: string): Promise<void> {
  return api<void>(`/v1/waitlist/${id}`, { method: "DELETE" });
}

/* ─── IPTV API ─── */
export interface IPTVWelcomeScreen {
  guest_name: string;
  villa_name: string;
  check_in: string;
  check_out: string;
  weather?: { temp: number; condition: string };
  message?: string;
  services?: { name: string; icon: string }[];
}

export async function getIPTVWelcome(villaId: string): Promise<IPTVWelcomeScreen> {
  return api<IPTVWelcomeScreen>(`/v1/iptv/welcome/${villaId}`);
}

export interface IPTVChannel {
  id: string;
  name: string;
  number: number;
  category: string;
  logo_url?: string;
  stream_url?: string;
  is_active: boolean;
}

export interface IPTVContent {
  id: string;
  title: string;
  type: string;
  description?: string;
  thumbnail_url?: string;
  duration?: number;
  is_active: boolean;
}

export interface IPTVRoomStatus {
  room_id: string;
  is_online: boolean;
  current_channel?: string;
  volume?: number;
  is_muted?: boolean;
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

/* ─── Agents API ─── */
export interface Agent {
  id: string;
  name: string;
  type: string;
  description?: string;
  is_active: boolean;
  config?: Record<string, any>;
  last_run_at?: string;
  created_at?: string;
  commission_pct?: number;
  contract_ref?: string;
  contact_email?: string;
  contact_phone?: string;
}

export async function getAgents(): Promise<Agent[]> {
  const data = await api<{ agents: Agent[] }>("/v1/agents");
  return data.agents || [];
}

export async function updateAgent(id: string, data: Partial<Agent>): Promise<Agent> {
  return api<Agent>(`/v1/agents/${id}`, { method: "PATCH", body: JSON.stringify(data) });
}

export async function deleteAgent(id: string): Promise<void> {
  return api<void>(`/v1/agents/${id}`, { method: "DELETE" });
}

/* ─── Room Status View API ─── */
export interface RoomStatusDay {
  date: string;
  status: string;
  reservation_id?: string;
  guest_name?: string;
}

export interface RoomDailyStatus {
  room_id: string;
  room_number: string;
  floor: string;
  type: string;
  days: RoomStatusDay[];
}

export interface RoomStatusView {
  date_from: string;
  date_to: string;
  rooms: RoomDailyStatus[];
}

export async function getRoomStatusView(dateFrom: string, dateTo: string): Promise<RoomStatusView> {
  return api<RoomStatusView>(`/v1/rooms/status-view?from=${dateFrom}&to=${dateTo}`);
}

/* ─── Rate Rules API ─── */
export interface RateRule {
  id: string;
  name: string;
  room_type: string;
  condition_type: string;
  condition_value: number;
  adjustment_type: string;
  adjustment_value: number;
  min_stay?: number;
  max_stay?: number;
  start_date?: string;
  end_date?: string;
  is_active: boolean;
  priority: number;
}

export interface RateRuleCreateRequest {
  name: string;
  room_type: string;
  condition_type: string;
  condition_value: number;
  adjustment_type: string;
  adjustment_value: number;
  min_stay?: number;
  max_stay?: number;
  start_date?: string;
  end_date?: string;
  priority?: number;
}

export async function getRateRules(): Promise<RateRule[]> {
  const data = await api<{ rules: RateRule[] }>("/v1/rate-rules");
  return data.rules || [];
}

export async function createRateRule(data: Partial<RateRuleCreateRequest>): Promise<RateRule> {
  return api<RateRule>("/v1/rate-rules", { method: "POST", body: JSON.stringify(data) });
}

export async function deleteRateRule(id: string): Promise<void> {
  return api<void>(`/v1/rate-rules/${id}`, { method: "DELETE" });
}

export async function calculateRate(params: { room_type: string; check_in: string; check_out: string }): Promise<{ total: number; breakdown: any[] }> {
  return api<{ total: number; breakdown: any[] }>("/v1/rate-rules/calculate", { method: "POST", body: JSON.stringify(params) });
}

/* ─── Audit Logs API ─── */
export interface AuditLog {
  id: string;
  action: string;
  entity_type: string;
  entity_id: string;
  user_id?: string;
  user_name?: string;
  user_email?: string;
  resource?: string;
  ip_address?: string;
  details?: Record<string, any>;
  created_at: string;
}

export interface AuditStats {
  total_actions: number;
  by_action: Record<string, number>;
  by_user: Record<string, number>;
  by_date: Record<string, number>;
  total_events?: number;
  today_events?: number;
  user_actions?: Record<string, number>;
  failed_actions?: number;
}

export async function getAuditLogs(): Promise<AuditLog[]> {
  const data = await api<{ logs: AuditLog[] }>("/v1/audit/logs");
  return data.logs || [];
}

export async function getAuditStats(): Promise<AuditStats> {
  return api<AuditStats>("/v1/audit/stats");
}

/* ─── Channel Manager API ─── */
export interface Channel {
  id: string;
  name: string;
  code: string;
  source?: string;
  display_name?: string;
  is_active: boolean;
  commission_pct?: number;
  commission_percent?: number;
  api_key?: string;
  last_sync_at?: string;
  last_sync_status?: string;
  sync_enabled: boolean;
}

export async function getChannels(): Promise<Channel[]> {
  const data = await api<{ channels: Channel[] }>("/v1/channels");
  return data.channels || [];
}

export async function updateChannel(id: string, data: Partial<Channel>): Promise<Channel> {
  return api<Channel>(`/v1/channels/${id}`, { method: "PATCH", body: JSON.stringify(data) });
}

export async function deleteChannel(id: string): Promise<void> {
  return api<void>(`/v1/channels/${id}`, { method: "DELETE" });
}

/* ─── Properties API ─── */
export interface Property {
  id: string;
  name: string;
  address?: string;
  city?: string;
  country?: string;
  phone?: string;
  email?: string;
  timezone?: string;
  currency?: string;
  is_active: boolean;
  room_count?: number;
}

export async function getProperties(): Promise<Property[]> {
  const data = await api<{ properties: Property[] }>("/v1/properties");
  return data.properties || [];
}

export async function createProperty(data: Partial<Property>): Promise<Property> {
  return api<Property>("/v1/properties", { method: "POST", body: JSON.stringify(data) });
}

/* ─── Competitors API ─── */
export interface CompetitorHotel {
  id: string;
  name: string;
  address?: string;
  city?: string;
  country?: string;
  star_rating?: number;
  is_active: boolean;
  rate_shop_url?: string;
}

export interface CompetitorRate {
  id: string;
  competitor_id: string;
  room_type: string;
  rate: number;
  currency: string;
  date: string;
  availability?: number;
  scraped_at: string;
}

export interface RateRecommendation {
  id: string;
  room_type: string;
  current_rate: number;
  recommended_rate: number;
  reason: string;
  confidence: number;
  expires_at?: string;
}

export interface RateShopConfig {
  id: string;
  is_active: boolean;
  frequency_hours: number;
  room_types: string[];
  notify_on_change: boolean;
}

export async function getCompetitors(): Promise<CompetitorHotel[]> {
  const data = await api<{ competitors: CompetitorHotel[] }>("/v1/competitors");
  return data.competitors || [];
}

export async function getCompetitorRates(competitorId?: string): Promise<CompetitorRate[]> {
  const qs = competitorId ? `?competitor_id=${competitorId}` : "";
  const data = await api<{ rates: CompetitorRate[] }>(`/v1/competitor-rates${qs}`);
  return data.rates || [];
}

export async function getRateRecommendations(): Promise<RateRecommendation[]> {
  const data = await api<{ recommendations: RateRecommendation[] }>("/v1/rate-recommendations");
  return data.recommendations || [];
}

export async function applyRateRecommendation(id: string): Promise<void> {
  return api<void>(`/v1/rate-recommendations/${id}/apply`, { method: "POST" });
}

export async function getRateShopConfig(): Promise<RateShopConfig> {
  return api<RateShopConfig>("/v1/rate-shop-config");
}

export async function updateRateShopConfig(data: Partial<RateShopConfig>): Promise<RateShopConfig> {
  return api<RateShopConfig>("/v1/rate-shop-config", { method: "PATCH", body: JSON.stringify(data) });
}

export async function createCompetitor(data: Partial<CompetitorHotel>): Promise<CompetitorHotel> {
  return api<CompetitorHotel>("/v1/competitors", { method: "POST", body: JSON.stringify(data) });
}

export async function deleteCompetitor(id: string): Promise<void> {
  return api<void>(`/v1/competitors/${id}`, { method: "DELETE" });
}
