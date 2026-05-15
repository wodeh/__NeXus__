// @ts-nocheck
"use client";

import { useState, useEffect } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import {
  Lock,
  Unlock,
  KeyRound,
  Plus,
  Search,
  Battery,
  BatteryWarning,
  BatteryMedium,
  Wifi,
  WifiOff,
  ShieldCheck,
  ShieldAlert,
  Eye,
  Trash2,
  RotateCcw,
  DoorOpen,
  ArrowRightLeft,
  Clock,
  AlertTriangle,
  Activity,
} from "lucide-react";
import {
  getLocks,
  getLockEvents,
  getLockAccessCodes,
  remoteUnlock,
  SmartLock,
  LockEvent,
  AccessCode,
} from "@/lib/api";

function ToggleSwitch({ checked }: { checked: boolean }) {
  return (
    <div className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors ${checked ? "bg-nexus-500" : "bg-slate-700"}`}>
      <span className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform ${checked ? "translate-x-4" : "translate-x-1"}`} />
    </div>
  );
}

export default function SmartLocksPage() {
  const { config } = useTenant();
  const [search, setSearch] = useState("");
  const [selectedLock, setSelectedLock] = useState<SmartLock | null>(null);
  const [tab, setTab] = useState<"overview" | "locks" | "codes" | "events">("overview");
  const [locks, setLocks] = useState<SmartLock[]>([]);
  const [events, setEvents] = useState<LockEvent[]>([]);
  const [accessCodes, setAccessCodes] = useState<AccessCode[]>([]);
  const [loading, setLoading] = useState(false);
  const [unlocking, setUnlocking] = useState(false);

  useEffect(() => {
    loadLocks();
  }, []);

  useEffect(() => {
    if (selectedLock) {
      loadLockData(selectedLock.id);
    }
  }, [selectedLock]);

  async function loadLocks() {
    try {
      const data = await getLocks();
      setLocks(data);
    } catch (e) {
      console.error("Failed to load locks", e);
    }
  }

  async function loadLockData(lockId: string) {
    setLoading(true);
    try {
      const [evts, codes] = await Promise.all([
        getLockEvents(lockId),
        getLockAccessCodes(lockId),
      ]);
      setEvents(evts);
      setAccessCodes(codes);
    } catch (e) {
      console.error("Failed to load lock data", e);
    } finally {
      setLoading(false);
    }
  }

  async function handleRemoteUnlock(lockId: string) {
    setUnlocking(true);
    try {
      await remoteUnlock(lockId);
      await loadLocks();
      if (selectedLock?.id === lockId) {
        await loadLockData(lockId);
      }
    } catch (e) {
      console.error("Failed to unlock", e);
    } finally {
      setUnlocking(false);
    }
  }

  const hasLocks = hasCapability(config, CAPABILITIES.OPERATIONS.SMART_LOCKS);

  if (!hasLocks) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-800">
          <Lock className="h-8 w-8 text-slate-500" />
        </div>
        <h2 className="mt-6 text-xl font-bold text-white">Smart Locks</h2>
        <p className="mt-2 max-w-md text-sm text-slate-400">
          Smart lock management requires Operations tier.
        </p>
      </div>
    );
  }

  const filteredLocks = locks.filter((l) =>
    l.room_number.toLowerCase().includes(search.toLowerCase()) ||
    l.serial_number.toLowerCase().includes(search.toLowerCase())
  );

  const onlineCount = locks.filter((l) => l.status === "online").length;
  const offlineCount = locks.filter((l) => l.status === "offline").length;
  const lowBatteryCount = locks.filter((l) => l.status === "low_battery").length;
  const warningCount = locks.filter((l) => l.status === "warning").length;
  const avgBattery = locks.length > 0
    ? Math.round(locks.reduce((s, l) => s + l.battery_level, 0) / locks.length)
    : 0;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Smart Locks</h2>
          <p className="text-sm text-slate-400">OrbitaTech OT-SL300 · Remote management</p>
        </div>
        <div className="flex gap-2">
          <button className="btn-secondary">
            <RotateCcw className="h-4 w-4" /> Sync All
          </button>
          <button className="btn-primary">
            <Plus className="h-4 w-4" /> Add Lock
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-5 gap-4">
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400">
            <Wifi className="h-4 w-4 text-emerald-400" /> Online
          </div>
          <p className="text-2xl font-bold text-white">{onlineCount}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400">
            <WifiOff className="h-4 w-4 text-rose-400" /> Offline
          </div>
          <p className="text-2xl font-bold text-white">{offlineCount}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400">
            <BatteryWarning className="h-4 w-4 text-amber-400" /> Low Battery
          </div>
          <p className="text-2xl font-bold text-white">{lowBatteryCount}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400">
            <AlertTriangle className="h-4 w-4 text-nexus-400" /> Warnings
          </div>
          <p className="text-2xl font-bold text-white">{warningCount}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400">
            <Battery className="h-4 w-4" /> Avg Battery
          </div>
          <p className="text-2xl font-bold text-white">{avgBattery}%</p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        {(["overview", "locks", "codes", "events"] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`rounded-lg px-4 py-2 text-sm font-medium ${tab === t ? "bg-slate-800 text-white" : "text-slate-400"}`}
          >
            {t === "overview" ? "Overview" : t === "locks" ? "All Locks" : t === "codes" ? "Access Codes" : "Event Log"}
          </button>
        ))}
      </div>

      {/* Overview */}
      {tab === "overview" && (
        <div className="grid grid-cols-3 gap-4">
          {filteredLocks.map((lock) => (
            <div key={lock.id} className="card space-y-3">
              <div className="flex items-start justify-between">
                <div className="flex items-center gap-3">
                  <div className={`flex h-10 w-10 items-center justify-center rounded-full ${
                    lock.status === "online" ? "bg-emerald-500/10" :
                    lock.status === "offline" ? "bg-rose-500/10" :
                    lock.status === "low_battery" ? "bg-amber-500/10" :
                    "bg-nexus-500/10"
                  }`}>
                    {lock.status === "online" ? <Wifi className="h-5 w-5 text-emerald-400" /> :
                     lock.status === "offline" ? <WifiOff className="h-5 w-5 text-rose-400" /> :
                     lock.status === "low_battery" ? <BatteryWarning className="h-5 w-5 text-amber-400" /> :
                     <ShieldAlert className="h-5 w-5 text-nexus-400" />}
                  </div>
                  <div>
                    <p className="text-sm font-medium text-white">Room {lock.room_number}</p>
                    <p className="text-xs text-slate-400">{lock.serial_number}</p>
                  </div>
                </div>
                <span className={`rounded px-2 py-0.5 text-[10px] ${
                  lock.status === "online" ? "bg-emerald-500/10 text-emerald-400" :
                  lock.status === "offline" ? "bg-rose-500/10 text-rose-400" :
                  lock.status === "low_battery" ? "bg-amber-500/10 text-amber-400" :
                  "bg-nexus-500/10 text-nexus-400"
                }`}>{lock.status}</span>
              </div>

              <div className="space-y-2">
                <div className="flex items-center justify-between text-xs">
                  <span className="text-slate-400">Battery</span>
                  <span className={`font-medium ${lock.battery_level < 20 ? "text-amber-400" : "text-white"}`}>
                    {lock.battery_level}%
                  </span>
                </div>
                <div className="h-1.5 rounded-full bg-slate-800">
                  <div className={`h-full rounded-full ${lock.battery_level < 20 ? "bg-amber-400" : "bg-emerald-400"}`}
                    style={{ width: `${lock.battery_level}%` }} />
                </div>
              </div>

              <div className="flex items-center justify-between text-xs text-slate-400">
                <span>Firmware: {lock.firmware_version}</span>
                <span>{lock.last_communication_at ? `Last comm: ${lock.last_communication_at}` : "No data"}</span>
              </div>

              <div className="flex gap-2">
                <button
                  onClick={() => handleRemoteUnlock(lock.id)}
                  disabled={unlocking || lock.status !== "online"}
                  className="btn-primary flex-1 gap-1 text-xs"
                >
                  <Unlock className="h-3 w-3" />
                  {unlocking ? "Unlocking..." : "Remote Unlock"}
                </button>
                <button onClick={() => { setSelectedLock(lock); setTab("codes"); }} className="btn-secondary gap-1 text-xs">
                  <KeyRound className="h-3 w-3" /> Codes
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Locks Table */}
      {tab === "locks" && (
        <div className="card">
          <div className="relative mb-4">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
            <input className="input w-full max-w-md pl-9" placeholder="Search locks..." value={search} onChange={(e) => setSearch(e.target.value)} />
          </div>
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-800 text-xs uppercase text-slate-400">
              <tr>
                <th className="px-4 py-3">Room</th>
                <th className="px-4 py-3">Serial</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Battery</th>
                <th className="px-4 py-3">Last Unlock</th>
                <th className="px-4 py-3">Remote</th>
                <th className="px-4 py-3">Auto-lock</th>
                <th className="px-4 py-3">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {filteredLocks.map((lock) => (
                <tr key={lock.id} className="hover:bg-slate-800/50">
                  <td className="px-4 py-3 text-white">{lock.room_number}</td>
                  <td className="px-4 py-3 text-slate-400">{lock.serial_number}</td>
                  <td className="px-4 py-3">
                    <span className={`rounded px-2 py-0.5 text-xs ${
                      lock.status === "online" ? "bg-emerald-500/10 text-emerald-400" :
                      lock.status === "offline" ? "bg-rose-500/10 text-rose-400" :
                      lock.status === "low_battery" ? "bg-amber-500/10 text-amber-400" :
                      "bg-nexus-500/10 text-nexus-400"
                    }`}>{lock.status}</span>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      {lock.battery_level < 20 ? <BatteryWarning className="h-4 w-4 text-amber-400" /> :
                       lock.battery_level < 50 ? <BatteryMedium className="h-4 w-4 text-slate-400" /> :
                       <Battery className="h-4 w-4 text-emerald-400" />}
                      <span className={`${lock.battery_level < 20 ? "text-amber-400" : "text-white"}`}>{lock.battery_level}%</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-slate-400">{lock.last_unlock_at || "—"}</td>
                  <td className="px-4 py-3"><ToggleSwitch checked={lock.remote_unlock_enabled} /></td>
                  <td className="px-4 py-3"><ToggleSwitch checked={lock.auto_lock_enabled} /></td>
                  <td className="px-4 py-3">
                    <div className="flex gap-2">
                      <button onClick={() => { setSelectedLock(lock); setTab("events"); }} className="text-slate-400 hover:text-white">
                        <Activity className="h-4 w-4" />
                      </button>
                      <button onClick={() => handleRemoteUnlock(lock.id)} disabled={unlocking} className="text-slate-400 hover:text-emerald-400">
                        <Unlock className="h-4 w-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Access Codes */}
      {tab === "codes" && (
        <div className="space-y-4">
          {selectedLock && (
            <div className="flex items-center justify-between">
              <p className="text-sm text-slate-400">Showing codes for <span className="text-white">Room {selectedLock.room_number}</span></p>
              <button className="btn-primary gap-1 text-xs"><Plus className="h-3 w-3" /> Add Code</button>
            </div>
          )}
          <div className="card">
            {loading ? (
              <p className="text-sm text-slate-500">Loading access codes...</p>
            ) : accessCodes.length === 0 ? (
              <p className="text-sm text-slate-500">No access codes found. Select a lock or add a new code.</p>
            ) : (
              <table className="w-full text-left text-sm">
                <thead className="bg-slate-800 text-xs uppercase text-slate-400">
                  <tr>
                    <th className="px-4 py-3">Code</th>
                    <th className="px-4 py-3">Label</th>
                    <th className="px-4 py-3">Status</th>
                    <th className="px-4 py-3">Valid From</th>
                    <th className="px-4 py-3">Valid Until</th>
                    <th className="px-4 py-3">Uses</th>
                    <th className="px-4 py-3">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800">
                  {accessCodes.map((code) => (
                    <tr key={code.id} className="hover:bg-slate-800/50">
                      <td className="px-4 py-3 font-mono text-white">{code.code}</td>
                      <td className="px-4 py-3 text-white">{code.label}</td>
                      <td className="px-4 py-3">
                        <span className={`rounded px-2 py-0.5 text-xs ${code.is_active ? "bg-emerald-500/10 text-emerald-400" : "bg-slate-500/10 text-slate-400"}`}>
                          {code.is_active ? "Active" : "Inactive"}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-slate-400">{code.valid_from}</td>
                      <td className="px-4 py-3 text-slate-400">{code.valid_until || "—"}</td>
                      <td className="px-4 py-3 text-slate-400">
                        {code.max_uses ? `${code.use_count}/${code.max_uses}` : `${code.use_count}`}
                      </td>
                      <td className="px-4 py-3">
                        <div className="flex gap-2">
                          <button className="text-slate-400 hover:text-white"><Eye className="h-4 w-4" /></button>
                          <button className="text-slate-400 hover:text-rose-400"><Trash2 className="h-4 w-4" /></button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}

      {/* Events */}
      {tab === "events" && (
        <div className="space-y-4">
          {selectedLock && (
            <div className="flex items-center justify-between">
              <p className="text-sm text-slate-400">Showing events for <span className="text-white">Room {selectedLock.room_number}</span></p>
            </div>
          )}
          <div className="card">
            {loading ? (
              <p className="text-sm text-slate-500">Loading events...</p>
            ) : events.length === 0 ? (
              <p className="text-sm text-slate-500">No events found. Select a lock or wait for new activity.</p>
            ) : (
              <table className="w-full text-left text-sm">
                <thead className="bg-slate-800 text-xs uppercase text-slate-400">
                  <tr>
                    <th className="px-4 py-3">Time</th>
                    <th className="px-4 py-3">Event</th>
                    <th className="px-4 py-3">Source</th>
                    <th className="px-4 py-3">Details</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800">
                  {events.map((evt) => (
                    <tr key={evt.id} className="hover:bg-slate-800/50">
                      <td className="px-4 py-3 text-slate-400">{evt.occurred_at}</td>
                      <td className="px-4 py-3">
                        <span className={`rounded px-2 py-0.5 text-xs ${
                          evt.event_type === "unlock" ? "bg-emerald-500/10 text-emerald-400" :
                          evt.event_type === "lock" ? "bg-slate-500/10 text-slate-400" :
                          evt.event_type === "alert" ? "bg-rose-500/10 text-rose-400" :
                          "bg-nexus-500/10 text-nexus-400"
                        }`}>{evt.event_type}</span>
                      </td>
                      <td className="px-4 py-3 text-white">{evt.event_source}</td>
                      <td className="px-4 py-3 text-slate-400">{evt.details}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
