"use client";

import { useState } from "react";
import {
  BedDouble,
  CheckCircle2,
  Clock,
  AlertTriangle,
  Star,
  ChevronDown,
  ChevronUp,
  Filter,
  Search,
  User,
  Plus,
  Save,
  RotateCcw,
  ArrowRightLeft,
  Sparkles,
  Wrench,
  Ban,
  Eye,
  X,
  Camera,
  MapPin,
} from "lucide-react";

type RoomStatus = "dirty" | "in_progress" | "inspected" | "ready" | "blocked" | "maintenance";

interface Room {
  id: string;
  number: string;
  floor: string;
  type: string;
  status: RoomStatus;
  assignedTo?: string;
  priority: boolean;
  vip: boolean;
  lateCheckout: boolean;
  estimatedTime?: number;
  lastCleaned?: string;
  notes?: string;
  damagePhotos?: number;
}

const mockRooms: Room[] = [
  { id: "101", number: "101", floor: "1", type: "Standard", status: "ready", assignedTo: "Maria K.", priority: false, vip: false, lateCheckout: false, lastCleaned: "2026-05-10 08:30" },
  { id: "102", number: "102", floor: "1", type: "Standard", status: "dirty", assignedTo: "Maria K.", priority: false, vip: false, lateCheckout: false, estimatedTime: 25 },
  { id: "103", number: "103", floor: "1", type: "Deluxe", status: "in_progress", assignedTo: "John D.", priority: true, vip: true, lateCheckout: false, estimatedTime: 30 },
  { id: "104", number: "104", floor: "1", type: "Standard", status: "blocked", assignedTo: undefined, priority: false, vip: false, lateCheckout: false },
  { id: "105", number: "105", floor: "1", type: "Standard", status: "maintenance", assignedTo: undefined, priority: false, vip: false, lateCheckout: false, notes: "Leaky faucet reported" },
  { id: "201", number: "201", floor: "2", type: "Deluxe King", status: "in_progress", assignedTo: "Sofia L.", priority: true, vip: true, lateCheckout: true, estimatedTime: 35, notes: "VIP arrival at 14:00" },
  { id: "202", number: "202", floor: "2", type: "Deluxe King", status: "dirty", assignedTo: undefined, priority: false, vip: false, lateCheckout: false, estimatedTime: 25 },
  { id: "203", number: "203", floor: "2", type: "Suite", status: "inspected", assignedTo: "Sofia L.", priority: false, vip: false, lateCheckout: false, lastCleaned: "2026-05-10 09:00" },
  { id: "204", number: "204", floor: "2", type: "Deluxe King", status: "ready", assignedTo: "John D.", priority: false, vip: false, lateCheckout: false, lastCleaned: "2026-05-10 07:45" },
  { id: "301", number: "301", floor: "3", type: "Suite", status: "dirty", assignedTo: "Maria K.", priority: false, vip: false, lateCheckout: false, estimatedTime: 40 },
  { id: "302", number: "302", floor: "3", type: "Suite", status: "ready", assignedTo: "John D.", priority: false, vip: false, lateCheckout: false, lastCleaned: "2026-05-10 08:00" },
  { id: "303", number: "303", floor: "3", type: "Deluxe", status: "in_progress", assignedTo: "Sofia L.", priority: false, vip: false, lateCheckout: false, estimatedTime: 30 },
  { id: "304", number: "304", floor: "3", type: "Standard", status: "dirty", assignedTo: undefined, priority: false, vip: false, lateCheckout: false, estimatedTime: 20 },
  { id: "401", number: "401", floor: "4", type: "Suite", status: "inspected", assignedTo: "Maria K.", priority: false, vip: false, lateCheckout: false, lastCleaned: "2026-05-10 09:15" },
  { id: "402", number: "402", floor: "4", type: "Deluxe", status: "dirty", assignedTo: "John D.", priority: true, vip: false, lateCheckout: true, estimatedTime: 30, notes: "Late checkout until 13:00" },
];

const staffMembers = ["Maria K.", "John D.", "Sofia L.", "Unassigned"];

const statusConfig: Record<RoomStatus, { label: string; color: string; bg: string; icon: React.ElementType }> = {
  dirty: { label: "Dirty", color: "text-rose-400", bg: "bg-rose-500/10", icon: BedDouble },
  in_progress: { label: "In Progress", color: "text-sky-400", bg: "bg-sky-500/10", icon: Clock },
  inspected: { label: "Inspected", color: "text-violet-400", bg: "bg-violet-500/10", icon: Eye },
  ready: { label: "Ready", color: "text-emerald-400", bg: "bg-emerald-500/10", icon: CheckCircle2 },
  blocked: { label: "Blocked", color: "text-slate-400", bg: "bg-slate-500/10", icon: Ban },
  maintenance: { label: "Maintenance", color: "text-amber-400", bg: "bg-amber-500/10", icon: Wrench },
};

const statusFlow: RoomStatus[] = ["dirty", "in_progress", "inspected", "ready"];

export default function HousekeepingPage() {
  const [rooms, setRooms] = useState(mockRooms);
  const [search, setSearch] = useState("");
  const [floorFilter, setFloorFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState<RoomStatus | "all">("all");
  const [staffFilter, setStaffFilter] = useState("all");
  const [viewMode, setViewMode] = useState<"grid" | "list" | "staff">("grid");
  const [expandedRoom, setExpandedRoom] = useState<string | null>(null);

  const filtered = rooms.filter((r) => {
    const matchesSearch = r.number.includes(search) || r.type.toLowerCase().includes(search.toLowerCase());
    const matchesFloor = floorFilter === "all" || r.floor === floorFilter;
    const matchesStatus = statusFilter === "all" || r.status === statusFilter;
    const matchesStaff = staffFilter === "all" || (r.assignedTo || "Unassigned") === staffFilter;
    return matchesSearch && matchesFloor && matchesStatus && matchesStaff;
  });

  const floors = Array.from(new Set(rooms.map((r) => r.floor))).sort();

  const updateStatus = (roomId: string, newStatus: RoomStatus) => {
    setRooms(rooms.map((r) => (r.id === roomId ? { ...r, status: newStatus, lastCleaned: newStatus === "ready" ? new Date().toISOString() : r.lastCleaned } : r)));
  };

  const assignStaff = (roomId: string, staff: string) => {
    setRooms(rooms.map((r) => (r.id === roomId ? { ...r, assignedTo: staff === "Unassigned" ? undefined : staff } : r)));
  };

  const nextStatus = (current: RoomStatus): RoomStatus | null => {
    const idx = statusFlow.indexOf(current);
    return idx >= 0 && idx < statusFlow.length - 1 ? statusFlow[idx + 1] : null;
  };

  const stats = {
    dirty: rooms.filter((r) => r.status === "dirty").length,
    inProgress: rooms.filter((r) => r.status === "in_progress").length,
    inspected: rooms.filter((r) => r.status === "inspected").length,
    ready: rooms.filter((r) => r.status === "ready").length,
    blocked: rooms.filter((r) => r.status === "blocked").length,
    maintenance: rooms.filter((r) => r.status === "maintenance").length,
    priority: rooms.filter((r) => r.priority).length,
    unassigned: rooms.filter((r) => !r.assignedTo && r.status === "dirty").length,
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Housekeeping</h1>
          <p className="text-sm text-slate-400">Room status · Staff assignments · Cleaning board</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => setViewMode("grid")} className={`rounded-lg px-3 py-2 text-xs ${viewMode === "grid" ? "bg-slate-800 text-white" : "text-slate-400"}`}>Grid</button>
          <button onClick={() => setViewMode("list")} className={`rounded-lg px-3 py-2 text-xs ${viewMode === "list" ? "bg-slate-800 text-white" : "text-slate-400"}`}>List</button>
          <button onClick={() => setViewMode("staff")} className={`rounded-lg px-3 py-2 text-xs ${viewMode === "staff" ? "bg-slate-800 text-white" : "text-slate-400"}`}>By Staff</button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-7 gap-2">
        {([
          { key: "dirty", label: "Dirty", value: stats.dirty, color: "rose" },
          { key: "inProgress", label: "In Progress", value: stats.inProgress, color: "sky" },
          { key: "inspected", label: "Inspected", value: stats.inspected, color: "violet" },
          { key: "ready", label: "Ready", value: stats.ready, color: "emerald" },
          { key: "blocked", label: "Blocked", value: stats.blocked, color: "slate" },
          { key: "maintenance", label: "Maintenance", value: stats.maintenance, color: "amber" },
          { key: "priority", label: "Priority", value: stats.priority, color: "nexus" },
        ] as const).map((s) => (
          <button
            key={s.key}
            onClick={() => setStatusFilter(s.key === "priority" ? "all" : (s.key === "inProgress" ? "in_progress" : s.key as RoomStatus))}
            className={`rounded-lg border p-2 text-center transition-colors ${
              statusFilter === (s.key === "inProgress" ? "in_progress" : s.key)
                ? `border-${s.color}-500 bg-${s.color}-500/10`
                : "border-slate-700 bg-slate-800/50 hover:border-slate-600"
            }`}
          >
            <p className="text-xs text-slate-400">{s.label}</p>
            <p className={`text-xl font-bold ${
              s.color === "rose" ? "text-rose-400" :
              s.color === "sky" ? "text-sky-400" :
              s.color === "violet" ? "text-violet-400" :
              s.color === "emerald" ? "text-emerald-400" :
              s.color === "slate" ? "text-slate-400" :
              s.color === "amber" ? "text-amber-400" :
              "text-nexus-400"
            }`}>{s.value}</p>
          </button>
        ))}
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
          <input className="input w-full pl-9" placeholder="Search room or type..." value={search} onChange={(e) => setSearch(e.target.value)} />
        </div>
        <select value={floorFilter} onChange={(e) => setFloorFilter(e.target.value)} className="input text-xs">
          <option value="all">All Floors</option>
          {floors.map((f) => (
            <option key={f} value={f}>Floor {f}</option>
          ))}
        </select>
        <select value={staffFilter} onChange={(e) => setStaffFilter(e.target.value)} className="input text-xs">
          <option value="all">All Staff</option>
          {staffMembers.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
        <button onClick={() => { setSearch(""); setFloorFilter("all"); setStatusFilter("all"); setStaffFilter("all"); }} className="btn-secondary text-xs gap-1">
          <RotateCcw className="h-3 w-3" /> Reset
        </button>
      </div>

      {/* Grid View */}
      {viewMode === "grid" && (
        <div className="grid grid-cols-5 gap-3">
          {filtered.map((room) => {
            const config = statusConfig[room.status];
            const Icon = config.icon;
            const next = nextStatus(room.status);
            return (
              <div
                key={room.id}
                className={`relative rounded-lg border p-4 transition-all hover:scale-[1.02] ${
                  room.priority ? "border-amber-500/30" : "border-slate-700"
                } bg-slate-800/50`}
              >
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <span className="text-lg font-bold text-white">{room.number}</span>
                    {room.vip && <Star className="h-4 w-4 text-amber-400" />}
                    {room.lateCheckout && <AlertTriangle className="h-4 w-4 text-amber-400" />}
                  </div>
                  <span className={`rounded px-2 py-0.5 text-[10px] font-medium ${config.bg} ${config.color}`}>
                    {config.label}
                  </span>
                </div>
                <p className="mt-1 text-xs text-slate-500">{room.type} · Floor {room.floor}</p>
                {room.assignedTo && (
                  <div className="mt-2 flex items-center gap-1 text-xs text-slate-400">
                    <User className="h-3 w-3" /> {room.assignedTo}
                  </div>
                )}
                {room.estimatedTime && room.status !== "ready" && (
                  <div className="mt-1 flex items-center gap-1 text-xs text-slate-500">
                    <Clock className="h-3 w-3" /> ~{room.estimatedTime} min
                  </div>
                )}
                {room.lastCleaned && (
                  <div className="mt-1 text-[10px] text-slate-500">Cleaned: {room.lastCleaned}</div>
                )}
                {room.notes && (
                  <div className="mt-2 rounded bg-amber-500/5 p-1.5 text-[10px] text-amber-400">{room.notes}</div>
                )}
                <div className="mt-3 flex gap-1">
                  {next && (
                    <button
                      onClick={() => updateStatus(room.id, next)}
                      className="flex-1 rounded bg-nexus-500/10 px-2 py-1 text-xs text-nexus-400 hover:bg-nexus-500/20 transition-colors"
                    >
                      Mark {statusConfig[next].label}
                    </button>
                  )}
                  {room.status === "ready" && (
                    <button
                      onClick={() => updateStatus(room.id, "dirty")}
                      className="flex-1 rounded bg-rose-500/10 px-2 py-1 text-xs text-rose-400 hover:bg-rose-500/20 transition-colors"
                    >
                      Reset to Dirty
                    </button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* List View */}
      {viewMode === "list" && (
        <div className="card overflow-hidden p-0">
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-800 text-xs uppercase text-slate-400">
              <tr>
                <th className="px-4 py-3">Room</th>
                <th className="px-4 py-3">Type</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Assigned To</th>
                <th className="px-4 py-3">Est. Time</th>
                <th className="px-4 py-3">Priority</th>
                <th className="px-4 py-3">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {filtered.map((room) => {
                const config = statusConfig[room.status];
                const Icon = config.icon;
                const next = nextStatus(room.status);
                return (
                  <tr key={room.id} className="hover:bg-slate-800/50">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <span className="font-medium text-white">{room.number}</span>
                        {room.vip && <Star className="h-3 w-3 text-amber-400" />}
                      </div>
                    </td>
                    <td className="px-4 py-3 text-slate-400">{room.type}</td>
                    <td className="px-4 py-3">
                      <span className={`inline-flex items-center gap-1 rounded px-2 py-0.5 text-xs ${config.bg} ${config.color}`}>
                        <Icon className="h-3 w-3" /> {config.label}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <select
                        value={room.assignedTo || "Unassigned"}
                        onChange={(e) => assignStaff(room.id, e.target.value)}
                        className="input text-xs py-1"
                      >
                        {staffMembers.map((s) => (
                          <option key={s} value={s}>{s}</option>
                        ))}
                      </select>
                    </td>
                    <td className="px-4 py-3 text-slate-400">{room.estimatedTime ? `${room.estimatedTime}m` : "—"}</td>
                    <td className="px-4 py-3">
                      {room.priority ? (
                        <span className="rounded bg-amber-500/10 px-2 py-0.5 text-xs text-amber-400">Priority</span>
                      ) : (
                        <span className="text-slate-500">—</span>
                      )}
                    </td>
                    <td className="px-4 py-3">
                      {next && (
                        <button onClick={() => updateStatus(room.id, next)} className="text-xs text-nexus-400 hover:text-nexus-300">
                          {statusConfig[next].label} →
                        </button>
                      )}
                    </td>
                  </tr>
                );
              })}
            </tbody>
          </table>
        </div>
      )}

      {/* Staff View */}
      {viewMode === "staff" && (
        <div className="space-y-4">
          {staffMembers.filter((s) => s !== "Unassigned").map((staff) => {
            const staffRooms = filtered.filter((r) => r.assignedTo === staff);
            return (
              <div key={staff} className="card space-y-3">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                    <User className="h-5 w-5 text-nexus-400" /> {staff}
                  </h3>
                  <span className="text-xs text-slate-400">{staffRooms.length} rooms</span>
                </div>
                <div className="grid grid-cols-5 gap-2">
                  {staffRooms.map((room) => {
                    const config = statusConfig[room.status];
                    return (
                      <div key={room.id} className="rounded border border-slate-700 bg-slate-800/50 p-3">
                        <div className="flex items-center justify-between">
                          <span className="font-bold text-white">{room.number}</span>
                          <span className={`rounded px-1.5 py-0.5 text-[10px] ${config.bg} ${config.color}`}>{config.label}</span>
                        </div>
                        <p className="text-xs text-slate-500">{room.type}</p>
                        {room.estimatedTime && (
                          <p className="text-xs text-slate-500">~{room.estimatedTime} min</p>
                        )}
                      </div>
                    );
                  })}
                </div>
                {staffRooms.length === 0 && <p className="text-sm text-slate-500">No rooms assigned</p>}
              </div>
            );
          })}
          {/* Unassigned */}
          <div className="card space-y-3">
            <h3 className="text-lg font-semibold text-white">Unassigned Rooms</h3>
            <div className="grid grid-cols-5 gap-2">
              {filtered.filter((r) => !r.assignedTo).map((room) => {
                const config = statusConfig[room.status];
                return (
                  <div key={room.id} className="rounded border border-slate-700 bg-slate-800/50 p-3">
                    <span className="font-bold text-white">{room.number}</span>
                    <span className={`ml-2 rounded px-1.5 py-0.5 text-[10px] ${config.bg} ${config.color}`}>{config.label}</span>
                    <div className="mt-2 flex gap-1">
                      {staffMembers.filter((s) => s !== "Unassigned").map((s) => (
                        <button key={s} onClick={() => assignStaff(room.id, s)} className="rounded bg-slate-700 px-2 py-0.5 text-[10px] text-white hover:bg-nexus-500">
                          {s.split(" ")[0]}
                        </button>
                      ))}
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
