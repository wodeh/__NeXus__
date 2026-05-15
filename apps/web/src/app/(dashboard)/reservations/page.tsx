// @ts-nocheck
"use client";
import OverbookingWidget from "@/components/OverbookingWidget";

import { useState, useEffect, useMemo, useCallback } from "react";
import Link from "next/link";
import {
  Plus, Search, CalendarDays, User, ChevronRight, AlertTriangle, CheckCircle2, Clock,
  BedDouble, ArrowRightLeft, DoorOpen, MapPin, X, Save, Filter, Grid3X3, List,
  ChevronLeft, ChevronRight as ChevronRightIcon, GripVertical, RefreshCw,
} from "lucide-react";
import {
  Reservation, Room, WaitlistEntry,
  getReservations, createReservation, checkInReservation, checkOutReservation,
  cancelReservation, assignRoom, getRooms, createGroupReservation,
  getWaitlist, addToWaitlist, assignWaitlistRoom, cancelWaitlistEntry,
} from "@/lib/api";

const statusBadge: Record<string, string> = {
  confirmed: "badge-blue",
  checked_in: "badge-green",
  checked_out: "badge-amber",
  cancelled: "badge-red",
  no_show: "badge-red",
};

const statusColors: Record<string, string> = {
  occupied: "bg-rose-500",
  vacant_clean: "bg-emerald-500",
  vacant_dirty: "bg-amber-500",
  blocked: "bg-slate-500",
  maintenance: "bg-orange-500",
};

const resColors = [
  "bg-sky-500", "bg-violet-500", "bg-amber-500", "bg-rose-500",
  "bg-emerald-500", "bg-orange-500", "bg-cyan-500", "bg-pink-500",
];

/* ─── Helpers ─── */
function hashStringToIndex(str: string, max: number): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) - hash) + str.charCodeAt(i);
    hash |= 0;
  }
  return Math.abs(hash) % max;
}

function addDays(dateStr: string, days: number): string {
  const d = new Date(dateStr + "T00:00:00");
  d.setDate(d.getDate() + days);
  return d.toISOString().split("T")[0];
}

function diffDays(a: string, b: string): number {
  const d1 = new Date(a + "T00:00:00").getTime();
  const d2 = new Date(b + "T00:00:00").getTime();
  return Math.round((d2 - d1) / (1000 * 60 * 60 * 24));
}

function formatDateShort(dateStr: string): string {
  const d = new Date(dateStr + "T00:00:00");
  return d.toLocaleDateString("en-US", { weekday: "short", month: "short", day: "numeric" });
}

function normalizeDate(d: string): string {
  return d.split("T")[0];
}

function isDateInRange(date: string, start: string, end: string): boolean {
  const t = normalizeDate(date);
  const s = normalizeDate(start);
  const e = normalizeDate(end);
  return t >= s && t < e;
}

/* ─── Main Page ─── */
export default function ReservationsPage() {
  const [view, setView] = useState<"list" | "tapechart" | "calendar">("list");
  const [reservations, setReservations] = useState<Reservation[]>([]);
  const [rooms, setRooms] = useState<Room[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [showQuickBook, setShowQuickBook] = useState(false);
  const [showGroupBook, setShowGroupBook] = useState(false);
  const [showWalkIn, setShowWalkIn] = useState(false);

  const today = useMemo(() => new Date().toISOString().split("T")[0], []);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [res, rms] = await Promise.all([getReservations(), getRooms()]);
      // Assign colors to reservations for tapechart
      const colored = res.map((r) => ({ ...r, color: r.color || resColors[hashStringToIndex(r.id, resColors.length)] }));
      setReservations(colored);
      setRooms(rms);
    } catch (e: any) {
      setError(e.message || "Failed to load data");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Reservations</h1>
          <p className="text-sm text-slate-400">Bookings, check-ins, room assignments</p>
        </div>
        <div className="flex items-center gap-2">
          {/* View Tabs */}
          <div className="flex rounded-lg border border-slate-700 bg-slate-800">
            <button onClick={() => setView("list")} className={`flex items-center gap-1.5 px-3 py-2 text-xs font-medium ${view === "list" ? "bg-slate-700 text-white" : "text-slate-400 hover:text-white"}`}>
              <List className="h-3.5 w-3.5" /> List
            </button>
            <Link href="/tapechart" className={`flex items-center gap-1.5 px-3 py-2 text-xs font-medium ${view === "tapechart" ? "bg-slate-700 text-white" : "text-slate-400 hover:text-white"}`}>
              <Grid3X3 className="h-3.5 w-3.5" /> Tapechart
            </Link>
            <button onClick={() => setView("calendar")} className={`flex items-center gap-1.5 px-3 py-2 text-xs font-medium ${view === "calendar" ? "bg-slate-700 text-white" : "text-slate-400 hover:text-white"}`}>
              <CalendarDays className="h-3.5 w-3.5" /> Calendar
            </button>
            <button onClick={() => setView("waitlist")} className={`flex items-center gap-1.5 px-3 py-2 text-xs font-medium ${view === "waitlist" ? "bg-slate-700 text-white" : "text-slate-400 hover:text-white"}`}>
              <Clock className="h-3.5 w-3.5" /> Waitlist
            </button>
            <button onClick={() => setView("readiness")} className={`flex items-center gap-1.5 px-3 py-2 text-xs font-medium ${view === "readiness" ? "bg-slate-700 text-white" : "text-slate-400 hover:text-white"}`}>
              <CheckCircle2 className="h-3.5 w-3.5" /> Readiness
            </button>
          </div>
          <button onClick={fetchData} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white" title="Refresh">
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
          </button>
          <button onClick={() => setShowQuickBook(true)} className="btn-primary">
            <Plus className="h-4 w-4" />
            Quick Book
          </button>
          <button onClick={() => setShowGroupBook(true)} className="btn-primary bg-violet-600 hover:bg-violet-500">
            <Plus className="h-4 w-4" />
            Group Book
          </button>
          <button onClick={() => setShowWalkIn(true)} className="btn-primary bg-emerald-600 hover:bg-emerald-500">
            <DoorOpen className="h-4 w-4" />
            Walk-in
          </button>
        </div>
      </div>

      {error && (
        <div className="rounded-lg border border-rose-500/20 bg-rose-500/5 px-4 py-3 text-sm text-rose-400">
          {error}
        </div>
      )}

      {loading ? (
        <div className="flex items-center justify-center py-20 text-slate-500">
          <RefreshCw className="mr-2 h-5 w-5 animate-spin" /> Loading...
        </div>
      ) : (
        <>
          {view === "list" && <ListView reservations={reservations} rooms={rooms} today={today} onRefresh={fetchData} />}
          {view === "calendar" && <CalendarView reservations={reservations} rooms={rooms} today={today} onRefresh={fetchData} />}
          {view === "waitlist" && <WaitlistView rooms={rooms} onRefresh={fetchData} />}
          {view === "readiness" && <ReadinessView reservations={reservations} rooms={rooms} today={today} onRefresh={fetchData} />}
        </>
      )}

      {/* Quick Book Modal */}
      {showQuickBook && (
        <QuickBookModal
          rooms={rooms}
          onClose={() => setShowQuickBook(false)}
          onCreate={async (payload) => {
            try {
              await createReservation(payload);
              fetchData();
              setShowQuickBook(false);
            } catch (e: any) {
              alert(e.message);
            }
          }}
        />
      )}

      {/* Group Book Modal */}
      {showGroupBook && (
        <GroupBookingModal
          rooms={rooms}
          onClose={() => setShowGroupBook(false)}
          onCreate={async (payload) => {
            try {
              await createGroupReservation(payload);
              fetchData();
              setShowGroupBook(false);
            } catch (e: any) {
              alert(e.message);
            }
          }}
        />
      )}

      {/* Walk-in Modal */}
      {showWalkIn && (
        <WalkInModal
          rooms={rooms}
          onClose={() => setShowWalkIn(false)}
          onCreate={async (payload) => {
            try {
              const res = await createReservation(payload);
              await checkInReservation(res.id);
              fetchData();
              setShowWalkIn(false);
            } catch (e: any) {
              alert(e.message);
            }
          }}
        />
      )}
    </div>
  );
}

function WalkInModal({ rooms, onClose, onCreate }: { rooms: Room[]; onClose: () => void; onCreate: (payload: any) => Promise<void> }) {
  const [guestName, setGuestName] = useState("");
  const [roomNumber, setRoomNumber] = useState("");
  const [nights, setNights] = useState(1);
  const [adults, setAdults] = useState(1);
  const [loading, setLoading] = useState(false);
  const today = new Date().toISOString().split("T")[0];
  const checkOut = addDays(today, nights);

  const vacantRooms = rooms.filter((r) => r.status === "vacant_clean");

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-sm rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-bold text-white">Walk-in Express</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>
        <div className="space-y-3">
          <input className="input w-full" placeholder="Guest name *" value={guestName} onChange={(e) => setGuestName(e.target.value)} autoFocus />
          <select className="input w-full" value={roomNumber} onChange={(e) => setRoomNumber(e.target.value)}>
            <option value="">Select vacant room...</option>
            {vacantRooms.map((r) => (
              <option key={r.number} value={r.number}>{r.number} · {r.type} · Floor {r.floor}</option>
            ))}
          </select>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Nights</label>
              <input type="number" min={1} max={30} className="input w-full" value={nights} onChange={(e) => setNights(parseInt(e.target.value) || 1)} />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Adults</label>
              <input type="number" min={1} max={6} className="input w-full" value={adults} onChange={(e) => setAdults(parseInt(e.target.value) || 1)} />
            </div>
          </div>
          <div className="rounded-lg border border-emerald-500/20 bg-emerald-500/5 p-2 text-xs text-emerald-400">
            <CheckCircle2 className="mr-1 inline h-3 w-3" />
            Auto check-in today ({today}) · checkout {checkOut}
          </div>
        </div>
        <div className="mt-4 flex justify-end gap-2">
          <button onClick={onClose} className="btn-secondary text-xs">Cancel</button>
          <button
            onClick={async () => {
              if (!guestName) return alert("Guest name required");
              if (!roomNumber) return alert("Select a room");
              setLoading(true);
              await onCreate({
                guest_name: guestName,
                room_number: roomNumber,
                room_type: vacantRooms.find((r) => r.number === roomNumber)?.type || "Standard",
                check_in: today,
                check_out: checkOut,
                adults,
                source: "walk_in",
              });
              setLoading(false);
            }}
            disabled={loading}
            className="btn-primary bg-emerald-600 hover:bg-emerald-500 text-xs disabled:opacity-50"
          >
            <DoorOpen className="mr-1 h-4 w-4" />
            {loading ? "Checking in..." : "Check In Now"}
          </button>
        </div>
      </div>
    </div>
  );
}

function WaitlistView({ rooms, onRefresh }: { rooms: Room[]; onRefresh: () => void }) {
  const [entries, setEntries] = useState<WaitlistEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAdd, setShowAdd] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  useEffect(() => {
    loadWaitlist();
  }, []);

  async function loadWaitlist() {
    setLoading(true);
    try {
      const data = await getWaitlist();
      setEntries(data);
    } catch (e: any) {
      alert(e.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <div className="text-sm text-slate-400">
          {entries.filter((e) => e.status === "waiting").length} waiting
        </div>
        <button onClick={() => setShowAdd(true)} className="btn-primary text-xs">
          <Plus className="h-4 w-4" /> Add to Waitlist
        </button>
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-20 text-slate-500">
          <RefreshCw className="mr-2 h-5 w-5 animate-spin" /> Loading...
        </div>
      ) : entries.length === 0 ? (
        <div className="rounded-lg border border-slate-700 bg-slate-800 p-8 text-center text-sm text-slate-400">
          No guests on the waitlist.
        </div>
      ) : (
        <div className="space-y-2">
          {entries.map((entry) => (
            <div key={entry.id} className="flex items-center justify-between rounded-lg border border-slate-700 bg-slate-800 p-3">
              <div className="flex items-center gap-3">
                <div className="flex h-8 w-8 items-center justify-center rounded-full bg-amber-500/10">
                  <Clock className="h-4 w-4 text-amber-400" />
                </div>
                <div>
                  <div className="text-sm font-medium text-white">{entry.guest_name}</div>
                  <div className="text-xs text-slate-400">
                    {entry.requested_check_in} → {entry.requested_check_out}
                    {entry.room_type && ` · ${entry.room_type}`}
                    {entry.adults > 0 && ` · ${entry.adults} adults`}
                    {entry.children > 0 && ` · ${entry.children} children`}
                  </div>
                </div>
              </div>
              <div className="flex items-center gap-2">
                {entry.assigned_room_number ? (
                  <span className="rounded bg-emerald-500/10 px-2 py-1 text-xs text-emerald-400">
                    Room {entry.assigned_room_number} suggested
                  </span>
                ) : (
                  <span className="rounded bg-slate-700 px-2 py-1 text-xs text-slate-400">No room available</span>
                )}
                {entry.assigned_room_number && (
                  <button
                    onClick={async () => {
                      setActionLoading(entry.id);
                      try {
                        await assignWaitlistRoom(entry.id, entry.assigned_room_number!);
                        loadWaitlist();
                        onRefresh();
                      } catch (e: any) {
                        alert(e.message);
                      } finally {
                        setActionLoading(null);
                      }
                    }}
                    disabled={actionLoading === entry.id}
                    className="btn-primary text-xs disabled:opacity-50"
                  >
                    {actionLoading === entry.id ? "Assigning..." : "Assign Room"}
                  </button>
                )}
                <button
                  onClick={async () => {
                    if (!confirm("Cancel this waitlist entry?")) return;
                    try {
                      await cancelWaitlistEntry(entry.id);
                      loadWaitlist();
                    } catch (e: any) {
                      alert(e.message);
                    }
                  }}
                  className="rounded border border-rose-500/20 px-2 py-1 text-xs text-rose-400 hover:bg-rose-500/10"
                >
                  Cancel
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {showAdd && (
        <WaitlistAddModal
          rooms={rooms}
          onClose={() => setShowAdd(false)}
          onAdd={async (payload) => {
            try {
              await addToWaitlist(payload);
              loadWaitlist();
              setShowAdd(false);
            } catch (e: any) {
              alert(e.message);
            }
          }}
        />
      )}
    </div>
  );
}

function WaitlistAddModal({
  rooms,
  onClose,
  onAdd,
}: {
  rooms: Room[];
  onClose: () => void;
  onAdd: (payload: any) => Promise<void>;
}) {
  const [guestName, setGuestName] = useState("");
  const [email, setEmail] = useState("");
  const [phone, setPhone] = useState("");
  const [roomType, setRoomType] = useState("");
  const [checkIn, setCheckIn] = useState("");
  const [checkOut, setCheckOut] = useState("");
  const [adults, setAdults] = useState(2);
  const [children, setChildren] = useState(0);
  const [priority, setPriority] = useState(0);
  const [notes, setNotes] = useState("");
  const [loading, setLoading] = useState(false);

  const roomTypes = Array.from(new Set(rooms.map((r) => r.type)));

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-bold text-white">Add to Waitlist</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white">
            <X className="h-5 w-5" />
          </button>
        </div>
        <div className="space-y-3">
          <input className="input w-full" placeholder="Guest name *" value={guestName} onChange={(e) => setGuestName(e.target.value)} />
          <div className="grid grid-cols-2 gap-3">
            <input className="input w-full" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} />
            <input className="input w-full" placeholder="Phone" value={phone} onChange={(e) => setPhone(e.target.value)} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <input type="date" className="input w-full" value={checkIn} onChange={(e) => setCheckIn(e.target.value)} />
            <input type="date" className="input w-full" value={checkOut} onChange={(e) => setCheckOut(e.target.value)} />
          </div>
          <select className="input w-full" value={roomType} onChange={(e) => setRoomType(e.target.value)}>
            <option value="">Any room type</option>
            {roomTypes.map((t) => (
              <option key={t} value={t}>{t}</option>
            ))}
          </select>
          <div className="grid grid-cols-3 gap-3">
            <input type="number" min={1} max={10} className="input w-full" value={adults} onChange={(e) => setAdults(parseInt(e.target.value) || 1)} placeholder="Adults" />
            <input type="number" min={0} max={10} className="input w-full" value={children} onChange={(e) => setChildren(parseInt(e.target.value) || 0)} placeholder="Children" />
            <input type="number" min={0} max={100} className="input w-full" value={priority} onChange={(e) => setPriority(parseInt(e.target.value) || 0)} placeholder="Priority" />
          </div>
          <textarea className="input w-full" rows={2} placeholder="Notes" value={notes} onChange={(e) => setNotes(e.target.value)} />
        </div>
        <div className="mt-4 flex justify-end gap-2">
          <button onClick={onClose} className="btn-secondary text-xs">Cancel</button>
          <button
            onClick={async () => {
              if (!guestName || !checkIn || !checkOut) return alert("Guest name, check in, and check out are required");
              setLoading(true);
              await onAdd({
                guest_name: guestName,
                email,
                phone,
                room_type: roomType,
                requested_check_in: checkIn,
                requested_check_out: checkOut,
                adults,
                children,
                priority,
                notes,
              });
              setLoading(false);
            }}
            disabled={loading}
            className="btn-primary text-xs disabled:opacity-50"
          >
            {loading ? "Adding..." : "Add to Waitlist"}
          </button>
        </div>
      </div>
    </div>
  );
}

/* ─── Readiness View ─── */
function ReadinessView({ reservations, rooms, today, onRefresh }: { reservations: Reservation[]; rooms: Room[]; today: string; onRefresh: () => void }) {
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const arrivals = reservations.filter(
    (r) => r.status === "confirmed" && (r.check_in === today || r.check_in === addDays(today, 1))
  );

  function isReady(r: Reservation) {
    const roomAssigned = !!r.room_number;
    const roomClean = rooms.find((room) => room.number === r.room_number)?.status === "vacant_clean";
    const hasDeposit = (r.deposit_paid || 0) > 0;
    const specialAck = !r.special_requests || r.special_requests_acknowledged;
    return roomAssigned && roomClean && hasDeposit && specialAck;
  }

  function readyItems(r: Reservation) {
    const items = [];
    if (r.room_number) items.push("Room assigned");
    if (rooms.find((room) => room.number === r.room_number)?.status === "vacant_clean") items.push("Room clean");
    if ((r.deposit_paid || 0) > 0) items.push("Deposit paid");
    if (!r.special_requests || r.special_requests_acknowledged) items.push("Requests OK");
    return items;
  }

  function missingItems(r: Reservation) {
    const items = [];
    if (!r.room_number) items.push("Assign room");
    else if (rooms.find((room) => room.number === r.room_number)?.status !== "vacant_clean") items.push("Room not clean");
    if ((r.deposit_paid || 0) === 0) items.push("No deposit");
    if (r.special_requests && !r.special_requests_acknowledged) items.push("Acknowledge requests");
    return items;
  }

  const todayArrivals = arrivals.filter((r) => r.check_in === today);
  const tomorrowArrivals = arrivals.filter((r) => r.check_in === addDays(today, 1));
  const readyCount = todayArrivals.filter(isReady).length;

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-3 gap-3">
        <StatCard label="Today's Arrivals" value={todayArrivals.length} icon={User} color="sky" />
        <StatCard label="Ready" value={readyCount} icon={CheckCircle2} color="emerald" />
        <StatCard label="Needs Attention" value={todayArrivals.length - readyCount} icon={AlertTriangle} color={todayArrivals.length - readyCount > 0 ? "rose" : "slate"} alert={todayArrivals.length - readyCount > 0} />
      </div>

      {todayArrivals.length === 0 && tomorrowArrivals.length === 0 ? (
        <div className="rounded-lg border border-slate-700 bg-slate-800 p-8 text-center text-sm text-slate-400">No upcoming arrivals.</div>
      ) : (
        <div className="space-y-4">
          {todayArrivals.length > 0 && (
            <div>
              <h3 className="mb-2 text-sm font-semibold text-white">Today</h3>
              <div className="space-y-2">
                {todayArrivals.map((r) => (
                  <div key={r.id} className="flex items-center justify-between rounded-lg border border-slate-700 bg-slate-800 p-3">
                    <div className="flex items-center gap-3">
                      <div className={`flex h-8 w-8 items-center justify-center rounded-full ${isReady(r) ? "bg-emerald-500/10" : "bg-amber-500/10"}`}>
                        {isReady(r) ? <CheckCircle2 className="h-4 w-4 text-emerald-400" /> : <AlertTriangle className="h-4 w-4 text-amber-400" />}
                      </div>
                      <div>
                        <div className="text-sm font-medium text-white">{r.guest_name}</div>
                        <div className="text-xs text-slate-400">
                          {r.room_number ? `Room ${r.room_number}` : "No room assigned"} · {r.room_type} · {r.adults} adults
                        </div>
                        <div className="mt-1 flex flex-wrap gap-1">
                          {readyItems(r).map((item) => (
                            <span key={item} className="rounded bg-emerald-500/10 px-1.5 py-0.5 text-[10px] text-emerald-400">{item}</span>
                          ))}
                          {missingItems(r).map((item) => (
                            <span key={item} className="rounded bg-amber-500/10 px-1.5 py-0.5 text-[10px] text-amber-400">{item}</span>
                          ))}
                        </div>
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      {r.room_number && (
                        <button
                          onClick={async () => {
                            setActionLoading(r.id);
                            try {
                              await checkInReservation(r.id);
                              onRefresh();
                            } catch (e: any) {
                              alert(e.message);
                            } finally {
                              setActionLoading(null);
                            }
                          }}
                          disabled={actionLoading === r.id}
                          className="btn-primary bg-emerald-600 hover:bg-emerald-500 text-xs disabled:opacity-50"
                        >
                          {actionLoading === r.id ? "Checking in..." : "Check In"}
                        </button>
                      )}
                      {!r.room_number && (
                        <Link href={`/tapechart?assign=${r.id}`} className="btn-secondary text-xs">Assign Room</Link>
                      )}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {tomorrowArrivals.length > 0 && (
            <div>
              <h3 className="mb-2 text-sm font-semibold text-slate-300">Tomorrow</h3>
              <div className="space-y-2">
                {tomorrowArrivals.map((r) => (
                  <div key={r.id} className="flex items-center justify-between rounded-lg border border-slate-700 bg-slate-800 p-3 opacity-70">
                    <div className="flex items-center gap-3">
                      <div className={`flex h-8 w-8 items-center justify-center rounded-full ${isReady(r) ? "bg-emerald-500/10" : "bg-amber-500/10"}`}>
                        {isReady(r) ? <CheckCircle2 className="h-4 w-4 text-emerald-400" /> : <AlertTriangle className="h-4 w-4 text-amber-400" />}
                      </div>
                      <div>
                        <div className="text-sm font-medium text-white">{r.guest_name}</div>
                        <div className="text-xs text-slate-400">
                          {r.room_number ? `Room ${r.room_number}` : "No room assigned"} · {r.room_type}
                        </div>
                        <div className="mt-1 flex flex-wrap gap-1">
                          {readyItems(r).map((item) => (
                            <span key={item} className="rounded bg-emerald-500/10 px-1.5 py-0.5 text-[10px] text-emerald-400">{item}</span>
                          ))}
                          {missingItems(r).map((item) => (
                            <span key={item} className="rounded bg-amber-500/10 px-1.5 py-0.5 text-[10px] text-amber-400">{item}</span>
                          ))}
                        </div>
                      </div>
                    </div>
                    <span className="text-xs text-slate-500">Arrives tomorrow</span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

/* ─── List View ─── */
function ListView({ reservations, rooms, today, onRefresh }: { reservations: Reservation[]; rooms: Room[]; today: string; onRefresh: () => void }) {
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [dateFilter, setDateFilter] = useState<string>("all");
  const [showAssignRoom, setShowAssignRoom] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const filtered = reservations.filter((r) => {
    const matchesSearch =
      r.guest_name.toLowerCase().includes(search.toLowerCase()) ||
      (r.room_number || "").includes(search) ||
      (r.email || "").toLowerCase().includes(search.toLowerCase());
    const matchesStatus = statusFilter === "all" || r.status === statusFilter;
    const matchesDate =
      dateFilter === "all"
        ? true
        : dateFilter === "arrivals"
        ? r.check_in === today && r.status !== "checked_in"
        : dateFilter === "departures"
        ? r.check_out === today && r.status === "checked_in"
        : dateFilter === "in_house"
        ? r.status === "checked_in"
        : dateFilter === "upcoming"
        ? r.status === "confirmed"
        : true;
    return matchesSearch && matchesStatus && matchesDate;
  });

  const stats = {
    total: reservations.length,
    inHouse: reservations.filter((r) => r.status === "checked_in").length,
    arrivalsToday: reservations.filter((r) => r.check_in === today && r.status !== "checked_in").length,
    departuresToday: reservations.filter((r) => r.check_out === today && r.status === "checked_in").length,
    unassigned: reservations.filter((r) => r.status === "confirmed" && !r.room_number).length,
    revenue: reservations.filter((r) => r.status === "checked_in").reduce((s, r) => s + r.total, 0),
  };

  async function doAction(id: string, action: () => Promise<unknown>) {
    setActionLoading(id);
    try {
      await action();
      onRefresh();
    } catch (e: any) {
      alert(e.message);
    } finally {
      setActionLoading(null);
    }
  }

  return (
    <div className="space-y-4">
      {/* Stats */}
      <div className="grid grid-cols-6 gap-3">
        <StatCard label="Total" value={stats.total} icon={BedDouble} color="sky" />
        <StatCard label="In House" value={stats.inHouse} icon={CheckCircle2} color="emerald" />
        <StatCard label="Arrivals Today" value={stats.arrivalsToday} icon={ArrowRightLeft} color="violet" />
        <StatCard label="Departures Today" value={stats.departuresToday} icon={DoorOpen} color="amber" />
        <StatCard label="Unassigned" value={stats.unassigned} icon={MapPin} color="rose" alert={stats.unassigned > 0} />
        <StatCard label="Revenue" value={`$${stats.revenue.toLocaleString()}`} icon={Clock} color="nexus" />
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
          <input className="input w-full pl-9" placeholder="Search guest, room, email..." value={search} onChange={(e) => setSearch(e.target.value)} />
        </div>
        <div className="flex gap-1">
          {["all", "arrivals", "departures", "in_house", "upcoming"].map((f) => (
            <button key={f} onClick={() => setDateFilter(f)}
              className={`rounded-lg px-3 py-2 text-xs font-medium transition-colors ${dateFilter === f ? "bg-nexus-500 text-white" : "bg-slate-800 text-slate-400 hover:bg-slate-700 hover:text-white"}`}>
              {f === "all" ? "All" : f === "in_house" ? "In House" : f.charAt(0).toUpperCase() + f.slice(1).replace("_", " ")}
            </button>
          ))}
        </div>
        <select value={statusFilter} onChange={(e) => setStatusFilter(e.target.value)} className="input text-xs">
          <option value="all">All Statuses</option>
          {["confirmed", "checked_in", "checked_out", "cancelled", "no_show"].map((s) => (
            <option key={s} value={s}>{s.replace("_", " ")}</option>
          ))}
        </select>
      </div>

      {/* Overbooking Warning */}
      {stats.unassigned > 0 && (
        <div className="flex items-center gap-2 rounded-lg border border-rose-500/20 bg-rose-500/5 px-4 py-3">
          <AlertTriangle className="h-4 w-4 text-rose-400" />
          <span className="text-sm text-rose-400">{stats.unassigned} reservation{stats.unassigned > 1 ? "s" : ""} confirmed but no room assigned</span>
        </div>
      )}

      {/* Table */}
      <div className="card overflow-hidden p-0">
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-800 text-xs uppercase text-slate-400">
            <tr>
              <th className="px-4 py-3">Guest</th>
              <th className="px-4 py-3">Room</th>
              <th className="px-4 py-3">Dates</th>
              <th className="px-4 py-3">Nights</th>
              <th className="px-4 py-3">Total</th>
              <th className="px-4 py-3">Balance</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            {filtered.map((r) => {
              const nights = diffDays(r.check_in, r.check_out);
              return (
                <tr key={r.id} className="hover:bg-slate-800/50">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-3">
                      <div className={`flex h-8 w-8 items-center justify-center rounded-full ${r.vip ? "bg-amber-500/10" : "bg-nexus-500/10"}`}>
                        <User className={`h-4 w-4 ${r.vip ? "text-amber-400" : "text-nexus-400"}`} />
                      </div>
                      <div>
                        <span className="font-medium text-white">{r.guest_name}</span>
                        {r.vip && <span className="ml-2 rounded bg-amber-500/10 px-1.5 py-0.5 text-[10px] text-amber-400">VIP</span>}
                        {r.group_name && <span className="ml-1 rounded bg-violet-500/10 px-1.5 py-0.5 text-[10px] text-violet-400">{r.group_name}</span>}
                        {r.email && <p className="text-[10px] text-slate-500">{r.email}</p>}
                      </div>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    {r.room_number ? (
                      <span className="rounded bg-slate-700/50 px-2 py-0.5 text-xs text-white">{r.room_number}</span>
                    ) : (
                      <button onClick={() => setShowAssignRoom(r.id)}
                        className="flex items-center gap-1 rounded bg-rose-500/10 px-2 py-0.5 text-xs text-rose-400 hover:bg-rose-500/20">
                        <MapPin className="h-3 w-3" /> Assign Room
                      </button>
                    )}
                  </td>
                  <td className="px-4 py-3 text-slate-300">
                    <div className="flex items-center gap-1 text-xs">
                      <CalendarDays className="h-3 w-3 text-slate-500" />
                      {r.check_in} → {r.check_out}
                    </div>
                  </td>
                  <td className="px-4 py-3 text-slate-300">{nights}</td>
                  <td className="px-4 py-3 text-slate-300">${r.total}</td>
                  <td className="px-4 py-3">
                    <span className={r.balance > 0 ? "text-amber-400" : "text-emerald-400"}>${r.balance}</span>
                  </td>
                  <td className="px-4 py-3"><span className={statusBadge[r.status]}>{r.status.replace("_", " ")}</span></td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-1">
                      {r.status === "confirmed" && (
                        <>
                          <button onClick={() => doAction(r.id, () => checkInReservation(r.id))}
                            disabled={actionLoading === r.id}
                            className="rounded bg-emerald-500/10 px-2 py-1 text-[10px] text-emerald-400 hover:bg-emerald-500/20 disabled:opacity-50">
                            {actionLoading === r.id ? "..." : "Check In"}
                          </button>
                          <button onClick={() => doAction(r.id, () => cancelReservation(r.id))}
                            disabled={actionLoading === r.id}
                            className="rounded bg-rose-500/10 px-2 py-1 text-[10px] text-rose-400 hover:bg-rose-500/20 disabled:opacity-50">
                            Cancel
                          </button>
                        </>
                      )}
                      {r.status === "checked_in" && (
                        <button onClick={() => doAction(r.id, () => checkOutReservation(r.id))}
                          disabled={actionLoading === r.id}
                          className="rounded bg-amber-500/10 px-2 py-1 text-[10px] text-amber-400 hover:bg-amber-500/20 disabled:opacity-50">
                          {actionLoading === r.id ? "..." : "Check Out"}
                        </button>
                      )}
                      <Link href={`/reservations/${r.id}`} className="text-slate-400 hover:text-white">
                        <ChevronRight className="h-4 w-4" />
                      </Link>
                    </div>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
        {filtered.length === 0 && <p className="py-8 text-center text-sm text-slate-500">No reservations found.</p>}
      </div>

      {/* Assign Room Modal */}
      {showAssignRoom && (
        <AssignRoomModal
          resId={showAssignRoom}
          rooms={rooms}
          onClose={() => setShowAssignRoom(null)}
          onAssign={async (roomNumber) => {
            setActionLoading(showAssignRoom);
            try {
              await assignRoom(showAssignRoom, roomNumber);
              onRefresh();
              setShowAssignRoom(null);
            } catch (e: any) {
              alert(e.message);
            } finally {
              setActionLoading(null);
            }
          }}
        />
      )}
    </div>
  );
}

/* ─── List View ─── */
function TapechartView({ reservations, rooms, today, onRefresh }: { reservations: Reservation[]; rooms: Room[]; today: string; onRefresh: () => void }) {
  const [startDate, setStartDate] = useState(today);
  const [dayCount, setDayCount] = useState(14);
  const [filterFloor, setFilterFloor] = useState<string>("all");
  const [filterType, setFilterType] = useState<string>("all");
  const [searchRoom, setSearchRoom] = useState("");
  const [draggingRes, setDraggingRes] = useState<Reservation | null>(null);
  const [hoveredCell, setHoveredCell] = useState<{ room: string; date: string } | null>(null);
  const [showNewRes, setShowNewRes] = useState<{ room: string; date: string } | null>(null);
  const [selectedRes, setSelectedRes] = useState<Reservation | null>(null);
  const [actionLoading, setActionLoading] = useState(false);

  const dates = useMemo(() => Array.from({ length: dayCount }, (_, i) => addDays(startDate, i)), [startDate, dayCount]);

  const filteredRooms = useMemo(() => {
    return rooms.filter((r) => {
      const matchFloor = filterFloor === "all" || r.floor === filterFloor;
      const matchType = filterType === "all" || r.type === filterType;
      const matchSearch = !searchRoom || r.number.includes(searchRoom) || r.type.toLowerCase().includes(searchRoom.toLowerCase());
      return matchFloor && matchType && matchSearch;
    });
  }, [rooms, filterFloor, filterType, searchRoom]);

  const floors = useMemo(() => [...new Set(rooms.map((r) => r.floor))].sort(), [rooms]);
  const types = useMemo(() => [...new Set(rooms.map((r) => r.type))].sort(), [rooms]);

  const getReservationsForRoomDate = useCallback((roomNumber: string, date: string) => {
    return reservations.filter((r) => {
      if (r.room_number !== roomNumber) return false;
      return isDateInRange(date, r.check_in, r.check_out);
    });
  }, [reservations]);

  const getSpanForRes = useCallback((res: Reservation, roomNumber: string, date: string) => {
    if (res.room_number !== roomNumber) return { span: 0, start: false };
    const resStart = normalizeDate(res.check_in);
    const resEnd = normalizeDate(res.check_out);
    const gridStart = startDate;
    const gridEnd = addDays(startDate, dayCount);
    if (resEnd <= gridStart || resStart >= gridEnd) return { span: 0, start: false };
    const visibleStart = resStart < gridStart ? gridStart : resStart;
    const isVisibleStart = date === visibleStart;
    if (!isVisibleStart) return { span: 0, start: false };
    const overlapStart = new Date(Math.max(new Date(gridStart + "T00:00:00").getTime(), new Date(resStart + "T00:00:00").getTime()));
    const overlapEnd = new Date(Math.min(new Date(gridEnd + "T00:00:00").getTime(), new Date(resEnd + "T00:00:00").getTime()));
    const days = Math.round((overlapEnd.getTime() - overlapStart.getTime()) / (1000 * 60 * 60 * 24));
    return { span: Math.max(0, days), start: true };
  }, [startDate, dayCount]);

  const handlePrev = () => setStartDate((d) => addDays(d, -7));
  const handleNext = () => setStartDate((d) => addDays(d, 7));
  const handleToday = () => setStartDate(today);

  const handleCellClick = (room: Room, date: string) => {
    const existing = getReservationsForRoomDate(room.number, date);
    if (existing.length > 0) {
      setSelectedRes(existing[0]);
    } else if (!draggingRes) {
      setShowNewRes({ room: room.number, date });
    }
  };

  const handleDrop = async (roomNumber: string, date: string) => {
    if (!draggingRes) return;
    setActionLoading(true);
    try {
      await assignRoom(draggingRes.id, roomNumber);
      onRefresh();
    } catch (e: any) {
      alert(e.message);
    } finally {
      setDraggingRes(null);
      setActionLoading(false);
    }
  };

  const occupancyStats = useMemo(() => {
    const totalRooms = filteredRooms.length;
    const occupiedToday = filteredRooms.filter((r) => r.status === "occupied").length;
    const availableToday = filteredRooms.filter((r) => r.status === "vacant_clean").length;
    const departingToday = reservations.filter((r) => normalizeDate(r.check_out) === today && r.status === "checked_in").length;
    const arrivingToday = reservations.filter((r) => normalizeDate(r.check_in) === today && r.status !== "checked_in").length;
    return { totalRooms, occupiedToday, availableToday, departingToday, arrivingToday, occupancyRate: totalRooms ? Math.round((occupiedToday / totalRooms) * 100) : 0 };
  }, [filteredRooms, reservations, today]);

  return (
    <div className="space-y-4">
      {/* Controls */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <div className="flex items-center rounded-lg border border-slate-700 bg-slate-800">
            <button onClick={handlePrev} className="p-2 text-slate-400 hover:text-white"><ChevronLeft className="h-4 w-4" /></button>
            <button onClick={handleToday} className="px-3 py-2 text-xs font-medium text-white hover:bg-slate-700">Today</button>
            <button onClick={handleNext} className="p-2 text-slate-400 hover:text-white"><ChevronRightIcon className="h-4 w-4" /></button>
          </div>
          <select value={dayCount} onChange={(e) => setDayCount(Number(e.target.value))} className="input text-xs">
            {[7, 14, 21, 30].map((d) => <option key={d} value={d}>{d} days</option>)}
          </select>
        </div>
        <div className="flex items-center gap-2">
          <div className="relative max-w-xs">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
            <input className="input w-48 pl-9 text-xs" placeholder="Search room..." value={searchRoom} onChange={(e) => setSearchRoom(e.target.value)} />
          </div>
          <select value={filterFloor} onChange={(e) => setFilterFloor(e.target.value)} className="input text-xs">
            <option value="all">All Floors</option>
            {floors.map((f) => <option key={f} value={f}>Floor {f}</option>)}
          </select>
          <select value={filterType} onChange={(e) => setFilterType(e.target.value)} className="input text-xs">
            <option value="all">All Types</option>
            {types.map((t) => <option key={t} value={t}>{t}</option>)}
          </select>
          {draggingRes && (
            <div className="flex items-center gap-2 rounded-lg border border-dashed border-nexus-500/30 bg-nexus-500/5 px-3 py-1.5">
              <GripVertical className="h-3 w-3 text-nexus-400" />
              <span className="text-xs text-nexus-400">Dragging: {draggingRes.guest_name}</span>
              <button onClick={() => setDraggingRes(null)} className="text-slate-500 hover:text-white"><X className="h-3 w-3" /></button>
            </div>
          )}
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-5 gap-3">
        <StatCard label="Total Rooms" value={occupancyStats.totalRooms} icon={BedDouble} color="sky" />
        <StatCard label="Occupied" value={occupancyStats.occupiedToday} icon={CheckCircle2} color="rose" />
        <StatCard label="Available" value={occupancyStats.availableToday} icon={Grid3X3} color="emerald" />
        <StatCard label="Arriving Today" value={occupancyStats.arrivingToday} icon={ArrowRightLeft} color="violet" />
        <StatCard label="Departing Today" value={occupancyStats.departingToday} icon={Clock} color="amber" />
      </div>

      {/* Grid */}
      <div className="overflow-hidden rounded-lg border border-slate-700 bg-slate-900">
        <div className="h-[500px] overflow-auto">
          <div className="min-w-max">
            {/* Date Header */}
            <div className="sticky top-0 z-20 flex border-b border-slate-700 bg-slate-800">
              <div className="sticky left-0 z-30 w-32 shrink-0 border-r border-slate-700 bg-slate-800 p-2">
                <span className="text-[10px] font-bold uppercase text-slate-400">Room</span>
              </div>
              {dates.map((date) => {
                const isToday = date === today;
                const isWeekend = new Date(date + "T00:00:00").getDay() % 6 === 0;
                return (
                  <div key={date} className={`w-20 shrink-0 border-r border-slate-700/50 p-2 text-center ${isToday ? "bg-nexus-500/10" : isWeekend ? "bg-slate-800/50" : ""}`}>
                    <div className={`text-[10px] font-bold ${isToday ? "text-nexus-400" : "text-slate-400"}`}>
                      {isToday ? "TODAY" : new Date(date + "T00:00:00").toLocaleDateString("en-US", { weekday: "short" }).toUpperCase()}
                    </div>
                    <div className={`text-xs ${isToday ? "font-bold text-white" : "text-slate-300"}`}>{new Date(date + "T00:00:00").getDate()}</div>
                    <div className="text-[9px] text-slate-500">{new Date(date + "T00:00:00").toLocaleDateString("en-US", { month: "short" })}</div>
                  </div>
                );
              })}
            </div>

            {/* Room Rows */}
            {floors.map((floor) => {
              const floorRooms = filteredRooms.filter((r) => r.floor === floor);
              if (floorRooms.length === 0) return null;
              return (
                <div key={floor}>
                  <div className="flex border-b border-slate-700/50 bg-slate-800/30">
                    <div className="sticky left-0 z-10 w-32 shrink-0 border-r border-slate-700/50 p-2">
                      <span className="text-[10px] font-bold uppercase text-slate-500">Floor {floor}</span>
                    </div>
                    <div className="flex">{dates.map((date) => <div key={date} className="w-20 shrink-0 border-r border-slate-700/30" />)}</div>
                  </div>
                  {floorRooms.map((room) => (
                    <div key={room.number} className="flex border-b border-slate-700/30 hover:bg-slate-800/20">
                      <div className="sticky left-0 z-10 w-32 shrink-0 border-r border-slate-700/50 bg-slate-900 p-2">
                        <div className="flex items-center gap-1.5">
                          <div className={`h-2 w-2 rounded-full ${statusColors[room.status]}`} />
                          <span className="text-xs font-bold text-white">{room.number}</span>
                        </div>
                        <div className="mt-0.5 text-[9px] text-slate-500">{room.type}</div>
                        <div className="text-[9px] text-slate-600">{room.bed_type}</div>
                      </div>
                      {dates.map((date) => {
                        const resList = getReservationsForRoomDate(room.number, date);
                        const isToday = date === today;
                        const isHovered = hoveredCell?.room === room.number && hoveredCell?.date === date;
                        const isDropTarget = draggingRes && isHovered && resList.length === 0;
                        const spanInfo = resList.map((r) => getSpanForRes(r, room.number, date)).find((s) => s.start);
                        const spanRes = spanInfo ? resList.find((r) => getSpanForRes(r, room.number, date).start) : null;
                        return (
                          <div
                            key={`${room.number}-${date}`}
                            className={`relative w-20 shrink-0 border-r border-slate-700/30 min-h-[48px] ${isToday ? "bg-nexus-500/5" : ""} ${isDropTarget ? "bg-emerald-500/10" : ""} ${!resList.length && !draggingRes ? "cursor-pointer hover:bg-slate-800/40" : ""}`}
                            onMouseEnter={() => setHoveredCell({ room: room.number, date })}
                            onMouseLeave={() => setHoveredCell(null)}
                            onClick={() => handleCellClick(room, date)}
                            onDragOver={(e) => { e.preventDefault(); setHoveredCell({ room: room.number, date }); }}
                            onDrop={(e) => { e.preventDefault(); handleDrop(room.number, date); }}
                          >
                            {spanRes && (
                              <div
                                draggable
                                onDragStart={() => setDraggingRes(spanRes)}
                                onDragEnd={() => setDraggingRes(null)}
                                className={`absolute inset-y-0.5 left-0.5 z-10 rounded cursor-move ${spanRes.color || "bg-sky-500"} hover:brightness-110`}
                                style={{ width: `${Math.max(spanInfo?.span || 1, 1) * 5 - 0.25}rem`, minWidth: "4.5rem" }}
                                onClick={(e) => { e.stopPropagation(); setSelectedRes(spanRes); }}
                              >
                                <div className="flex h-full items-center px-1.5 overflow-hidden">
                                  <span className="text-[9px] font-medium text-white truncate">{spanRes.guest_name}</span>
                                  {spanRes.vip && <span className="ml-1 text-[7px] text-amber-300">★</span>}
                                </div>
                              </div>
                            )}
                            {!resList.length && !draggingRes && isHovered && (
                              <div className="absolute inset-0 flex items-center justify-center">
                                <Plus className="h-3 w-3 text-slate-600 opacity-0 hover:opacity-100" />
                              </div>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  ))}
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Legend */}
      <div className="flex items-center gap-4 text-[10px] text-slate-500">
        <span className="font-bold uppercase">Room Status:</span>
        {Object.entries(statusColors).map(([status, color]) => (
          <span key={status} className="flex items-center gap-1"><div className={`h-2 w-2 rounded-full ${color}`} /> {status.replace("_", " ")}</span>
        ))}
        <span className="ml-auto">Drag reservation blocks to reassign rooms</span>
      </div>

      {/* New Reservation Modal */}
      {showNewRes && (
        <NewReservationModal
          room={showNewRes.room}
          date={showNewRes.date}
          rooms={rooms}
          onClose={() => setShowNewRes(null)}
          onCreate={async (payload) => {
            setActionLoading(true);
            try {
              await createReservation(payload);
              onRefresh();
              setShowNewRes(null);
            } catch (e: any) {
              alert(e.message);
            } finally {
              setActionLoading(false);
            }
          }}
        />
      )}

      {/* Reservation Detail Modal */}
      {selectedRes && (
        <ReservationDetailModal
          reservation={selectedRes}
          onClose={() => setSelectedRes(null)}
          onAction={async (action) => {
            setActionLoading(true);
            try {
              if (action === "checkin") await checkInReservation(selectedRes.id);
              else if (action === "checkout") await checkOutReservation(selectedRes.id);
              else if (action === "cancel") await cancelReservation(selectedRes.id);
              onRefresh();
              setSelectedRes(null);
            } catch (e: any) {
              alert(e.message);
            } finally {
              setActionLoading(false);
            }
          }}
          loading={actionLoading}
        />
      )}
    </div>
  );
}


/* ─── Calendar View ─── */
function CalendarView({ reservations, rooms, today, onRefresh }: { reservations: Reservation[]; rooms: Room[]; today: string; onRefresh: () => void }) {
  const [currentDate, setCurrentDate] = useState(() => new Date(today + "T00:00:00"));
  const [selectedRes, setSelectedRes] = useState<Reservation | null>(null);
  const [actionLoading, setActionLoading] = useState(false);

  const year = currentDate.getFullYear();
  const month = currentDate.getMonth();

  const monthNames = ["January","February","March","April","May","June","July","August","September","October","November","December"];
  const dayNames = ["Sun","Mon","Tue","Wed","Thu","Fri","Sat"];

  const goToPrevMonth = () => setCurrentDate(new Date(year, month - 1, 1));
  const goToNextMonth = () => setCurrentDate(new Date(year, month + 1, 1));
  const goToToday = () => setCurrentDate(new Date(today + "T00:00:00"));

  const firstDayOfMonth = new Date(year, month, 1);
  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const startDayOfWeek = firstDayOfMonth.getDay();

  const days: { date: number; dateStr: string; isToday: boolean; isCurrentMonth: boolean }[] = [];

  // Previous month padding
  const prevMonthDays = new Date(year, month, 0).getDate();
  for (let i = startDayOfWeek - 1; i >= 0; i--) {
    const d = prevMonthDays - i;
    const prevMonth = month === 0 ? 11 : month - 1;
    const prevYear = month === 0 ? year - 1 : year;
    const dateStr = `${prevYear}-${String(prevMonth + 1).padStart(2, '0')}-${String(d).padStart(2, '0')}`;
    days.push({ date: d, dateStr, isToday: false, isCurrentMonth: false });
  }

  // Current month
  for (let i = 1; i <= daysInMonth; i++) {
    const dateStr = `${year}-${String(month + 1).padStart(2, '0')}-${String(i).padStart(2, '0')}`;
    days.push({ date: i, dateStr, isToday: dateStr === today, isCurrentMonth: true });
  }

  // Next month padding (fill 6 rows = 42 cells)
  const remainingCells = 42 - days.length;
  for (let i = 1; i <= remainingCells; i++) {
    const nextMonth = month + 1;
    const nextYear = nextMonth > 11 ? year + 1 : year;
    const actualNextMonth = nextMonth % 12;
    const dateStr = `${nextYear}-${String(actualNextMonth + 1).padStart(2, '0')}-${String(i).padStart(2, '0')}`;
    days.push({ date: i, dateStr, isToday: false, isCurrentMonth: false });
  }

  const getResForDay = (dateStr: string) => {
    return reservations.filter((r) => isDateInRange(dateStr, r.check_in, r.check_out));
  };

  const inHouse = reservations.filter((r) => r.status === "checked_in").length;
  const upcoming = reservations.filter((r) => r.status === "confirmed").length;

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="flex items-center rounded-lg border border-slate-700 bg-slate-800">
            <button onClick={goToPrevMonth} className="p-2 text-slate-400 hover:text-white"><ChevronLeft className="h-4 w-4" /></button>
            <button onClick={goToToday} className="px-3 py-2 text-xs font-medium text-white hover:bg-slate-700">Today</button>
            <button onClick={goToNextMonth} className="p-2 text-slate-400 hover:text-white"><ChevronRightIcon className="h-4 w-4" /></button>
          </div>
          <h2 className="text-lg font-bold text-white">{monthNames[month]} {year}</h2>
        </div>
        <div className="text-xs text-slate-400">{inHouse} in-house · {upcoming} upcoming</div>
      </div>

      {/* Calendar Grid */}
      <div className="rounded-lg border border-slate-700 bg-slate-900 overflow-hidden">
        <div className="grid grid-cols-7 border-b border-slate-700 bg-slate-800">
          {dayNames.map((d) => (
            <div key={d} className="px-2 py-2 text-center text-[10px] font-bold uppercase text-slate-400">{d}</div>
          ))}
        </div>
        <div className="grid grid-cols-7 auto-rows-fr">
          {days.map((day, idx) => {
            const dayRes = getResForDay(day.dateStr);
            const isWeekend = idx % 7 === 0 || idx % 7 === 6;
            return (
              <div
                key={idx}
                className={`min-h-[110px] border-r border-b border-slate-700/30 p-1.5 ${day.isCurrentMonth ? 'bg-slate-900' : 'bg-slate-800/30'} ${day.isToday ? 'ring-1 ring-inset ring-nexus-500/40' : ''} ${isWeekend && day.isCurrentMonth ? 'bg-slate-800/10' : ''}`}
              >
                <div className={`flex items-center justify-between text-xs font-medium ${day.isToday ? 'text-nexus-400' : day.isCurrentMonth ? 'text-slate-300' : 'text-slate-600'}`}>
                  <span>{day.date}</span>
                  {day.isToday && <span className="rounded bg-nexus-500/20 px-1 text-[9px]">TODAY</span>}
                </div>
                <div className="mt-1 space-y-0.5">
                  {dayRes.slice(0, 4).map((r) => (
                    <div
                      key={r.id}
                      className={`rounded px-1.5 py-0.5 text-[9px] truncate cursor-pointer hover:brightness-110 ${r.color || resColors[hashStringToIndex(r.id, resColors.length)]} text-white`}
                      onClick={() => setSelectedRes(r)}
                      title={`${r.guest_name} · ${r.room_number || 'Unassigned'} · ${r.status.replace('_', ' ')}`}
                    >
                      <span className="flex items-center gap-1">
                        {r.vip && <span className="text-[7px] text-amber-300">★</span>}
                        <span className="truncate">{r.guest_name}</span>
                      </span>
                    </div>
                  ))}
                  {dayRes.length > 4 && <div className="px-1 text-[9px] text-slate-500">+{dayRes.length - 4} more</div>}
                </div>
              </div>
            );
          })}
        </div>
      </div>

      {/* Legend */}
      <div className="flex items-center gap-4 text-[10px] text-slate-500">
        <span className="font-bold uppercase">Status:</span>
        <span className="flex items-center gap-1"><div className="h-2 w-2 rounded-full bg-sky-500" /> Confirmed</span>
        <span className="flex items-center gap-1"><div className="h-2 w-2 rounded-full bg-emerald-500" /> Checked In</span>
        <span className="flex items-center gap-1"><div className="h-2 w-2 rounded-full bg-amber-500" /> Checked Out</span>
        <span className="flex items-center gap-1"><div className="h-2 w-2 rounded-full bg-rose-500" /> Cancelled</span>
        <span className="ml-auto">Click a block for details · Use Quick Book to create</span>
      </div>

      {/* Reservation Detail Modal */}
      {selectedRes && (
        <ReservationDetailModal
          reservation={selectedRes}
          onClose={() => setSelectedRes(null)}
          onAction={async (action) => {
            setActionLoading(true);
            try {
              if (action === "checkin") await checkInReservation(selectedRes.id);
              else if (action === "checkout") await checkOutReservation(selectedRes.id);
              else if (action === "cancel") await cancelReservation(selectedRes.id);
              onRefresh();
              setSelectedRes(null);
            } catch (e: any) {
              alert(e.message);
            } finally {
              setActionLoading(false);
            }
          }}
          loading={actionLoading}
        />
      )}
    </div>
  );
}

/* ─── Quick Book Modal ─── */
function QuickBookModal({ rooms, onClose, onCreate }: { rooms: Room[]; onClose: () => void; onCreate: (payload: any) => Promise<void> }) {
  const [form, setForm] = useState({ guest_name: "", email: "", phone: "", nights: 1, adults: 2, children: 0, special_requests: "", room_number: "", check_in: new Date().toISOString().split("T")[0] });
  const [loading, setLoading] = useState(false);

  const roomData = rooms.find((r) => r.number === form.room_number);
  const rate = roomData?.type === "Suite" ? 229 : roomData?.type === "Deluxe" || roomData?.type === "Deluxe King" ? 129 : 89;
  const total = form.nights * rate;
  const checkOut = addDays(form.check_in, form.nights);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">Quick Book</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>
        <div className="mt-4 space-y-3">
          <div>
            <label className="mb-1 block text-xs text-slate-400">Guest Name *</label>
            <input className="input w-full" value={form.guest_name} onChange={(e) => setForm({ ...form, guest_name: e.target.value })} placeholder="Full name" />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Email</label>
              <input className="input w-full" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} placeholder="guest@email.com" />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Phone</label>
              <input className="input w-full" value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} placeholder="+1-555-0000" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Check In</label>
              <input type="date" className="input w-full" value={form.check_in} onChange={(e) => setForm({ ...form, check_in: e.target.value })} />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Nights</label>
              <input type="number" min={1} max={30} className="input w-full" value={form.nights} onChange={(e) => setForm({ ...form, nights: parseInt(e.target.value) || 1 })} />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Room</label>
              <select className="input w-full" value={form.room_number} onChange={(e) => setForm({ ...form, room_number: e.target.value })}>
                <option value="">Select room...</option>
                {rooms.filter((r) => r.status === "vacant_clean").map((r) => (
                  <option key={r.number} value={r.number}>{r.number} · {r.type}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Guests</label>
              <div className="flex items-center gap-2">
                <input type="number" min={1} max={6} className="input w-16" value={form.adults} onChange={(e) => setForm({ ...form, adults: parseInt(e.target.value) || 1 })} />
                <span className="text-xs text-slate-500">+</span>
                <input type="number" min={0} max={4} className="input w-16" value={form.children} onChange={(e) => setForm({ ...form, children: parseInt(e.target.value) || 0 })} />
              </div>
            </div>
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Special Requests</label>
            <textarea className="input w-full" rows={2} value={form.special_requests} onChange={(e) => setForm({ ...form, special_requests: e.target.value })} placeholder="Any special needs..." />
          </div>
        </div>
        <div className="mt-4 flex items-center justify-between border-t border-slate-700 pt-4">
          <div className="text-sm text-slate-400">{form.nights} nights · <span className="font-bold text-white">${total}</span></div>
          <div className="flex gap-2">
            <button onClick={onClose} className="btn-secondary text-xs">Cancel</button>
            <button
              onClick={async () => {
                if (!form.guest_name) return alert("Guest name is required");
                if (!form.room_number) return alert("Room is required");
                setLoading(true);
                await onCreate({
                  guest_name: form.guest_name,
                  email: form.email,
                  phone: form.phone,
                  room_type: roomData?.type || "Standard",
                  room_number: form.room_number,
                  check_in: form.check_in,
                  check_out: checkOut,
                  adults: form.adults,
                  children: form.children,
                  source: "direct",
                  special_requests: form.special_requests,
                });
                setLoading(false);
              }}
              disabled={loading}
              className="btn-primary text-xs disabled:opacity-50"
            >
              <Save className="h-4 w-4" />
              {loading ? "Creating..." : "Create Booking"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

/* ─── Group Booking Modal ─── */
function GroupBookingModal({ rooms, onClose, onCreate }: { rooms: Room[]; onClose: () => void; onCreate: (payload: { group_name: string; members: any[] }) => Promise<void> }) {
  const [groupName, setGroupName] = useState("");
  const [members, setMembers] = useState([{ guest_name: "", room_number: "", adults: 2, children: 0 }]);
  const [checkIn, setCheckIn] = useState("");
  const [nights, setNights] = useState(1);
  const [loading, setLoading] = useState(false);

  const vacantRooms = rooms.filter((r) => r.status === "vacant_clean");
  const checkOut = checkIn ? addDays(checkIn, nights) : "";

  const addMember = () => setMembers([...members, { guest_name: "", room_number: "", adults: 2, children: 0 }]);
  const removeMember = (i: number) => setMembers(members.filter((_, idx) => idx !== i));
  const updateMember = (i: number, field: string, value: any) => {
    const next = [...members];
    next[i] = { ...next[i], [field]: value };
    setMembers(next);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-lg rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl max-h-[80vh] overflow-auto">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-bold text-white">Group Booking</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>
        <div className="space-y-4">
          <div>
            <label className="mb-1 block text-xs text-slate-400">Group Name</label>
            <input className="input w-full" value={groupName} onChange={(e) => setGroupName(e.target.value)} placeholder="e.g. Johnson Wedding Party" />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Check In</label>
              <input type="date" className="input w-full" value={checkIn} onChange={(e) => setCheckIn(e.target.value)} />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Nights</label>
              <input type="number" min={1} max={30} className="input w-full" value={nights} onChange={(e) => setNights(parseInt(e.target.value) || 1)} />
            </div>
          </div>
          <div className="border-t border-slate-700 pt-4">
            <div className="flex items-center justify-between mb-2">
              <span className="text-xs font-bold uppercase text-slate-400">Members</span>
              <button onClick={addMember} className="text-xs text-violet-400 hover:text-violet-300">+ Add Room</button>
            </div>
            <div className="space-y-3">
              {members.map((m, i) => (
                <div key={i} className="rounded-lg border border-slate-700 bg-slate-800 p-3 space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-xs text-slate-500">Room {i + 1}</span>
                    {members.length > 1 && (
                      <button onClick={() => removeMember(i)} className="text-[10px] text-rose-400 hover:text-rose-300">Remove</button>
                    )}
                  </div>
                  <div className="grid grid-cols-2 gap-2">
                    <input className="input text-xs" placeholder="Guest name" value={m.guest_name} onChange={(e) => updateMember(i, "guest_name", e.target.value)} />
                    <select className="input text-xs" value={m.room_number} onChange={(e) => updateMember(i, "room_number", e.target.value)}>
                      <option value="">Select room...</option>
                      {vacantRooms.map((r) => (
                        <option key={r.number} value={r.number}>{r.number} · {r.type}</option>
                      ))}
                    </select>
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-[10px] text-slate-500">Adults</span>
                    <input type="number" min={1} max={6} className="input w-16 text-xs" value={m.adults} onChange={(e) => updateMember(i, "adults", parseInt(e.target.value) || 1)} />
                    <span className="text-[10px] text-slate-500">Children</span>
                    <input type="number" min={0} max={4} className="input w-16 text-xs" value={m.children} onChange={(e) => updateMember(i, "children", parseInt(e.target.value) || 0)} />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
        <div className="mt-4 flex items-center justify-between border-t border-slate-700 pt-4">
          <div className="text-sm text-slate-400">{members.length} rooms · {nights} nights</div>
          <div className="flex gap-2">
            <button onClick={onClose} className="btn-secondary text-xs">Cancel</button>
            <button
              onClick={async () => {
                if (!groupName) return alert("Group name is required");
                if (!checkIn) return alert("Check in date is required");
                if (members.some((m) => !m.guest_name || !m.room_number)) return alert("All members need a guest name and room");
                setLoading(true);
                const payload = {
                  group_name: groupName,
                  members: members.map((m) => ({
                    guest_name: m.guest_name,
                    room_number: m.room_number,
                    room_type: vacantRooms.find((r) => r.number === m.room_number)?.type || "Standard",
                    check_in: checkIn,
                    check_out: checkOut,
                    adults: m.adults,
                    children: m.children,
                    source: "direct",
                  })),
                };
                await onCreate(payload);
                setLoading(false);
              }}
              disabled={loading}
              className="btn-primary text-xs disabled:opacity-50"
            >
              <Save className="h-4 w-4" />
              {loading ? "Creating..." : "Create Group Booking"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

/* ─── Shared Components ─── */
function StatCard({ label, value, icon: Icon, color, alert }: { label: string; value: string | number; icon: React.ElementType; color: string; alert?: boolean }) {
  const colorMap: Record<string, string> = {
    sky: "bg-sky-500/10 text-sky-400 border-sky-500/20",
    emerald: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
    violet: "bg-violet-500/10 text-violet-400 border-violet-500/20",
    amber: "bg-amber-500/10 text-amber-400 border-amber-500/20",
    rose: "bg-rose-500/10 text-rose-400 border-rose-500/20",
    nexus: "bg-nexus-500/10 text-nexus-400 border-nexus-500/20",
  };
  return (
    <div className={`rounded-lg border p-3 ${colorMap[color]} ${alert ? "animate-pulse" : ""}`}>
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium opacity-80">{label}</span>
        <Icon className="h-4 w-4 opacity-60" />
      </div>
      <p className="mt-1 text-2xl font-bold">{value}</p>
    </div>
  );
}

function AssignRoomModal({ resId, rooms, onClose, onAssign }: { resId: string; rooms: Room[]; onClose: () => void; onAssign: (roomNumber: string) => void }) {
  const vacantRooms = rooms.filter((r) => r.status === "vacant_clean");
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">Assign Room</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>
        <p className="mt-1 text-xs text-slate-400">Select a vacant clean room</p>
        <div className="mt-4 grid grid-cols-3 gap-2">
          {vacantRooms.map((r) => (
            <button key={r.number} onClick={() => onAssign(r.number)}
              className="rounded-lg border border-slate-700 bg-slate-800 p-3 text-left hover:border-nexus-500 hover:bg-slate-700">
              <p className="text-lg font-bold text-white">{r.number}</p>
              <p className="text-xs text-slate-400">{r.type}</p>
              <span className="mt-1 inline-block rounded bg-emerald-500/10 px-1.5 py-0.5 text-[10px] text-emerald-400">Vacant Clean</span>
            </button>
          ))}
        </div>
        {vacantRooms.length === 0 && (
          <div className="mt-4 rounded-lg border border-rose-500/20 bg-rose-500/5 p-4 text-center">
            <AlertTriangle className="mx-auto h-6 w-6 text-rose-400" />
            <p className="mt-2 text-sm text-rose-400">No vacant clean rooms available</p>
          </div>
        )}
      </div>
    </div>
  );
}

function NewReservationModal({ room, date, rooms, onClose, onCreate }: { room: string; date: string; rooms: Room[]; onClose: () => void; onCreate: (payload: any) => Promise<void> }) {
  const roomData = rooms.find((r) => r.number === room);
  const [form, setForm] = useState({ guest_name: "", email: "", phone: "", nights: 1, adults: 2, children: 0, special_requests: "" });
  const [loading, setLoading] = useState(false);

  const rate = roomData?.type === "Suite" ? 229 : roomData?.type === "Deluxe" || roomData?.type === "Deluxe King" ? 129 : 89;
  const total = form.nights * rate;
  const checkOut = addDays(date, form.nights);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-bold text-white">New Booking</h3>
            <p className="text-xs text-slate-400">Room {room} · {roomData?.type} · {formatDateShort(date)}</p>
          </div>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>
        <div className="mt-4 space-y-3">
          <div>
            <label className="mb-1 block text-xs text-slate-400">Guest Name *</label>
            <input className="input w-full" value={form.guest_name} onChange={(e) => setForm({ ...form, guest_name: e.target.value })} placeholder="Full name" />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Email</label>
              <input className="input w-full" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} placeholder="guest@email.com" />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Phone</label>
              <input className="input w-full" value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} placeholder="+1-555-0000" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Nights</label>
              <input type="number" min={1} max={30} className="input w-full" value={form.nights} onChange={(e) => setForm({ ...form, nights: parseInt(e.target.value) || 1 })} />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Guests</label>
              <div className="flex items-center gap-2">
                <input type="number" min={1} max={6} className="input w-16" value={form.adults} onChange={(e) => setForm({ ...form, adults: parseInt(e.target.value) || 1 })} />
                <span className="text-xs text-slate-500">+</span>
                <input type="number" min={0} max={4} className="input w-16" value={form.children} onChange={(e) => setForm({ ...form, children: parseInt(e.target.value) || 0 })} />
              </div>
            </div>
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Special Requests</label>
            <textarea className="input w-full" rows={2} value={form.special_requests} onChange={(e) => setForm({ ...form, special_requests: e.target.value })} placeholder="Any special needs..." />
          </div>
        </div>
        <div className="mt-4 flex items-center justify-between border-t border-slate-700 pt-4">
          <div className="text-sm text-slate-400">{form.nights} nights · <span className="font-bold text-white">${total}</span></div>
          <div className="flex gap-2">
            <button onClick={onClose} className="btn-secondary text-xs">Cancel</button>
            <button
              onClick={async () => {
                if (!form.guest_name) return alert("Guest name is required");
                setLoading(true);
                await onCreate({
                  guest_name: form.guest_name,
                  email: form.email,
                  phone: form.phone,
                  room_type: roomData?.type || "Standard",
                  room_number: room,
                  check_in: date,
                  check_out: checkOut,
                  adults: form.adults,
                  children: form.children,
                  source: "direct",
                  special_requests: form.special_requests,
                });
                setLoading(false);
              }}
              disabled={loading}
              className="btn-primary text-xs disabled:opacity-50"
            >
              <Save className="h-4 w-4" />
              {loading ? "Creating..." : "Create Booking"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

function ReservationDetailModal({ reservation, onClose, onAction, loading }: { reservation: Reservation; onClose: () => void; onAction: (action: string) => Promise<void>; loading: boolean }) {
  const nights = diffDays(reservation.check_in, reservation.check_out);
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className={`flex h-10 w-10 items-center justify-center rounded-full ${reservation.vip ? "bg-amber-500/10" : "bg-slate-700"}`}>
              <User className={`h-5 w-5 ${reservation.vip ? "text-amber-400" : "text-slate-400"}`} />
            </div>
            <div>
              <h3 className="text-lg font-bold text-white flex items-center gap-2">
                {reservation.guest_name}
                {reservation.vip && <span className="rounded bg-amber-500/10 px-1.5 py-0.5 text-[10px] text-amber-400">VIP</span>}
              </h3>
              <p className="text-xs text-slate-400">{reservation.email}</p>
            </div>
          </div>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>
        <div className="mt-4 grid grid-cols-2 gap-3 text-sm">
          <div className="rounded-lg bg-slate-800 p-3">
            <div className="text-[10px] uppercase text-slate-500">Room</div>
            <div className="font-medium text-white">{reservation.room_number || "Unassigned"}</div>
            <div className="text-xs text-slate-400">{reservation.room_type}</div>
          </div>
          <div className="rounded-lg bg-slate-800 p-3">
            <div className="text-[10px] uppercase text-slate-500">Status</div>
            <div className="font-medium text-white capitalize">{reservation.status.replace("_", " ")}</div>
            <div className="text-xs text-slate-400 capitalize">{reservation.source.replace("_", " ")}</div>
          </div>
          <div className="rounded-lg bg-slate-800 p-3">
            <div className="text-[10px] uppercase text-slate-500">Check In</div>
            <div className="font-medium text-white">{formatDateShort(reservation.check_in)}</div>
          </div>
          <div className="rounded-lg bg-slate-800 p-3">
            <div className="text-[10px] uppercase text-slate-500">Check Out</div>
            <div className="font-medium text-white">{formatDateShort(reservation.check_out)}</div>
          </div>
          <div className="rounded-lg bg-slate-800 p-3">
            <div className="text-[10px] uppercase text-slate-500">Guests</div>
            <div className="font-medium text-white">{reservation.adults} adults {reservation.children > 0 && `+ ${reservation.children} children`}</div>
          </div>
          <div className="rounded-lg bg-slate-800 p-3">
            <div className="text-[10px] uppercase text-slate-500">Balance</div>
            <div className={`font-medium ${reservation.balance > 0 ? "text-amber-400" : "text-emerald-400"}`}>${reservation.balance}</div>
          </div>
        </div>
        <div className="mt-4 flex gap-2">
          <Link href={`/reservations/${reservation.id}`} className="flex-1 btn-primary text-center text-xs">Open Full Profile</Link>
          {reservation.status === "confirmed" && (
            <>
              <button onClick={() => onAction("checkin")} disabled={loading} className="btn-primary text-xs disabled:opacity-50">{loading ? "..." : "Check In"}</button>
              <button onClick={() => onAction("cancel")} disabled={loading} className="rounded-lg border border-rose-500/20 bg-rose-500/10 px-3 py-2 text-xs text-rose-400 hover:bg-rose-500/20 disabled:opacity-50">{loading ? "..." : "Cancel"}</button>
            </>
          )}
          {reservation.status === "checked_in" && (
            <button onClick={() => onAction("checkout")} disabled={loading} className="btn-primary text-xs disabled:opacity-50">{loading ? "..." : "Check Out"}</button>
          )}
        </div>
      </div>
    </div>
  );
}
