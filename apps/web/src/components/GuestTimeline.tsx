// @ts-nocheck

"use client";

import { useEffect, useState } from "react";
import { Clock, CheckCircle2, Mail, DoorOpen, LogOut, Star, AlertTriangle, MessageSquare } from "lucide-react";
import { getGuestTimeline, addGuestJourneyEvent, GuestJourneyEvent } from "@/lib/api";

const eventIcons: Record<string, React.ElementType> = {
  booking_confirmed: CheckCircle2,
  pre_arrival_email_sent: Mail,
  check_in_completed: DoorOpen,
  mid_stay_check: MessageSquare,
  check_out_completed: LogOut,
  post_stay_review_requested: Star,
  post_stay_review_received: Star,
  complaint_raised: AlertTriangle,
  complaint_resolved: CheckCircle2,
};

const eventColors: Record<string, string> = {
  booking_confirmed: "text-emerald-400 bg-emerald-500/10",
  pre_arrival_email_sent: "text-sky-400 bg-sky-500/10",
  check_in_completed: "text-violet-400 bg-violet-500/10",
  mid_stay_check: "text-amber-400 bg-amber-500/10",
  check_out_completed: "text-rose-400 bg-rose-500/10",
  post_stay_review_requested: "text-nexus-400 bg-nexus-500/10",
  post_stay_review_received: "text-emerald-400 bg-emerald-500/10",
  complaint_raised: "text-rose-400 bg-rose-500/10",
  complaint_resolved: "text-emerald-400 bg-emerald-500/10",
};

export default function GuestTimeline({ reservationId, guestName }: { reservationId: string; guestName: string }) {
  const [timeline, setTimeline] = useState<GuestJourneyEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAdd, setShowAdd] = useState(false);
  const [eventType, setEventType] = useState("mid_stay_check");
  const [eventNote, setEventNote] = useState("");

  async function fetchTimeline() {
    setLoading(true);
    try {
      const data = await getGuestTimeline(reservationId);
      setTimeline(data.events);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    fetchTimeline();
  }, [reservationId]);

  async function handleAddEvent() {
    try {
      await addGuestJourneyEvent({
        reservation_id: reservationId,
        event_type: eventType,
        event_data: { note: eventNote },
        created_by: "staff",
      });
      setShowAdd(false);
      setEventNote("");
      fetchTimeline();
    } catch (e: any) {
      alert(e.message);
    }
  }

  if (loading) return <div className="text-sm text-slate-400">Loading timeline...</div>;

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-white">Guest Journey</h3>
        <button onClick={() => setShowAdd(true)} className="text-xs text-nexus-400 hover:text-nexus-300">+ Add Event</button>
      </div>

      {timeline.length === 0 ? (
        <div className="text-xs text-slate-500">No journey events yet.</div>
      ) : (
        <div className="relative space-y-4 pl-4">
          <div className="absolute left-1.5 top-2 bottom-2 w-px bg-slate-700" />
          {timeline.map((event) => {
            const Icon = eventIcons[event.event_type] || Clock;
            const colorClass = eventColors[event.event_type] || "text-slate-400 bg-slate-700";
            return (
              <div key={event.id} className="relative flex items-start gap-3">
                <div className={`relative z-10 flex h-6 w-6 shrink-0 items-center justify-center rounded-full ${colorClass}`}>
                  <Icon className="h-3 w-3" />
                </div>
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <span className="text-sm font-medium text-white capitalize">{event.event_type.replace(/_/g, " ")}</span>
                    <span className="text-xs text-slate-500">{new Date(event.occurred_at).toLocaleString()}</span>
                  </div>
                  {event.event_data?.note && (
                    <div className="text-xs text-slate-400">{event.event_data.note}</div>
                  )}
                  <div className="text-xs text-slate-500">by {event.created_by}</div>
                </div>
              </div>
            );
          })}
        </div>
      )}

      {showAdd && (
        <div className="rounded-lg border border-slate-700 bg-slate-800 p-3 space-y-2">
          <select className="input w-full text-xs" value={eventType} onChange={(e) => setEventType(e.target.value)}>
            <option value="mid_stay_check">Mid-stay Check</option>
            <option value="complaint_raised">Complaint Raised</option>
            <option value="complaint_resolved">Complaint Resolved</option>
            <option value="post_stay_review_requested">Review Requested</option>
          </select>
          <textarea className="input w-full text-xs" placeholder="Note (optional)" value={eventNote} onChange={(e) => setEventNote(e.target.value)} rows={2} />
          <div className="flex gap-2">
            <button onClick={() => setShowAdd(false)} className="btn-secondary text-xs">Cancel</button>
            <button onClick={handleAddEvent} className="btn-primary text-xs">Add</button>
          </div>
        </div>
      )}
    </div>
  );
}
