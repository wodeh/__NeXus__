"use client";

import { useState, useEffect, useCallback } from "react";
import {
  Smartphone, MapPin, Thermometer, ArrowUp, Battery, RefreshCw,
  CheckCircle2, AlertTriangle, Clock, Home, Wifi, WifiOff,
} from "lucide-react";
import {
  getVillas, getSensorLogs, recordSensorLog, getLatestSensorLog,
  VillaProperty, CleanerSensorLog,
} from "@/lib/api";

export default function CleanerTrackingPage() {
  const [villas, setVillas] = useState<VillaProperty[]>([]);
  const [selectedVilla, setSelectedVilla] = useState<string>("");
  const [logs, setLogs] = useState<CleanerSensorLog[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [tracking, setTracking] = useState(false);
  const [cleanerName, setCleanerName] = useState("");
  const [cleanerId, setCleanerId] = useState("");

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

    // Start geolocation + sensor simulation
    if (!navigator.geolocation) {
      alert("Geolocation not supported");
      return;
    }

    const villa = villas.find((v) => v.id === selectedVilla);
    if (!villa) return;

    const watchId = navigator.geolocation.watchPosition(
      async (position) => {
        const { latitude, longitude, altitude } = position.coords;
        // Derive floor from altitude difference
        const floor = altitude ? Math.round((altitude - villa.elevation) / 3.5) : 0;
        // Mock temperature based on altitude (simpler for demo)
        const temp = altitude && altitude > villa.elevation + 2 ? 31.5 : 23.0;
        const locationType = temp > 25 ? "outside" : floor === 0 ? "inside" : floor > 0 ? "inside" : "outside";

        const payload = {
          villa_id: selectedVilla,
          villa_name: villa.name,
          cleaner_id: id,
          cleaner_name: cleanerName,
          temperature: temp,
          latitude,
          longitude,
          altitude: altitude || villa.elevation,
          floor: Math.max(0, floor),
          location_type: locationType,
          battery_level: 85,
          recorded_at: new Date().toISOString(),
        };

        try {
          await recordSensorLog(payload);
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

    // Stop after 8 hours (work shift)
    setTimeout(() => {
      navigator.geolocation.clearWatch(watchId);
      setTracking(false);
    }, 8 * 60 * 60 * 1000);

    return () => navigator.geolocation.clearWatch(watchId);
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
      {/* Mobile Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-bold text-white flex items-center gap-2">
          <Smartphone className="h-6 w-6 text-nexus-400" />
          Cleaner Tracker
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

      {/* Cleaner Info */}
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

      {/* Active Tracking */}
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

      {/* Logs History */}
      <div className="space-y-3">
        <p className="text-xs font-medium text-slate-400 uppercase">Session History</p>
        {logs.map((log) => (
          <LogCard key={log.id} log={log} />
        ))}
        {logs.length === 0 && (
          <p className="text-center text-sm text-slate-500 py-8">No tracking data yet</p>
        )}
      </div>
    </div>
  );
}

function LatestPositionCard({ log, villa }: { log: CleanerSensorLog; villa?: VillaProperty }) {
  return (
    <div className="grid grid-cols-2 gap-3">
      <div className={`rounded-lg p-3 ${log.temperature > 25 ? "bg-amber-500/10" : "bg-sky-500/10"}`}>
        <div className="flex items-center gap-2">
          <Thermometer className={`h-4 w-4 ${log.temperature > 25 ? "text-amber-400" : "text-sky-400"}`} />
          <span className={`text-lg font-bold ${log.temperature > 25 ? "text-amber-400" : "text-sky-400"}`}>
            {log.temperature}°C
          </span>
        </div>
        <p className="text-[10px] text-slate-400">
          {log.temperature > 25 ? "Outside cleaning" : "Inside cleaning"}
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
