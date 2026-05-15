// @ts-nocheck
"use client";

import React, { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import {
  ArrowLeft,
  Edit3,
  Save,
  X,
  CheckCircle2,
  DoorOpen,
  DoorClosed,
  Ban,
  Printer,
  FileText,
  CreditCard,
  Plus,
  ArrowRightLeft,
  User,
  Mail,
  Phone,
  MapPin,
  BedDouble,
  CalendarDays,
  Clock,
  Star,
  Trash2,
  Receipt,
  Banknote,
  Wifi,
  KeyRound,
  AlertTriangle,
  RefreshCw,
} from "lucide-react";
import { Reservation as ApiReservation, Room, getReservations, getRooms, getReservationDetail, checkInReservation, checkOutReservation, cancelReservation, assignRoom, restoreReservation } from "@/lib/api";

interface Charge {
  id: string;
  description: string;
  amount: number;
  qty: number;
  total: number;
  category: "room" | "service" | "minibar" | "restaurant" | "spa" | "late" | "other";
  postedAt: string;
  staffName?: string;
}

interface Reservation extends ApiReservation {
  address?: string;
  country?: string;
  nights?: number;
  paid?: number;
  deposit?: number;
  createdAt?: string;
  checkedInAt?: string;
  checkedOutAt?: string;
  loyalty_id?: string;
}

const mockCharges: Charge[] = [
  { id: "ch-1", description: "Room Charge", amount: 129, qty: 4, total: 516, category: "room", postedAt: "2026-05-10" },
  { id: "ch-2", description: "Room Service - Breakfast", amount: 24, qty: 2, total: 48, category: "restaurant", postedAt: "2026-05-11", staffName: "Maria K." },
];

const statusColors: Record<string, string> = {
  confirmed: "bg-sky-500/10 text-sky-400 border-sky-500/20",
  checked_in: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
  checked_out: "bg-amber-500/10 text-amber-400 border-amber-500/20",
  cancelled: "bg-rose-500/10 text-rose-400 border-rose-500/20",
  no_show: "bg-rose-500/10 text-rose-400 border-rose-500/20",
};

export default function ReservationDetailPage(props: { params: Promise<{ id: string }> }) {
  const params = React.use(props.params);
  const id = params?.id || "";
  const [res, setRes] = useState<Reservation | null>(null);
  const [rooms, setRooms] = useState<Room[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [charges, setCharges] = useState<Charge[]>(mockCharges);
  const [activeTab, setActiveTab] = useState<"overview" | "charges" | "invoice" | "history">("overview");
  const [editingGuest, setEditingGuest] = useState(false);
  const [showChangeRoom, setShowChangeRoom] = useState(false);
  const [showAddCharge, setShowAddCharge] = useState(false);
  const [showInvoice, setShowInvoice] = useState(false);
  const [showCheckInOut, setShowCheckInOut] = useState<"checkin" | "checkout" | null>(null);
  const [showCancel, setShowCancel] = useState(false);

  const fetchData = useCallback(async () => {
    if (!id) return;
    setLoading(true);
    setError(null);
    try {
      const [detail, allRooms] = await Promise.all([
        getReservationDetail(id).catch(() => null),
        getRooms(),
      ]);
      
      let found: Reservation | null = null;
      
      if (detail) {
        found = {
          ...detail,
          id: detail.id,
          guest_name: detail.guest.name,
          email: detail.guest.email,
          phone: detail.guest.phone,
          room_number: detail.room.room_number,
          room_type: detail.room.room_type,
          check_in: detail.dates.check_in,
          check_out: detail.dates.check_out,
          adults: detail.party.adults,
          children: detail.party.children,
          status: detail.status as any,
          source: detail.source as any,
          total: detail.financials.total,
          balance: detail.financials.balance,
          special_requests: detail.special_requests,
          vip: detail.guest.vip,
          created_at: detail.created_at,
          updated_at: detail.updated_at,
          // Enrich with fields API doesn't have yet
          address: detail.guest.address || "123 Park Avenue, New York",
          country: detail.guest.country || "United States",
          nights: detail.room.total_nights,
          paid: detail.financials.paid || detail.financials.total - detail.financials.balance,
          deposit: detail.financials.deposit_paid || 100,
          createdAt: detail.created_at?.split("T")[0] || "",
          checkedInAt: detail.status === "checked_in" ? detail.updated_at : undefined,
          checkedOutAt: detail.status === "checked_out" ? detail.updated_at : undefined,
          loyalty_id: undefined,
        };
      } else {
        // Fallback to old method
        const allRes = await getReservations();
        const r = allRes.find((r) => r.id === id);
        if (r) {
          found = {
            ...r,
            address: "123 Park Avenue, New York",
            country: "United States",
            nights: Math.max(1, Math.round((new Date(r.check_out).getTime() - new Date(r.check_in).getTime()) / (1000 * 60 * 60 * 24))),
            paid: r.total - r.balance,
            deposit: 100,
            createdAt: r.created_at?.split("T")[0] || "",
          };
        }
      }
      
      if (!found) {
        setError("Reservation not found");
        return;
      }
      setRes(found);
      setRooms(allRooms);
    } catch (e: any) {
      setError(e.message || "Failed to load reservation");
    } finally {
      setLoading(false);
    }
  }, [id]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  if (loading) {
    return (
      <div className="flex h-96 items-center justify-center">
        <RefreshCw className="h-8 w-8 animate-spin text-nexus-400" />
      </div>
    );
  }

  if (error || !res) {
    return (
      <div className="flex h-96 flex-col items-center justify-center gap-4">
        <AlertTriangle className="h-10 w-10 text-rose-400" />
        <p className="text-rose-400">{error || "Reservation not found"}</p>
        <div className="flex gap-2">
          <Link href="/reservations" className="btn-secondary">Back to Reservations</Link>
          <button onClick={fetchData} className="btn-primary">Retry</button>
        </div>
      </div>
    );
  }

  const subtotal = charges.reduce((s, c) => s + c.total, 0);
  const tax = subtotal * 0.12;
  const grandTotal = subtotal + tax;
  const balanceDue = grandTotal - (res.paid || 0);
  const availableRooms = rooms.filter((r) => r.status === "vacant_clean" && r.number !== res.room_number);

  const handleGuestSave = (updated: Partial<Reservation>) => {
    setRes({ ...res, ...updated });
    setEditingGuest(false);
  };

  const handleRoomChange = async (newRoom: string) => {
    try {
      await assignRoom(res.id, newRoom);
      setRes({ ...res, room_number: newRoom });
      setShowChangeRoom(false);
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleAddCharge = (charge: Charge) => {
    setCharges([...charges, charge]);
    setShowAddCharge(false);
  };

  const handleCheckIn = async () => {
    try {
      await checkInReservation(res.id);
      setRes({ ...res, status: "checked_in", checkedInAt: new Date().toISOString() });
      setShowCheckInOut(null);
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleCheckOut = async () => {
    try {
      await checkOutReservation(res.id);
      setRes({ ...res, status: "checked_out", checkedOutAt: new Date().toISOString() });
      setShowCheckInOut(null);
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleCancel = async () => {
    try {
      await cancelReservation(res.id);
      setRes({ ...res, status: "cancelled" });
      setShowCancel(false);
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleRestore = async () => {
    try {
      await restoreReservation(res.id);
      setRes({ ...res, status: "confirmed" });
    } catch (e: any) {
      alert(e.message);
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link href="/reservations" className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white">
            <ArrowLeft className="h-4 w-4" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold text-white">Reservation {res.id.toUpperCase()}</h1>
            <p className="text-sm text-slate-400">Created {res.createdAt}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {res.status === "cancelled" && (
            <button onClick={handleRestore} className="btn-primary gap-2 bg-emerald-600 hover:bg-emerald-500">
              <CheckCircle2 className="h-4 w-4" /> Restore
            </button>
          )}
          {res.status === "confirmed" && (
            <button onClick={() => setShowCheckInOut("checkin")} className="btn-primary gap-2">
              <DoorOpen className="h-4 w-4" /> Check In
            </button>
          )}
          {res.status === "checked_in" && (
            <button onClick={() => setShowCheckInOut("checkout")} className="btn-primary gap-2 bg-amber-600 hover:bg-amber-500">
              <DoorClosed className="h-4 w-4" /> Check Out
            </button>
          )}
          {res.status !== "cancelled" && res.status !== "checked_out" && (
            <button onClick={() => setShowCancel(true)} className="btn-secondary gap-2 text-rose-400 hover:text-rose-300">
              <Ban className="h-4 w-4" /> Cancel
            </button>
          )}
          <button onClick={() => setShowInvoice(true)} className="btn-secondary gap-2">
            <Printer className="h-4 w-4" /> Invoice
          </button>
        </div>
      </div>

      {/* Status Banner */}
      <div className={`flex items-center gap-3 rounded-lg border px-4 py-3 ${statusColors[res.status]}`}>
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-white/10">
          {res.status === "checked_in" ? <CheckCircle2 className="h-5 w-5" /> :
           res.status === "checked_out" ? <DoorClosed className="h-5 w-5" /> :
           res.status === "cancelled" ? <Ban className="h-5 w-5" /> :
           <Clock className="h-5 w-5" />}
        </div>
        <div>
          <p className="text-sm font-medium">{res.status.replace("_", " ").toUpperCase()}</p>
          {res.checkedInAt && <p className="text-xs opacity-80">Checked in: {res.checkedInAt}</p>}
          {res.checkedOutAt && <p className="text-xs opacity-80">Checked out: {res.checkedOutAt}</p>}
        </div>
        <div className="ml-auto flex items-center gap-4">
          <div className="text-right">
            <p className="text-xs opacity-80">Balance Due</p>
            <p className={`text-lg font-bold ${balanceDue > 0 ? "text-amber-400" : "text-emerald-400"}`}>
              ${balanceDue.toFixed(2)}
            </p>
          </div>
          <div className="text-right">
            <p className="text-xs opacity-80">Total</p>
            <p className="text-lg font-bold">${grandTotal.toFixed(2)}</p>
          </div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        {(["overview", "charges", "invoice", "history"] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setActiveTab(tab)}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              activeTab === tab ? "bg-slate-800 text-white" : "text-slate-400 hover:bg-slate-800 hover:text-white"
            }`}
          >
            {tab === "overview" ? "Overview" : tab === "charges" ? "Charges & Folio" : tab === "invoice" ? "Invoice Preview" : "History"}
          </button>
        ))}
      </div>

      {/* Overview Tab */}
      {activeTab === "overview" && (
        <div className="grid grid-cols-2 gap-4">
          {/* Guest Card */}
          <div className="card space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <User className="h-5 w-5 text-nexus-400" /> Guest Details
              </h3>
              <button onClick={() => setEditingGuest(true)} className="text-xs text-nexus-400 hover:text-nexus-300 flex items-center gap-1">
                <Edit3 className="h-3 w-3" /> Edit
              </button>
            </div>
            <div className="space-y-3">
              <div className="flex items-center gap-3">
                <div className={`flex h-12 w-12 items-center justify-center rounded-full ${res.vip ? "bg-amber-500/10" : "bg-slate-800"}`}>
                  <User className={`h-6 w-6 ${res.vip ? "text-amber-400" : "text-slate-400"}`} />
                </div>
                <div>
                  <p className="text-lg font-bold text-white">{res.guest_name}</p>
                  {res.vip && <span className="rounded bg-amber-500/10 px-2 py-0.5 text-xs text-amber-400">VIP Guest</span>}
                  {res.loyalty_id && <span className="ml-2 rounded bg-sky-500/10 px-2 py-0.5 text-xs text-sky-400">{res.loyalty_id}</span>}
                </div>
              </div>
              <div className="grid grid-cols-2 gap-3 text-sm">
                <div className="flex items-center gap-2 text-slate-300">
                  <Mail className="h-4 w-4 text-slate-500" /> {res.email}
                </div>
                <div className="flex items-center gap-2 text-slate-300">
                  <Phone className="h-4 w-4 text-slate-500" /> {res.phone}
                </div>
                <div className="flex items-center gap-2 text-slate-300 col-span-2">
                  <MapPin className="h-4 w-4 text-slate-500" /> {res.address}, {res.country}
                </div>
              </div>
            </div>
          </div>

          {/* Stay Card */}
          <div className="card space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <BedDouble className="h-5 w-5 text-nexus-400" /> Stay Details
              </h3>
              <button onClick={() => setShowChangeRoom(true)} className="text-xs text-nexus-400 hover:text-nexus-300 flex items-center gap-1">
                <ArrowRightLeft className="h-3 w-3" /> Change Room
              </button>
            </div>
            <div className="space-y-3 text-sm">
              <div className="flex justify-between">
                <span className="text-slate-400">Room</span>
                <span className="font-medium text-white">{res.room_number} — {res.room_type}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Check In</span>
                <span className="text-white">{res.check_in}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Check Out</span>
                <span className="text-white">{res.check_out}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Nights</span>
                <span className="text-white">{res.nights}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Guests</span>
                <span className="text-white">{res.adults} adults{res.children > 0 ? `, ${res.children} children` : ""}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-slate-400">Source</span>
                <span className="rounded bg-slate-800 px-2 py-0.5 text-xs text-slate-300">{res.source.replace("_", " ")}</span>
              </div>
              {res.special_requests && (
                <div className="rounded bg-slate-800/50 p-2 text-xs text-slate-300">
                  <span className="text-slate-500">Special Requests:</span> {res.special_requests}
                </div>
              )}
            </div>
          </div>

          {/* Folio Summary */}
          <div className="card col-span-2">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <Receipt className="h-5 w-5 text-nexus-400" /> Folio Summary
              </h3>
              <button onClick={() => setShowAddCharge(true)} className="btn-primary text-xs gap-1">
                <Plus className="h-3 w-3" /> Add Charge
              </button>
            </div>
            <div className="grid grid-cols-4 gap-4">
              <div className="rounded bg-slate-800/50 p-3 text-center">
                <p className="text-xs text-slate-400">Room Charges</p>
                <p className="text-lg font-bold text-white">${charges.filter(c => c.category === "room").reduce((s, c) => s + c.total, 0).toFixed(2)}</p>
              </div>
              <div className="rounded bg-slate-800/50 p-3 text-center">
                <p className="text-xs text-slate-400">Services</p>
                <p className="text-lg font-bold text-white">${charges.filter(c => c.category !== "room").reduce((s, c) => s + c.total, 0).toFixed(2)}</p>
              </div>
              <div className="rounded bg-slate-800/50 p-3 text-center">
                <p className="text-xs text-slate-400">Paid</p>
                <p className="text-lg font-bold text-emerald-400">${(res.paid || 0).toFixed(2)}</p>
              </div>
              <div className="rounded bg-slate-800/50 p-3 text-center">
                <p className="text-xs text-slate-400">Balance</p>
                <p className={`text-lg font-bold ${balanceDue > 0 ? "text-amber-400" : "text-emerald-400"}`}>${balanceDue.toFixed(2)}</p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Charges Tab */}
      {activeTab === "charges" && (
        <div className="card space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold text-white">Folio Charges</h3>
            <button onClick={() => setShowAddCharge(true)} className="btn-primary gap-2 text-xs">
              <Plus className="h-3 w-3" /> Add Charge
            </button>
          </div>
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-800 text-xs uppercase text-slate-400">
              <tr>
                <th className="px-4 py-2">Date</th>
                <th className="px-4 py-2">Description</th>
                <th className="px-4 py-2">Category</th>
                <th className="px-4 py-2">Qty</th>
                <th className="px-4 py-2">Unit</th>
                <th className="px-4 py-2 text-right">Total</th>
                <th className="px-4 py-2 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {charges.map((c) => (
                <tr key={c.id} className="hover:bg-slate-800/50">
                  <td className="px-4 py-2 text-slate-400">{c.postedAt}</td>
                  <td className="px-4 py-2 text-white">{c.description}{c.staffName && <span className="text-slate-500 text-xs"> · {c.staffName}</span>}</td>
                  <td className="px-4 py-2">
                    <span className="rounded bg-slate-800 px-2 py-0.5 text-xs text-slate-300 capitalize">{c.category}</span>
                  </td>
                  <td className="px-4 py-2 text-slate-300">{c.qty}</td>
                  <td className="px-4 py-2 text-slate-300">${c.amount.toFixed(2)}</td>
                  <td className="px-4 py-2 text-right font-medium text-white">${c.total.toFixed(2)}</td>
                  <td className="px-4 py-2 text-right">
                    <button className="text-rose-400 hover:text-rose-300"><Trash2 className="h-4 w-4" /></button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          <div className="flex justify-end gap-6 border-t border-slate-700 pt-4 text-sm">
            <div className="text-right">
              <p className="text-slate-400">Subtotal</p>
              <p className="font-medium text-white">${subtotal.toFixed(2)}</p>
            </div>
            <div className="text-right">
              <p className="text-slate-400">Tax (12%)</p>
              <p className="font-medium text-white">${tax.toFixed(2)}</p>
            </div>
            <div className="text-right">
              <p className="text-slate-400">Total</p>
              <p className="font-bold text-white">${grandTotal.toFixed(2)}</p>
            </div>
          </div>
        </div>
      )}

      {/* Invoice Tab */}
      {activeTab === "invoice" && (
        <InvoicePreview res={res} charges={charges} subtotal={subtotal} tax={tax} grandTotal={grandTotal} balanceDue={balanceDue} />
      )}

      {/* History Tab */}
      {activeTab === "history" && (
        <div className="card space-y-3">
          <h3 className="text-lg font-semibold text-white">Activity History</h3>
          <div className="space-y-2">
            {[
              { time: "2026-05-10 14:32", action: "Checked in", user: "Front Desk", icon: DoorOpen, color: "emerald" },
              { time: "2026-05-10 14:30", action: "Access code generated", user: "System", icon: KeyRound, color: "sky" },
              { time: "2026-05-10 09:00", action: "Pre-arrival message sent", user: "System", icon: Mail, color: "sky" },
              { time: "2026-04-28 10:15", action: "Reservation created", user: "Online Booking", icon: FileText, color: "nexus" },
            ].map((h, i) => (
              <div key={i} className="flex items-center gap-3 rounded bg-slate-800/50 p-3">
                <div className={`flex h-8 w-8 items-center justify-center rounded-full bg-${h.color}-500/10`}>
                  <h.icon className={`h-4 w-4 text-${h.color}-400`} />
                </div>
                <div className="flex-1">
                  <p className="text-sm text-white">{h.action}</p>
                  <p className="text-xs text-slate-400">{h.user}</p>
                </div>
                <span className="text-xs text-slate-500">{h.time}</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Modals */}
      {editingGuest && <EditGuestModal res={res} onSave={handleGuestSave} onClose={() => setEditingGuest(false)} />}
      {showChangeRoom && <ChangeRoomModal currentRoom={res.room_number || ""} rooms={availableRooms} onSelect={handleRoomChange} onClose={() => setShowChangeRoom(false)} />}
      {showAddCharge && <AddChargeModal onAdd={handleAddCharge} onClose={() => setShowAddCharge(false)} />}
      {showInvoice && <InvoiceModal res={res} charges={charges} subtotal={subtotal} tax={tax} grandTotal={grandTotal} balanceDue={balanceDue} onClose={() => setShowInvoice(false)} />}
      {showCheckInOut === "checkin" && <CheckInModal res={res} onConfirm={handleCheckIn} onClose={() => setShowCheckInOut(null)} />}
      {showCheckInOut === "checkout" && <CheckOutModal res={res} balanceDue={balanceDue} onConfirm={handleCheckOut} onClose={() => setShowCheckInOut(null)} />}
      {showCancel && <CancelModal res={res} onConfirm={handleCancel} onClose={() => setShowCancel(false)} />}
    </div>
  );
}

function EditGuestModal({ res, onSave, onClose }: { res: Reservation; onSave: (u: Partial<Reservation>) => void; onClose: () => void }) {
  const [form, setForm] = useState({ ...res });
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">Edit Guest</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>
        <div className="space-y-3">
          <div>
            <label className="mb-1 block text-xs text-slate-400">Name</label>
            <input className="input w-full" value={form.guest_name} onChange={(e) => setForm({ ...form, guest_name: e.target.value })} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Email</label>
              <input className="input w-full" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Phone</label>
              <input className="input w-full" value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} />
            </div>
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Address</label>
            <input className="input w-full" value={form.address} onChange={(e) => setForm({ ...form, address: e.target.value })} />
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Special Requests</label>
            <textarea className="input w-full" rows={2} value={form.special_requests} onChange={(e) => setForm({ ...form, special_requests: e.target.value })} />
          </div>
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <button onClick={onClose} className="btn-secondary">Cancel</button>
          <button onClick={() => onSave(form)} className="btn-primary gap-2"><Save className="h-4 w-4" /> Save</button>
        </div>
      </div>
    </div>
  );
}

function ChangeRoomModal({ currentRoom, rooms, onSelect, onClose }: { currentRoom: string; rooms: Room[]; onSelect: (r: string) => void; onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">Change Room</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>
        <p className="text-sm text-slate-400">Current: <span className="text-white font-medium">{currentRoom}</span></p>
        <div className="grid grid-cols-2 gap-2">
          {rooms.map((r) => (
            <button key={r.number} onClick={() => onSelect(r.number)} className="rounded-lg border border-slate-700 bg-slate-800 p-3 text-left hover:border-nexus-500 hover:bg-slate-700 transition-colors">
              <p className="text-lg font-bold text-white">{r.number}</p>
              <p className="text-xs text-slate-400">{r.type} · Floor {r.floor}</p>
              <span className="mt-1 inline-block rounded bg-emerald-500/10 px-1.5 py-0.5 text-[10px] text-emerald-400">Vacant Clean</span>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}

function AddChargeModal({ onAdd, onClose }: { onAdd: (c: Charge) => void; onClose: () => void }) {
  const [form, setForm] = useState({ description: "", amount: 0, qty: 1, category: "service" as Charge["category"] });
  const cats = ["room", "service", "minibar", "restaurant", "spa", "late", "other"];
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">Add Charge</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>
        <div className="space-y-3">
          <div>
            <label className="mb-1 block text-xs text-slate-400">Description</label>
            <input className="input w-full" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Amount</label>
              <input type="number" className="input w-full" value={form.amount} onChange={(e) => setForm({ ...form, amount: parseFloat(e.target.value) || 0 })} />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Quantity</label>
              <input type="number" className="input w-full" value={form.qty} onChange={(e) => setForm({ ...form, qty: parseInt(e.target.value) || 1 })} />
            </div>
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Category</label>
            <div className="flex flex-wrap gap-2">
              {cats.map((c) => (
                <button key={c} onClick={() => setForm({ ...form, category: c as Charge["category"] })} className={`rounded px-3 py-1 text-xs capitalize ${form.category === c ? "bg-nexus-500 text-white" : "bg-slate-800 text-slate-400 hover:text-white"}`}>
                  {c.replace("_", " ")}
                </button>
              ))}
            </div>
          </div>
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <button onClick={onClose} className="btn-secondary">Cancel</button>
          <button onClick={() => onAdd({ ...form, id: `ch-${Date.now()}`, total: form.amount * form.qty, postedAt: new Date().toISOString().split("T")[0] })} className="btn-primary gap-2"><Plus className="h-4 w-4" /> Add</button>
        </div>
      </div>
    </div>
  );
}

function InvoicePreview({ res, charges, subtotal, tax, grandTotal, balanceDue }: { res: Reservation; charges: Charge[]; subtotal: number; tax: number; grandTotal: number; balanceDue: number }) {
  return (
    <div className="card space-y-4 max-w-3xl mx-auto">
      <div className="flex items-center justify-between border-b border-slate-700 pb-4">
        <div>
          <h2 className="text-xl font-bold text-white">NEXUS HOTELS</h2>
          <p className="text-xs text-slate-400">123 Hospitality Blvd, New York, NY 10001</p>
          <p className="text-xs text-slate-400">Tel: +1-555-HOTEL · info@nexushotels.com</p>
        </div>
        <div className="text-right">
          <p className="text-sm text-slate-400">Invoice</p>
          <p className="text-lg font-bold text-white">#{res.id.toUpperCase()}</p>
          <p className="text-xs text-slate-400">{new Date().toISOString().split("T")[0]}</p>
        </div>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <p className="text-xs text-slate-500 uppercase">Bill To</p>
          <p className="text-sm font-medium text-white">{res.guest_name}</p>
          <p className="text-xs text-slate-400">{res.email}</p>
          <p className="text-xs text-slate-400">{res.phone}</p>
          <p className="text-xs text-slate-400">{res.address}</p>
        </div>
        <div className="text-right">
          <p className="text-xs text-slate-500 uppercase">Stay Details</p>
          <p className="text-sm text-white">Room {res.room_number} — {res.room_type}</p>
          <p className="text-xs text-slate-400">{res.check_in} → {res.check_out} ({res.nights} nights)</p>
        </div>
      </div>
      <table className="w-full text-sm">
        <thead className="border-b border-slate-700 text-xs uppercase text-slate-400">
          <tr>
            <th className="py-2 text-left">Description</th>
            <th className="py-2 text-right">Qty</th>
            <th className="py-2 text-right">Unit</th>
            <th className="py-2 text-right">Total</th>
          </tr>
        </thead>
        <tbody className="divide-y divide-slate-800">
          {charges.map((c) => (
            <tr key={c.id}>
              <td className="py-2 text-white">{c.description}</td>
              <td className="py-2 text-right text-slate-300">{c.qty}</td>
              <td className="py-2 text-right text-slate-300">${c.amount.toFixed(2)}</td>
              <td className="py-2 text-right font-medium text-white">${c.total.toFixed(2)}</td>
            </tr>
          ))}
        </tbody>
      </table>
      <div className="border-t border-slate-700 pt-4 space-y-1 text-sm">
        <div className="flex justify-between"><span className="text-slate-400">Subtotal</span><span className="text-white">${subtotal.toFixed(2)}</span></div>
        <div className="flex justify-between"><span className="text-slate-400">Tax (12%)</span><span className="text-white">${tax.toFixed(2)}</span></div>
        <div className="flex justify-between"><span className="text-slate-400">Deposit</span><span className="text-white">-${(res.deposit || 0).toFixed(2)}</span></div>
        <div className="flex justify-between"><span className="text-slate-400">Paid</span><span className="text-emerald-400">-${(res.paid || 0).toFixed(2)}</span></div>
        <div className="flex justify-between border-t border-slate-700 pt-2 text-lg font-bold">
          <span className={balanceDue > 0 ? "text-amber-400" : "text-white"}>Balance Due</span>
          <span className={balanceDue > 0 ? "text-amber-400" : "text-emerald-400"}>${balanceDue.toFixed(2)}</span>
        </div>
      </div>
      <div className="flex justify-end gap-2">
        <button className="btn-secondary gap-2"><Printer className="h-4 w-4" /> Print</button>
        <button className="btn-primary gap-2"><FileText className="h-4 w-4" /> Download PDF</button>
      </div>
    </div>
  );
}

function InvoiceModal(props: React.ComponentProps<typeof InvoicePreview> & { onClose: () => void }) {
  const { onClose, ...rest } = props;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4 overflow-auto">
      <div className="w-full max-w-4xl">
        <div className="mb-4 flex justify-end">
          <button onClick={onClose} className="rounded-lg bg-slate-800 p-2 text-white hover:bg-slate-700"><X className="h-5 w-5" /></button>
        </div>
        <InvoicePreview {...rest} />
      </div>
    </div>
  );
}

function CheckInModal({ res, onConfirm, onClose }: { res: Reservation; onConfirm: () => void; onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-sm rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl space-y-4">
        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-emerald-500/10 mx-auto">
          <DoorOpen className="h-6 w-6 text-emerald-400" />
        </div>
        <h3 className="text-center text-lg font-bold text-white">Check In Guest</h3>
        <div className="text-center text-sm text-slate-400 space-y-1">
          <p>{res.guest_name}</p>
          <p>Room {res.room_number} · {res.room_type}</p>
          <p>{res.check_in} → {res.check_out}</p>
        </div>
        <div className="flex gap-2">
          <button onClick={onClose} className="flex-1 btn-secondary">Cancel</button>
          <button onClick={onConfirm} className="flex-1 btn-primary gap-2"><DoorOpen className="h-4 w-4" /> Confirm</button>
        </div>
      </div>
    </div>
  );
}

function CheckOutModal({ res, balanceDue, onConfirm, onClose }: { res: Reservation; balanceDue: number; onConfirm: () => void; onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-sm rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl space-y-4">
        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-amber-500/10 mx-auto">
          <DoorClosed className="h-6 w-6 text-amber-400" />
        </div>
        <h3 className="text-center text-lg font-bold text-white">Check Out Guest</h3>
        <div className="text-center text-sm text-slate-400 space-y-1">
          <p>{res.guest_name}</p>
          <p>Room {res.room_number}</p>
          {balanceDue > 0 && (
            <div className="flex items-center justify-center gap-2 rounded bg-rose-500/10 p-2 text-rose-400">
              <AlertTriangle className="h-4 w-4" /> Outstanding balance: ${balanceDue.toFixed(2)}
            </div>
          )}
        </div>
        <div className="flex gap-2">
          <button onClick={onClose} className="flex-1 btn-secondary">Cancel</button>
          <button onClick={onConfirm} className="flex-1 btn-primary gap-2 bg-amber-600 hover:bg-amber-500"><DoorClosed className="h-4 w-4" /> Confirm</button>
        </div>
      </div>
    </div>
  );
}

function CancelModal({ res, onConfirm, onClose }: { res: Reservation; onConfirm: () => void; onClose: () => void }) {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-sm rounded-xl border border-rose-500/20 bg-slate-900 p-6 shadow-2xl space-y-4">
        <div className="flex h-12 w-12 items-center justify-center rounded-full bg-rose-500/10 mx-auto">
          <Ban className="h-6 w-6 text-rose-400" />
        </div>
        <h3 className="text-center text-lg font-bold text-white">Cancel Reservation</h3>
        <p className="text-center text-sm text-slate-400">
          This will cancel the reservation for <span className="text-white font-medium">{res.guest_name}</span>. This action cannot be undone.
        </p>
        <div className="flex gap-2">
          <button onClick={onClose} className="flex-1 btn-secondary">Keep Reservation</button>
          <button onClick={onConfirm} className="flex-1 btn-secondary gap-2 bg-rose-500/10 text-rose-400 hover:bg-rose-500/20"><Ban className="h-4 w-4" /> Cancel</button>
        </div>
      </div>
    </div>
  );
}
