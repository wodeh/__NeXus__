"use client";

import { useState } from "react";
import { Search, ClipboardList, Sparkles, Wrench, CheckCircle } from "lucide-react";

const mockTasks = [
  { id: "hk-001", room: "412", type: "cleaning", status: "pending", priority: "high", assigned_to: "Maria G." },
  { id: "hk-002", room: "305", type: "inspection", status: "in_progress", priority: "normal", assigned_to: "John D." },
  { id: "hk-003", room: "108", type: "maintenance", status: "completed", priority: "normal", assigned_to: "Tech Team" },
  { id: "hk-004", room: "201", type: "cleaning", status: "in_progress", priority: "high", assigned_to: "Maria G." },
];

const typeIcons: Record<string, React.ElementType> = {
  cleaning: Sparkles,
  inspection: CheckCircle,
  maintenance: Wrench,
};

const statusColors: Record<string, string> = {
  pending: "badge-amber",
  in_progress: "badge-blue",
  completed: "badge-green",
};

export default function HousekeepingPage() {
  const [search, setSearch] = useState("");
  const filtered = mockTasks.filter((t) => t.room.includes(search));

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">Housekeeping</h2>
        <p className="text-sm text-slate-400">Room cleaning and maintenance tasks</p>
      </div>

      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
          <input className="input w-full pl-9" placeholder="Search by room..." value={search} onChange={(e) => setSearch(e.target.value)} />
        </div>
        <button className="btn-primary">
          <ClipboardList className="h-4 w-4" />
          New Task
        </button>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
        {filtered.map((t) => {
          const Icon = typeIcons[t.type] || ClipboardList;
          return (
            <div key={t.id} className="card-hover">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <Icon className="h-4 w-4 text-nexus-400" />
                  <span className="font-medium text-white">Room {t.room}</span>
                </div>
                <span className={statusColors[t.status]}>{t.status.replace("_", " ")}</span>
              </div>
              <div className="mt-3 space-y-1 text-sm text-slate-400">
                <p>Type: {t.type}</p>
                <p>Priority: {t.priority}</p>
                <p>Assigned: {t.assigned_to}</p>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}
