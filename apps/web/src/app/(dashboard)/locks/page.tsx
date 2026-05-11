"use client";

import { useState } from "react";
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

interface SmartLock {
  id: string;
  roomId: string;
  roomNumber: string;
  serialNumber: string;
  model: string;
  manufacturer: string;
  status: "online" | "offline" | "low_battery" | "warning";
  batteryLevel: number;
  lastCommunicationAt?: string;
  lastUnlockAt?: string;
  lastLockAt?: string;
  firmwareVersion: string;
  remoteUnlockEnabled: boolean;
  autoLockEnabled: boolean;
  accessCodeCount: number;
}

interface LockEvent {
  id: string;
  eventType: string;
  eventSource: string;
  details: string;
  occurredAt: string;
}

interface AccessCode {
  id: string;
  code: string;
  label: string;
  isActive: boolean;
  validFrom: string;
  validUntil: string;
  maxUses?: number;
  useCount: number;
}

const mockLocks: SmartLock[] = [
  { id: "l-101", roomId: "r-101", roomNumber: "101", serialNumber: "OT-101-2026", model: "OT-SL300", manufacturer: "OrbitaTech", status: "online", batteryLevel: 87, lastCommunicationAt: "2026-05-10 08:45", lastUnlockAt: "2026-05-10 07:30", firmwareVersion: "3.2.1", remoteUnlockEnabled: true, autoLockEnabled: true, accessCodeCount: 3 },
  { id: "l-102", roomId: "r-102", roomNumber: "102", serialNumber: "OT-102-2026", model: "OT-SL300", manufacturer: "OrbitaTech", status: "online", batteryLevel: 92, lastCommunicationAt: "2026-05-10 09:00", firmwareVersion: "3.2.1", remoteUnlockEnabled: true, autoLockEnabled: true, accessCodeCount: 0 },
  { id: "l-103", roomId: "r-103", roomNumber: "103", serialNumber: "OT-103-2026", model: "OT-SL300", manufacturer: "OrbitaTech", status: "low_battery", batteryLevel: 12, lastCommunicationAt: "2026-05-10 06:00", firmwareVersion: "3.1.0", remoteUnlockEnabled: true, autoLockEnabled: true, accessCodeCount: 1 },
  { id: "l-201", roomId: "r-201", roomNumber: "201", serialNumber: "OT-201-2026", model: "OT-SL300", manufacturer: "OrbitaTech", status: "online", batteryLevel: 76, lastCommunicationAt: "2026-05-10 09:15", lastUnlockAt: "2026-05-10 08:00", firmwareVersion: "3.2.1", remoteUnlockEnabled: true, autoLockEnabled: false, accessCodeCount: 2 },
  { id: "l-202", roomId: "r-202", roomNumber: "202", serialNumber: "OT-202-2026", model: "OT-SL300", manufacturer: "OrbitaTech", status: "online", batteryLevel: 88, lastCommunicationAt: "2026-05-10 09:20", firmwareVersion: "3.2.1", remoteUnlockEnabled: true, autoLockEnabled: true, accessCodeCount: 0 },
  { id: "l-203", roomId: "r-203", roomNumber: "203", serialNumber: "OT-203-2026", model: "OT-SL300", manufacturer: "OrbitaTech", status: "warning", batteryLevel: 45, lastCommunicationAt: "2026-05-10 04:00", lastUnlockAt: "2026-05-10 03:00", firmwareVersion: "3.1.0", remoteUnlockEnabled: false, autoLockEnabled: true, accessCodeCount: 4 },
  { id: "l-301", roomId: "r-301", roomNumber: "301", serialNumber: "OT-301-2026", model: "OT-SL300", manufacturer: "OrbitaTech", status: "online", batteryLevel: 91, lastCommunicationAt: "2026-05-10 09:10", firmwareVersion: "3.2.1", remoteUnlockEnabled: true, autoLockEnabled: true, accessCodeCount: 0 },
  { id: "l-302", roomId: "r-302", roomNumber: "302", serialNumber: "OT-302-2026", model: "OT-SL300", manufacturer: "OrbitaTech", status: "offline", batteryLevel: 0, lastCommunicationAt: "2026-05-09 14:00", firmwareVersion: "3.0.0", remoteUnlockEnabled: false, autoLockEnabled: true, accessCodeCount: 0 },
];

const mockEvents: Record<string, LockEvent[]> = {
  "l-101": [
    { id: "e1", eventType: "unlock", eventSource: "guest_keycard", details: "Room 101 - Guest keycard", occurredAt: "2026-05-10 07:30" },
    { id: "e2", eventType: "lock", eventSource: "auto_lock", details: "Auto-locked after 30s", occurredAt: "2026-05-10 07:31" },
    { id: "e3", eventType: "unlock", eventSource: "staff_master", details: "Housekeeping master key", occurredAt: "2026-05-10 08:45" },
    { id: "e4", eventType: "lock", eventSource: "staff_master", details: "Housekeeping master key", occurredAt: "2026-05-10 09:00" },
  ],
  "l-203": [
    { id: "e1", eventType: "unlock", eventSource: "guest_keycard", details: "Room 203 - Guest keycard", occurredAt: "2026-05-10 03:00" },
    { id: "e2", eventType: "lock", eventSource: "auto_lock", details: "Auto-locked after 30s", occurredAt: "2026-05-10 03:01" },
  ],
};

const mockAccessCodes: Record<string, AccessCode[]> = {
  "l-101": [
    { id: "ac1", code: "3749", label: "Guest Primary", isActive: true, validFrom: "2026-05-10", validUntil: "2026-05-14", useCount: 3 },
    { id: "ac2", code: "2847", label: "Housekeeping", isActive: true, validFrom: "2026-05-10", validUntil: "2026-05-10", maxUses: 10, useCount: 2 },
    { id: "ac3", code: "9102", label: "Maintenance", isActive: false, validFrom: "2026-05-10", validUntil: "2026-05-10", useCount: 0 },
  ],
};

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

  const hasLocks = hasCapability(config, CAPABILITIES.OPERATIONS.SMART_LOCKS);

  if (!hasLocks) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-800">
          <Lock className="h-8 w-8 text-slate-500" />
        </div>
        <h2 className="mt-6 text-xl font-bold text-white">Smart Locks</h2>
        <p className="mt-2 max-w-md text-sm text-slate-400">
          Upgrade to Operations or Enterprise tier to manage smart lock access.
        </p>
      </div>
    );
  }

  const filteredLocks = mockLocks.filter((l) =>
    l.roomNumber.includes(search) ||
    l.serialNumber.toLowerCase().includes(search.toLowerCase())
  );

  const stats = {
    total: mockLocks.length,
    online: mockLocks.filter((l) => l.status === "online").length,
    offline: mockLocks.filter((l) => l.status === "offline").length,
    lowBattery: mockLocks.filter((l) => l.batteryLevel < 20).length,
    warning: mockLocks.filter((l) => l.status === "warning").length,
    activeCodes: mockLocks.reduce((sum, l) => sum + l.accessCodeCount, 0),
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Smart Locks</h1>
          <p className="text-sm text-slate-400">OrbitaTech lock management · Remote control · Access codes</p>
        </div>
        <div className="flex items-center gap-2">
          <button className="btn-secondary gap-2 text-xs">
            <RotateCcw className="h-4 w-4" />
            Sync All
          </button>
          <button className="btn-primary gap-2 text-xs">
            <Plus className="h-4 w-4" />
            Add Lock
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-5 gap-3">
        <StatCard label="Total Locks" value={stats.total} icon={Lock} color="sky" />
        <StatCard label="Online" value={stats.online} icon={Wifi} color="emerald" />
        <StatCard label="Offline" value={stats.offline} icon={WifiOff} color="rose" />
        <StatCard label="Low Battery" value={stats.lowBattery} icon={BatteryWarning} color="amber" alert />
        <StatCard label="Active Codes" value={stats.activeCodes} icon={KeyRound} color="violet" />
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        {(["overview", "locks", "codes", "events"] as const).map((t) => (
          <button
            key={t}
            onClick={() => { setTab(t); setSelectedLock(null); }}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              tab === t ? "bg-slate-800 text-white" : "text-slate-400 hover:bg-slate-800 hover:text-white"
            }`}
          >
            {t === "overview" ? "Overview" : t === "locks" ? "All Locks" : t === "codes" ? "Access Codes" : "Events"}
          </button>
        ))}
      </div>

      {tab === "overview" && <OverviewTab locks={mockLocks} onSelect={setSelectedLock} />}
      {tab === "locks" && (
        <>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
            <input
              placeholder="Search by room or serial number..."
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              className="input w-full pl-10"
            />
          </div>
          <div className="grid grid-cols-4 gap-4">
            {filteredLocks.map((lock) => (
              <LockCard key={lock.id} lock={lock} onClick={() => setSelectedLock(lock)} isSelected={selectedLock?.id === lock.id} />
            ))}
          </div>
          {selectedLock && <LockDetailPanel lock={selectedLock} />}
        </>
      )}
      {tab === "codes" && <AccessCodesTab locks={mockLocks} />}
      {tab === "events" && <EventsTab locks={mockLocks} />}
    </div>
  );
}

function StatCard({ label, value, icon: Icon, color, alert }: { label: string; value: number; icon: React.ElementType; color: string; alert?: boolean }) {
  const colorMap: Record<string, string> = {
    sky: "bg-sky-500/10 text-sky-400 border-sky-500/20",
    emerald: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
    rose: "bg-rose-500/10 text-rose-400 border-rose-500/20",
    amber: "bg-amber-500/10 text-amber-400 border-amber-500/20",
    violet: "bg-violet-500/10 text-violet-400 border-violet-500/20",
  };
  return (
    <div className={`rounded-lg border p-3 ${colorMap[color]} ${alert ? "animate-pulse" : ""}`}>
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium opacity-80">{label}</span>
        <Icon className="h-4 w-4 opacity-60" />
      </div>
      <p className="mt-1 text-2xl font-bold">{value}</p>
    </div>
  );
}

function LockCard({ lock, onClick, isSelected }: { lock: SmartLock; onClick: () => void; isSelected: boolean }) {
  const statusColor: Record<string, string> = {
    online: "text-emerald-400",
    offline: "text-rose-400",
    low_battery: "text-amber-400",
    warning: "text-amber-400",
  };

  return (
    <button
      onClick={onClick}
      className={`rounded-lg border p-4 text-left transition-all hover:scale-[1.02] ${
        isSelected ? "border-nexus-400 ring-1 ring-nexus-400" : "border-slate-700 bg-slate-800/50"
      }`}
    >
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <DoorOpen className="h-5 w-5 text-slate-400" />
          <span className="text-lg font-bold text-white">Room {lock.roomNumber}</span>
        </div>
        {lock.status === "online" ? (
          <Wifi className="h-4 w-4 text-emerald-400" />
        ) : (
          <WifiOff className={`h-4 w-4 ${statusColor[lock.status]}`} />
        )}
      </div>
      <p className="mt-1 text-xs text-slate-500">{lock.serialNumber}</p>
      <div className="mt-2 flex items-center gap-2">
        {lock.batteryLevel > 50 ? (
          <Battery className="h-3 w-3 text-emerald-400" />
        ) : lock.batteryLevel > 20 ? (
          <BatteryMedium className="h-3 w-3 text-amber-400" />
        ) : (
          <BatteryWarning className="h-3 w-3 text-rose-400" />
        )}
        <span className={`text-xs ${statusColor[lock.status]}`}>
          {lock.batteryLevel}%
        </span>
        <span className="ml-auto text-[10px] text-slate-500">{lock.firmwareVersion}</span>
      </div>
      <div className="mt-2 flex items-center gap-2">
        <span className={`rounded px-1.5 py-0.5 text-[10px] font-medium ${statusColor[lock.status]} bg-slate-700/50`}>
          {lock.status.replace("_", " ").toUpperCase()}
        </span>
        {lock.accessCodeCount > 0 && (
          <span className="rounded px-1.5 py-0.5 text-[10px] font-medium text-violet-400 bg-slate-700/50">
            {lock.accessCodeCount} codes
          </span>
        )}
      </div>
    </button>
  );
}

function LockDetailPanel({ lock }: { lock: SmartLock }) {
  return (
    <div className="mt-4 rounded-lg border border-slate-700 bg-slate-800/50 p-4">
      <div className="flex items-center justify-between">
        <h3 className="text-lg font-bold text-white">Room {lock.roomNumber} Details</h3>
        <div className="flex items-center gap-2">
          {lock.remoteUnlockEnabled && (
            <button className="btn-primary gap-2 text-xs bg-amber-500 hover:bg-amber-600">
              <Unlock className="h-4 w-4" />
              Unlock
            </button>
          )}
          <button className="btn-secondary gap-2 text-xs">
            <KeyRound className="h-4 w-4" />
            New Code
          </button>
        </div>
      </div>

      <div className="mt-4 grid grid-cols-3 gap-4">
        <div className="rounded bg-slate-700/30 p-3">
          <p className="text-xs text-slate-400">Serial Number</p>
          <p className="text-sm font-medium text-white">{lock.serialNumber}</p>
        </div>
        <div className="rounded bg-slate-700/30 p-3">
          <p className="text-xs text-slate-400">Model</p>
          <p className="text-sm font-medium text-white">{lock.model}</p>
        </div>
        <div className="rounded bg-slate-700/30 p-3">
          <p className="text-xs text-slate-400">Firmware</p>
          <p className="text-sm font-medium text-white">{lock.firmwareVersion}</p>
        </div>
        <div className="rounded bg-slate-700/30 p-3">
          <p className="text-xs text-slate-400">Last Communication</p>
          <p className="text-sm font-medium text-white">{lock.lastCommunicationAt || "Never"}</p>
        </div>
        <div className="rounded bg-slate-700/30 p-3">
          <p className="text-xs text-slate-400">Last Unlock</p>
          <p className="text-sm font-medium text-white">{lock.lastUnlockAt || "Never"}</p>
        </div>
        <div className="rounded bg-slate-700/30 p-3">
          <p className="text-xs text-slate-400">Last Lock</p>
          <p className="text-sm font-medium text-white">{lock.lastLockAt || "Never"}</p>
        </div>
      </div>

      <div className="mt-4 flex items-center gap-6">
        <div className="flex items-center gap-2">
          <ToggleSwitch checked={lock.remoteUnlockEnabled} />
          <span className="text-xs text-slate-300">Remote Unlock</span>
        </div>
        <div className="flex items-center gap-2">
          <ToggleSwitch checked={lock.autoLockEnabled} />
          <span className="text-xs text-slate-300">Auto-Lock</span>
        </div>
      </div>
    </div>
  );
}

function OverviewTab({ locks, onSelect }: { locks: SmartLock[]; onSelect: (l: SmartLock) => void }) {
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="card space-y-2">
          <p className="text-sm text-slate-400">Recent Activity</p>
          <div className="space-y-2">
            {locks.slice(0, 5).map((l) => (
              <button key={l.id} onClick={() => onSelect(l)} className="flex w-full items-center justify-between rounded bg-slate-700/30 px-3 py-2 text-left hover:bg-slate-700/50">
                <div className="flex items-center gap-2">
                  <DoorOpen className="h-4 w-4 text-slate-400" />
                  <span className="text-sm text-white">Room {l.roomNumber}</span>
                </div>
                <span className="text-xs text-slate-400">{l.lastUnlockAt || "No recent activity"}</span>
              </button>
            ))}
          </div>
        </div>
        <div className="card space-y-2">
          <p className="text-sm text-slate-400">Battery Overview</p>
          <div className="space-y-2">
            {locks.map((l) => (
              <div key={l.id} className="flex items-center gap-3">
                <span className="w-12 text-xs text-slate-400">{l.roomNumber}</span>
                <div className="flex-1 h-2 rounded-full bg-slate-700">
                  <div
                    className={`h-full rounded-full ${l.batteryLevel > 50 ? "bg-emerald-400" : l.batteryLevel > 20 ? "bg-amber-400" : "bg-rose-400"}`}
                    style={{ width: `${l.batteryLevel}%` }}
                  />
                </div>
                <span className={`w-8 text-right text-xs ${l.batteryLevel > 20 ? "text-slate-400" : "text-rose-400"}`}>{l.batteryLevel}%</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}

function AccessCodesTab({ locks }: { locks: SmartLock[] }) {
  const [selectedLockId, setSelectedLockId] = useState(locks[0]?.id || "");
  const codes = mockAccessCodes[selectedLockId] || [];

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <span className="text-sm text-slate-400">Room:</span>
        <div className="flex gap-1">
          {locks.map((l) => (
            <button
              key={l.id}
              onClick={() => setSelectedLockId(l.id)}
              className={`rounded px-3 py-1 text-xs font-medium ${
                selectedLockId === l.id ? "bg-nexus-500 text-white" : "bg-slate-800 text-slate-400 hover:text-white"
              }`}
            >
              {l.roomNumber}
            </button>
          ))}
        </div>
      </div>

      <div className="space-y-2">
        {codes.length === 0 ? (
          <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-8 text-center">
            <KeyRound className="mx-auto h-8 w-8 text-slate-500" />
            <p className="mt-2 text-sm text-slate-400">No access codes for this room</p>
            <button className="btn-primary mt-3 gap-2 text-xs">
              <Plus className="h-4 w-4" />
              Create Code
            </button>
          </div>
        ) : (
          codes.map((code) => (
            <div key={code.id} className="flex items-center justify-between rounded-lg border border-slate-700 bg-slate-800/50 p-4">
              <div className="flex items-center gap-4">
                <div className="flex h-12 w-16 items-center justify-center rounded bg-slate-700 text-lg font-bold tracking-widest text-white">
                  {code.code}
                </div>
                <div>
                  <p className="text-sm font-medium text-white">{code.label}</p>
                  <p className="text-xs text-slate-400">{code.validFrom} &rarr; {code.validUntil} · Used {code.useCount} {code.maxUses && `/ ${code.maxUses}`} times</p>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <ToggleSwitch checked={code.isActive} />
                <button className="rounded p-1 text-slate-400 hover:text-white"><Eye className="h-4 w-4" /></button>
                <button className="rounded p-1 text-rose-400 hover:text-rose-300"><Trash2 className="h-4 w-4" /></button>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}

function EventsTab({ locks }: { locks: SmartLock[] }) {
  const [selectedLockId, setSelectedLockId] = useState(locks[0]?.id || "");
  const events = mockEvents[selectedLockId] || [];

  const eventIcon: Record<string, React.ElementType> = {
    unlock: Unlock,
    lock: Lock,
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <span className="text-sm text-slate-400">Room:</span>
        <div className="flex gap-1">
          {locks.map((l) => (
            <button
              key={l.id}
              onClick={() => setSelectedLockId(l.id)}
              className={`rounded px-3 py-1 text-xs font-medium ${
                selectedLockId === l.id ? "bg-nexus-500 text-white" : "bg-slate-800 text-slate-400 hover:text-white"
              }`}
            >
              {l.roomNumber}
            </button>
          ))}
        </div>
      </div>

      <div className="space-y-2">
        {events.length === 0 ? (
          <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-8 text-center">
            <Activity className="mx-auto h-8 w-8 text-slate-500" />
            <p className="mt-2 text-sm text-slate-400">No events for this room</p>
          </div>
        ) : (
          events.map((e) => {
            const Icon = eventIcon[e.eventType] || Activity;
            return (
              <div key={e.id} className="flex items-center gap-3 rounded-lg border border-slate-700 bg-slate-800/50 p-3">
                <div className={`flex h-8 w-8 items-center justify-center rounded-full ${
                  e.eventType === "unlock" ? "bg-emerald-500/10" : "bg-sky-500/10"
                }`}>
                  <Icon className={`h-4 w-4 ${e.eventType === "unlock" ? "text-emerald-400" : "text-sky-400"}`} />
                </div>
                <div className="flex-1">
                  <p className="text-sm text-white">{e.eventType.toUpperCase()} · {e.eventSource.replace("_", " ")}</p>
                  <p className="text-xs text-slate-400">{e.details}</p>
                </div>
                <span className="text-xs text-slate-500">{e.occurredAt}</span>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
