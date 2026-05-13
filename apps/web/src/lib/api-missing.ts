
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

/* ─── Missing Exports (continued) ─── */

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

export async function getRoomStatusView(): Promise<RoomStatusView[]> {
  const data = await api<{ rooms: RoomStatusView[] }>('/v1/rooms/status-view');
  return data.rooms || [];
}
