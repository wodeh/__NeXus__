"use client";

import { useState, useCallback } from "react";
import Link from "next/link";
import {
  Plus,
  Search,
  Filter,
  CalendarDays,
  User,
  ChevronRight,
  AlertTriangle,
  CheckCircle2,
  Clock,
  BedDouble,
  ArrowRightLeft,
  DoorOpen,
  Ban,
  X,
  Save,
  MapPin,
} from "lucide-react";

interface Reservation {
  id: string;
  guest_name: string;
  email?: string;
  phone?: string;
  room_number: string;
  room_type: string;
  check_in: string;
  check_out: string;
  nights: number;
  adults: number;
  children: number;
  status: "confirmed" | "checked_in" | "checked_out" | "cancelled" | "no_show";
  source: "walk_in" | "ota" | "direct" | "agent";
  total: number;
  balance: number;
  special_requests?: string;
  vip?: boolean;
}

const mockReservations: Reservation[] = [
  { id: "r-001", guest_name: "Alice Chen", email: "alice@email.com", phone: "+1-555-0101", room_number: "201", room_type: "Deluxe King", check_in: "2026-05-10", check_out: "2026-05-14", nights: 4, adults: 2, children: 0, status: "checked_in", source: "direct", total: 129*4, balance: 0, special_requests: "Late checkout requested", vip: true },
  { id: "r-002", guest_name: "Bob Smith", email: "bob@email.com", phone: "+1-555-0102", room_number: "305", room_type: "Standard", check_in: "2026-05-12", check_out: "2026-05-15", nights: 3, adults: 1, children: 0, status: "confirmed", source: "ota", total: 89*3, balance: 267, vip: false },
  { id: "r-003", guest_name: "Carol Jones", email: "carol@email.com", phone: "+1-555-0103", room_number: "412", room_type: "Suite", check_in: "2026-05-08", check_out: "2026-05-11", nights: 3, adults: 2, children: 2, status: "checked_out", source: "agent", total: 229*3, balance: 0, vip: false },
  { id: "r-004", guest_name: "David Lee", email: "david@email.com", room_number: "103", room_type: "Deluxe", check_in: "2026-05-10", check_out: "2026-05-15", nights: 5, adults: 2, children: 0, status: "checked_in", source: "walk_in", total: 129*5, balance: 129, vip: false },
  { id: "r-005", guest_name: "Emma Wilson", email: "emma@email.com", room_number: "202", room_type: "Deluxe King", check_in: "2026-05-14", check_out: "2026-05-18", nights: 4, adults: 2, children: 1, status: "confirmed", source: "ota", total: 129*4, balance: 516, vip: true },
  { id: "r-006", guest_name: "Frank Brown", room_number: "", room_type: "Standard", check_in: "2026-05-10", check_out: "2026-05-12", nights: 2, adults: 1, children: 0, status: "confirmed", source: "direct", total: 89*2, balance: 178, vip: false },
  { id: "r-007", guest_name: "Grace Kim", room_number: "301", room_type: "Suite", check_in: "2026-05-09", check_out: "2026-05-13", nights: 4, adults: 2, children: 0, status: "no_show", source: "ota", total: 229*4, balance: 916, vip: false },
  { id: "r-008", guest_name: "Henry Park", room_number: "204", room_type: "Deluxe King", check_in: "2026-05-11", check_out: "2026-05-16", nights: 5, adults: 2, children: 0, status: "confirmed", source: "walk_in", total: 129*5, balance: 645, vip: false },
];

const mockRooms = [
  { number: "101", type: "Standard", status: "occupied", floor: "1" },
  { number: "102", type: "Standard", status: "vacant_clean", floor: "1" },
  { number: "103", type: "Deluxe", status: "occupied", floor: "1" },
  { number: "104", type: "Standard", status: "blocked", floor: "1" },
  { number: "201", type: "Deluxe King", status: "occupied", floor: "2" },
  { number: "202", type: "Deluxe King", status: "blocked", floor: "2" },
  { number: "203", type: "Suite", status: "occupied", floor: "2" },
  { number: "204", type: "Deluxe King", status: "vacant_clean", floor: "2" },
  { number: "301", type: "Suite", status: "occupied", floor: "3" },
  { number: "302", type: "Suite", status: "blocked", floor: "3" },
  { number: "305", type: "Standard", status: "blocked", floor: "3" },
  { number: "412", type: "Suite", status: "vacant_dirty", floor: "4" },
];

const statusBadge: Record<string, string> = {
  confirmed: "badge-blue",
  checked_in: "badge-green",
  checked_out: "badge-amber",
  cancelled: "badge-red",
  no_show: "badge-red",
};

export default function ReservationsPage() {
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");
  const [dateFilter, setDateFilter] = useState<string>("all");
  const [showQuickBook, setShowQuickBook] = useState(false);
  const [showAssignRoom, setShowAssignRoom] = useState<string | null>(null);
  const [draggingResId, setDraggingResId] = useState<string | null>(null);

  const today = "2026-05-10";

  const filtered = mockReservations.filter((r) => {
    const matchesSearch =
      r.guest_name.toLowerCase().includes(search.toLowerCase()) ||
      r.room_number.includes(search) ||
      r.email?.toLowerCase().includes(search.toLowerCase());
    const matchesStatus = statusFilter === "all" || r.status === statusFilter;
    const matchesDate =
      dateFilter === "all"
        ? true
        : dateFilter === "arrivals"
        ? r.check_in === today
        : dateFilter === "departures"
        ? r.check_out === today
        : dateFilter === "in_house"
        ? r.status === "checked_in"
        : dateFilter === "upcoming"
        ? r.status === "confirmed"
        : true;
    return matchesSearch && matchesStatus && matchesDate;
  });

  const stats = {
    total: mockReservations.length,
    inHouse: mockReservations.filter((r) => r.status === "checked_in").length,
    arrivalsToday: mockReservations.filter((r) => r.check_in === today && r.status !== "checked_in").length,
    departuresToday: mockReservations.filter((r) => r.check_out === today && r.status === "checked_in").length,
    unassigned: mockReservations.filter((r) => r.status === "confirmed" && !r.room_number).length,
    overbooked: mockReservations.filter((r) => r.status === "confirmed" && !r.room_number).length,
    revenue: mockReservations.filter((r) => r.status === "checked_in").reduce((s, r) => s + r.total, 0),
  };

  const handleRoomDrop = useCallback((resId: string, roomNumber: string) => {
    // In production: API call to assign room
    setShowAssignRoom(null);
    setDraggingResId(null);
  }, []);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Reservations</h1>
          <p className="text-sm text-slate-400">Bookings, check-ins, room assignments</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => setShowQuickBook(true)} className="btn-primary">
            <Plus className="h-4 w-4" />
            Quick Book (Walk-in)
          </button>
          <Link href="/reservations/new" className="btn-secondary">
            <CalendarDays className="h-4 w-4" />
            Full Booking
          </Link>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-6 gap-3">
        <StatCard label="Total" value={stats.total} icon={BedDouble} color="sky" />
        <StatCard label="In House" value={stats.inHouse} icon={CheckCircle2} color="emerald" />
        <StatCard label="Arrivals Today" value={stats.arrivalsToday} icon={ArrowRightLeft} color="violet" />
        <StatCard label="Departures Today" value={stats.departuresToday} icon={DoorOpen} color="amber" />
        <StatCard label="Unassigned" value={stats.unassigned} icon={MapPin} color="rose" alert={stats.unassigned > 0} />
        <StatCard label="In-House Revenue" value={`$${stats.revenue.toLocaleString()}`} icon={Clock} color="nexus" />
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
          <input
            className="input w-full pl-9"
            placeholder="Search guest, room, email..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <div className="flex gap-1">
          {[
            { key: "all", label: "All" },
            { key: "arrivals", label: "Arrivals" },
            { key: "departures", label: "Departures" },
            { key: "in_house", label: "In House" },
            { key: "upcoming", label: "Upcoming" },
          ].map((f) => (
            <button
              key={f.key}
              onClick={() => setDateFilter(f.key)}
              className={`rounded-lg px-3 py-2 text-xs font-medium transition-colors ${
                dateFilter === f.key
                  ? "bg-nexus-500 text-white"
                  : "bg-slate-800 text-slate-400 hover:bg-slate-700 hover:text-white"
              }`}
            >
              {f.label}
            </button>
          ))}
        </div>
        <select
          value={statusFilter}
          onChange={(e) => setStatusFilter(e.target.value)}
          className="input text-xs"
        >
          <option value="all">All Statuses</option>
          <option value="confirmed">Confirmed</option>
          <option value="checked_in">Checked In</option>
          <option value="checked_out">Checked Out</option>
          <option value="cancelled">Cancelled</option>
          <option value="no_show">No Show</option>
        </select>
      </div>

      {/* Overbooking Warning */}
      {stats.unassigned > 0 && (
        <div className="flex items-center gap-2 rounded-lg border border-rose-500/20 bg-rose-500/5 px-4 py-3">
          <AlertTriangle className="h-4 w-4 text-rose-400" />
          <span className="text-sm text-rose-400">
            {stats.unassigned} reservation{stats.unassigned > 1 ? "s" : ""} confirmed but no room assigned — potential overbooking
          </span>
        </div>
      )}

      {/* Room Assignment Panel (drop zone) */}
      {draggingResId && (
        <div className="rounded-lg border border-dashed border-nexus-500/30 bg-nexus-500/5 p-3">
          <p className="text-xs text-nexus-400">Drag to a room below to assign, or select:</p>
          <div className="mt-2 flex flex-wrap gap-2">
            {mockRooms
              .filter((rm) => rm.status === "vacant_clean")
              .map((rm) => (
                <button
                  key={rm.number}
                  onClick={() => handleRoomDrop(draggingResId, rm.number)}
                  className="rounded border border-slate-700 bg-slate-800 px-3 py-1.5 text-xs text-white hover:border-nexus-500 hover:bg-slate-700"
                >
                  {rm.number} · {rm.type}
                </button>
              ))}
          </div>
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
              <th className="px-4 py-3">Source</th>
              <th className="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            {filtered.map((r) => (
              <tr
                key={r.id}
                className="hover:bg-slate-800/50"
                draggable
                onDragStart={() => setDraggingResId(r.id)}
                onDragEnd={() => setDraggingResId(null)}
              >
                <td className="px-4 py-3">
                  <div className="flex items-center gap-3">
                    <div className={`flex h-8 w-8 items-center justify-center rounded-full ${r.vip ? "bg-amber-500/10" : "bg-nexus-500/10"}`}>
                      <User className={`h-4 w-4 ${r.vip ? "text-amber-400" : "text-nexus-400"}`} />
                    </div>
                    <div>
                      <span className="font-medium text-white">{r.guest_name}</span>
                      {r.vip && <span className="ml-2 rounded bg-amber-500/10 px-1.5 py-0.5 text-[10px] text-amber-400">VIP</span>}
                      {r.email && <p className="text-[10px] text-slate-500">{r.email}</p>}
                    </div>
                  </div>
                </td>
                <td className="px-4 py-3">
                  {r.room_number ? (
                    <span className="rounded bg-slate-700/50 px-2 py-0.5 text-xs text-white">{r.room_number}</span>
                  ) : (
                    <button
                      onClick={() => setShowAssignRoom(r.id)}
                      className="flex items-center gap-1 rounded bg-rose-500/10 px-2 py-0.5 text-xs text-rose-400 hover:bg-rose-500/20"
                    >
                      <MapPin className="h-3 w-3" />
                      Assign Room
                    </button>
                  )}
                </td>
                <td className="px-4 py-3 text-slate-300">
                  <div className="flex items-center gap-1 text-xs">
                    <CalendarDays className="h-3 w-3 text-slate-500" />
                    {r.check_in} → {r.check_out}
                  </div>
                </td>
                <td className="px-4 py-3 text-slate-300">{r.nights}</td>
                <td className="px-4 py-3 text-slate-300">${r.total}</td>
                <td className="px-4 py-3">
                  <span className={r.balance > 0 ? "text-amber-400" : "text-emerald-400"}>
                    ${r.balance}
                  </span>
                </td>
                <td className="px-4 py-3">
                  <span className={statusBadge[r.status]}>{r.status.replace("_", " ")}</span>
                </td>
                <td className="px-4 py-3">
                  <span className="rounded bg-slate-700/50 px-2 py-0.5 text-[10px] uppercase text-slate-400">
                    {r.source.replace("_", " ")}
                  </span>
                </td>
                <td className="px-4 py-3 text-right">
                  <Link href={`/reservations/${r.id}`} className="text-slate-400 hover:text-white">
                    <ChevronRight className="h-4 w-4" />
                  </Link>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {filtered.length === 0 && (
          <p className="py-8 text-center text-sm text-slate-500">No reservations found.</p>
        )}
      </div>

      {/* Quick Book Modal */}
      {showQuickBook && <QuickBookModal onClose={() => setShowQuickBook(false)} />}

      {/* Assign Room Modal */}
      {showAssignRoom && (
        <AssignRoomModal
          resId={showAssignRoom}
          rooms={mockRooms}
          onClose={() => setShowAssignRoom(null)}
          onAssign={(roomNumber) => handleRoomDrop(showAssignRoom, roomNumber)}
        />
      )}
    </div>
  );
}

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

function QuickBookModal({ onClose }: { onClose: () => void }) {
  const [form, setForm] = useState({
    guest_name: "",
    phone: "",
    email: "",
    room_type: "Standard",
    check_in: "2026-05-10",
    check_out: "2026-05-12",
    adults: 1,
    children: 0,
    special_requests: "",
  });

  const roomTypes = ["Standard", "Deluxe", "Deluxe King", "Suite", "Accessible"];
  const nights = Math.max(1, Math.ceil((new Date(form.check_out).getTime() - new Date(form.check_in).getTime()) / (1000 * 60 * 60 * 24)));

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-lg rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">Quick Book — Walk-in</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white">
            <X className="h-5 w-5" />
          </button>
        </div>
        <p className="mt-1 text-xs text-slate-400">Instant booking for guests arriving now</p>

        <div className="mt-4 space-y-3">
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Guest Name *</label>
              <input
                className="input w-full"
                value={form.guest_name}
                onChange={(e) => setForm({ ...form, guest_name: e.target.value })}
                placeholder="Full name"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Phone</label>
              <input
                className="input w-full"
                value={form.phone}
                onChange={(e) => setForm({ ...form, phone: e.target.value })}
                placeholder="+1-555-0000"
              />
            </div>
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Email</label>
            <input
              className="input w-full"
              value={form.email}
              onChange={(e) => setForm({ ...form, email: e.target.value })}
              placeholder="guest@email.com"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Room Type</label>
              <select
                className="input w-full"
                value={form.room_type}
                onChange={(e) => setForm({ ...form, room_type: e.target.value })}
              >
                {roomTypes.map((t) => (
                  <option key={t} value={t}>{t}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Guests</label>
              <div className="flex items-center gap-2">
                <input
                  type="number"
                  min={1}
                  max={4}
                  className="input w-20"
                  value={form.adults}
                  onChange={(e) => setForm({ ...form, adults: parseInt(e.target.value) || 1 })}
                />
                <span className="text-xs text-slate-500">adults</span>
                <input
                  type="number"
                  min={0}
                  max={3}
                  className="input w-20"
                  value={form.children}
                  onChange={(e) => setForm({ ...form, children: parseInt(e.target.value) || 0 })}
                />
                <span className="text-xs text-slate-500">children</span>
              </div>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Check In</label>
              <input
                type="date"
                className="input w-full"
                value={form.check_in}
                onChange={(e) => setForm({ ...form, check_in: e.target.value })}
              />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Check Out</label>
              <input
                type="date"
                className="input w-full"
                value={form.check_out}
                onChange={(e) => setForm({ ...form, check_out: e.target.value })}
              />
            </div>
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Special Requests</label>
            <textarea
              className="input w-full"
              rows={2}
              value={form.special_requests}
              onChange={(e) => setForm({ ...form, special_requests: e.target.value })}
              placeholder="Any special needs..."
            />
          </div>
        </div>

        <div className="mt-4 flex items-center justify-between border-t border-slate-700 pt-4">
          <div className="text-sm text-slate-400">
            {nights} nights · Est. total: <span className="font-bold text-white">${nights * (form.room_type === "Suite" ? 229 : form.room_type === "Deluxe" || form.room_type === "Deluxe King" ? 129 : 89)}</span>
          </div>
          <div className="flex gap-2">
            <button onClick={onClose} className="btn-secondary">Cancel</button>
            <button className="btn-primary">
              <Save className="h-4 w-4" />
              Book Now
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

function AssignRoomModal({
  rooms,
  onClose,
  onAssign,
}: {
  resId: string;
  rooms: { number: string; type: string; status: string }[];
  onClose: () => void;
  onAssign: (roomNumber: string) => void;
}) {
  const vacantRooms = rooms.filter((r) => r.status === "vacant_clean");
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">Assign Room</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white">
            <X className="h-5 w-5" />
          </button>
        </div>
        <p className="mt-1 text-xs text-slate-400">Select a vacant clean room</p>

        <div className="mt-4 grid grid-cols-3 gap-2">
          {vacantRooms.map((r) => (
            <button
              key={r.number}
              onClick={() => onAssign(r.number)}
              className="rounded-lg border border-slate-700 bg-slate-800 p-3 text-left transition-colors hover:border-nexus-500 hover:bg-slate-700"
            >
              <p className="text-lg font-bold text-white">{r.number}</p>
              <p className="text-xs text-slate-400">{r.type}</p>
              <span className="mt-1 inline-block rounded bg-emerald-500/10 px-1.5 py-0.5 text-[10px] text-emerald-400">
                Vacant Clean
              </span>
            </button>
          ))}
        </div>
        {vacantRooms.length === 0 && (
          <div className="mt-4 rounded-lg border border-rose-500/20 bg-rose-500/5 p-4 text-center">
            <AlertTriangle className="mx-auto h-6 w-6 text-rose-400" />
            <p className="mt-2 text-sm text-rose-400">No vacant clean rooms available</p>
            <p className="text-xs text-slate-500">Housekeeping needed or all rooms occupied</p>
          </div>
        )}
      </div>
    </div>
  );
}
