"use client";

import { useState, useMemo, useCallback, useEffect } from "react";
import Link from "next/link";
import {
  ChevronLeft, ChevronRight, CalendarDays, Plus, CheckCircle2, X, RefreshCw,
  Phone, MessageSquare, Globe, Walk, ArrowRightLeft, Eye, AlertTriangle,
} from "lucide-react";
import {
  getVillaReservations, getVillas, updateVillaReservation, createVillaReservation,
  VillaReservation, VillaProperty,
} from "@/lib/api";

/* ─── Helpers ─── */
function formatDateLocal(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, "0")}-${String(d.getDate()).padStart(2, "0")}`;
}
function addDays(dateStr: string, days: number): string {
  const [y, m, d] = dateStr.split("-").map(Number);
  const date = new Date(y, m - 1, d);
  date.setDate(date.getDate() + days);
  return formatDateLocal(date);
}
function isDateInRange(date: string, start: string, end: string): boolean {
  const t = new Date(date + "T00:00:00").getTime();
  const s = new Date(start + "T00:00:00").getTime();
  const e = new Date(end + "T00:00:00").getTime();
  return t >= s && t < e;
}

const resColors = [
  "bg-sky-500", "bg-violet-500", "bg-amber-500", "bg-rose-500",
  "bg-emerald-500", "bg-orange-500", "bg-cyan-500", "bg-pink-500",
];

function hashStringToIndex(str: string, max: number): number {
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) - hash) + str.charCodeAt(i);
    hash |= 0;
  }
  return Math.abs(hash) % max;
}

export default function VillaTapechartPage() {
  const [reservations, setReservations] = useState<VillaReservation[]>([]);
  const [villas, setVillas] = useState<VillaProperty[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<"week" | "month">("week");
  const [viewStart, setViewStart] = useState(() => {
    const d = new Date();
    d.setDate(d.getDate() - d.getDay()); // Start of week
    return formatDateLocal(d);
  });
  const [selectedVilla, setSelectedVilla] = useState<string>("all");
  const [showCreate, setShowCreate] = useState(false);
  const [createDefaults, setCreateDefaults] = useState<{villaId?: string; date?: string}>({});
  const [selectedRes, setSelectedRes] = useState<VillaReservation | null>(null);

  const dayCount = viewMode === "week" ? 7 : 30;

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [r, v] = await Promise.all([getVillaReservations(), getVillas()]);
      setReservations(r);
      setVillas(v);
    } catch (e: any) {
      setError(e.message || "Failed to load");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => { fetchData(); }, [fetchData]);

  const days = useMemo(() => {
    const arr: string[] = [];
    for (let i = 0; i < dayCount; i++) arr.push(addDays(viewStart, i));
    return arr;
  }, [viewStart, dayCount]);

  const visibleVillas = useMemo(() => {
    if (selectedVilla === "all") return villas;
    return villas.filter((v) => v.id === selectedVilla);
  }, [villas, selectedVilla]);

  const nav = (dir: number) => setViewStart(addDays(viewStart, dir * dayCount));

  const handleApprove = async (res: VillaReservation) => {
    try {
      await updateVillaReservation(res.id, { status: "reserved" });
      fetchData();
    } catch (e: any) { alert(e.message); }
  };

  const handleDeny = async (res: VillaReservation) => {
    if (!confirm(`Deny reservation for ${res.guest_name}?`)) return;
    try {
      await updateVillaReservation(res.id, { status: "cancelled" });
      fetchData();
    } catch (e: any) { alert(e.message); }
  };

  const handleCreate = async (data: Partial<VillaReservation>) => {
    try {
      await createVillaReservation(data as any);
      setShowCreate(false);
      fetchData();
    } catch (e: any) { alert(e.message); }
  };

  /* occupancy helper for suggesting alternatives */
  const getAlternativeVillas = (checkIn: string, checkOut: string, excludeVillaId: string) => {
    return villas
      .filter((v) => v.id !== excludeVillaId && v.status === "available")
      .map((v) => {
        const conflicts = reservations.filter(
          (r) => r.villa_id === v.id && r.status !== "cancelled" &&
            isDateInRange(checkIn, r.check_in_date, r.check_out_date) ||
            isDateInRange(checkOut, r.check_in_date, r.check_out_date) ||
            (new Date(checkIn) >= new Date(r.check_in_date) && new Date(checkOut) <= new Date(r.check_out_date))
        );
        return { villa: v, available: conflicts.length === 0 };
      })
      .filter((x) => x.available)
      .map((x) => x.villa);
  };

  if (loading) {
    return (
      <div className="flex h-96 items-center justify-center">
        <RefreshCw className="h-8 w-8 animate-spin text-nexus-400" />
      </div>
    );
  }
  if (error) {
    return (
      <div className="flex h-96 flex-col items-center justify-center gap-4">
        <AlertTriangle className="h-10 w-10 text-rose-400" />
        <p className="text-rose-400">{error}</p>
        <button onClick={fetchData} className="btn-primary">Retry</button>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <CalendarDays className="h-6 w-6 text-nexus-400" /> Villa Tapechart
          </h1>
          <p className="text-sm text-slate-400">Weekly / monthly occupancy · Approve or deny phone requests · Suggest alternatives</p>
        </div>
        <div className="flex items-center gap-2">
          <Link href="/villa-reservations" className="btn-secondary gap-2 text-sm">
            <Eye className="h-4 w-4" /> List View
          </Link>
          <button onClick={() => setShowCreate(true)} className="btn-primary gap-2 text-sm">
            <Plus className="h-4 w-4" /> New Booking
          </button>
        </div>
      </div>

      {/* Controls */}
      <div className="flex items-center justify-between gap-4">
        <div className="flex items-center gap-2">
          <button onClick={() => nav(-1)} className="rounded-lg border border-slate-700 p-2 hover:bg-slate-800">
            <ChevronLeft className="h-4 w-4 text-slate-300" />
          </button>
          <span className="text-sm font-medium text-white w-48 text-center">
            {viewMode === "week"
              ? `${viewStart} → ${addDays(viewStart, 6)}`
              : `${viewStart.slice(0, 7)}`}
          </span>
          <button onClick={() => nav(1)} className="rounded-lg border border-slate-700 p-2 hover:bg-slate-800">
            <ChevronRight className="h-4 w-4 text-slate-300" />
          </button>
          <button
            onClick={() => setViewStart(formatDateLocal(new Date()))}
            className="rounded-lg border border-slate-700 px-3 py-2 text-xs text-slate-300 hover:bg-slate-800"
          >
            Today
          </button>
        </div>

        <div className="flex items-center gap-2">
          {/* Villa filter */}
          <select
            className="rounded-lg border border-slate-700 bg-slate-900 px-3 py-2 text-sm text-white"
            value={selectedVilla}
            onChange={(e) => setSelectedVilla(e.target.value)}
          >
            <option value="all">All Villas</option>
            {villas.map((v) => (
              <option key={v.id} value={v.id}>{v.name}</option>
            ))}
          </select>

          {/* View toggle */}
          <div className="flex rounded-lg border border-slate-700 overflow-hidden">
            <button
              onClick={() => setViewMode("week")}
              className={`px-3 py-2 text-xs font-medium ${viewMode === "week" ? "bg-nexus-500/20 text-nexus-400" : "text-slate-400 hover:text-white"}`}
            >
              Week
            </button>
            <button
              onClick={() => setViewMode("month")}
              className={`px-3 py-2 text-xs font-medium ${viewMode === "month" ? "bg-nexus-500/20 text-nexus-400" : "text-slate-400 hover:text-white"}`}
            >
              Month
            </button>
          </div>
        </div>
      </div>

      {/* Tapechart Grid */}
      <div className="card overflow-hidden">
        <div className="overflow-x-auto">
          <div className="min-w-max">
            {/* Header Row: Days */}
            <div className="grid" style={{ gridTemplateColumns: `200px repeat(${dayCount}, minmax(${viewMode === "week" ? 120 : 40}px, 1fr))` }}>
              <div className="sticky left-0 z-10 border-b border-r border-slate-700 bg-slate-900 px-4 py-3 text-xs font-medium text-slate-400 uppercase">
                Villa
              </div>
              {days.map((d) => {
                const date = new Date(d + "T00:00:00");
                const isWeekend = date.getDay() === 0 || date.getDay() === 6;
                const isToday = d === formatDateLocal(new Date());
                return (
                  <div
                    key={d}
                    className={`border-b border-r border-slate-700 px-2 py-3 text-center text-xs ${
                      isToday ? "bg-nexus-500/10 font-bold text-nexus-400" : isWeekend ? "bg-slate-800/50 text-slate-500" : "text-slate-400"
                    }`}
                  >
                    <div>{date.toLocaleDateString("en-US", { weekday: viewMode === "week" ? "short" : undefined })}</div>
                    <div className={isToday ? "font-bold" : ""}>{date.getDate()}</div>
                    {viewMode === "week" && <div className="text-[10px] opacity-60">{date.toLocaleDateString("en-US", { month: "short" })}</div>}
                  </div>
                );
              })}
            </div>

            {/* Villa Rows */}
            {visibleVillas.map((villa) => {
              const villaReservations = reservations.filter((r) => r.villa_id === villa.id && r.status !== "cancelled");
              return (
                <div
                  key={villa.id}
                  className="grid"
                  style={{ gridTemplateColumns: `200px repeat(${dayCount}, minmax(${viewMode === "week" ? 120 : 40}px, 1fr))` }}
                >
                  {/* Villa Label */}
                  <div className="sticky left-0 z-10 border-b border-r border-slate-700 bg-slate-900 px-4 py-3">
                    <p className="text-sm font-medium text-white">{villa.name}</p>
                    <p className="text-[10px] text-slate-500">${villa.price_per_night}/night · {villa.bedrooms}BR</p>
                  </div>

                  {/* Day Cells */}
                  {days.map((day) => {
                    const dayRes = villaReservations.filter((r) => isDateInRange(day, r.check_in_date, r.check_out_date));
                    const isToday = day === formatDateLocal(new Date());
                    const hasRes = dayRes.length > 0;
                    const res = hasRes ? dayRes[0] : null;
                    const isFirstDay = res ? day === res.check_in_date : false;
                    const isPending = res?.status === "pending";

                    return (
                      <div
                        key={`${villa.id}-${day}`}
                        className={`relative border-b border-r border-slate-700 min-h-[${viewMode === "week" ? 64 : 48}px] ${
                          isToday ? "bg-nexus-500/5" : ""
                        } ${!hasRes ? "hover:bg-slate-800/30 cursor-pointer" : ""}`}
                        onClick={() => {
                          if (!hasRes) {
                            setCreateDefaults({ villaId: villa.id, date: day });
                            setShowCreate(true);
                          } else {
                            setSelectedRes(res);
                          }
                        }}
                      >
                        {hasRes && isFirstDay && res && (
                          <div
                            className={`absolute inset-1 rounded-md px-2 py-1 text-xs text-white shadow-sm ${
                              isPending ? "bg-amber-500 ring-1 ring-amber-400/50" :
                              res.status === "reserved" ? resColors[hashStringToIndex(res.id, resColors.length)] :
                              "bg-slate-500"
                            }`}
                            style={{
                              width: `calc(${Math.min(
                                Math.ceil((new Date(res.check_out_date).getTime() - new Date(day).getTime()) / (86400000)),
                                dayCount - days.indexOf(day)
                              )} * 100% + ${Math.min(
                                Math.ceil((new Date(res.check_out_date).getTime() - new Date(day).getTime()) / (86400000)),
                                dayCount - days.indexOf(day)
                              ) - 1}px)`,
                            }}
                          >
                            <div className="flex items-center gap-1 truncate">
                              {res.source === "phone" && <Phone className="h-3 w-3" />}
                              {res.source === "whatsapp" && <MessageSquare className="h-3 w-3" />}
                              {res.source === "website" && <Globe className="h-3 w-3" />}
                              {res.source === "walkin" && <Walk className="h-3 w-3" />}
                              <span className="truncate font-medium">{res.guest_name}</span>
                            </div>
                            {viewMode === "week" && (
                              <div className="text-[10px] opacity-80 mt-0.5">
                                {res.guest_count} guests · ${Math.round(res.total_amount)}
                              </div>
                            )}
                            {isPending && viewMode === "week" && (
                              <div className="mt-1 flex gap-1">
                                <button
                                  onClick={(e) => { e.stopPropagation(); handleApprove(res); }}
                                  className="rounded bg-white/20 px-1.5 py-0.5 text-[10px] hover:bg-white/30"
                                >
                                  ✓ Approve
                                </button>
                                <button
                                  onClick={(e) => { e.stopPropagation(); handleDeny(res); }}
                                  className="rounded bg-black/20 px-1.5 py-0.5 text-[10px] hover:bg-black/30"
                                >
                                  ✕ Deny
                                </button>
                              </div>
                            )}
                          </div>
                        )}
                        {!hasRes && viewMode === "week" && (
                          <div className="flex h-full items-center justify-center text-slate-600">
                            <Plus className="h-4 w-4 opacity-0 hover:opacity-100" />
                          </div>
                        )}
                      </div>
                    );
                  })}
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Reservation Detail Modal */}
      {selectedRes && (
        <ResDetailModal
          res={selectedRes}
          villas={villas}
          alternatives={getAlternativeVillas(selectedRes.check_in_date, selectedRes.check_out_date, selectedRes.villa_id)}
          onApprove={handleApprove}
          onDeny={handleDeny}
          onClose={() => setSelectedRes(null)}
          onUpdate={(id, data) => { updateVillaReservation(id, data).then(fetchData); setSelectedRes(null); }}
        />
      )}

      {/* Create Modal */}
      {showCreate && (
        <CreateReservationModal
          villas={villas}
          defaultVillaId={createDefaults.villaId}
          defaultDate={createDefaults.date}
          onSave={handleCreate}
          onClose={() => { setShowCreate(false); setCreateDefaults({}); }}
        />
      )}
    </div>
  );
}

/* ─── Res Detail Modal ─── */
function ResDetailModal({ res, villas, alternatives, onApprove, onDeny, onClose, onUpdate }: {
  res: VillaReservation;
  villas: VillaProperty[];
  alternatives: VillaProperty[];
  onApprove: (r: VillaReservation) => void;
  onDeny: (r: VillaReservation) => void;
  onClose: () => void;
  onUpdate: (id: string, data: Partial<VillaReservation>) => void;
}) {
  const villa = villas.find((v) => v.id === res.villa_id);
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">Reservation Details</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>

        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Guest</span>
            <span className="text-sm font-medium text-white">{res.guest_name}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Phone</span>
            <span className="text-sm text-white">{res.guest_phone}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Villa</span>
            <span className="text-sm font-medium text-white">{res.villa_name}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Dates</span>
            <span className="text-sm text-white">{res.check_in_date} → {res.check_out_date} ({res.nights} nights)</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Amount</span>
            <span className="text-sm font-medium text-white">${Math.round(res.total_amount).toLocaleString()}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Status</span>
            <span className={`rounded px-2 py-0.5 text-xs capitalize ${
              res.status === "pending" ? "bg-amber-500/10 text-amber-400" :
              res.status === "reserved" ? "bg-emerald-500/10 text-emerald-400" :
              res.status === "completed" ? "bg-sky-500/10 text-sky-400" :
              "bg-rose-500/10 text-rose-400"
            }`}>
              {res.status}
            </span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-400">Source</span>
            <span className="flex items-center gap-1 text-xs text-slate-400 capitalize">
              {res.source === "phone" && <Phone className="h-3 w-3" />}
              {res.source === "whatsapp" && <MessageSquare className="h-3 w-3" />}
              {res.source}
            </span>
          </div>
        </div>

        {/* Alternative Villas */}
        {alternatives.length > 0 && (
          <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-3 space-y-2">
            <p className="text-xs font-medium text-slate-300">Alternative Villas Available:</p>
            <div className="flex flex-wrap gap-2">
              {alternatives.slice(0, 4).map((alt) => (
                <button
                  key={alt.id}
                  onClick={() => onUpdate(res.id, { villa_id: alt.id, villa_name: alt.name })}
                  className="rounded bg-slate-700 px-2 py-1 text-xs text-white hover:bg-nexus-500/20 hover:text-nexus-400"
                >
                  <ArrowRightLeft className="h-3 w-3 inline mr-1" />
                  {alt.name} · ${alt.price_per_night}
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Actions */}
        <div className="flex gap-2 pt-2">
          {res.status === "pending" && (
            <>
              <button onClick={() => onApprove(res)} className="btn-primary flex-1 gap-2">
                <CheckCircle2 className="h-4 w-4" /> Approve
              </button>
              <button onClick={() => onDeny(res)} className="btn-secondary flex-1 bg-rose-500/10 text-rose-400 border-rose-500/30 hover:bg-rose-500/20">
                <X className="h-4 w-4" /> Deny
              </button>
            </>
          )}
          {res.status !== "pending" && (
            <button onClick={onClose} className="btn-secondary flex-1">Close</button>
          )}
        </div>
      </div>
    </div>
  );
}

/* ─── Create Modal ─── */
function CreateReservationModal({ villas, defaultVillaId, defaultDate, onSave, onClose }: {
  villas: VillaProperty[];
  defaultVillaId?: string;
  defaultDate?: string;
  onSave: (data: Partial<VillaReservation>) => void;
  onClose: () => void;
}) {
  const [form, setForm] = useState({
    villa_id: defaultVillaId || villas[0]?.id || "",
    guest_name: "",
    guest_phone: "",
    guest_email: "",
    guest_count: 2,
    check_in_date: defaultDate || formatDateLocal(new Date()),
    check_out_date: defaultDate ? addDays(defaultDate, 3) : addDays(formatDateLocal(new Date()), 3),
    nights: 3,
    total_amount: 0,
    currency: "USD",
    status: "pending",
    source: "phone",
    balance_due: 0,
  });

  const selectedVilla = villas.find((v) => v.id === form.villa_id);

  useEffect(() => {
    if (selectedVilla && form.check_in_date && form.check_out_date) {
      const nights = Math.ceil((new Date(form.check_out_date).getTime() - new Date(form.check_in_date).getTime()) / (1000 * 60 * 60 * 24));
      const total = nights * (selectedVilla.price_per_night || 0);
      setForm((f) => ({ ...f, nights: Math.max(1, nights), total_amount: total, balance_due: total }));
    }
  }, [form.check_in_date, form.check_out_date, form.villa_id, selectedVilla]);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-lg rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl space-y-4 max-h-[90vh] overflow-auto">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">New Villa Booking</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>

        <div className="space-y-3">
          <div>
            <label className="mb-1 block text-xs text-slate-400">Villa</label>
            <select className="input w-full" value={form.villa_id} onChange={(e) => setForm({ ...form, villa_id: e.target.value })}>
              {villas.map((v) => (
                <option key={v.id} value={v.id}>{v.name} - ${v.price_per_night}/night</option>
              ))}
            </select>
          </div>

          <input className="input w-full" placeholder="Guest Name" value={form.guest_name} onChange={(e) => setForm({ ...form, guest_name: e.target.value })} />
          <div className="grid grid-cols-2 gap-3">
            <input className="input w-full" placeholder="Phone" value={form.guest_phone} onChange={(e) => setForm({ ...form, guest_phone: e.target.value })} />
            <input className="input w-full" placeholder="Email" value={form.guest_email} onChange={(e) => setForm({ ...form, guest_email: e.target.value })} />
          </div>
          <input type="number" className="input w-full" placeholder="Guest Count" value={form.guest_count} onChange={(e) => setForm({ ...form, guest_count: parseInt(e.target.value) || 1 })} />

          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Check-in</label>
              <input type="date" className="input w-full" value={form.check_in_date} onChange={(e) => setForm({ ...form, check_in_date: e.target.value })} />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Check-out</label>
              <input type="date" className="input w-full" value={form.check_out_date} onChange={(e) => setForm({ ...form, check_out_date: e.target.value })} />
            </div>
          </div>

          <div className="grid grid-cols-3 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Nights</label>
              <input type="number" className="input w-full bg-slate-800" value={form.nights} readOnly />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Total</label>
              <input type="number" className="input w-full bg-slate-800" value={form.total_amount} readOnly />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Balance Due</label>
              <input type="number" className="input w-full bg-slate-800" value={form.balance_due} readOnly />
            </div>
          </div>

          <div className="grid grid-cols-2 gap-3">
            <select className="input w-full" value={form.status} onChange={(e) => setForm({ ...form, status: e.target.value })}>
              <option value="pending">Pending (needs approval)</option>
              <option value="reserved">Reserved</option>
              <option value="completed">Completed</option>
              <option value="cancelled">Cancelled</option>
            </select>
            <select className="input w-full" value={form.source} onChange={(e) => setForm({ ...form, source: e.target.value })}>
              <option value="phone">Phone</option>
              <option value="whatsapp">WhatsApp</option>
              <option value="walkin">Walk-in</option>
              <option value="website">Website</option>
            </select>
          </div>
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <button onClick={onClose} className="btn-secondary">Cancel</button>
          <button onClick={() => onSave(form)} className="btn-primary gap-2"><CheckCircle2 className="h-4 w-4" /> Save</button>
        </div>
      </div>
    </div>
  );
}
