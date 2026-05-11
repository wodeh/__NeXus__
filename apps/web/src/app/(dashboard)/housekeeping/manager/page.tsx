"use client";

import { useState } from "react";
import {
  Users, ArrowLeft, ChevronRight, Clock, BedDouble, CheckCircle2, User, Plus, Trash2,
  RotateCcw, Sparkles, AlertTriangle, Star, Filter, Search, Grid3X3, List, Wand2,
  Calendar, ChevronDown, MapPin
} from "lucide-react";
import Link from "next/link";

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

const mockStaff: Staff[] = [
  { id: "s1", name: "Maria K.", role: "cleaner", active: true, shift: "morning", floors: ["1", "2"], maxRooms: 8, currentLoad: 3, rating: 4.8, completedToday: 5 },
  { id: "s2", name: "John D.", role: "cleaner", active: true, shift: "morning", floors: ["2", "3"], maxRooms: 8, currentLoad: 4, rating: 4.5, completedToday: 4 },
  { id: "s3", name: "Sofia L.", role: "cleaner", active: true, shift: "afternoon", floors: ["3", "4"], maxRooms: 6, currentLoad: 2, rating: 4.9, completedToday: 6 },
  { id: "s4", name: "Ahmed R.", role: "inspector", active: true, shift: "morning", floors: ["1", "2", "3", "4"], maxRooms: 12, currentLoad: 5, rating: 4.7, completedToday: 3 },
  { id: "s5", name: "Chen W.", role: "cleaner", active: false, shift: "night", floors: ["1"], maxRooms: 5, currentLoad: 0, rating: 4.2, completedToday: 0 },
];

const mockTasks: RoomTask[] = [
  { id: "101", number: "101", floor: "1", type: "Standard", status: "dirty", priority: false, vip: false, estimatedTime: 25, taskType: "clean" },
  { id: "102", number: "102", floor: "1", type: "Standard", status: "dirty", priority: true, vip: false, estimatedTime: 25, taskType: "clean" },
  { id: "103", number: "103", floor: "1", type: "Deluxe", status: "in_progress", assignedTo: "s2", priority: true, vip: true, estimatedTime: 30, taskType: "clean" },
  { id: "105", number: "105", floor: "1", type: "Standard", status: "maintenance", priority: false, vip: false, estimatedTime: 15, taskType: "maintenance" },
  { id: "201", number: "201", floor: "2", type: "Deluxe King", status: "dirty", priority: true, vip: true, estimatedTime: 35, taskType: "clean" },
  { id: "202", number: "202", floor: "2", type: "Deluxe King", status: "dirty", priority: false, vip: false, estimatedTime: 25, taskType: "clean" },
  { id: "203", number: "203", floor: "2", type: "Suite", status: "inspected", assignedTo: "s3", priority: false, vip: false, estimatedTime: 10, taskType: "inspection" },
  { id: "301", number: "301", floor: "3", type: "Suite", status: "dirty", priority: false, vip: false, estimatedTime: 40, taskType: "clean" },
  { id: "302", number: "302", floor: "3", type: "Suite", status: "ready", assignedTo: "s1", priority: false, vip: false, estimatedTime: 20, taskType: "refill" },
  { id: "303", number: "303", floor: "3", type: "Deluxe", status: "in_progress", assignedTo: "s3", priority: false, vip: false, estimatedTime: 30, taskType: "clean" },
  { id: "304", number: "304", floor: "3", type: "Standard", status: "dirty", priority: false, vip: false, estimatedTime: 20, taskType: "clean" },
  { id: "401", number: "401", floor: "4", type: "Suite", status: "dirty", priority: false, vip: false, estimatedTime: 40, taskType: "clean" },
  { id: "402", number: "402", floor: "4", type: "Deluxe", status: "dirty", priority: true, vip: false, estimatedTime: 30, taskType: "clean" },
];

const statusConfig: Record<RoomStatus, { label: string; color: string; bg: string }> = {
  dirty: { label: "Dirty", color: "text-rose-400", bg: "bg-rose-500/10" },
  in_progress: { label: "In Progress", color: "text-sky-400", bg: "bg-sky-500/10" },
  inspected: { label: "Inspected", color: "text-violet-400", bg: "bg-violet-500/10" },
  ready: { label: "Ready", color: "text-emerald-400", bg: "bg-emerald-500/10" },
  blocked: { label: "Blocked", color: "text-slate-400", bg: "bg-slate-500/10" },
  maintenance: { label: "Maintenance", color: "text-amber-400", bg: "bg-amber-500/10" },
};

export default function HousekeepingManagerPage() {
  const [tasks, setTasks] = useState<RoomTask[]>(mockTasks);
  const [staff, setStaff] = useState<Staff[]>(mockStaff);
  const [shiftFilter, setShiftFilter] = useState<Shift | "all">("all");
  const [floorFilter, setFloorFilter] = useState("all");
  const [search, setSearch] = useState("");
  const [selectedTasks, setSelectedTasks] = useState<string[]>([]);
  const [bulkAssignStaff, setBulkAssignStaff] = useState("");
  const [viewMode, setViewMode] = useState<"floor" | "staff">("floor");

  const floors = Array.from(new Set(tasks.map((t) => t.floor))).sort();

  const filteredTasks = tasks.filter((t) => {
    const matchesFloor = floorFilter === "all" || t.floor === floorFilter;
    const matchesSearch = t.number.includes(search) || t.type.toLowerCase().includes(search.toLowerCase());
    return matchesFloor && matchesSearch;
  });

  const assignTask = (taskId: string, staffId: string) => {
    setTasks(tasks.map((t) => t.id === taskId ? { ...t, assignedTo: staffId || undefined } : t));
    setStaff(staff.map((s) => {
      const task = tasks.find((t) => t.id === taskId);
      if (!task) return s;
      if (s.id === staffId) return { ...s, currentLoad: s.currentLoad + 1 };
      if (s.id === task.assignedTo) return { ...s, currentLoad: Math.max(0, s.currentLoad - 1) };
      return s;
    }));
  };

  const autoAssign = () => {
    const unassigned = tasks.filter((t) => !t.assignedTo && t.status === "dirty");
    const activeStaff = staff.filter((s) => s.active && s.role === "cleaner");
    let updatedTasks = [...tasks];
    let updatedStaff = [...staff];

    for (const task of unassigned) {
      const candidates = activeStaff.filter((s) => {
        const workload = updatedStaff.find((us) => us.id === s.id)?.currentLoad || s.currentLoad;
        return workload < s.maxRooms && s.floors.includes(task.floor);
      }).sort((a, b) => {
        const aLoad = updatedStaff.find((us) => us.id === a.id)?.currentLoad || a.currentLoad;
        const bLoad = updatedStaff.find((us) => us.id === b.id)?.currentLoad || b.currentLoad;
        return aLoad - bLoad;
      });

      if (candidates.length > 0) {
        const chosen = candidates[0];
        updatedTasks = updatedTasks.map((t) => t.id === task.id ? { ...t, assignedTo: chosen.id } : t);
        updatedStaff = updatedStaff.map((s) => s.id === chosen.id ? { ...s, currentLoad: s.currentLoad + 1 } : s);
      }
    }
    setTasks(updatedTasks);
    setStaff(updatedStaff);
  };

  const bulkAssign = () => {
    if (!bulkAssignStaff) return;
    let updatedTasks = [...tasks];
    let updatedStaff = [...staff];
    for (const taskId of selectedTasks) {
      const task = tasks.find((t) => t.id === taskId);
      if (!task || task.assignedTo === bulkAssignStaff) continue;
      updatedTasks = updatedTasks.map((t) => t.id === taskId ? { ...t, assignedTo: bulkAssignStaff } : t);
      updatedStaff = updatedStaff.map((s) => {
        if (s.id === bulkAssignStaff) return { ...s, currentLoad: s.currentLoad + 1 };
        if (s.id === task.assignedTo) return { ...s, currentLoad: Math.max(0, s.currentLoad - 1) };
        return s;
      });
    }
    setTasks(updatedTasks);
    setStaff(updatedStaff);
    setSelectedTasks([]);
    setBulkAssignStaff("");
  };

  const activeStaff = staff.filter((s) => s.active && (shiftFilter === "all" || s.shift === shiftFilter));

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
          <button onClick={autoAssign} className="btn-primary text-xs gap-1.5">
            <Wand2 className="h-3.5 w-3.5" /> Auto Assign
          </button>
          <Link href="/housekeeping/staff" className="btn-secondary text-xs gap-1.5">
            <Users className="h-3.5 w-3.5" /> Cleaner Tablet
          </Link>
        </div>
      </div>

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
                            className="input w-full text-[10px] py-1"
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
                          <button onClick={() => assignTask(task.id, s.id)} className="mt-2 w-full rounded bg-nexus-500/10 py-1 text-[10px] text-nexus-400 hover:bg-nexus-500/20">
                            Assign to {s.name.split(" ")[0]}
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
                        <button key={s.id} onClick={() => assignTask(task.id, s.id)} className="rounded bg-slate-700 px-2 py-0.5 text-[10px] text-white hover:bg-nexus-500">
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
