"use client";

import { useState, useEffect } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import {
  LayoutGrid,
  Lock,
  Unlock,
  KeyRound,
  Eye,
  UserCheck,
  UserX,
  Wrench,
  AlertTriangle,
  CheckCircle2,
  Ban,
  Clock,
  CalendarDays,
  Search,
  Tv,
  Battery,
  BatteryWarning,
  Wifi,
  WifiOff,
  ArrowRightLeft,
  BedDouble,
  Sparkles,
  Phone,
  ChevronRight,
  Crown,
} from "lucide-react";
import {
  getRooms,
  getReservations,
  getLocks,
  getHousekeepingTasks,
  remoteUnlock,
  Room,
  Reservation,
  SmartLock,
  HousekeepingTask,
} from "@/lib/api";

interface RoomControl {
  id: string;
  number: string;
  floor: string;
  type: string;
  status: "vacant_clean" | "vacant_dirty" | "occupied" | "blocked" | "maintenance" | "out_of_order";
  guest?: string;
  arrival?: string;
  departure?: string;
  nights?: number;
  balance?: number;
  lockStatus?: "locked" | "unlocked" | "offline";
  batteryLevel?: number;
  housekeepingStatus?: "clean" | "dirty" | "inspected";
  iptvOnline?: boolean;
  vip?: boolean;
  notes?: string;
}

const statusConfig: Record<string, { label: string; color: string; icon: React.ElementType; bg: string }> = {
  vacant_clean: { label: "Vacant Clean", color: "text-emerald-400", icon: CheckCircle2, bg: "bg-emerald-500/5 border-emerald-500/10" },
  vacant_dirty: { label: "Vacant Dirty", color: "text-rose-400", icon: Sparkles, bg: "bg-rose-500/5 border-rose-500/10" },
  occupied: { label: "Occupied", color: "text-sky-400", icon: BedDouble, bg: "bg-sky-500/5 border-sky-500/10" },
  blocked: { label: "Blocked", color: "text-violet-400", icon: Ban, bg: "bg-violet-500/5 border-violet-500/10" },
  maintenance: { label: "Maintenance", color: "text-amber-400", icon: Wrench, bg: "bg-amber-500/5 border-amber-500/10" },
  out_of_order: { label: "OOO", color: "text-slate-400", icon: AlertTriangle, bg: "bg-slate-500/5 border-slate-500/10" },
};

export default function CommandCenterPage() {
  const { config } = useTenant();
  const [filter, setFilter] = useState<string>("all");
  const [search, setSearch] = useState("");
  const [selectedRoom, setSelectedRoom] = useState<RoomControl | null>(null);
  const [rooms, setRooms] = useState<RoomControl[]>([]);
  const [loading, setLoading] = useState(false);
  const [unlocking, setUnlocking] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    setLoading(true);
    try {
      const [roomList, reservations, locks, hkTasks] = await Promise.all([
        getRooms(),
        getReservations(),
        getLocks(),
        getHousekeepingTasks(),
      ]);

      const today = new Date().toISOString().split("T")[0];
      const activeReservations = reservations.filter((r) =>
        r.status === "confirmed" || r.status === "checked_in"
      );

      const merged: RoomControl[] = roomList.map((room) => {
        const res = activeReservations.find((r) => r.room_number === room.number);
        const lock = locks.find((l) => l.room_number === room.number);
        const hk = hkTasks.find((t) => t.room_number === room.number);

        const checkIn = res?.check_in;
        const checkOut = res?.check_out;
        const nights = checkIn && checkOut
          ? Math.ceil((new Date(checkOut).getTime() - new Date(checkIn).getTime()) / (1000 * 60 * 60 * 24))
          : undefined;

        return {
          id: room.id,
          number: room.number,
          floor: room.floor,
          type: room.type,
          status: (room.status as RoomControl["status"]) || "vacant_clean",
          guest: res?.guest_name,
          arrival: checkIn,
          departure: checkOut,
          nights,
          balance: res?.total,
          lockStatus: lock
            ? lock.status === "offline"
              ? "offline"
              : "locked"
            : undefined,
          batteryLevel: lock?.battery_level,
          housekeepingStatus: (hk?.status as RoomControl["housekeepingStatus"]) || undefined,
          iptvOnline: true, // TODO: wire IPTV status
          vip: room.config?.vip === true,
          notes: (room.config?.notes as string) || res?.special_requests || undefined,
        };
      });

      setRooms(merged);
    } catch (e) {
      console.error("Failed to load command center data", e);
    } finally {
      setLoading(false);
    }
  }

  async function handleRemoteUnlock() {
    if (!selectedRoom) return;
    const lock = await getLocks().then((locks) => locks.find((l) => l.room_number === selectedRoom.number));
    if (!lock) return;
    setUnlocking(true);
    try {
      await remoteUnlock(lock.id);
      await loadData();
    } catch (e) {
      console.error("Failed to unlock", e);
    } finally {
      setUnlocking(false);
    }
  }

  const hasCommand = hasCapability(config, CAPABILITIES.OPERATIONS.FRONT_DESK);

  if (!hasCommand) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-800">
          <Lock className="h-8 w-8 text-slate-500" />
        </div>
        <h2 className="mt-6 text-xl font-bold text-white">Command Center Locked</h2>
        <p className="mt-2 max-w-md text-sm text-slate-400">
          The hotel command center requires the Operations or Enterprise license tier.
        </p>
        <button className="btn-primary mt-6">
          <ArrowRightLeft className="h-4 w-4" />
          Upgrade License
        </button>
      </div>
    );
  }

  const floors = Array.from(new Set(rooms.map((r) => r.floor))).sort();

  const filteredRooms = rooms.filter((r) => {
    if (filter !== "all" && r.status !== filter) return false;
    if (search) {
      const q = search.toLowerCase();
      return (
        r.number.includes(q) ||
        r.guest?.toLowerCase().includes(q) ||
        r.type.toLowerCase().includes(q) ||
        r.notes?.toLowerCase().includes(q)
      );
    }
    return true;
  });

  const todayStr = new Date().toISOString().split("T")[0];
  const stats = {
    total: rooms.length,
    occupied: rooms.filter((r) => r.status === "occupied").length,
    vacantClean: rooms.filter((r) => r.status === "vacant_clean").length,
    vacantDirty: rooms.filter((r) => r.status === "vacant_dirty").length,
    blocked: rooms.filter((r) => r.status === "blocked").length,
    maintenance: rooms.filter((r) => ["maintenance", "out_of_order"].includes(r.status)).length,
    arrivalsToday: rooms.filter((r) => r.arrival === todayStr).length,
    departuresToday: rooms.filter((r) => r.departure === todayStr).length,
    lowBattery: rooms.filter((r) => (r.batteryLevel || 100) < 20).length,
    offlineLocks: rooms.filter((r) => r.lockStatus === "offline").length,
    revenueAtRisk: rooms.filter((r) => r.status === "vacant_dirty").length * 129,
  };

  const todayArrivals = rooms.filter((r) => r.arrival === todayStr && r.status !== "occupied");
  const todayDepartures = rooms.filter((r) => r.departure === todayStr && r.status === "occupied");

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Command Center</h1>
          <p className="text-sm text-slate-400">Real-time hotel control — every room, every guest, every lock.</p>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 rounded-lg bg-slate-800 px-3 py-2">
            <Clock className="h-4 w-4 text-slate-400" />
            <span className="text-sm font-mono text-white">{new Date().toLocaleTimeString()}</span>
          </div>
          <div className="flex items-center gap-2 rounded-lg bg-emerald-500/10 px-3 py-2">
            <Wifi className="h-4 w-4 text-emerald-400" />
            <span className="text-sm text-emerald-400">{rooms.filter(r => r.lockStatus !== "offline").length}/{rooms.length} Online</span>
          </div>
        </div>
      </div>

      {/* Stats Row */}
      <div className="grid grid-cols-6 gap-3">
        <StatCard label="Occupied" value={stats.occupied} total={stats.total} color="sky" icon={BedDouble} />
        <StatCard label="Vacant Clean" value={stats.vacantClean} total={stats.total} color="emerald" icon={CheckCircle2} />
        <StatCard label="Vacant Dirty" value={stats.vacantDirty} total={stats.total} color="rose" icon={Sparkles} />
        <StatCard label="Arrivals Today" value={stats.arrivalsToday} color="violet" icon={CalendarDays} />
        <StatCard label="Departures Today" value={stats.departuresToday} color="amber" icon={ArrowRightLeft} />
        <StatCard label="Alerts" value={stats.lowBattery + stats.offlineLocks} color="rose" icon={AlertTriangle} alert />
      </div>

      {/* Search + Filters */}
      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            placeholder="Search room, guest, or note..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="w-full rounded-lg border border-slate-700 bg-slate-800 py-2 pl-10 pr-4 text-sm text-white placeholder-slate-500 focus:border-nexus-500 focus:outline-none"
          />
        </div>
        <div className="flex gap-1">
          {["all", "occupied", "vacant_clean", "vacant_dirty", "blocked", "maintenance"].map((f) => (
            <button
              key={f}
              onClick={() => setFilter(f)}
              className={`rounded-lg px-3 py-2 text-xs font-medium transition-colors ${
                filter === f
                  ? "bg-nexus-500 text-white"
                  : "bg-slate-800 text-slate-400 hover:bg-slate-700 hover:text-white"
              }`}
            >
              {f.replace("_", " ").replace(/\b\w/g, (l) => l.toUpperCase())}
            </button>
          ))}
        </div>
      </div>

      <div className="flex gap-6">
        {/* Room Grid */}
        <div className="flex-1 space-y-4">
          {loading ? (
            <p className="text-sm text-slate-500">Loading rooms...</p>
          ) : (
            floors.map((floor) => {
              const floorRooms = filteredRooms.filter((r) => r.floor === floor);
              if (floorRooms.length === 0) return null;
              return (
                <div key={floor}>
                  <h3 className="mb-2 text-xs font-semibold uppercase tracking-wider text-slate-500">Floor {floor}</h3>
                  <div className="grid grid-cols-5 gap-3">
                    {floorRooms.map((room) => {
                      const st = statusConfig[room.status];
                      const StatusIcon = st.icon;
                      return (
                        <button
                          key={room.id}
                          onClick={() => { setSelectedRoom(room); }}
                          className={`relative rounded-lg border p-3 text-left transition-all hover:scale-[1.02] ${
                            selectedRoom?.id === room.id
                              ? "border-nexus-400 ring-1 ring-nexus-400"
                              : st.bg
                          }`}
                        >
                          <div className="flex items-center justify-between">
                            <span className={`text-lg font-bold ${st.color}`}>{room.number}</span>
                            {room.vip && <Crown className="h-3 w-3 text-amber-400" />}
                          </div>
                          <div className="mt-1 flex items-center gap-1">
                            <StatusIcon className={`h-3 w-3 ${st.color}`} />
                            <span className={`text-[10px] font-medium ${st.color}`}>{st.label}</span>
                          </div>
                          {room.guest && (
                            <p className="mt-1 truncate text-[10px] text-slate-300">{room.guest}</p>
                          )}
                          <div className="mt-2 flex items-center gap-2">
                            {room.lockStatus === "offline" ? (
                              <WifiOff className="h-3 w-3 text-rose-400" />
                            ) : room.lockStatus === "unlocked" ? (
                              <Unlock className="h-3 w-3 text-amber-400" />
                            ) : (
                              <Lock className="h-3 w-3 text-emerald-400" />
                            )}
                            {(room.batteryLevel || 0) < 20 ? (
                              <BatteryWarning className="h-3 w-3 text-rose-400" />
                            ) : (
                              <Battery className="h-3 w-3 text-slate-500" />
                            )}
                            {!room.iptvOnline && <Tv className="h-3 w-3 text-rose-400" />}
                          </div>
                          {room.notes && (
                            <p className="mt-1 text-[9px] text-slate-500 line-clamp-1">{room.notes}</p>
                          )}
                        </button>
                      );
                    })}
                  </div>
                </div>
              );
            })
          )}
        </div>

        {/* Side Panel */}
        <div className="w-80 shrink-0 space-y-4">
          {selectedRoom && (
            <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-4">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-bold text-white">Room {selectedRoom.number}</h3>
                <button onClick={() => setSelectedRoom(null)} className="text-slate-500 hover:text-white">
                  <ChevronRight className="h-4 w-4 rotate-90" />
                </button>
              </div>
              <p className="text-sm text-slate-400">{selectedRoom.type} · Floor {selectedRoom.floor}</p>

              {selectedRoom.guest && (
                <div className="mt-3 rounded-lg bg-slate-700/50 p-3">
                  <div className="flex items-center gap-2">
                    <UserCheck className="h-4 w-4 text-sky-400" />
                    <span className="text-sm font-medium text-white">{selectedRoom.guest}</span>
                  </div>
                  <div className="mt-1 flex items-center gap-3 text-xs text-slate-400">
                    <span>{selectedRoom.arrival} → {selectedRoom.departure}</span>
                    <span>{selectedRoom.nights} nights</span>
                  </div>
                  {selectedRoom.balance !== undefined && (
                    <div className="mt-1 text-xs text-slate-400">Balance: ${selectedRoom.balance.toFixed(2)}</div>
                  )}
                </div>
              )}

              <div className="mt-3 grid grid-cols-2 gap-2">
                {selectedRoom.status === "occupied" && (
                  <>
                    <QuickActionButton icon={ArrowRightLeft} label="Extend Stay" color="sky" />
                    <QuickActionButton icon={UserX} label="Check Out" color="rose" />
                  </>
                )}
                {selectedRoom.status === "vacant_dirty" && (
                  <>
                    <QuickActionButton icon={Sparkles} label="Mark Clean" color="emerald" />
                    <QuickActionButton icon={UserCheck} label="Walk-in" color="sky" />
                  </>
                )}
                {selectedRoom.status === "vacant_clean" && (
                  <>
                    <QuickActionButton icon={UserCheck} label="Check In" color="sky" />
                    <QuickActionButton icon={Ban} label="Block Room" color="violet" />
                  </>
                )}
                <QuickActionButton
                  icon={selectedRoom.lockStatus === "locked" ? Unlock : Lock}
                  label={selectedRoom.lockStatus === "locked" ? "Unlock" : "Lock"}
                  color={selectedRoom.lockStatus === "locked" ? "amber" : "emerald"}
                  onClick={handleRemoteUnlock}
                />
                <QuickActionButton icon={KeyRound} label="New Code" color="slate" />
              </div>
            </div>
          )}

          <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-4">
            <div className="flex items-center gap-2 mb-3">
              <CalendarDays className="h-4 w-4 text-violet-400" />
              <h4 className="text-sm font-semibold text-white">Arrivals Today</h4>
              <span className="ml-auto rounded-full bg-violet-500/10 px-2 py-0.5 text-xs text-violet-400">{todayArrivals.length}</span>
            </div>
            <div className="space-y-2">
              {todayArrivals.length === 0 ? (
                <p className="text-xs text-slate-500">No arrivals scheduled</p>
              ) : (
                todayArrivals.map((r) => (
                  <div key={r.id} className="flex items-center justify-between rounded bg-slate-700/30 px-2 py-1.5">
                    <div>
                      <span className="text-xs font-medium text-white">Room {r.number}</span>
                      {r.guest && <span className="ml-2 text-[10px] text-slate-400">{r.guest}</span>}
                    </div>
                    <button className="rounded px-2 py-0.5 text-[10px] bg-sky-500/10 text-sky-400 hover:bg-sky-500/20">Check In</button>
                  </div>
                ))
              )}
            </div>
          </div>

          <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-4">
            <div className="flex items-center gap-2 mb-3">
              <ArrowRightLeft className="h-4 w-4 text-amber-400" />
              <h4 className="text-sm font-semibold text-white">Departures Today</h4>
              <span className="ml-auto rounded-full bg-amber-500/10 px-2 py-0.5 text-xs text-amber-400">{todayDepartures.length}</span>
            </div>
            <div className="space-y-2">
              {todayDepartures.length === 0 ? (
                <p className="text-xs text-slate-500">No departures scheduled</p>
              ) : (
                todayDepartures.map((r) => (
                  <div key={r.id} className="flex items-center justify-between rounded bg-slate-700/30 px-2 py-1.5">
                    <div>
                      <span className="text-xs font-medium text-white">Room {r.number}</span>
                      <span className="ml-2 text-[10px] text-slate-400">{r.guest}</span>
                    </div>
                    <button className="rounded px-2 py-0.5 text-[10px] bg-rose-500/10 text-rose-400 hover:bg-rose-500/20">Check Out</button>
                  </div>
                ))
              )}
            </div>
          </div>

          <div className="rounded-lg border border-rose-500/20 bg-rose-500/5 p-4">
            <div className="flex items-center gap-2 mb-3">
              <AlertTriangle className="h-4 w-4 text-rose-400" />
              <h4 className="text-sm font-semibold text-white">Alerts</h4>
            </div>
            <div className="space-y-2">
              {rooms.filter(r => (r.batteryLevel || 100) < 20 || r.lockStatus === "offline").map(r => (
                <div key={r.id} className="flex items-center gap-2 text-xs">
                  {r.lockStatus === "offline" ? (
                    <WifiOff className="h-3 w-3 text-rose-400" />
                  ) : (
                    <BatteryWarning className="h-3 w-3 text-rose-400" />
                  )}
                  <span className="text-slate-300">Room {r.number}</span>
                  <span className="text-slate-500">
                    {r.lockStatus === "offline" ? "Lock offline" : `Battery ${r.batteryLevel}%`}
                  </span>
                </div>
              ))}
              {stats.lowBattery + stats.offlineLocks === 0 && (
                <p className="text-xs text-slate-500">All systems nominal</p>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function StatCard({ label, value, total, color, icon: Icon, alert }: { label: string; value: number; total?: number; color: string; icon: React.ElementType; alert?: boolean }) {
  const colorMap: Record<string, string> = {
    sky: "bg-sky-500/10 text-sky-400 border-sky-500/20",
    emerald: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
    rose: "bg-rose-500/10 text-rose-400 border-rose-500/20",
    violet: "bg-violet-500/10 text-violet-400 border-violet-500/20",
    amber: "bg-amber-500/10 text-amber-400 border-amber-500/20",
  };
  return (
    <div className={`rounded-lg border p-3 ${colorMap[color]} ${alert ? "animate-pulse" : ""}`}>
      <div className="flex items-center justify-between">
        <span className="text-xs font-medium opacity-80">{label}</span>
        <Icon className="h-4 w-4 opacity-60" />
      </div>
      <div className="mt-1 flex items-baseline gap-1">
        <span className="text-2xl font-bold">{value}</span>
        {total !== undefined && <span className="text-xs opacity-60">/ {total}</span>}
      </div>
    </div>
  );
}

function QuickActionButton({ icon: Icon, label, color, onClick }: { icon: React.ElementType; label: string; color: string; onClick?: () => void }) {
  const colorMap: Record<string, string> = {
    sky: "bg-sky-500/10 text-sky-400 hover:bg-sky-500/20",
    emerald: "bg-emerald-500/10 text-emerald-400 hover:bg-emerald-500/20",
    rose: "bg-rose-500/10 text-rose-400 hover:bg-rose-500/20",
    violet: "bg-violet-500/10 text-violet-400 hover:bg-violet-500/20",
    amber: "bg-amber-500/10 text-amber-400 hover:bg-amber-500/20",
    slate: "bg-slate-500/10 text-slate-400 hover:bg-slate-500/20",
  };
  return (
    <button 
      onClick={onClick}
      className={`flex items-center gap-1.5 rounded px-2 py-1.5 text-[10px] font-medium transition-colors ${colorMap[color]}`}
    >
      <Icon className="h-3 w-3" />
      {label}
    </button>
  );
}
