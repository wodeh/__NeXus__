"use client";

import { useState, useEffect, useCallback, useMemo } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import { useDragScroll } from "@/hooks/useDragScroll";
import { Lock, Filter, Square, CheckCircle2, Wrench, Ban, RefreshCw, AlertTriangle, Grid3X3, List, Columns3 } from "lucide-react";
import { Room, Reservation, getRooms, getReservations } from "@/lib/api";

interface RoomStatus {
  id: string;
  number: string;
  floor: string;
  type: string;
  status: string;
  guest?: string;
  arrival?: string;
  departure?: string;
  blockReason?: string;
}

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
  const [rooms, setRooms] = useState<Room[]>([]);
  const [reservations, setReservations] = useState<Reservation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<"grid" | "horizontal" | "list">("horizontal");
  const dragRef = useDragScroll();

  const hasFloor = hasCapability(config, CAPABILITIES.OPERATIONS.FLOOR_DASHBOARD);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [rms, res] = await Promise.all([getRooms(), getReservations()]);
      setRooms(rms);
      setReservations(res);
    } catch (e: any) {
      setError(e.message || "Failed to load floor data");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

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

  // Build room status from real data
  const roomStatuses: RoomStatus[] = useMemo(() => {
    return rooms.map((r) => {
      const activeRes = reservations.find(
        (res) => res.room_number === r.number && res.status === "checked_in"
      );
      return {
        id: r.id,
        number: r.number,
        floor: r.floor,
        type: r.type,
        status: r.status,
        guest: activeRes?.guest_name,
        arrival: activeRes?.check_in,
        departure: activeRes?.check_out,
        blockReason: r.status === "blocked" ? "Blocked" : undefined,
      };
    });
  }, [rooms, reservations]);

  const floors = useMemo(() => Array.from(new Set(roomStatuses.map((r) => r.floor))).sort(), [roomStatuses]);
  const filteredRooms = filter === "all" ? roomStatuses : roomStatuses.filter((r) => r.status === filter);

  const occupancy = useMemo(() => ({
    total: roomStatuses.length,
    occupied: roomStatuses.filter((r) => r.status === "occupied").length,
    vacantClean: roomStatuses.filter((r) => r.status === "vacant_clean").length,
    vacantDirty: roomStatuses.filter((r) => r.status === "vacant_dirty").length,
    blocked: roomStatuses.filter((r) => r.status === "blocked").length,
    maintenance: roomStatuses.filter((r) => ["maintenance", "out_of_order"].includes(r.status)).length,
  }), [roomStatuses]);

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
          <h2 className="text-2xl font-bold text-white">Floor Dashboard</h2>
          <p className="text-sm text-slate-400">Real-time room status and occupancy</p>
        </div>
        <div className="flex items-center gap-2 text-sm">
          <span className="text-slate-400">Occupancy:</span>
          <span className="font-semibold text-white">{Math.round((occupancy.occupied / occupancy.total) * 100)}%</span>
          <span className="text-slate-500">({occupancy.occupied}/{occupancy.total})</span>
          <button onClick={fetchData} className="ml-2 rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white" title="Refresh">
            <RefreshCw className="h-4 w-4" />
          </button>
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

      {/* View toggle + filters */}
      <div className="flex items-center justify-between">
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
        <div className="flex rounded-lg border border-slate-700 bg-slate-800">
          <button onClick={() => setViewMode("grid")} className={`p-2 ${viewMode === "grid" ? "text-white bg-slate-700" : "text-slate-400 hover:text-white"}`} title="Grid view"><Grid3X3 className="h-4 w-4"/></button>
          <button onClick={() => setViewMode("horizontal")} className={`p-2 ${viewMode === "horizontal" ? "text-white bg-slate-700" : "text-slate-400 hover:text-white"}`} title="Horizontal scroll"><Columns3 className="h-4 w-4"/></button>
          <button onClick={() => setViewMode("list")} className={`p-2 ${viewMode === "list" ? "text-white bg-slate-700" : "text-slate-400 hover:text-white"}`} title="List view"><List className="h-4 w-4"/></button>
        </div>
      </div>

      {/* Floor grids */}
      <div className="space-y-6">
        {floors.map((floor) => {
          const floorRooms = filteredRooms.filter((r) => r.floor === floor);
          if (floorRooms.length === 0) return null;
          return (
            <div key={floor} className="space-y-3">
              <h3 className="text-sm font-semibold text-slate-300">Floor {floor} <span className="text-slate-500 text-xs">({floorRooms.length} rooms)</span></h3>
              {viewMode === "horizontal" ? (
                <div ref={dragRef} className="overflow-x-auto scrollbar-hide cursor-grab">
                  <div className="flex gap-3 min-w-max">
                    {floorRooms.map((room) => {
                      const cfg = statusConfig[room.status] || statusConfig.vacant_clean;
                      const Icon = cfg.icon;
                      return (
                        <button
                          key={room.id}
                          onClick={() => setSelectedRoom(room)}
                          className={`relative flex-shrink-0 w-32 rounded-lg border p-3 text-left transition-all hover:scale-[1.02] ${cfg.color}`}
                        >
                          <div className="flex items-center justify-between">
                            <span className="text-lg font-bold">{room.number}</span>
                            <Icon className="h-4 w-4 opacity-60" />
                          </div>
                          <p className="mt-1 text-xs opacity-80">{room.type}</p>
                          {room.guest && (
                            <p className="mt-2 text-xs font-medium truncate">{room.guest}</p>
                          )}
                          {room.departure && (
                            <p className="mt-1 text-[10px] opacity-60">Out: {room.departure}</p>
                          )}
                        </button>
                      );
                    })}
                  </div>
                </div>
              ) : viewMode === "grid" ? (
                <div className="grid grid-cols-5 gap-3">
                  {floorRooms.map((room) => {
                    const cfg = statusConfig[room.status] || statusConfig.vacant_clean;
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
              ) : (
                <div className="rounded-lg border border-slate-700 bg-slate-800 overflow-hidden">
                  <table className="w-full text-sm">
                    <thead className="bg-slate-800 border-b border-slate-700">
                      <tr>
                        <th className="px-4 py-2 text-left text-xs text-slate-500">Room</th>
                        <th className="px-4 py-2 text-left text-xs text-slate-500">Type</th>
                        <th className="px-4 py-2 text-left text-xs text-slate-500">Status</th>
                        <th className="px-4 py-2 text-left text-xs text-slate-500">Guest</th>
                        <th className="px-4 py-2 text-left text-xs text-slate-500">Departure</th>
                      </tr>
                    </thead>
                    <tbody>
                      {floorRooms.map((room) => {
                        const cfg = statusConfig[room.status] || statusConfig.vacant_clean;
                        return (
                          <tr key={room.id} onClick={() => setSelectedRoom(room)} className="border-b border-slate-700/50 hover:bg-slate-700/20 cursor-pointer">
                            <td className="px-4 py-2 font-bold text-white">{room.number}</td>
                            <td className="px-4 py-2 text-slate-300">{room.type}</td>
                            <td className="px-4 py-2"><span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${cfg.color.split(" ").slice(0,2).join(" ")}`}>{cfg.label}</span></td>
                            <td className="px-4 py-2 text-slate-300">{room.guest || "—"}</td>
                            <td className="px-4 py-2 text-slate-400">{room.departure || "—"}</td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>
                </div>
              )}
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
              <div className="flex justify-between"><span className="text-slate-400">Status</span><span className={(statusConfig[selectedRoom.status]?.color.split(" ")[2]) || "text-slate-400"}>{statusConfig[selectedRoom.status]?.label || selectedRoom.status}</span></div>
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
