// @ts-nocheck
"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import {
  BedDouble, CheckCircle2, Clock, AlertTriangle, Star, Filter, Search, User, Plus, Save,
  RotateCcw, ArrowRightLeft, Sparkles, Wrench, Ban, Eye, X, Camera, MapPin, ClipboardList,
  Droplets, Package, Activity, Users, ShieldCheck, ChevronRight, Settings, Play, Square,
  RefreshCw,
} from "lucide-react";
import {
  Room as ApiRoom,
  HousekeepingTask, HousekeepingStaff,
  getRooms, updateRoomStatus,
  getHousekeepingTasks, createHousekeepingTask, updateHousekeepingTask, getHousekeepingStaff,
} from "@/lib/api";

type RoomStatus = "dirty" | "in_progress" | "inspected" | "ready" | "blocked" | "maintenance";

interface Room {
  id: string; number: string; floor: string; type: string; status: RoomStatus;
  assignedTo?: string; priority: boolean; vip: boolean; lateCheckout: boolean;
  estimatedTime?: number; lastCleaned?: string; notes?: string; damagePhotos?: number;
  damageReport?: string; inspectionNotes?: string; suppliesNeeded?: string[];
}

function mapApiRoom(r: ApiRoom): Room {
  const statusMap: Record<string, RoomStatus> = {
    vacant_dirty: "dirty",
    occupied: "in_progress",
    vacant_clean: "ready",
    blocked: "blocked",
    maintenance: "maintenance",
  };
  return {
    id: r.id,
    number: r.number,
    floor: r.floor,
    type: r.type,
    status: statusMap[r.status] || "dirty",
    priority: false,
    vip: false,
    lateCheckout: false,
  };
}

function mapRoomStatusToApi(status: RoomStatus): string {
  const map: Record<string, string> = {
    dirty: "vacant_dirty",
    in_progress: "occupied",
    inspected: "vacant_clean",
    ready: "vacant_clean",
    blocked: "blocked",
    maintenance: "maintenance",
  };
  return map[status] || "vacant_dirty";
}

const statusConfig: Record<RoomStatus, { label: string; color: string; bg: string; icon: React.ElementType }> = {
  dirty: { label: "Dirty", color: "text-rose-400", bg: "bg-rose-500/10", icon: BedDouble },
  in_progress: { label: "In Progress", color: "text-sky-400", bg: "bg-sky-500/10", icon: Clock },
  inspected: { label: "Inspected", color: "text-violet-400", bg: "bg-violet-500/10", icon: Eye },
  ready: { label: "Ready", color: "text-emerald-400", bg: "bg-emerald-500/10", icon: CheckCircle2 },
  blocked: { label: "Blocked", color: "text-slate-400", bg: "bg-slate-500/10", icon: Ban },
  maintenance: { label: "Maintenance", color: "text-amber-400", bg: "bg-amber-500/10", icon: Wrench },
};

const statusFlow: RoomStatus[] = ["dirty", "in_progress", "inspected", "ready"];

const supplyItems = ["Shampoo", "Conditioner", "Body lotion", "Soap", "Dental kit", "Slippers", "Coffee", "Tea", "Sugar", "Minibar: Coke", "Minibar: Water", "Minibar: Nuts", "Towels", "Toilet paper"];

const inspectionChecklist = [
  "Beds made & linens fresh", "Bathroom sanitized", "Towels folded & stocked", "Minibar restocked",
  "Coffee/tea station filled", "Dust surfaces", "Vacuum floor", "Windows clean",
  "TV & remote working", "AC functioning", "Lights all working", "Trash emptied",
];

export default function HousekeepingPage() {
  const [rooms, setRooms] = useState<Room[]>([]);
  const [tasks, setTasks] = useState<HousekeepingTask[]>([]);
  const [staff, setStaff] = useState<HousekeepingStaff[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [floorFilter, setFloorFilter] = useState("all");
  const [statusFilter, setStatusFilter] = useState<RoomStatus | "all">("all");
  const [staffFilter, setStaffFilter] = useState("all");
  const [viewMode, setViewMode] = useState<"grid" | "list" | "staff">("grid");
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const [inspectModal, setInspectModal] = useState<string | null>(null);
  const [inspectChecks, setInspectChecks] = useState<boolean[]>(new Array(inspectionChecklist.length).fill(false));
  const [inspectNotes, setInspectNotes] = useState("");

  const [damageModal, setDamageModal] = useState<string | null>(null);
  const [damageText, setDamageText] = useState("");
  const [damageUrgency, setDamageUrgency] = useState("low");

  const [suppliesModal, setSuppliesModal] = useState<string | null>(null);
  const [selectedSupplies, setSelectedSupplies] = useState<string[]>([]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [apiRooms, apiTasks, apiStaff] = await Promise.all([getRooms(), getHousekeepingTasks(), getHousekeepingStaff()]);
      const mapped = apiRooms.map(mapApiRoom);
      // Merge task data into rooms for assignment info
      apiTasks.forEach((task) => {
        const room = mapped.find((r) => r.number === task.room_number);
        if (room && task.assigned_to) {
          const s = apiStaff.find((st) => st.id === task.assigned_to);
          if (s) room.assignedTo = s.name;
          if (task.status === "in_progress") room.status = "in_progress";
          if (task.priority === "high" || task.priority === "urgent") room.priority = true;
        }
      });
      setRooms(mapped);
      setTasks(apiTasks);
      setStaff(apiStaff);
    } catch (e: any) {
      setError(e.message || "Failed to load housekeeping data");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const filtered = rooms.filter((r) => {
    const matchesSearch = r.number.includes(search) || r.type.toLowerCase().includes(search.toLowerCase());
    const matchesFloor = floorFilter === "all" || r.floor === floorFilter;
    const matchesStatus = statusFilter === "all" || r.status === statusFilter;
    const matchesStaff = staffFilter === "all" || (r.assignedTo || "Unassigned") === staffFilter;
    return matchesSearch && matchesFloor && matchesStatus && matchesStaff;
  });

  const floors = Array.from(new Set(rooms.map((r) => r.floor))).sort();
  const staffNames = ["Unassigned", ...staff.map((s) => s.name)];

  async function updateStatus(roomId: string, newStatus: RoomStatus) {
    const room = rooms.find((r) => r.id === roomId);
    if (!room) return;
    setActionLoading(roomId);
    try {
      await updateRoomStatus(room.number, mapRoomStatusToApi(newStatus));
      // Create/update housekeeping task
      const existingTask = tasks.find((t) => t.room_number === room.number && t.status !== "completed");
      if (newStatus === "in_progress" && !existingTask) {
        await createHousekeepingTask({ room_number: room.number, task_type: "full_clean", priority: room.priority ? "high" : "normal" });
      } else if ((newStatus === "inspected" || newStatus === "ready") && existingTask) {
        await updateHousekeepingTask(existingTask.id, { status: "completed" });
      }
      setRooms(rooms.map((r) => (r.id === roomId ? { ...r, status: newStatus, lastCleaned: newStatus === "ready" ? new Date().toISOString() : r.lastCleaned } : r)));
      await fetchData();
    } catch (e: any) {
      alert(e.message);
    } finally {
      setActionLoading(null);
    }
  }

  function assignStaff(roomId: string, staffName: string) {
    const s = staff.find((st) => st.name === staffName);
    setRooms(rooms.map((r) => (r.id === roomId ? { ...r, assignedTo: staffName === "Unassigned" ? undefined : staffName } : r)));
    if (s) {
      const room = rooms.find((r) => r.id === roomId);
      if (room) {
        createHousekeepingTask({ room_number: room.number, task_type: "full_clean", assigned_to: s.id }).then(fetchData).catch((e) => alert(e.message));
      }
    }
  }

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

  const openInspect = (roomId: string) => {
    setInspectModal(roomId);
    setInspectChecks(new Array(inspectionChecklist.length).fill(false));
    setInspectNotes("");
  };

  const submitInspection = async () => {
    if (!inspectModal) return;
    const passed = inspectChecks.filter(Boolean).length;
    const total = inspectionChecklist.length;
    if (passed >= total * 0.8) {
      await updateStatus(inspectModal, "inspected");
    }
    setRooms(rooms.map((r) => r.id === inspectModal ? { ...r, inspectionNotes: inspectNotes } : r));
    setInspectModal(null);
  };

  const submitDamage = () => {
    if (!damageModal) return;
    setRooms(rooms.map((r) => r.id === damageModal ? { ...r, damageReport: damageText, status: damageUrgency === "high" ? "maintenance" : r.status } : r));
    const room = rooms.find((r) => r.id === damageModal);
    if (room && damageUrgency === "high") {
      updateRoomStatus(room.number, "maintenance").catch((e) => alert(e.message));
    }
    setDamageModal(null);
    setDamageText("");
  };

  const submitSupplies = () => {
    if (!suppliesModal) return;
    setRooms(rooms.map((r) => r.id === suppliesModal ? { ...r, suppliesNeeded: selectedSupplies } : r));
    setSuppliesModal(null);
    setSelectedSupplies([]);
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20 text-slate-500">
        <RefreshCw className="mr-2 h-5 w-5 animate-spin" /> Loading housekeeping data...
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Housekeeping</h1>
          <p className="text-sm text-slate-400">Room status · Staff assignments · Cleaning board</p>
        </div>
        <div className="flex items-center gap-3">
          <button onClick={fetchData} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white" title="Refresh">
            <RefreshCw className="h-4 w-4" />
          </button>
          <Link href="/housekeeping/manager" className="btn-secondary text-xs gap-1.5">
            <Users className="h-3.5 w-3.5" /> Manager View
          </Link>
          <Link href="/housekeeping/staff" className="btn-secondary text-xs gap-1.5">
            <Play className="h-3.5 w-3.5" /> Cleaner Tablet
          </Link>
          <div className="flex items-center gap-1 rounded-lg bg-slate-800 p-1">
            <button onClick={() => setViewMode("grid")} className={`rounded-md px-3 py-1.5 text-xs ${viewMode === "grid" ? "bg-slate-700 text-white" : "text-slate-400"}`}>Grid</button>
            <button onClick={() => setViewMode("list")} className={`rounded-md px-3 py-1.5 text-xs ${viewMode === "list" ? "bg-slate-700 text-white" : "text-slate-400"}`}>List</button>
            <button onClick={() => setViewMode("staff")} className={`rounded-md px-3 py-1.5 text-xs ${viewMode === "staff" ? "bg-slate-700 text-white" : "text-slate-400"}`}>By Staff</button>
          </div>
        </div>
      </div>

      {error && (
        <div className="rounded-lg border border-rose-500/20 bg-rose-500/5 px-4 py-3 text-sm text-rose-400">
          {error}
        </div>
      )}

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
          {staffNames.map((s) => (
            <option key={s} value={s}>{s}</option>
          ))}
        </select>
        <button onClick={() => { setSearch(""); setFloorFilter("all"); setStatusFilter("all"); setStaffFilter("all"); }} className="btn-secondary text-xs gap-1">
          <RotateCcw className="h-3 w-3" /> Reset
        </button>
      </div>

      <div className="grid grid-cols-3 gap-6">
        <div className="col-span-2 space-y-6">
          {viewMode === "grid" && (
            <div className="grid grid-cols-4 gap-3">
              {filtered.map((room) => {
                const config = statusConfig[room.status];
                const Icon = config.icon;
                const next = nextStatus(room.status);
                return (
                  <div key={room.id} className={`relative rounded-lg border p-4 transition-all hover:scale-[1.02] ${room.priority ? "border-amber-500/30" : "border-slate-700"} bg-slate-800/50`}>
                    <div className="flex items-center justify-between">
                      <div className="flex items-center gap-2">
                        <span className="text-lg font-bold text-white">{room.number}</span>
                        {room.vip && <Star className="h-4 w-4 text-amber-400" />}
                        {room.lateCheckout && <AlertTriangle className="h-4 w-4 text-amber-400" />}
                      </div>
                      <span className={`rounded px-2 py-0.5 text-[10px] font-medium ${config.bg} ${config.color}`}>{config.label}</span>
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
                    {room.damageReport && (
                      <div className="mt-1 rounded bg-rose-500/5 p-1.5 text-[10px] text-rose-400 flex items-center gap-1">
                        <Wrench className="h-3 w-3" /> {room.damageReport}
                      </div>
                    )}
                    {room.suppliesNeeded && room.suppliesNeeded.length > 0 && (
                      <div className="mt-1 rounded bg-sky-500/5 p-1.5 text-[10px] text-sky-400 flex items-center gap-1">
                        <Package className="h-3 w-3" /> {room.suppliesNeeded.join(", ")}
                      </div>
                    )}
                    <div className="mt-3 flex flex-wrap gap-1">
                      {next && (
                        <button
                          onClick={() => updateStatus(room.id, next)}
                          disabled={actionLoading === room.id}
                          className="flex-1 rounded bg-nexus-500/10 px-2 py-1 text-xs text-nexus-400 hover:bg-nexus-500/20 transition-colors disabled:opacity-50"
                        >
                          {actionLoading === room.id ? "..." : `Mark ${statusConfig[next].label}`}
                        </button>
                      )}
                      <button onClick={() => openInspect(room.id)} className="rounded bg-violet-500/10 px-2 py-1 text-xs text-violet-400 hover:bg-violet-500/20">
                        <ClipboardList className="h-3 w-3 inline" /> Inspect
                      </button>
                      <button onClick={() => { setDamageModal(room.id); setDamageText(room.damageReport || ""); }} className="rounded bg-rose-500/10 px-2 py-1 text-xs text-rose-400 hover:bg-rose-500/20">
                        <Wrench className="h-3 w-3 inline" />
                      </button>
                      <button onClick={() => { setSuppliesModal(room.id); setSelectedSupplies(room.suppliesNeeded || []); }} className="rounded bg-sky-500/10 px-2 py-1 text-xs text-sky-400 hover:bg-sky-500/20">
                        <Package className="h-3 w-3 inline" />
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}

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
                    <th className="px-4 py-3">Notes</th>
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
                          <select value={room.assignedTo || "Unassigned"} onChange={(e) => assignStaff(room.id, e.target.value)} className="input text-xs py-1">
                            {staffNames.map((s) => (
                              <option key={s} value={s}>{s}</option>
                            ))}
                          </select>
                        </td>
                        <td className="px-4 py-3 text-slate-400">{room.estimatedTime ? `${room.estimatedTime}m` : "—"}</td>
                        <td className="px-4 py-3">
                          {room.damageReport && <span className="text-rose-400 text-xs">{room.damageReport}</span>}
                          {room.suppliesNeeded && <span className="text-sky-400 text-xs">{room.suppliesNeeded.join(", ")}</span>}
                        </td>
                        <td className="px-4 py-3 flex gap-1">
                          {next && <button onClick={() => updateStatus(room.id, next)} disabled={actionLoading === room.id} className="text-xs text-nexus-400 disabled:opacity-50">{actionLoading === room.id ? "..." : `${statusConfig[next].label} →`}</button>}
                          <button onClick={() => openInspect(room.id)} className="text-xs text-violet-400">Inspect</button>
                          <button onClick={() => { setDamageModal(room.id); setDamageText(room.damageReport || ""); }} className="text-xs text-rose-400">Damage</button>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}

          {viewMode === "staff" && (
            <div className="space-y-4">
              {staffNames.filter((s) => s !== "Unassigned").map((staffName) => {
                const staffRooms = filtered.filter((r) => r.assignedTo === staffName);
                return (
                  <div key={staffName} className="card space-y-3">
                    <div className="flex items-center justify-between">
                      <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                        <User className="h-5 w-5 text-nexus-400" /> {staffName}
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
                            {room.estimatedTime && <p className="text-xs text-slate-500">~{room.estimatedTime} min</p>}
                          </div>
                        );
                      })}
                    </div>
                    {staffRooms.length === 0 && <p className="text-sm text-slate-500">No rooms assigned</p>}
                  </div>
                );
              })}
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
                          {staffNames.filter((s) => s !== "Unassigned").map((s) => (
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

        <div className="space-y-4">
          <div className="card">
            <div className="flex items-center gap-2 mb-3">
              <Activity className="h-4 w-4 text-nexus-400" />
              <h3 className="font-semibold text-white text-sm">Tasks</h3>
            </div>
            <div className="space-y-3 max-h-[320px] overflow-y-auto">
              {tasks.length === 0 && <p className="text-xs text-slate-500">No active tasks</p>}
              {tasks.map((task) => (
                <div key={task.id} className="flex items-start gap-2 text-xs">
                  <div className={`mt-0.5 h-2 w-2 rounded-full ${
                    task.status === "completed" ? "bg-emerald-400" :
                    task.status === "in_progress" ? "bg-sky-400" :
                    task.priority === "urgent" || task.priority === "high" ? "bg-rose-400" :
                    "bg-amber-400"
                  }`} />
                  <div>
                    <p className="text-slate-300"><span className="font-medium text-white">{task.room_number}</span> · {task.task_type.replace("_", " ")}</p>
                    <p className="text-slate-500">{task.status} · {task.priority}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="card space-y-3">
            <h3 className="font-semibold text-white text-sm">Today's Summary</h3>
            <div className="space-y-2 text-xs">
              <div className="flex justify-between"><span className="text-slate-400">Rooms ready</span><span className="text-white font-medium">{rooms.filter(r => r.status === "ready").length}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">In progress</span><span className="text-white font-medium">{rooms.filter(r => r.status === "in_progress").length}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Staff on duty</span><span className="text-white font-medium">{staff.length}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Damage reports</span><span className="text-rose-400 font-medium">{rooms.filter(r => r.damageReport).length}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Supply requests</span><span className="text-sky-400 font-medium">{rooms.filter(r => r.suppliesNeeded && r.suppliesNeeded.length > 0).length}</span></div>
            </div>
          </div>

          <div className="card space-y-2">
            <h3 className="font-semibold text-white text-sm flex items-center gap-2">
              <AlertTriangle className="h-4 w-4 text-amber-400" /> Priority
            </h3>
            {rooms.filter(r => r.priority).map(r => (
              <div key={r.id} className="flex items-center justify-between text-xs">
                <span className="text-white font-medium">{r.number}</span>
                <span className="text-slate-400">{r.status.replace("_", " ")}</span>
              </div>
            ))}
            {rooms.filter(r => r.priority).length === 0 && <p className="text-xs text-slate-500">No priority rooms</p>}
          </div>
        </div>
      </div>

      {/* Inspection Modal */}
      {inspectModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold text-white">Inspect Room {rooms.find(r => r.id === inspectModal)?.number}</h3>
              <button onClick={() => setInspectModal(null)} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
            </div>
            <div className="space-y-2 max-h-[300px] overflow-y-auto mb-4">
              {inspectionChecklist.map((item, i) => (
                <label key={i} className="flex items-center gap-2 text-sm text-slate-300 cursor-pointer">
                  <input type="checkbox" checked={inspectChecks[i]} onChange={(e) => {
                    const next = [...inspectChecks]; next[i] = e.target.checked; setInspectChecks(next);
                  }} className="rounded border-slate-600 bg-slate-800 text-nexus-500" />
                  {item}
                </label>
              ))}
            </div>
            <textarea className="input w-full text-xs mb-4" rows={2} placeholder="Inspection notes..." value={inspectNotes} onChange={(e) => setInspectNotes(e.target.value)} />
            <div className="flex gap-2">
              <button onClick={() => setInspectModal(null)} className="btn-secondary flex-1 text-xs">Cancel</button>
              <button onClick={submitInspection} className="btn-primary flex-1 text-xs gap-1">
                <ShieldCheck className="h-3.5 w-3.5" /> Submit Inspection
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Damage Report Modal */}
      {damageModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold text-white">Damage Report — Room {rooms.find(r => r.id === damageModal)?.number}</h3>
              <button onClick={() => setDamageModal(null)} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
            </div>
            <textarea className="input w-full text-sm mb-3" rows={3} placeholder="Describe the damage or issue..." value={damageText} onChange={(e) => setDamageText(e.target.value)} />
            <div className="mb-4">
              <label className="text-xs text-slate-400 mb-1 block">Urgency</label>
              <div className="flex gap-2">
                {["low", "medium", "high"].map((u) => (
                  <button key={u} onClick={() => setDamageUrgency(u)} className={`flex-1 rounded-lg py-1.5 text-xs capitalize ${
                    damageUrgency === u ? u === "high" ? "bg-rose-500/20 text-rose-400 border border-rose-500/30" : u === "medium" ? "bg-amber-500/20 text-amber-400 border border-amber-500/30" : "bg-emerald-500/20 text-emerald-400 border border-emerald-500/30"
                    : "bg-slate-800 text-slate-400 border border-slate-700"
                  }`}>
                    {u}
                  </button>
                ))}
              </div>
            </div>
            <div className="flex gap-2">
              <button onClick={() => setDamageModal(null)} className="btn-secondary flex-1 text-xs">Cancel</button>
              <button onClick={submitDamage} className="btn-primary flex-1 text-xs gap-1 bg-rose-500 hover:bg-rose-600">
                <Wrench className="h-3.5 w-3.5" /> Report Damage
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Supplies Request Modal */}
      {suppliesModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-bold text-white">Request Supplies — Room {rooms.find(r => r.id === suppliesModal)?.number}</h3>
              <button onClick={() => setSuppliesModal(null)} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
            </div>
            <div className="grid grid-cols-2 gap-2 max-h-[280px] overflow-y-auto mb-4">
              {supplyItems.map((item) => (
                <label key={item} className={`flex items-center gap-2 rounded-lg border p-2 text-xs cursor-pointer transition-colors ${
                  selectedSupplies.includes(item) ? "border-nexus-500 bg-nexus-500/10 text-white" : "border-slate-700 bg-slate-800 text-slate-400"
                }`}>
                  <input type="checkbox" className="hidden" checked={selectedSupplies.includes(item)} onChange={(e) => {
                    if (e.target.checked) setSelectedSupplies([...selectedSupplies, item]);
                    else setSelectedSupplies(selectedSupplies.filter(s => s !== item));
                  }} />
                  {item}
                </label>
              ))}
            </div>
            <div className="flex gap-2">
              <button onClick={() => setSuppliesModal(null)} className="btn-secondary flex-1 text-xs">Cancel</button>
              <button onClick={submitSupplies} className="btn-primary flex-1 text-xs gap-1">
                <Package className="h-3.5 w-3.5" /> Request
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
