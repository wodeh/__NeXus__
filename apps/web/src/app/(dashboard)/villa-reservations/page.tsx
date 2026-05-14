"use client";

import { useState, useEffect, useCallback } from "react";
import {
  Calendar, Plus, Edit3, Trash2, CheckCircle2, X, RefreshCw, AlertTriangle,
  Phone, MessageSquare, DollarSign, Clock, User, ChevronRight,
  CreditCard, Banknote, ArrowRightLeft, Wallet, LayoutGrid,
} from "lucide-react";
import Link from "next/link";
import {
  getVillaReservations, getVillas, createVillaReservation, updateVillaReservation, deleteVillaReservation,
  VillaReservation, VillaProperty,
} from "@/lib/api";

export default function VillaReservationsPage() {
  const [reservations, setReservations] = useState<VillaReservation[]>([]);
  const [villas, setVillas] = useState<VillaProperty[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [editing, setEditing] = useState<VillaReservation | null>(null);
  const [filterStatus, setFilterStatus] = useState<string>("all");

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

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleCreate = async (res: Partial<VillaReservation>) => {
    try {
      await createVillaReservation(res as any);
      setShowCreate(false);
      fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleUpdate = async (id: string, updates: Partial<VillaReservation>) => {
    try {
      await updateVillaReservation(id, updates);
      setEditing(null);
      fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this reservation?")) return;
    try {
      await deleteVillaReservation(id);
      fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleMarkDownPayment = async (res: VillaReservation, method: string) => {
    const amount = prompt(`Enter down payment amount for ${res.guest_name}:`, res.balance_due?.toString());
    if (!amount) return;
    const updates: Partial<VillaReservation> = {
      status: "reserved",
      down_payment: {
        amount: parseFloat(amount),
        method,
        status: "received",
        received_at: new Date().toISOString(),
        reference: `${method}-${Date.now()}`,
        notes: "",
      },
    };
    await handleUpdate(res.id, updates);
  };

  const filtered = filterStatus === "all" ? reservations : reservations.filter((r) => r.status === filterStatus);

  const stats = {
    total: reservations.length,
    pending: reservations.filter((r) => r.status === "pending").length,
    reserved: reservations.filter((r) => r.status === "reserved").length,
    completed: reservations.filter((r) => r.status === "completed").length,
    revenue: reservations.filter((r) => r.status === "reserved" || r.status === "completed").reduce((s, r) => s + r.total_amount, 0),
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
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Calendar className="h-6 w-6 text-nexus-400" /> Reservations
          </h1>
          <p className="text-sm text-slate-400">Day-based villa bookings · Down payment tracking · WhatsApp/Phone sources</p>
        </div>
        <div className="flex items-center gap-2">
          <Link href="/villa-tapechart" className="btn-secondary gap-2 text-sm">
            <LayoutGrid className="h-4 w-4" /> Tapechart View
          </Link>
          <button onClick={() => setShowCreate(true)} className="btn-primary gap-2">
            <Plus className="h-4 w-4" /> New Booking
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-5 gap-4">
        <StatCard label="Total" value={stats.total} color="nexus" />
        <StatCard label="Pending" value={stats.pending} color="amber" />
        <StatCard label="Reserved" value={stats.reserved} color="emerald" />
        <StatCard label="Completed" value={stats.completed} color="sky" />
        <StatCard label="Revenue" value={`$${Math.round(stats.revenue).toLocaleString()}`} color="violet" />
      </div>

      {/* Filters */}
      <div className="flex items-center gap-2">
        {["all", "pending", "reserved", "completed", "cancelled"].map((status) => (
          <button
            key={status}
            onClick={() => setFilterStatus(status)}
            className={`rounded-lg px-3 py-1.5 text-xs font-medium capitalize ${
              filterStatus === status ? "bg-slate-700 text-white" : "text-slate-400 hover:bg-slate-800 hover:text-white"
            }`}
          >
            {status}
          </button>
        ))}
      </div>

      {/* Reservations Table */}
      <div className="card space-y-4">
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-800 text-xs uppercase text-slate-400">
            <tr>
              <th className="px-4 py-2">Guest</th>
              <th className="px-4 py-2">Villa</th>
              <th className="px-4 py-2">Dates</th>
              <th className="px-4 py-2">Nights</th>
              <th className="px-4 py-2">Amount</th>
              <th className="px-4 py-2">Down Payment</th>
              <th className="px-4 py-2">Balance</th>
              <th className="px-4 py-2">Status</th>
              <th className="px-4 py-2">Source</th>
              <th className="px-4 py-2">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            {filtered.map((res) => (
              <tr key={res.id} className="hover:bg-slate-800/50">
                <td className="px-4 py-3">
                  <p className="text-white font-medium">{res.guest_name}</p>
                  <p className="text-xs text-slate-500">{res.guest_phone}</p>
                </td>
                <td className="px-4 py-3 text-slate-300">{res.villa_name}</td>
                <td className="px-4 py-3 text-slate-400">
                  {res.check_in_date} → {res.check_out_date}
                </td>
                <td className="px-4 py-3 text-slate-400">{res.nights}n</td>
                <td className="px-4 py-3 text-white font-medium">${Math.round(res.total_amount).toLocaleString()}</td>
                <td className="px-4 py-3">
                  {res.down_payment ? (
                    <div className="flex items-center gap-1 text-xs">
                      <CheckCircle2 className="h-3 w-3 text-emerald-400" />
                      <span className="text-emerald-400">${res.down_payment.amount}</span>
                      <span className="text-slate-500 capitalize">({res.down_payment.method})</span>
                    </div>
                  ) : (
                    <span className="text-xs text-amber-400">Pending</span>
                  )}
                </td>
                <td className="px-4 py-3 text-slate-400">${Math.round(res.balance_due).toLocaleString()}</td>
                <td className="px-4 py-3">
                  <span className={`rounded px-2 py-0.5 text-xs capitalize ${
                    res.status === "pending" ? "bg-amber-500/10 text-amber-400" :
                    res.status === "reserved" ? "bg-emerald-500/10 text-emerald-400" :
                    res.status === "completed" ? "bg-sky-500/10 text-sky-400" :
                    "bg-rose-500/10 text-rose-400"
                  }`}>
                    {res.status}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span className="flex items-center gap-1 text-xs text-slate-400 capitalize">
                    {res.source === "phone" && <Phone className="h-3 w-3" />}
                    {res.source === "whatsapp" && <MessageSquare className="h-3 w-3" />}
                    {res.source}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <div className="flex items-center gap-1">
                    {res.status === "pending" && (
                      <div className="relative group">
                        <button className="rounded bg-emerald-500/10 px-2 py-1 text-[10px] text-emerald-400">Mark Paid</button>
                        <div className="absolute right-0 top-full z-10 hidden w-32 rounded bg-slate-800 shadow-lg group-hover:block">
                          <button onClick={() => handleMarkDownPayment(res, "cash")} className="flex w-full items-center gap-2 px-3 py-2 text-xs text-white hover:bg-slate-700">
                            <Banknote className="h-3 w-3" /> Cash
                          </button>
                          <button onClick={() => handleMarkDownPayment(res, "bank_transfer")} className="flex w-full items-center gap-2 px-3 py-2 text-xs text-white hover:bg-slate-700">
                            <ArrowRightLeft className="h-3 w-3" /> Bank Transfer
                          </button>
                          <button onClick={() => handleMarkDownPayment(res, "third_party")} className="flex w-full items-center gap-2 px-3 py-2 text-xs text-white hover:bg-slate-700">
                            <Wallet className="h-3 w-3" /> Third Party
                          </button>
                          <button onClick={() => handleMarkDownPayment(res, "visa")} className="flex w-full items-center gap-2 px-3 py-2 text-xs text-white hover:bg-slate-700">
                            <CreditCard className="h-3 w-3" /> Visa
                          </button>
                        </div>
                      </div>
                    )}
                    <button onClick={() => setEditing(res)} className="rounded p-1 text-slate-400 hover:text-white">
                      <Edit3 className="h-4 w-4" />
                    </button>
                    <button onClick={() => handleDelete(res.id)} className="rounded p-1 text-slate-400 hover:text-rose-400">
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {showCreate && <ReservationModal villas={villas} onSave={handleCreate} onClose={() => setShowCreate(false)} />}
      {editing && <ReservationModal villas={villas} reservation={editing} onSave={(r) => handleUpdate(editing.id, r)} onClose={() => setEditing(null)} />}
    </div>
  );
}

function StatCard({ label, value, color }: { label: string; value: string | number; color: string }) {
  const colorMap: Record<string, string> = {
    nexus: "text-nexus-400 bg-nexus-500/10",
    emerald: "text-emerald-400 bg-emerald-500/10",
    amber: "text-amber-400 bg-amber-500/10",
    sky: "text-sky-400 bg-sky-500/10",
    violet: "text-violet-400 bg-violet-500/10",
  };
  return (
    <div className="card space-y-2">
      <span className="text-xs text-slate-400">{label}</span>
      <p className={`text-xl font-bold ${colorMap[color].split(" ")[0]}`}>{value}</p>
    </div>
  );
}

function ReservationModal({ villas, reservation, onSave, onClose }: { villas: VillaProperty[]; reservation?: VillaReservation; onSave: (r: Partial<VillaReservation>) => void; onClose: () => void }) {
  const [form, setForm] = useState({
    villa_id: reservation?.villa_id || villas[0]?.id || "",
    guest_name: reservation?.guest_name || "",
    guest_phone: reservation?.guest_phone || "",
    guest_email: reservation?.guest_email || "",
    guest_count: reservation?.guest_count || 2,
    check_in_date: reservation?.check_in_date || "",
    check_out_date: reservation?.check_out_date || "",
    nights: reservation?.nights || 1,
    total_amount: reservation?.total_amount || 0,
    currency: reservation?.currency || "USD",
    status: reservation?.status || "pending",
    source: reservation?.source || "phone",
    internal_notes: reservation?.internal_notes || "",
    balance_due: reservation?.balance_due || 0,
  });

  const selectedVilla = villas.find((v) => v.id === form.villa_id);

  const calculateNights = () => {
    if (!form.check_in_date || !form.check_out_date) return;
    const start = new Date(form.check_in_date);
    const end = new Date(form.check_out_date);
    const nights = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24));
    const total = nights * (selectedVilla?.price_per_night || 0);
    setForm((f) => ({ ...f, nights: Math.max(1, nights), total_amount: total, balance_due: total }));
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-lg rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl space-y-4 max-h-[90vh] overflow-auto">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">{reservation ? "Edit Reservation" : "New Booking"}</h3>
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
              <input type="date" className="input w-full" value={form.check_in_date} onChange={(e) => { setForm({ ...form, check_in_date: e.target.value }); setTimeout(calculateNights, 0); }} />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Check-out</label>
              <input type="date" className="input w-full" value={form.check_out_date} onChange={(e) => { setForm({ ...form, check_out_date: e.target.value }); setTimeout(calculateNights, 0); }} />
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
              <option value="pending">Pending</option>
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

          <textarea className="input w-full" rows={2} placeholder="Internal Notes" value={form.internal_notes} onChange={(e) => setForm({ ...form, internal_notes: e.target.value })} />
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <button onClick={onClose} className="btn-secondary">Cancel</button>
          <button onClick={() => onSave(form)} className="btn-primary gap-2"><CheckCircle2 className="h-4 w-4" /> Save</button>
        </div>
      </div>
    </div>
  );
}
