"use client";

import { useState } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import { Lock, Filter, Square, CheckCircle2, Wrench, Ban } from "lucide-react";

interface RoomStatus {
  id: string;
  number: string;
  floor: string;
  type: string;
  status: "vacant_clean" | "vacant_dirty" | "occupied" | "blocked" | "maintenance" | "out_of_order";
  guest?: string;
  arrival?: string;
  departure?: string;
  blockReason?: string;
}

const mockRooms: RoomStatus[] = [
  { id: "r-101", number: "101", floor: "1", type: "Standard", status: "occupied", guest: "Alice Chen", arrival: "2026-05-10", departure: "2026-05-14" },
  { id: "r-102", number: "102", floor: "1", type: "Standard", status: "vacant_clean" },
  { id: "r-103", number: "103", floor: "1", type: "Deluxe", status: "vacant_dirty" },
  { id: "r-104", number: "104", floor: "1", type: "Standard", status: "blocked", blockReason: "VIP Hold - Mr. Smith" },
  { id: "r-105", number: "105", floor: "1", type: "Accessible", status: "maintenance", blockReason: "Plumbing repair" },
  { id: "r-201", number: "201", floor: "2", type: "Deluxe King", status: "occupied", guest: "Bob Jones", arrival: "2026-05-08", departure: "2026-05-11" },
  { id: "r-202", number: "202", floor: "2", type: "Deluxe King", status: "vacant_clean" },
  { id: "r-203", number: "203", floor: "2", type: "Suite", status: "occupied", guest: "Carol White", arrival: "2026-05-09", departure: "2026-05-12" },
  { id: "r-204", number: "204", floor: "2", type: "Deluxe King", status: "out_of_order", blockReason: "Renovation" },
  { id: "r-205", number: "205", floor: "2", type: "Standard", status: "vacant_dirty" },
  { id: "r-301", number: "301", floor: "3", type: "Suite", status: "vacant_clean" },
  { id: "r-302", number: "302", floor: "3", type: "Suite", status: "blocked", blockReason: "Group Block - Wedding Party" },
  { id: "r-303", number: "303", floor: "3", type: "Deluxe King", status: "occupied", guest: "David Lee", arrival: "2026-05-10", departure: "2026-05-15" },
  { id: "r-304", number: "304", floor: "3", type: "Standard", status: "vacant_clean" },
  { id: "r-305", number: "305", floor: "3", type: "Standard", status: "vacant_dirty" },
];

const statusConfig: Record<string, { label: string; color: string; icon: React.ElementType }> = {
  vacant_clean: { label: "Vacant Clean", color: "bg-emerald-500/10 border-emerald-500/20 text-emerald-400", icon: CheckCircle2 },
  vacant_dirty: { label: "Vacant Dirty", color: "bg-red-500/10 border-red-500/20 text-red-400", icon: Square },
  occupied: { label: "Occupied", color: "bg-blue-500/10 border-blue-500/20 text-blue-400", icon: CheckCircle2 },
  blocked: { label: "Blocked", color: "bg-purple-500/10 border-purple-500/20 text-purple-400", icon: Ban },
  maintenance: { label: "Maintenance", color: "bg-amber-500/10 border-amber-500/20 text-amber-400", icon: Wrench },
  out_of_order: { label: "OOO", color: "bg-slate-700 border-slate-600 text-slate-400", icon: Ban },
};

export default function FloorPage() {
  const { config } = useTenant();
  const [filter, setFilter] = useState<string>("all");
  const [selectedRoom, setSelectedRoom] = useState<RoomStatus | null>(null);

  const hasFloor = hasCapability(config, CAPABILITIES.OPERATIONS.FLOOR_DASHBOARD);

  if (!hasFloor) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-800">
          <Lock className="h-8 w-8 text-slate-500" />
        </div>
        <h2 className="mt-6 text-xl font-bold text-white">Floor Dashboard Locked</h2>
        <p className="mt-2 max-w-md text-sm text-slate-400">
          The interactive floor plan requires the Operations or Enterprise license tier.
        </p>
      </div>
    );
  }

  const floors = Array.from(new Set(mockRooms.map((r) => r.floor))).sort();
  const filteredRooms = filter === "all" ? mockRooms : mockRooms.filter((r) => r.status === filter);

  const occupancy = {
    total: mockRooms.length,
    occupied: mockRooms.filter((r) => r.status === "occupied").length,
    vacantClean: mockRooms.filter((r) => r.status === "vacant_clean").length,
    vacantDirty: mockRooms.filter((r) => r.status === "vacant_dirty").length,
    blocked: mockRooms.filter((r) => r.status === "blocked").length,
    maintenance: mockRooms.filter((r) => ["maintenance", "out_of_order"].includes(r.status)).length,
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Floor Dashboard</h2>
          <p className="text-sm text-slate-400">Real-time room status and occupancy</p>
        </div>
        <div className="flex items-center gap-2 text-sm">
          <span className="text-slate-400">Occupancy:</span>
          <span className="font-semibold text-white">{Math.round((occupancy.occupied / occupancy.total) * 100)}%</span>
          <span className="text-slate-500">({occupancy.occupied}/{occupancy.total})</span>
        </div>
      </div>

      {/* Status summary */}
      <div className="grid grid-cols-6 gap-3">
        {[
          { label: "Total", value: occupancy.total, color: "text-white", bg: "bg-slate-800" },
          { label: "Occupied", value: occupancy.occupied, color: "text-blue-400", bg: "bg-blue-500/10" },
          { label: "Vacant Clean", value: occupancy.vacantClean, color: "text-emerald-400", bg: "bg-emerald-500/10" },
          { label: "Vacant Dirty", value: occupancy.vacantDirty, color: "text-red-400", bg: "bg-red-500/10" },
          { label: "Blocked", value: occupancy.blocked, color: "text-purple-400", bg: "bg-purple-500/10" },
          { label: "Maint / OOO", value: occupancy.maintenance, color: "text-amber-400", bg: "bg-amber-500/10" },
        ].map((stat) => (
          <button
            key={stat.label}
            onClick={() => setFilter(stat.label === "Total" ? "all" : stat.label.toLowerCase().replace(/\s+/g, "_"))}
            className={`rounded-lg ${stat.bg} p-3 text-center transition-colors hover:opacity-80`}
          >
            <p className={`text-2xl font-bold ${stat.color}`}>{stat.value}</p>
            <p className="mt-1 text-xs text-slate-400">{stat.label}</p>
          </button>
        ))}
      </div>

      {/* Filters */}
      <div className="flex items-center gap-2">
        <Filter className="h-4 w-4 text-slate-500" />
        {["all", "vacant_clean", "vacant_dirty", "occupied", "blocked", "maintenance"].map((s) => (
          <button
            key={s}
            onClick={() => setFilter(s)}
            className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
              filter === s
                ? "bg-nexus-500 text-white"
                : "bg-slate-800 text-slate-400 hover:bg-slate-700"
            }`}
          >
            {s === "all" ? "All" : statusConfig[s]?.label || s}
          </button>
        ))}
      </div>

      {/* Floor grids */}
      <div className="space-y-6">
        {floors.map((floor) => {
          const floorRooms = filteredRooms.filter((r) => r.floor === floor);
          if (floorRooms.length === 0) return null;
          return (
            <div key={floor} className="space-y-3">
              <h3 className="text-sm font-semibold text-slate-300">Floor {floor}</h3>
              <div className="grid grid-cols-5 gap-3">
                {floorRooms.map((room) => {
                  const cfg = statusConfig[room.status];
                  const Icon = cfg.icon;
                  return (
                    <button
                      key={room.id}
                      onClick={() => setSelectedRoom(room)}
                      className={`relative rounded-lg border p-4 text-left transition-all hover:scale-[1.02] ${cfg.color}`}
                    >
                      <div className="flex items-center justify-between">
                        <span className="text-lg font-bold">{room.number}</span>
                        <Icon className="h-4 w-4 opacity-60" />
                      </div>
                      <p className="mt-1 text-xs opacity-80">{room.type}</p>
                      {room.guest && (
                        <p className="mt-2 text-xs font-medium truncate">{room.guest}</p>
                      )}
                      {room.blockReason && (
                        <p className="mt-2 text-xs opacity-80 truncate">{room.blockReason}</p>
                      )}
                      {room.departure && (
                        <p className="mt-1 text-[10px] opacity-60">Out: {room.departure}</p>
                      )}
                    </button>
                  );
                })}
              </div>
            </div>
          );
        })}
      </div>

      {/* Room detail modal */}
      {selectedRoom && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={() => setSelectedRoom(null)}>
          <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between">
              <h3 className="text-xl font-bold text-white">Room {selectedRoom.number}</h3>
              <button onClick={() => setSelectedRoom(null)} className="text-slate-400 hover:text-white">×</button>
            </div>
            <div className="mt-4 space-y-3 text-sm">
              <div className="flex justify-between"><span className="text-slate-400">Type</span><span className="text-white">{selectedRoom.type}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Floor</span><span className="text-white">{selectedRoom.floor}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Status</span><span className={statusConfig[selectedRoom.status].color.split(" ")[2]}>{statusConfig[selectedRoom.status].label}</span></div>
              {selectedRoom.guest && (
                <>
                  <div className="flex justify-between"><span className="text-slate-400">Guest</span><span className="text-white">{selectedRoom.guest}</span></div>
                  <div className="flex justify-between"><span className="text-slate-400">Arrival</span><span className="text-white">{selectedRoom.arrival}</span></div>
                  <div className="flex justify-between"><span className="text-slate-400">Departure</span><span className="text-white">{selectedRoom.departure}</span></div>
                </>
              )}
              {selectedRoom.blockReason && (
                <div className="flex justify-between"><span className="text-slate-400">Reason</span><span className="text-white">{selectedRoom.blockReason}</span></div>
              )}
            </div>
            <div className="mt-6 flex gap-2">
              {selectedRoom.status === "vacant_clean" && (
                <button className="btn-primary flex-1">Assign Guest</button>
              )}
              {selectedRoom.status === "vacant_dirty" && (
                <button className="btn-primary flex-1">Mark Clean</button>
              )}
              {selectedRoom.status === "occupied" && (
                <>
                  <button className="btn-primary flex-1">Extend Stay</button>
                  <button className="btn-secondary flex-1">Check Out</button>
                </>
              )}
              {(selectedRoom.status === "vacant_clean" || selectedRoom.status === "vacant_dirty") && (
                <button className="btn-secondary flex-1">Block Room</button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
