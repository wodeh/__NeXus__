"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import {
  Users, ArrowLeft, ChevronRight, Clock, BedDouble, CheckCircle2, User, Plus, Trash2,
  RotateCcw, Sparkles, AlertTriangle, Star, Filter, Search, Grid3X3, List, Wand2,
  Calendar, ChevronDown, MapPin, RefreshCw,
} from "lucide-react";
import {
  Room as ApiRoom, HousekeepingTask, HousekeepingStaff,
  getRooms, getHousekeepingTasks, getHousekeepingStaff,
  createHousekeepingTask, updateHousekeepingTask, updateRoomStatus,
} from "@/lib/api";

type RoomStatus = "dirty" | "in_progress" | "inspected" | "ready" | "blocked" | "maintenance";
type Shift = "morning" | "afternoon" | "evening" | "night";

interface Staff {
  id: string; name: string; role: "cleaner" | "inspector" | "manager";
  active: boolean; shift: Shift; floors: string[]; maxRooms: number;
  currentLoad: number; rating: number; completedToday: number;
}

interface RoomTask {
  id: string; number: string; floor: string; type: string; status: RoomStatus;
  assignedTo?: string; priority: boolean; vip: boolean; estimatedTime: number;
  taskType: "clean" | "refill" | "maintenance" | "inspection";
}

const shifts: Shift[] = ["morning", "afternoon", "evening", "night"];

const statusConfig: Record<RoomStatus, { label: string; color: string; bg: string }> = {
  dirty: { label: "Dirty", color: "text-rose-400", bg: "bg-rose-500/10" },
  in_progress: { label: "In Progress", color: "text-sky-400", bg: "bg-sky-500/10" },
  inspected: { label: "Inspected", color: "text-violet-400", bg: "bg-violet-500/10" },
  ready: { label: "Ready", color: "text-emerald-400", bg: "bg-emerald-500/10" },
  blocked: { label: "Blocked", color: "text-slate-400", bg: "bg-slate-500/10" },
  maintenance: { label: "Maintenance", color: "text-amber-400", bg: "bg-amber-500/10" },
};

function mapApiRoomStatus(status: string): RoomStatus {
  const map: Record<string, RoomStatus> = {
    vacant_dirty: "dirty",
    occupied: "in_progress",
    vacant_clean: "ready",
    blocked: "blocked",
    maintenance: "maintenance",
  };
  return map[status] || "dirty";
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

export default function HousekeepingManagerPage() {
  const [tasks, setTasks] = useState<RoomTask[]>([]);
  const [staff, setStaff] = useState<Staff[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [shiftFilter, setShiftFilter] = useState<Shift | "all">("all");
  const [floorFilter, setFloorFilter] = useState("all");
  const [search, setSearch] = useState("");
  const [selectedTasks, setSelectedTasks] = useState<string[]>([]);
  const [bulkAssignStaff, setBulkAssignStaff] = useState("");
  const [viewMode, setViewMode] = useState<"floor" | "staff">("floor");
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [apiRooms, apiTasks, apiStaff] = await Promise.all([
        getRooms(), getHousekeepingTasks(), getHousekeepingStaff(),
      ]);

      // Map API tasks to RoomTask shape
      const mappedTasks: RoomTask[] = apiTasks.map((t) => {
        const room = apiRooms.find((r) => r.number === t.room_number);
        return {
          id: t.id,
          number: t.room_number,
          floor: room?.floor || "?",
          type: room?.type || "Standard",
          status: mapApiRoomStatus(room?.status || "vacant_dirty"),
          assignedTo: t.assigned_to || undefined,
          priority: t.priority === "urgent" || t.priority === "high",
          vip: false,
          estimatedTime: 25,
          taskType: (t.task_type.replace("_", "") as any) || "clean",
        };
      });

      // Build staff with computed load
      const mappedStaff: Staff[] = apiStaff.map((s) => {
        const assignedCount = apiTasks.filter(
          (t) => t.assigned_to === s.id && t.status !== "completed"
        ).length;
        const completedCount = apiTasks.filter(
          (t) => t.assigned_to === s.id && t.status === "completed"
        ).length;
        return {
          id: s.id,
          name: s.name,
          role: (s.role as any) || "cleaner",
          active: s.active,
          shift: (s.active_shift as Shift) || "morning",
          floors: ["1", "2", "3", "4"], // API doesn't have floors per staff yet
          maxRooms: s.max_rooms_per_day || 8,
          currentLoad: assignedCount,
          rating: 4.5,
          completedToday: completedCount,
        };
      });

      setTasks(mappedTasks);
      setStaff(mappedStaff);
    } catch (e: any) {
      setError(e.message || "Failed to load data");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const floors = Array.from(new Set(tasks.map((t) => t.floor))).sort();

  const filteredTasks = tasks.filter((t) => {
    const matchesFloor = floorFilter === "all" || t.floor === floorFilter;
    const matchesSearch = t.number.includes(search) || t.type.toLowerCase().includes(search.toLowerCase());
    return matchesFloor && matchesSearch;
  });

  const activeStaff = staff.filter((s) => s.active && (shiftFilter === "all" || s.shift === shiftFilter));

  async function assignTask(taskId: string, staffId: string) {
    const task = tasks.find((t) => t.id === taskId);
    if (!task) return;
    setActionLoading(taskId);
    try {
      await updateHousekeepingTask(taskId, { assigned_to: staffId || undefined });
      // Also update room status if assigning cleaner
      if (staffId) {
        const s = staff.find((st) => st.id === staffId);
        if (s && s.role === "cleaner" && task.status === "dirty") {
          await updateRoomStatus(task.number, "occupied");
        }
      }
      setTasks(tasks.map((t) =>
        t.id === taskId ? { ...t, assignedTo: staffId || undefined, status: staffId && task.status === "dirty" ? "in_progress" : t.status } : t
      ));
      await fetchData();
    } catch (e: any) {
      alert(e.message);
    } finally {
      setActionLoading(null);
    }
  }

  async function autoAssign() {
    const unassigned = tasks.filter((t) => !t.assignedTo && t.status === "dirty");
    const cleaners = staff.filter((s) => s.active && s.role === "cleaner");
    try {
      for (const task of unassigned) {
        const candidates = cleaners
          .filter((s) => s.currentLoad < s.maxRooms)
          .sort((a, b) => a.currentLoad - b.currentLoad);
        if (candidates.length > 0) {
          const chosen = candidates[0];
          await updateHousekeepingTask(task.id, { assigned_to: chosen.id });
          await updateRoomStatus(task.number, "occupied");
        }
      }
      await fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  }

  async function bulkAssign() {
    if (!bulkAssignStaff) return;
    try {
      for (const taskId of selectedTasks) {
        const task = tasks.find((t) => t.id === taskId);
        if (!task || task.assignedTo === bulkAssignStaff) continue;
        await updateHousekeepingTask(taskId, { assigned_to: bulkAssignStaff });
        if (task.status === "dirty") {
          await updateRoomStatus(task.number, "occupied");
        }
      }
      setSelectedTasks([]);
      setBulkAssignStaff("");
      await fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center py-20 text-slate-500">
        <RefreshCw className="mr-2 h-5 w-5 animate-spin" /> Loading manager board...
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Link href="/housekeeping" className="text-slate-400 hover:text-white">
            <ArrowLeft className="h-5 w-5" />
          </Link>
          <div>
            <h1 className="text-2xl font-bold text-white">Housekeeping Manager</h1>
            <p className="text-sm text-slate-400">Assign tasks · Manage shifts · Monitor workload</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={fetchData} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white" title="Refresh">
            <RefreshCw className="h-4 w-4" />
          </button>
          <button onClick={autoAssign} className="btn-primary text-xs gap-1.5">
            <Wand2 className="h-3.5 w-3.5" /> Auto Assign
          </button>
          <Link href="/housekeeping/staff" className="btn-secondary text-xs gap-1.5">
            <Users className="h-3.5 w-3.5" /> Cleaner Tablet
          </Link>
        </div>
      </div>

      {error && (
        <div className="rounded-lg border border-rose-500/20 bg-rose-500/5 px-4 py-3 text-sm text-rose-400">
          {error}
        </div>
      )}

      {/* Staff Overview Cards */}
      <div className="grid grid-cols-5 gap-3">
        {activeStaff.map((s) => {
          const assigned = tasks.filter((t) => t.assignedTo === s.id && t.status !== "ready").length;
          const completed = s.completedToday;
          const capacity = Math.round((assigned / s.maxRooms) * 100);
          return (
            <div key={s.id} className={`rounded-lg border p-3 ${assigned >= s.maxRooms ? "border-amber-500/30 bg-amber-500/5" : "border-slate-700 bg-slate-800/50"}`}>
              <div className="flex items-center justify-between">
                <span className="text-sm font-semibold text-white">{s.name}</span>
                <span className={`text-[10px] uppercase ${s.role === "inspector" ? "text-violet-400" : "text-sky-400"}`}>{s.role}</span>
              </div>
              <div className="mt-2 flex items-center justify-between text-xs text-slate-400">
                <span>Shift: {s.shift}</span>
                <span>Rating: {s.rating}</span>
              </div>
              <div className="mt-2">
                <div className="flex justify-between text-[10px] text-slate-400 mb-1">
                  <span>Load</span>
                  <span>{assigned}/{s.maxRooms}</span>
                </div>
                <div className="h-1.5 rounded-full bg-slate-700">
                  <div className={`h-1.5 rounded-full ${capacity > 90 ? "bg-rose-500" : capacity > 70 ? "bg-amber-500" : "bg-emerald-500"}`} style={{ width: `${Math.min(capacity, 100)}%` }} />
                </div>
              </div>
              <div className="mt-2 flex items-center gap-3 text-[10px] text-slate-400">
                <span className="flex items-center gap-1"><CheckCircle2 className="h-3 w-3 text-emerald-400" /> {completed}</span>
                <span className="flex items-center gap-1"><Clock className="h-3 w-3 text-sky-400" /> {assigned} active</span>
              </div>
              <div className="mt-2 flex flex-wrap gap-1">
                {s.floors.map((f) => (
                  <span key={f} className="rounded bg-slate-700 px-1.5 py-0.5 text-[10px] text-slate-300">Floor {f}</span>
                ))}
              </div>
            </div>
          );
        })}
      </div>

      {/* Filters + View Toggle */}
      <div className="flex items-center justify-between gap-3">
        <div className="flex items-center gap-2 flex-1">
          <div className="relative flex-1 max-w-xs">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
            <input className="input w-full pl-9 text-xs" placeholder="Search room..." value={search} onChange={(e) => setSearch(e.target.value)} />
          </div>
          <select value={floorFilter} onChange={(e) => setFloorFilter(e.target.value)} className="input text-xs">
            <option value="all">All Floors</option>
            {floors.map((f) => <option key={f} value={f}>Floor {f}</option>)}
          </select>
          <select value={shiftFilter} onChange={(e) => setShiftFilter(e.target.value as Shift | "all")} className="input text-xs">
            <option value="all">All Shifts</option>
            {shifts.map((s) => <option key={s} value={s}>{s}</option>)}
          </select>
        </div>
        <div className="flex items-center gap-1 rounded-lg bg-slate-800 p-1">
          <button onClick={() => setViewMode("floor")} className={`rounded-md px-3 py-1.5 text-xs ${viewMode === "floor" ? "bg-slate-700 text-white" : "text-slate-400"}`}>
            <Grid3X3 className="h-3.5 w-3.5 inline mr-1" />By Floor
          </button>
          <button onClick={() => setViewMode("staff")} className={`rounded-md px-3 py-1.5 text-xs ${viewMode === "staff" ? "bg-slate-700 text-white" : "text-slate-400"}`}>
            <Users className="h-3.5 w-3.5 inline mr-1" />By Staff
          </button>
        </div>
      </div>

      {/* Bulk Assign Bar */}
      {selectedTasks.length > 0 && (
        <div className="flex items-center gap-3 rounded-lg border border-nexus-500/30 bg-nexus-500/5 p-3">
          <span className="text-sm text-white">{selectedTasks.length} tasks selected</span>
          <select value={bulkAssignStaff} onChange={(e) => setBulkAssignStaff(e.target.value)} className="input text-xs py-1">
            <option value="">Assign to...</option>
            {activeStaff.filter((s) => s.role === "cleaner").map((s) => (
              <option key={s.id} value={s.id}>{s.name} ({s.currentLoad}/{s.maxRooms})</option>
            ))}
          </select>
          <button onClick={bulkAssign} disabled={!bulkAssignStaff} className="btn-primary text-xs disabled:opacity-50">
            Assign
          </button>
          <button onClick={() => setSelectedTasks([])} className="btn-secondary text-xs ml-auto">
            Clear
          </button>
        </div>
      )}

      {/* Floor Board View */}
      {viewMode === "floor" && (
        <div className="space-y-6">
          {floors.map((floor) => {
            const floorTasks = filteredTasks.filter((t) => t.floor === floor);
            return (
              <div key={floor} className="card space-y-3">
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                    <MapPin className="h-5 w-5 text-nexus-400" /> Floor {floor}
                  </h3>
                  <span className="text-xs text-slate-400">{floorTasks.length} rooms</span>
                </div>
                <div className="grid grid-cols-6 gap-3">
                  {floorTasks.map((task) => {
                    const config = statusConfig[task.status];
                    const assigned = staff.find((s) => s.id === task.assignedTo);
                    return (
                      <div key={task.id} className={`relative rounded-lg border p-3 ${task.priority ? "border-amber-500/30" : "border-slate-700"} bg-slate-800/50`}>
                        <div className="flex items-center gap-2 mb-1">
                          <input
                            type="checkbox"
                            checked={selectedTasks.includes(task.id)}
                            onChange={(e) => {
                              if (e.target.checked) setSelectedTasks([...selectedTasks, task.id]);
                              else setSelectedTasks(selectedTasks.filter((id) => id !== task.id));
                            }}
                            className="rounded border-slate-600"
                          />
                          <span className="font-bold text-white">{task.number}</span>
                          {task.vip && <Star className="h-3 w-3 text-amber-400" />}
                        </div>
                        <p className="text-[10px] text-slate-500">{task.type} · {task.taskType}</p>
                        <span className={`mt-1 inline-block rounded px-1.5 py-0.5 text-[10px] ${config.bg} ${config.color}`}>{config.label}</span>
                        {task.estimatedTime > 0 && (
                          <p className="mt-1 text-[10px] text-slate-500">~{task.estimatedTime} min</p>
                        )}
                        <div className="mt-2">
                          <select
                            value={task.assignedTo || ""}
                            onChange={(e) => assignTask(task.id, e.target.value)}
                            disabled={actionLoading === task.id}
                            className="input w-full text-[10px] py-1 disabled:opacity-50"
                          >
                            <option value="">Unassigned</option>
                            {activeStaff.filter((s) => s.role === "cleaner").map((s) => (
                              <option key={s.id} value={s.id}>{s.name}</option>
                            ))}
                          </select>
                        </div>
                        {assigned && (
                          <p className="mt-1 text-[10px] text-slate-400">Assigned: {assigned.name}</p>
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>
            );
          })}
        </div>
      )}

      {/* Staff Board View */}
      {viewMode === "staff" && (
        <div className="space-y-6">
          {activeStaff.filter((s) => s.role === "cleaner").map((s) => {
            const sTasks = filteredTasks.filter((t) => t.assignedTo === s.id || (!t.assignedTo && s.floors.includes(t.floor)));
            return (
              <div key={s.id} className="card space-y-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <User className="h-5 w-5 text-nexus-400" />
                    <h3 className="text-lg font-semibold text-white">{s.name}</h3>
                    <span className="text-xs text-slate-400">({s.currentLoad}/{s.maxRooms})</span>
                  </div>
                  <span className="text-xs text-slate-400">{sTasks.filter((t) => t.assignedTo === s.id).length} assigned</span>
                </div>
                <div className="grid grid-cols-6 gap-3">
                  {sTasks.map((task) => {
                    const config = statusConfig[task.status];
                    return (
                      <div key={task.id} className={`rounded-lg border p-3 ${task.assignedTo === s.id ? "border-nexus-500/30" : "border-slate-700 border-dashed"} bg-slate-800/50`}>
                        <div className="flex items-center justify-between">
                          <span className="font-bold text-white">{task.number}</span>
                          <span className={`rounded px-1.5 py-0.5 text-[10px] ${config.bg} ${config.color}`}>{config.label}</span>
                        </div>
                        <p className="text-[10px] text-slate-500">{task.type} · Floor {task.floor}</p>
                        {task.assignedTo !== s.id && (
                          <button onClick={() => assignTask(task.id, s.id)} disabled={actionLoading === task.id} className="mt-2 w-full rounded bg-nexus-500/10 py-1 text-[10px] text-nexus-400 hover:bg-nexus-500/20 disabled:opacity-50">
                            {actionLoading === task.id ? "..." : `Assign to ${s.name.split(" ")[0]}`}
                          </button>
                        )}
                      </div>
                    );
                  })}
                </div>
              </div>
            );
          })}
          {/* Unassigned Pool */}
          <div className="card space-y-3">
            <h3 className="text-lg font-semibold text-white">Unassigned Tasks</h3>
            <div className="grid grid-cols-6 gap-3">
              {filteredTasks.filter((t) => !t.assignedTo).map((task) => {
                const config = statusConfig[task.status];
                return (
                  <div key={task.id} className="rounded-lg border border-dashed border-slate-600 bg-slate-800/30 p-3">
                    <div className="flex items-center justify-between">
                      <span className="font-bold text-white">{task.number}</span>
                      <span className={`rounded px-1.5 py-0.5 text-[10px] ${config.bg} ${config.color}`}>{config.label}</span>
                    </div>
                    <p className="text-[10px] text-slate-500">{task.type} · Floor {task.floor}</p>
                    <div className="mt-2 flex flex-wrap gap-1">
                      {activeStaff.filter((s) => s.role === "cleaner" && s.floors.includes(task.floor)).map((s) => (
                        <button key={s.id} onClick={() => assignTask(task.id, s.id)} disabled={actionLoading === task.id} className="rounded bg-slate-700 px-2 py-0.5 text-[10px] text-white hover:bg-nexus-500 disabled:opacity-50">
                          {s.name.split(" ")[0]}
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
