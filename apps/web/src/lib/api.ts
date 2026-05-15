/* ─── Guest Journey Timeline API ─── */
export interface GuestJourneyEvent {
  id: string;
  reservation_id: string;
  event_type: string;
  event_data: Record<string, any>;
  occurred_at: string;
  created_by: string;
}

export interface GuestJourneyTimeline {
  reservation_id: string;
  guest_name: string;
  events: GuestJourneyEvent[];
}

export async function getGuestTimeline(reservationId: string): Promise<GuestJourneyTimeline> {
  return api<GuestJourneyTimeline>(`/v1/guest-journey/${reservationId}`);
}

export async function addGuestJourneyEvent(payload: { reservation_id: string; event_type: string; event_data?: Record<string, any>; created_by: string }): Promise<GuestJourneyEvent> {
  return api<GuestJourneyEvent>("/v1/guest-journey/events", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}
