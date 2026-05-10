"use client";

import { useState } from "react";
import { Plus, Search, DoorOpen, Wrench, Sparkles, ChevronRight } from "lucide-react";

const mockRooms = [
  { id: "rm-001", number: "201", type: "Deluxe King", floor: "2", status: "occupied", housekeeping: "clean" },
  { id: "rm-002", number: "305", type: "Standard Twin", floor: "3", status: "vacant", housekeeping: "clean" },
  { id: "rm-003", number: "412", type: "Suite", floor: "4", status: "occupied", housekeeping: "dirty" },
  { id: "rm-004", number: "108", type: "Standard Queen", floor: "1", status: "maintenance", housekeeping: "clean" },
];

const statusColors: Record<string, string> = {
  occupied: "badge-green",
  vacant: "badge-blue",
  maintenance: "badge-amber",
  out_of_order: "badge-red",
};

const hkColors: Record<string, string> = {
  clean: "badge-green",
  dirty: "badge-red",
  inspected: "badge-blue",
  in_progress: "badge-amber",
};

export default function RoomsPage() {
  const [search, setSearch] = useState("");
  const filtered = mockRooms.filter((r) => r.number.includes(search) || r.type.toLowerCase().includes(search.toLowerCase()));

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Rooms</h2>
          <p className="text-sm text-slate-400">Room inventory and status</p>
        </div>
        <button className="btn-primary">
          <Plus className="h-4 w-4" />
          Add Room
        </button>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
        <input className="input w-full max-w-md pl-9" placeholder="Search rooms..." value={search} onChange={(e) => setSearch(e.target.value)} />
      </div>

      <div className="card overflow-hidden p-0">
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-800 text-xs uppercase text-slate-400">
            <tr>
              <th className="px-4 py-3">Room</th>
              <th className="px-4 py-3">Type</th>
              <th className="px-4 py-3">Floor</th>
              <th className="px-4 py-3">Status</th>
              <th className="px-4 py-3">Housekeeping</th>
              <th className="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            {filtered.map((r) => (
              <tr key={r.id} className="hover:bg-slate-800/50">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-3">
                    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-nexus-500/10">
                      <DoorOpen className="h-4 w-4 text-nexus-400" />
                    </div>
                    <span className="font-medium text-white">{r.number}</span>
                  </div>
                </td>
                <td className="px-4 py-3 text-slate-300">{r.type}</td>
                <td className="px-4 py-3 text-slate-300">{r.floor}</td>
                <td className="px-4 py-3"><span className={statusColors[r.status]}>{r.status}</span></td>
                <td className="px-4 py-3"><span className={hkColors[r.housekeeping]}>{r.housekeeping}</span></td>
                <td className="px-4 py-3 text-right">
                  <button className="text-slate-400 hover:text-white"><ChevronRight className="h-4 w-4" /></button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
