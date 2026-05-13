"use client";

import { useState, useEffect, useCallback } from "react";
import {
  Smartphone, MapPin, Thermometer, ArrowUp, Battery, RefreshCw,
  CheckCircle2, AlertTriangle, Clock, Home, Wifi, WifiOff,
  CalendarDays, Plus, X, User, ListTodo, History, Radio,
} from "lucide-react";
import {
  getVillas, getSensorLogs, recordSensorLog,
  VillaProperty, CleanerSensorLog,
} from "@/lib/api";

interface CleaningSchedule {
  id: string;
  villa_id: string;
  villa_name: string;
  cleaner_name: string;
  scheduled_date: string;
  status: "scheduled" | "in_progress" | "completed" | "skipped";
  notes: string;
  completed_at?: string;
}

export default function CleanerTrackingPage() {
  const [villas, setVillas] = useState<VillaProperty[]>([]);
  const [selectedVilla, setSelectedVilla] = useState<string>("");
  const [logs, setLogs] = useState<CleanerSensorLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tracking, setTracking] = useState(false);
  const [cleanerName, setCleanerName] = useState("");
  const [cleanerId, setCleanerId] = useState("");
  const [tab, setTab] = useState<"tracking" | "schedule" | "history">("tracking");
  const [schedules, setSchedules] = useState<CleaningSchedule[]>([]);
  const [showScheduleForm, setShowScheduleForm] = useState(false);

  const fetchData = useCallback(async () => {
    try {
      const v = await getVillas();
      setVillas(v);
      if (v.length > 0 && !selectedVilla) {
        setSelectedVilla(v[0].id);
      }
    } catch (e: any) {
      setError(e.message);
    } finally {
      setLoading(false);
    }
  }, [selectedVilla]);

  useEffect(() => {
    fetchData();
    // Load demo schedules
    setSchedules([
      { id: "sch-1", villa_id: "villa-001", villa_name: "Villa Al-Mashta", cleaner_name: "Muhammad", scheduled_date: "2024-06-14", status: "completed", notes: "Deep clean after checkout", completed_at: "2024-06-14T11:30:00Z" },
      { id: "sch-2", villa_id: "villa-002", villa_name: "Chalet Al-Balad", cleaner_name: "Ahmad", scheduled_date: "2024-06-15", status: "scheduled", notes: "Regular maintenance" },
      { id: "sch-3", villa_id: "villa-001", villa_name: "Villa Al-Mashta", cleaner_name: "Muhammad", scheduled_date: "2024-06-20", status: "scheduled", notes: "Post-guest cleaning" },
      { id: "sch-4", villa_id: "villa-003", villa_name: "Villa Al-Reef", cleaner_name: "Layla", scheduled_date: "2024-06-16", status: "in_progress", notes: "Pool area focus" },
    ]);
  }, [fetchData]);

  useEffect(() => {
    if (!selectedVilla) return;
    getSensorLogs(selectedVilla).then(setLogs).catch(console.error);
  }, [selectedVilla]);

  const startTracking = async () => {
    if (!selectedVilla || !cleanerName) {
      alert("Select villa and enter your name");
      return;
    }

    const id = `cleaner-${Date.now()}`;
    setCleanerId(id);
    setTracking(true);

    if (!navigator.geolocation) {
      alert("Geolocation not supported");
      return;
    }

    const villa = villas.find((v) => v.id === selectedVilla);
    if (!villa) return;

    const watchId = navigator.geolocation.watchPosition(
      async (position) => {
        const { latitude, longitude, altitude } = position.coords;
        const elevation = villa.elevation ?? 0;
        const floor = altitude ? Math.round((altitude - elevation) / 3.5) : 0;
        const temp = altitude && altitude > elevation + 2 ? 31.5 : 23.0;
        const locationType = temp > 25 ? "outside" : floor === 0 ? "inside" : floor > 0 ? "inside" : "outside";

        const payload = {
          villa_id: selectedVilla,
          villa_name: villa.name,
          cleaner_id: id,
          cleaner_name: cleanerName,
          temperature: temp,
          latitude,
          longitude,
          altitude: altitude || elevation,
          floor: Math.max(0, floor),
          location_type: locationType,
          battery_level: 85,
          recorded_at: new Date().toISOString(),
        };

        try {
          await recordSensorLog(selectedVilla, payload);
          const updated = await getSensorLogs(selectedVilla);
          setLogs(updated);
        } catch (e) {
          console.error("Failed to record sensor:", e);
        }
      },
      (err) => {
        console.error("Geolocation error:", err);
      },
      { enableHighAccuracy: true, maximumAge: 30000, timeout: 27000 }
    );

    setTimeout(() => {
      navigator.geolocation.clearWatch(watchId);
      setTracking(false);
    }, 8 * 60 * 60 * 1000);

    return () => navigator.geolocation.clearWatch(watchId);
  };

  const addSchedule = (schedule: Omit<CleaningSchedule, "id">) => {
    const newSchedule: CleaningSchedule = {
      ...schedule,
      id: `sch-${Date.now()}`,
    };
    setSchedules((prev) => [...prev, newSchedule]);
    setShowScheduleForm(false);
  };

  const updateScheduleStatus = (id: string, status: CleaningSchedule["status"]) => {
    setSchedules((prev) =>
      prev.map((s) =>
        s.id === id
          ? { ...s, status, completed_at: status === "completed" ? new Date().toISOString() : s.completed_at }
          : s
      )
    );
  };

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <RefreshCw className="h-8 w-8 animate-spin text-nexus-400" />
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-slate-950 p-4 space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold text-white flex items-center gap-2">
          <Smartphone className="h-6 w-6 text-nexus-400" />
          Cleaner Tracking
        </h1>
        <div className="flex items-center gap-2 text-xs text-slate-400">
          {tracking ? (
            <span className="flex items-center gap-1 text-emerald-400">
              <Wifi className="h-3 w-3" /> Live
            </span>
          ) : (
            <span className="flex items-center gap-1 text-slate-500">
              <WifiOff className="h-3 w-3" /> Off
            </span>
          )}
        </div>
      </div>

      {/* Villa Selector */}
      <div className="space-y-2">
        <label className="text-xs text-slate-400">Select Villa</label>
        <select
          className="input w-full"
          value={selectedVilla}
          onChange={(e) => setSelectedVilla(e.target.value)}
        >
          {villas.map((v) => (
            <option key={v.id} value={v.id}>
              {v.name} ({v.city})
            </option>
          ))}
        </select>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        {(["tracking", "schedule", "history"] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              tab === t ? "bg-slate-800 text-white" : "text-slate-400 hover:bg-slate-800 hover:text-white"
            }`}
          >
            {t === "tracking" && <Radio className="h-4 w-4 inline mr-1" />}
            {t === "schedule" && <ListTodo className="h-4 w-4 inline mr-1" />}
            {t === "history" && <History className="h-4 w-4 inline mr-1" />}
            {t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>

      {/* Tracking Tab */}
      {tab === "tracking" && (
        <div className="space-y-6">
          {!tracking && (
            <div className="space-y-3 rounded-lg bg-slate-900 p-4">
              <input
                className="input w-full"
                placeholder="Your Name"
                value={cleanerName}
                onChange={(e) => setCleanerName(e.target.value)}
              />
              <button onClick={startTracking} className="btn-primary w-full gap-2">
                <CheckCircle2 className="h-4 w-4" />
                Start Shift
              </button>
            </div>
          )}

          {tracking && (
            <div className="space-y-4 rounded-lg bg-slate-900 p-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm font-medium text-white">{cleanerName}</p>
                  <p className="text-xs text-emerald-400">Tracking active</p>
                </div>
                <div className="h-3 w-3 animate-pulse rounded-full bg-emerald-400" />
              </div>

              {logs.length > 0 && (
                <div className="space-y-2">
                  <p className="text-xs text-slate-400">Latest Position</p>
                  <LatestPositionCard log={logs[0]} villa={villas.find((v) => v.id === selectedVilla)} />
                </div>
              )}
            </div>
          )}

          {/* Live Logs */}
          <div className="space-y-3">
            <p className="text-xs font-medium text-slate-400 uppercase">Live Session</p>
            {logs.slice(0, 5).map((log) => (
              <LogCard key={log.id} log={log} />
            ))}
            {logs.length === 0 && (
              <p className="text-center text-sm text-slate-500 py-8">No tracking data yet. Start a shift.</p>
            )}
          </div>
        </div>
      )}

      {/* Schedule Tab */}
      {tab === "schedule" && (
        <div className="space-y-4">
          <div className="flex items-center justify-between">
            <h2 className="text-sm font-medium text-white">Cleaning Schedules</h2>
            <button onClick={() => setShowScheduleForm(true)} className="btn-primary gap-2 text-xs">
              <Plus className="h-3 w-3" /> Add Schedule
            </button>
          </div>

          {showScheduleForm && (
            <ScheduleForm
              villas={villas}
              onSave={addSchedule}
              onClose={() => setShowScheduleForm(false)}
            />
          )}

          <div className="space-y-2">
            {schedules
              .filter((s) => selectedVilla === "" || s.villa_id === selectedVilla)
              .map((schedule) => (
                <div
                  key={schedule.id}
                  className="flex items-center justify-between rounded-lg border border-slate-700 bg-slate-800/50 p-3"
                >
                  <div className="flex items-center gap-3">
                    <div
                      className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-full ${
                        schedule.status === "completed"
                          ? "bg-emerald-500/10"
                          : schedule.status === "in_progress"
                          ? "bg-amber-500/10"
                          : schedule.status === "scheduled"
                          ? "bg-sky-500/10"
                          : "bg-rose-500/10"
                      }`}
                    >
                      <CheckCircle2
                        className={`h-5 w-5 ${
                          schedule.status === "completed"
                            ? "text-emerald-400"
                            : schedule.status === "in_progress"
                            ? "text-amber-400"
                            : schedule.status === "scheduled"
                            ? "text-sky-400"
                            : "text-rose-400"
                        }`}
                      />
                    </div>
                    <div>
                      <p className="text-sm font-medium text-white">{schedule.villa_name}</p>
                      <p className="text-xs text-slate-400">
                        {schedule.cleaner_name} · {schedule.scheduled_date}
                      </p>
                      <p className="text-xs text-slate-500">{schedule.notes}</p>
                    </div>
                  </div>
                  <div className="flex items-center gap-2">
                    <span
                      className={`rounded-full px-2 py-0.5 text-[10px] font-medium capitalize ${
                        schedule.status === "completed"
                          ? "bg-emerald-500/10 text-emerald-400"
                          : schedule.status === "in_progress"
                          ? "bg-amber-500/10 text-amber-400"
                          : schedule.status === "scheduled"
                          ? "bg-sky-500/10 text-sky-400"
                          : "bg-rose-500/10 text-rose-400"
                      }`}
                    >
                      {schedule.status.replace("_", " ")}
                    </span>
                    {schedule.status === "scheduled" && (
                      <>
                        <button
                          onClick={() => updateScheduleStatus(schedule.id, "in_progress")}
                          className="rounded bg-amber-500/10 px-2 py-1 text-[10px] text-amber-400 hover:bg-amber-500/20"
                        >
                          Start
                        </button>
                        <button
                          onClick={() => updateScheduleStatus(schedule.id, "skipped")}
                          className="rounded bg-rose-500/10 px-2 py-1 text-[10px] text-rose-400 hover:bg-rose-500/20"
                        >
                          Skip
                        </button>
                      </>
                    )}
                    {schedule.status === "in_progress" && (
                      <button
                        onClick={() => updateScheduleStatus(schedule.id, "completed")}
                        className="rounded bg-emerald-500/10 px-2 py-1 text-[10px] text-emerald-400 hover:bg-emerald-500/20"
                      >
                        Complete
                      </button>
                    )}
                  </div>
                </div>
              ))}
            {schedules.length === 0 && (
              <p className="text-center text-sm text-slate-500 py-8">No schedules yet</p>
            )}
          </div>
        </div>
      )}

      {/* History Tab */}
      {tab === "history" && (
        <div className="space-y-3">
          <p className="text-xs font-medium text-slate-400 uppercase">All Session History</p>
          {logs.map((log) => (
            <LogCard key={log.id} log={log} />
          ))}
          {logs.length === 0 && (
            <p className="text-center text-sm text-slate-500 py-8">No tracking data yet</p>
          )}
        </div>
      )}
    </div>
  );
}

function ScheduleForm({
  villas,
  onSave,
  onClose,
}: {
  villas: VillaProperty[];
  onSave: (schedule: Omit<CleaningSchedule, "id">) => void;
  onClose: () => void;
}) {
  const [form, setForm] = useState({
    villa_id: villas[0]?.id || "",
    cleaner_name: "",
    scheduled_date: new Date().toISOString().split("T")[0],
    notes: "",
  });

  const selectedVilla = villas.find((v) => v.id === form.villa_id);

  return (
    <div className="rounded-lg border border-slate-700 bg-slate-900 p-4 space-y-3">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-medium text-white">Add Cleaning Schedule</h3>
        <button onClick={onClose} className="text-slate-400 hover:text-white">
          <X className="h-4 w-4" />
        </button>
      </div>

      <select
        className="input w-full"
        value={form.villa_id}
        onChange={(e) => setForm({ ...form, villa_id: e.target.value })}
      >
        {villas.map((v) => (
          <option key={v.id} value={v.id}>
            {v.name}
          </option>
        ))}
      </select>

      <input
        className="input w-full"
        placeholder="Cleaner Name"
        value={form.cleaner_name}
        onChange={(e) => setForm({ ...form, cleaner_name: e.target.value })}
      />

      <input
        type="date"
        className="input w-full"
        value={form.scheduled_date}
        onChange={(e) => setForm({ ...form, scheduled_date: e.target.value })}
      />

      <textarea
        className="input w-full"
        rows={2}
        placeholder="Notes (e.g., Deep clean, pool focus...)"
        value={form.notes}
        onChange={(e) => setForm({ ...form, notes: e.target.value })}
      />

      <button
        onClick={() =>
          onSave({
            villa_id: form.villa_id,
            villa_name: selectedVilla?.name || "",
            cleaner_name: form.cleaner_name,
            scheduled_date: form.scheduled_date,
            status: "scheduled",
            notes: form.notes,
          })
        }
        className="btn-primary w-full gap-2"
      >
        <Plus className="h-4 w-4" /> Add Schedule
      </button>
    </div>
  );
}

function LatestPositionCard({ log, villa }: { log: CleanerSensorLog; villa?: VillaProperty }) {
  return (
    <div className="grid grid-cols-2 gap-3">
      <div className={`rounded-lg p-3 ${(log.temperature ?? 0) > 25 ? "bg-amber-500/10" : "bg-sky-500/10"}`}>
        <div className="flex items-center gap-2">
          <Thermometer className={`h-4 w-4 ${(log.temperature ?? 0) > 25 ? "text-amber-400" : "text-sky-400"}`} />
          <span className={`text-lg font-bold ${(log.temperature ?? 0) > 25 ? "text-amber-400" : "text-sky-400"}`}>
            {log.temperature}°C
          </span>
        </div>
        <p className="text-[10px] text-slate-400">
          {log.temperature && log.temperature > 25 ? "Outside cleaning" : "Inside cleaning"}
        </p>
      </div>

      <div className="rounded-lg bg-slate-800 p-3">
        <div className="flex items-center gap-2">
          <ArrowUp className="h-4 w-4 text-violet-400" />
          <span className="text-lg font-bold text-violet-400">
            Floor {log.floor}
          </span>
        </div>
        <p className="text-[10px] text-slate-400">
          Alt: {Math.round(log.altitude)}m
        </p>
      </div>

      <div className="rounded-lg bg-slate-800 p-3">
        <div className="flex items-center gap-2">
          <MapPin className="h-4 w-4 text-emerald-400" />
          <span className="text-xs text-emerald-400">
            {log.location_type}
          </span>
        </div>
        <p className="text-[10px] text-slate-400">
          {log.latitude.toFixed(4)}, {log.longitude.toFixed(4)}
        </p>
      </div>

      <div className="rounded-lg bg-slate-800 p-3">
        <div className="flex items-center gap-2">
          <Battery className="h-4 w-4 text-slate-400" />
          <span className="text-xs text-white">{log.battery_level}%</span>
        </div>
        <p className="text-[10px] text-slate-400">
          {new Date(log.recorded_at).toLocaleTimeString()}
        </p>
      </div>
    </div>
  );
}

function LogCard({ log }: { log: CleanerSensorLog }) {
  return (
    <div className="flex items-center gap-3 rounded-lg bg-slate-900 p-3">
      <div
        className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-full ${
          log.temperature > 25 ? "bg-amber-500/10" : "bg-sky-500/10"
        }`}
      >
        <Thermometer
          className={`h-5 w-5 ${log.temperature > 25 ? "text-amber-400" : "text-sky-400"}`}
        />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className="text-sm font-medium text-white">{log.cleaner_name}</span>
          <span className="text-[10px] text-slate-500">
            Floor {log.floor}
          </span>
        </div>
        <p className="text-xs text-slate-400 truncate">
          {log.temperature > 25 ? "Outside" : "Inside"} · {log.location_type} ·{" "}
          {new Date(log.recorded_at).toLocaleTimeString()}
        </p>
      </div>
      <div className="text-xs text-slate-500">
        {Math.round(log.altitude)}m
      </div>
    </div>
  );
}
