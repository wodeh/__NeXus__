"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import {
  ArrowLeft, Play, Square, CheckCircle2, Clock, BedDouble, Package,
  Coffee, Droplets, Sparkles, Wrench, ChevronRight, User, AlertTriangle,
  Star, Wifi, Thermometer, Battery, MapPin, ClipboardList, RefreshCw,
} from "lucide-react";
import {
  HousekeepingTask, HousekeepingStaff,
  getHousekeepingTasks, getHousekeepingStaff, updateHousekeepingTask,
} from "@/lib/api";

type TaskStatus = "pending" | "in_progress" | "completed" | "paused";
type CardMode = "clean_full" | "clean_refill" | "maintenance";

interface CleanerTask {
  id: string;
  roomId: string;
  roomNumber: string;
  floor: string;
  type: string;
  status: TaskStatus;
  taskType: "clean" | "refill" | "maintenance" | "inspection";
  priority: boolean;
  vip: boolean;
  estimatedMin: number;
  startedAt?: string;
  checklist: ChecklistItem[];
  notes?: string;
  suppliesNeeded?: string[];
}

interface ChecklistItem {
  id: string;
  description: string;
  completed: boolean;
}

const defaultChecklist: Record<string, ChecklistItem[]> = {
  clean_full: [
    { id: "c1", description: "Strip beds & replace linens", completed: false },
    { id: "c2", description: "Clean bathroom & replace towels", completed: false },
    { id: "c3", description: "Dust all surfaces", completed: false },
    { id: "c4", description: "Vacuum / mop floor", completed: false },
    { id: "c5", description: "Restock minibar", completed: false },
    { id: "c6", description: "Refill coffee / tea station", completed: false },
    { id: "c7", description: "Check TV, AC, lights", completed: false },
    { id: "c8", description: "Empty trash", completed: false },
    { id: "c9", description: "Final inspection walk", completed: false },
  ],
  clean_refill: [
    { id: "r1", description: "Replace used towels", completed: false },
    { id: "r2", description: "Refill toiletries (shampoo, soap)", completed: false },
    { id: "r3", description: "Refill coffee / tea / sugar", completed: false },
    { id: "r4", description: "Top up minibar items", completed: false },
    { id: "r5", description: "Replace water bottles", completed: false },
    { id: "r6", description: "Quick wipe surfaces", completed: false },
  ],
  maintenance: [
    { id: "m1", description: "Assess reported issue", completed: false },
    { id: "m2", description: "Fix or escalate", completed: false },
    { id: "m3", description: "Test functionality", completed: false },
    { id: "m4", description: "Document work done", completed: false },
  ],
};

function mapApiTaskToCleanerTask(t: HousekeepingTask): CleanerTask {
  const mode: CardMode =
    t.task_type === "refill" ? "clean_refill" :
    t.task_type === "maintenance" ? "maintenance" :
    "clean_full";
  return {
    id: t.id,
    roomId: t.room_number,
    roomNumber: t.room_number,
    floor: "?",
    type: "Standard",
    status: t.status as TaskStatus,
    taskType: (t.task_type.replace("_", "") as any) || "clean",
    priority: t.priority === "urgent" || t.priority === "high",
    vip: false,
    estimatedMin: 25,
    startedAt: t.status === "in_progress" ? new Date().toISOString() : undefined,
    checklist: JSON.parse(JSON.stringify(defaultChecklist[mode] || defaultChecklist.clean_full)),
    notes: t.notes,
  };
}

export default function HousekeepingStaffPage() {
  const [tasks, setTasks] = useState<CleanerTask[]>([]);
  const [staff, setStaff] = useState<HousekeepingStaff[]>([]);
  const [selectedStaffId, setSelectedStaffId] = useState<string>("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTaskId, setActiveTaskId] = useState<string | null>(null);
  const [cardMode, setCardMode] = useState<CardMode>("clean_full");
  const [showModePicker, setShowModePicker] = useState(false);
  const [actionLoading, setActionLoading] = useState<string | null>(null);

  const activeTask = tasks.find((t) => t.id === activeTaskId);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [apiTasks, apiStaff] = await Promise.all([
        getHousekeepingTasks(), getHousekeepingStaff(),
      ]);
      setStaff(apiStaff);
      const mapped = apiTasks.map(mapApiTaskToCleanerTask);
      setTasks(mapped);
      // If no staff selected yet, default to first active cleaner
      if (!selectedStaffId && apiStaff.length > 0) {
        setSelectedStaffId(apiStaff[0].id);
      }
    } catch (e: any) {
      setError(e.message || "Failed to load tasks");
    } finally {
      setLoading(false);
    }
  }, [selectedStaffId]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const myTasks = tasks.filter((t) => {
    // We don't have per-cleaner auth, so we show ALL tasks but highlight ones that might be theirs
    // In a real setup, tasks would be filtered by assigned_to === currentUser.staffId
    return true;
  });

  const pendingTasks = myTasks.filter((t) => t.status === "pending");
  const inProgressTasks = myTasks.filter((t) => t.status === "in_progress");
  const completedToday = myTasks.filter((t) => t.status === "completed").length;

  async function startTask(taskId: string) {
    setActionLoading(taskId);
    try {
      await updateHousekeepingTask(taskId, { status: "in_progress" });
      setTasks(tasks.map((t) =>
        t.id === taskId
          ? { ...t, status: "in_progress" as TaskStatus, startedAt: new Date().toISOString() }
          : t.status === "in_progress"
          ? { ...t, status: "paused" as TaskStatus }
          : t
      ));
      setActiveTaskId(taskId);
      await fetchData();
    } catch (e: any) {
      alert(e.message);
    } finally {
      setActionLoading(null);
    }
  }

  function toggleChecklistItem(taskId: string, itemId: string) {
    setTasks(tasks.map((t) =>
      t.id === taskId
        ? {
            ...t,
            checklist: t.checklist.map((c) =>
              c.id === itemId ? { ...c, completed: !c.completed } : c
            ),
          }
        : t
    ));
  }

  async function completeTask(taskId: string) {
    const task = tasks.find((t) => t.id === taskId);
    if (!task) return;
    const completedCount = task.checklist.filter((c) => c.completed).length;
    const total = task.checklist.length;
    if (completedCount < total * 0.5) {
      if (!confirm(`Only ${completedCount}/${total} checklist items done. Complete anyway?`)) return;
    }
    setActionLoading(taskId);
    try {
      await updateHousekeepingTask(taskId, { status: "completed" });
      setTasks(tasks.map((t) => (t.id === taskId ? { ...t, status: "completed" as TaskStatus } : t)));
      setActiveTaskId(null);
      await fetchData();
    } catch (e: any) {
      alert(e.message);
    } finally {
      setActionLoading(null);
    }
  }

  const modeLabel: Record<CardMode, string> = {
    clean_full: "Full Clean",
    clean_refill: "Refill Only",
    maintenance: "Maintenance",
  };

  const modeIcon: Record<CardMode, React.ElementType> = {
    clean_full: Sparkles,
    clean_refill: Package,
    maintenance: Wrench,
  };

  const modeColor: Record<CardMode, string> = {
    clean_full: "bg-nexus-500",
    clean_refill: "bg-sky-500",
    maintenance: "bg-amber-500",
  };

  const currentStaff = staff.find((s) => s.id === selectedStaffId);

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-900 flex items-center justify-center text-slate-500">
        <RefreshCw className="mr-2 h-5 w-5 animate-spin" /> Loading cleaner tablet...
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-900">
      {/* Top Bar */}
      <div className="sticky top-0 z-40 border-b border-slate-800 bg-slate-900/95 backdrop-blur">
        <div className="flex items-center justify-between px-4 py-3">
          <div className="flex items-center gap-3">
            <Link href="/housekeeping" className="text-slate-400 hover:text-white">
              <ArrowLeft className="h-6 w-6" />
            </Link>
            <div>
              <h1 className="text-lg font-bold text-white">Cleaner Tablet</h1>
              <p className="text-xs text-slate-400">
                {currentStaff ? `${currentStaff.name} · ${currentStaff.shift} Shift` : "Select staff member"}
              </p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <select
              value={selectedStaffId}
              onChange={(e) => setSelectedStaffId(e.target.value)}
              className="input text-xs py-1.5"
            >
              <option value="">Select staff...</option>
              {staff.map((s) => (
                <option key={s.id} value={s.id}>{s.name} · {s.department}</option>
              ))}
            </select>
            <button onClick={fetchData} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white" title="Refresh">
              <RefreshCw className="h-4 w-4" />
            </button>
            <button
              onClick={() => setShowModePicker(true)}
              className={`flex items-center gap-2 rounded-lg px-4 py-2 text-sm font-medium text-white ${modeColor[cardMode]}`}
            >
              {(() => {
                const Icon = modeIcon[cardMode];
                return <Icon className="h-4 w-4" />;
              })()}
              {modeLabel[cardMode]}
            </button>
          </div>
        </div>
      </div>

      {error && (
        <div className="mx-4 mt-3 rounded-lg border border-rose-500/20 bg-rose-500/5 px-4 py-3 text-sm text-rose-400">
          {error}
        </div>
      )}

      {/* Stats Row */}
      <div className="grid grid-cols-3 gap-2 px-4 py-3">
        <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-3 text-center">
          <p className="text-xs text-slate-400">Pending</p>
          <p className="text-2xl font-bold text-white">{pendingTasks.length}</p>
        </div>
        <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-3 text-center">
          <p className="text-xs text-slate-400">Active</p>
          <p className="text-2xl font-bold text-sky-400">{inProgressTasks.length}</p>
        </div>
        <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-3 text-center">
          <p className="text-xs text-slate-400">Done Today</p>
          <p className="text-2xl font-bold text-emerald-400">{completedToday}</p>
        </div>
      </div>

      <div className="px-4 pb-6 space-y-4">
        {/* Active Task Panel */}
        {activeTask && (
          <div className="rounded-xl border border-nexus-500/30 bg-nexus-500/5 p-4">
            <div className="flex items-center justify-between mb-3">
              <div className="flex items-center gap-2">
                <Play className="h-5 w-5 text-nexus-400" />
                <span className="text-lg font-bold text-white">Room {activeTask.roomNumber}</span>
                {activeTask.vip && <Star className="h-4 w-4 text-amber-400" />}
                {activeTask.priority && <AlertTriangle className="h-4 w-4 text-amber-400" />}
              </div>
              <button
                onClick={() => completeTask(activeTask.id)}
                disabled={actionLoading === activeTask.id}
                className="rounded-lg bg-emerald-500 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-600 disabled:opacity-50"
              >
                <CheckCircle2 className="h-4 w-4 inline mr-1" />
                {actionLoading === activeTask.id ? "..." : "Done"}
              </button>
            </div>
            <p className="text-xs text-slate-400 mb-3">
              {activeTask.type} · Floor {activeTask.floor} · {activeTask.taskType} · Started{" "}
              {activeTask.startedAt
                ? new Date(activeTask.startedAt).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })
                : "—"}
            </p>
            {activeTask.notes && (
              <div className="mb-3 rounded bg-amber-500/5 p-2 text-xs text-amber-400">{activeTask.notes}</div>
            )}
            <div className="space-y-2">
              {activeTask.checklist.map((item) => (
                <button
                  key={item.id}
                  onClick={() => toggleChecklistItem(activeTask.id, item.id)}
                  className={`flex w-full items-center gap-3 rounded-lg border p-3 text-left transition-all ${
                    item.completed
                      ? "border-emerald-500/30 bg-emerald-500/10"
                      : "border-slate-700 bg-slate-800/50"
                  }`}
                >
                  <div
                    className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full border-2 ${
                      item.completed
                        ? "border-emerald-500 bg-emerald-500 text-white"
                        : "border-slate-600"
                    }`}
                  >
                    {item.completed && <CheckCircle2 className="h-4 w-4" />}
                  </div>
                  <span className={`text-sm ${item.completed ? "text-emerald-400 line-through" : "text-white"}`}>
                    {item.description}
                  </span>
                </button>
              ))}
            </div>
          </div>
        )}

        {/* Pending Tasks List */}
        <div className="space-y-3">
          <h2 className="text-sm font-semibold text-slate-400 uppercase tracking-wider">My Tasks</h2>
          {pendingTasks.map((task) => {
            const checklistDone = task.checklist.filter((c) => c.completed).length;
            const checklistTotal = task.checklist.length;
            return (
              <button
                key={task.id}
                onClick={() => startTask(task.id)}
                disabled={actionLoading === task.id}
                className="flex w-full items-center gap-4 rounded-xl border border-slate-700 bg-slate-800/50 p-4 text-left hover:border-slate-600 transition-colors disabled:opacity-50"
              >
                <div className="flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-slate-700">
                  <BedDouble className="h-6 w-6 text-slate-400" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center gap-2">
                    <span className="text-lg font-bold text-white">{task.roomNumber}</span>
                    {task.vip && <Star className="h-4 w-4 text-amber-400" />}
                    {task.priority && <AlertTriangle className="h-4 w-4 text-amber-400" />}
                    <span className="ml-auto rounded bg-slate-700 px-2 py-0.5 text-[10px] text-slate-300">
                      {task.taskType}
                    </span>
                  </div>
                  <p className="text-xs text-slate-400">
                    {task.type} · Floor {task.floor} · ~{task.estimatedMin} min
                  </p>
                  {task.suppliesNeeded && task.suppliesNeeded.length > 0 && (
                    <p className="text-xs text-sky-400 mt-0.5">
                      <Package className="h-3 w-3 inline mr-1" />
                      {task.suppliesNeeded.join(", ")}
                    </p>
                  )}
                  {checklistDone > 0 && (
                    <div className="mt-1 h-1 rounded-full bg-slate-700">
                      <div
                        className="h-1 rounded-full bg-nexus-500"
                        style={{ width: `${(checklistDone / checklistTotal) * 100}%` }}
                      />
                    </div>
                  )}
                </div>
                <ChevronRight className="h-5 w-5 text-slate-500" />
              </button>
            );
          })}
          {pendingTasks.length === 0 && (
            <div className="rounded-xl border border-slate-700 bg-slate-800/30 p-8 text-center">
              <CheckCircle2 className="h-10 w-10 text-emerald-400 mx-auto mb-2" />
              <p className="text-white font-medium">All tasks completed!</p>
              <p className="text-xs text-slate-400">Great work today.</p>
            </div>
          )}
        </div>
      </div>

      {/* Mode Picker Modal */}
      {showModePicker && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm px-4">
          <div className="w-full max-w-sm rounded-2xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
            <h3 className="text-lg font-bold text-white mb-4">Select Card Mode</h3>
            <div className="space-y-3">
              {(["clean_full", "clean_refill", "maintenance"] as CardMode[]).map((mode) => {
                const Icon = modeIcon[mode];
                return (
                  <button
                    key={mode}
                    onClick={() => {
                      setCardMode(mode);
                      setShowModePicker(false);
                    }}
                    className={`flex w-full items-center gap-4 rounded-xl border p-4 text-left transition-all ${
                      cardMode === mode
                        ? "border-nexus-500 bg-nexus-500/10"
                        : "border-slate-700 bg-slate-800/50 hover:border-slate-600"
                    }`}
                  >
                    <div className={`flex h-12 w-12 shrink-0 items-center justify-center rounded-xl ${modeColor[mode]}`}>
                      <Icon className="h-6 w-6 text-white" />
                    </div>
                    <div>
                      <p className="font-semibold text-white">{modeLabel[mode]}</p>
                      <p className="text-xs text-slate-400">
                        {mode === "clean_full"
                          ? "Full room cleaning with checklist"
                          : mode === "clean_refill"
                          ? "Restock amenities & minibar only"
                          : "Maintenance & repair tasks"}
                      </p>
                    </div>
                  </button>
                );
              })}
            </div>
            <button
              onClick={() => setShowModePicker(false)}
              className="mt-4 w-full rounded-lg border border-slate-700 py-2 text-sm text-slate-400 hover:text-white"
            >
              Cancel
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
