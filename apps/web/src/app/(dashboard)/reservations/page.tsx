"use client";

import { useState } from "react";
import { Plus, Search, Filter, CalendarDays, User, ChevronRight } from "lucide-react";

const tenant = "demo";

interface Reservation {
  id: string;
  guest_name: string;
  room_number: string;
  check_in: string;
  check_out: string;
  status: "confirmed" | "checked_in" | "checked_out" | "cancelled";
}

const mockReservations: Reservation[] = [
  { id: "r-001", guest_name: "Alice Chen", room_number: "201", check_in: "2026-05-10", check_out: "2026-05-14", status: "checked_in" },
  { id: "r-002", guest_name: "Bob Smith", room_number: "305", check_in: "2026-05-12", check_out: "2026-05-15", status: "confirmed" },
  { id: "r-003", guest_name: "Carol Jones", room_number: "412", check_in: "2026-05-08", check_out: "2026-05-11", status: "checked_out" },
];

const statusBadge: Record<string, string> = {
  confirmed: "badge-blue",
  checked_in: "badge-green",
  checked_out: "badge-amber",
  cancelled: "badge-red",
};

export default function ReservationsPage() {
  const [search, setSearch] = useState("");
  const filtered = mockReservations.filter((r) =>
    r.guest_name.toLowerCase().includes(search.toLowerCase()) ||
    r.room_number.includes(search)
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Reservations</h2>
          <p className="text-sm text-slate-400">Manage bookings and check-ins</p>
        </div>
        <button className="btn-primary">
          <Plus className="h-4 w-4" />
          New Reservation
        </button>
      </div>

      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
          <input
            className="input w-full pl-9"
            placeholder="Search by guest name or room..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <button className="btn-secondary">
          <Filter className="h-4 w-4" />
          Filter
        </button>
      </div>

      <div className="card overflow-hidden p-0">
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-800 text-xs uppercase text-slate-400">
            <tr>
              <th className="px-4 py-3">Guest</th>
              <th className="px-4 py-3">Room</th>
              <th className="px-4 py-3">Check In</th>
              <th className="px-4 py-3">Check Out</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            {filtered.map((r) => (
              <tr key={r.id} className="hover:bg-slate-800/50">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-3">
                    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-nexus-500/10">
                      <User className="h-4 w-4 text-nexus-400" />
                    </div>
                    <span className="font-medium text-white">{r.guest_name}</span>
                  </div>
                </td>
                <td className="px-4 py-3 text-slate-300">{r.room_number}</td>
                <td className="px-4 py-3 text-slate-300">{r.check_in}</td>
                <td className="px-4 py-3 text-slate-300">{r.check_out}</td>
                <td className="px-4 py-3">
                  <span className={statusBadge[r.status]}>{r.status.replace("_", " ")}</span>
                </td>
                <td className="px-4 py-3 text-right">
                  <button className="text-slate-400 hover:text-white">
                    <ChevronRight className="h-4 w-4" />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {filtered.length === 0 && (
          <p className="py-8 text-center text-sm text-slate-500">No reservations found.</p>
        )}
      </div>
    </div>
  );
}
