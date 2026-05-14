/* ─── Missing Exports ─── */

export async function deleteVillaReservation(id: string): Promise<void> {
  return api(`/v1/villa-reservations/${id}`, { method: "DELETE" });
}

export interface UpsellOffer {
  id: string;
  tenant_id: string;
  name: string;
  description?: string;
  category: string;
  price: number;
  image_url?: string;
  is_active: boolean;
  inventory_count?: number;
  max_per_guest?: number;
  requires_approval: boolean;
  created_at: string;
  updated_at: string;
}

export interface UpsellPurchase {
  id: string;
  tenant_id: string;
  offer_id: string;
  offer_name: string;
  reservation_id: string;
  guest_name: string;
  room_number?: string;
  quantity: number;
  unit_price: number;
  total_price: number;
  status: "pending" | "approved" | "delivered" | "cancelled" | "refunded";
  requested_at: string;
  fulfilled_at?: string;
}

export async function getUpsellOffers(): Promise<UpsellOffer[]> {
  const data = await api<{ offers: UpsellOffer[] }>("/v1/upsells");
  return data.offers || [];
}

export async function getUpsellPurchases(): Promise<UpsellPurchase[]> {
  const data = await api<{ purchases: UpsellPurchase[] }>("/v1/upsells/purchases");
  return data.purchases || [];
}

export async function createUpsellOffer(payload: Omit<UpsellOffer, "id" | "tenant_id" | "created_at" | "updated_at">): Promise<UpsellOffer> {
  return api("/v1/upsells", { method: "POST", body: JSON.stringify(payload) });
}

export async function updateUpsellOffer(id: string, payload: Partial<UpsellOffer>): Promise<UpsellOffer> {
  return api(`/v1/upsells/${id}`, { method: "PATCH", body: JSON.stringify(payload) });
}

export async function deleteUpsellOffer(id: string): Promise<void> {
  return api(`/v1/upsells/${id}`, { method: "DELETE" });
}

export async function getTenantConfig(): Promise<TenantConfig> {
  return api("/v1/tenant/config");
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
  dates: {
    check_in: string;
    check_out: string;
    nights: number;
  };
  party: {
    adults: number;
    children: number;
  };
  status: string;
  source: string;
  financials: {
    subtotal: number;
    tax: number;
    total: number;
    paid: number;
    balance: number;
    deposit_paid?: number;
  };
  special_requests?: string;
  created_at: string;
  updated_at: string;
}

export async function getReservationDetail(id: string): Promise<ReservationDetail> {
  return api(`/v1/reservations/${id}/detail`);
}

export interface RoomStatusView {
  room_number: string;
  room_type: string;
  floor?: string;
  status: string;
  guest_name?: string;
  check_in?: string;
  check_out?: string;
  housekeeping_status?: string;
  is_vip: boolean;
  notes?: string;
}

/* ─── Guest Journey API ─── */

export interface GuestJourney {
  id: string;
  tenant_id: string;
  reservation_id: string;
  guest_name: string;
  name?: string;
  stage: string;
  status: string;
  trigger?: string;
  is_active?: boolean;
  steps?: Array<{ name: string; completed: boolean; channel?: string; id?: string; template_id?: string; delay_hours?: number; condition?: string; upsell_offer_id?: string; is_active?: boolean }>;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface JourneyExecution {
  id: string;
  journey_id: string;
  reservation_id: string;
  guest_name: string;
  guest_phone?: string;
  stage: string;
  status: string;
  current_step?: number;
  completed_steps?: number;
  total_steps?: number;
  started_at?: string;
  next_trigger_at?: string;
  executed_at?: string;
  completed_at?: string;
}

export async function getJourneys(): Promise<GuestJourney[]> {
  const data = await api<{ journeys: GuestJourney[] }>('/v1/journeys');
  return data.journeys || [];
}

export async function updateJourney(id: string, payload: Partial<GuestJourney>): Promise<GuestJourney> {
  return api(`/v1/journeys/${id}`, { method: 'PATCH', body: JSON.stringify(payload) });
}

export async function deleteJourney(id: string): Promise<void> {
  return api(`/v1/journeys/${id}`, { method: 'DELETE' });
}

export async function getJourneyExecutions(): Promise<JourneyExecution[]> {
  const data = await api<{ executions: JourneyExecution[] }>('/v1/journeys/executions');
  return data.executions || [];
}

export async function createJourney(payload: Omit<GuestJourney, "id" | "tenant_id" | "created_at" | "updated_at">): Promise<GuestJourney> {
  return api('/v1/journeys', { method: 'POST', body: JSON.stringify(payload) });
}

/* ─── Competitor & Rate Shop API ─── */

export interface Competitor {
  id: string;
  tenant_id: string;
  name: string;
  website?: string;
  address?: string;
  city?: string;
  country?: string;
  star_rating?: number;
  room_count?: number;
  last_scraped?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface RateShopConfig {
  id: string;
  tenant_id: string;
  enabled: boolean;
  frequency: string;
  frequency_hours?: number;
  lookahead_days?: number;
  auto_adjust?: boolean;
  max_adjustment_pct?: number;
  competitors: string[];
  room_types: string[];
  alert_threshold_pct: number;
  created_at: string;
  updated_at: string;
}

export async function getCompetitors(): Promise<Competitor[]> {
  const data = await api<{ competitors: Competitor[] }>('/v1/competitors');
  return data.competitors || [];
}

export async function createCompetitor(payload: Omit<Competitor, "id" | "tenant_id" | "created_at" | "updated_at">): Promise<Competitor> {
  return api('/v1/competitors', { method: 'POST', body: JSON.stringify(payload) });
}

export async function updateCompetitor(id: string, payload: Partial<Competitor>): Promise<Competitor> {
  return api(`/v1/competitors/${id}`, { method: 'PATCH', body: JSON.stringify(payload) });
}

export async function deleteCompetitor(id: string): Promise<void> {
  return api(`/v1/competitors/${id}`, { method: 'DELETE' });
}

export async function getRateShopConfig(): Promise<RateShopConfig> {
  return api('/v1/rate-shop/config');
}

export async function updateRateShopConfig(payload: Partial<RateShopConfig>): Promise<RateShopConfig> {
  return api('/v1/rate-shop/config', { method: 'PATCH', body: JSON.stringify(payload) });
}

/* ─── Rate Recommendations API ─── */

export interface RateRecommendation {
  id: string;
  room_type: string;
  current_rate: number;
  recommended_rate: number;
  competitor_rate: number;
  confidence: number;
  reason: string;
  factors?: string[];
  date: string;
  applied: boolean;
}

export async function getRateRecommendations(): Promise<RateRecommendation[]> {
  const data = await api<{ recommendations: RateRecommendation[] }>('/v1/rate-shop/recommendations');
  return data.recommendations || [];
}

export async function applyRateRecommendation(id: string): Promise<{ status: string }> {
  return api(`/v1/rate-shop/recommendations/${id}/apply`, { method: 'POST' });
}

/* ─── Competitor Rates API ─── */

export interface CompetitorRate {
  id: string;
  competitor_id: string;
  competitor_name: string;
  room_type: string;
  rate: number;
  is_promo?: boolean;
  availability?: number;
  min_stay?: number;
  source?: string;
  date: string;
  scraped_at: string;
}

export async function getCompetitorRates(): Promise<CompetitorRate[]> {
  const data = await api<{ rates: CompetitorRate[] }>('/v1/competitors/rates');
  return data.rates || [];
}

/* ─── Type aliases for competitor page ─── */
export type CompetitorHotel = Competitor;


/* ─── GuestJourneyExecution type alias ─── */
export type GuestJourneyExecution = JourneyExecution;

/* ─── IPTV Types ─── */
export interface IPTVChannel {
  id: string;
  number: number;
  name: string;
  category: string;
  language: string;
  is_premium: boolean;
  is_active: boolean;
}

export interface IPTVContent {
  id: string;
  title: string;
  category: string;
  type: "movie" | "series" | "live";
  duration?: number;
  is_active: boolean;
}

export interface IPTVRoomStatus {
  id: string;
  room_number: string;
  is_online: boolean;
  current_channel?: string;
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

/* ─── Room Daily Status (for room status view) ─── */
export interface RoomDailyStatus {
  room_number: string;
  room_type: string;
  status: string;
  guest_name?: string;
  check_in?: string;
  check_out?: string;
  rate?: number;
}

export interface RoomStatusDay {
  date: string;
  occupancy: number;
  arrivals: number;
  departures: number;
  stayovers: number;
  revenue: number;
  rooms: RoomDailyStatus[];
}

export async function getRoomStatusView(date?: string, view?: string): Promise<{ dates: RoomStatusDay[] }> {
  const params = new URLSearchParams();
  if (date) params.append("date", date);
  if (view) params.append("view", view);
  const query = params.toString() ? `?${params.toString()}` : "";
  return api(`/v1/rooms/status-view${query}`);
}
