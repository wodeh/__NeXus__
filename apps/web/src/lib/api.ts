

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
