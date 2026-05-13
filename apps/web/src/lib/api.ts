
/* ─── Upsell API ─── */

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
