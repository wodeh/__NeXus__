"use client";

import { useState, useEffect, useMemo, useCallback } from "react";
import Link from "next/link";
import {
  BedDouble,
  Users,
  User,
  ArrowRightLeft,
  DoorOpen,
  DollarSign,
  TrendingUp,
  TrendingDown,
  AlertTriangle,
  Clock,
  CheckCircle2,
  CalendarDays,
  ChevronRight,
  Wifi,
  WifiOff,
  BatteryWarning,
  KeyRound,
  Plus,
  Star,
  Activity,
  RefreshCw,
} from "lucide-react";
import { Reservation, Room, getReservations, getRooms } from "@/lib/api";

const today = new Date().toISOString().split("T")[0];

/* ─── Mock data for features without backend APIs yet ─── */
const mockWeeklyRevenue = [
  { day: "Mon", revenue: 6200, occupancy: 65 },
  { day: "Tue", revenue: 7100, occupancy: 72 },
  { day: "Wed", revenue: 6800, occupancy: 68 },
  { day: "Thu", revenue: 7500, occupancy: 75 },
  { day: "Fri", revenue: 8900, occupancy: 85 },
  { day: "Sat", revenue: 9200, occupancy: 90 },
  { day: "Sun", revenue: 8432, occupancy: 73 },
];

const mockRecentActivity = [
  { time: "09:15", event: "Check-in", detail: "Room 305 — James Wilson", icon: DoorOpen, color: "emerald" },
  { time: "09:02", event: "Lock offline", detail: "Room 302 — OT-302-2026", icon: WifiOff, color: "rose" },
  { time: "08:45", event: "Housekeeping done", detail: "Room 104 — Vacant Clean", icon: CheckCircle2, color: "sky" },
  { time: "08:30", event: "New reservation", detail: "Booking.com — Room 305", icon: Plus, color: "nexus" },
  { time: "08:15", event: "Low battery alert", detail: "Room 103 — 12% remaining", icon: BatteryWarning, color: "amber" },
];

export default function HotelDashboardPage() {
  const [reservations, setReservations] = useState<Reservation[]>([]);
  const [rooms, setRooms] = useState<Room[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [res, rms] = await Promise.all([getReservations(), getRooms()]);
      setReservations(res);
      setRooms(rms);
    } catch (e: any) {
      setError(e.message || "Failed to load dashboard data");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const stats = useMemo(() => {
    const totalRooms = rooms.length;
    const occupied = rooms.filter((r) => r.status === "occupied").length;
    const available = rooms.filter((r) => r.status === "vacant_clean").length;
    const blocked = rooms.filter((r) => r.status === "blocked").length;
    const maintenance = rooms.filter((r) => r.status === "maintenance").length;
    const arrivalsToday = (reservations || []).filter((r) => r.check_in === today && r.status !== "checked_in").length;
    const departuresToday = (reservations || []).filter((r) => r.check_out === today && r.status === "checked_in").length;
    const inHouse = (reservations || []).filter((r) => r.status === "checked_in").length;
    const vipArrivals = (reservations || []).filter((r) => r.check_in === today && r.vip).length;
    const occupancyRate = totalRooms ? Math.round((occupied / totalRooms) * 100) : 0;

    return {
      totalRooms,
      occupied,
      available,
      blocked,
      maintenance,
      arrivalsToday,
      departuresToday,
      inHouse,
      vipArrivals,
      occupancyRate,
      revenueToday: 8432,
      revenueYesterday: 7650,
      pendingBalance: (reservations || []).reduce((sum, r) => sum + (r.balance || 0), 0),
    };
  }, [rooms, reservations]);

  const arrivals = useMemo(() =>
    (reservations || [])
      .filter((r) => r.check_in === today && r.status !== "checked_in")
      .map((r) => ({
        id: r.id,
        guest: r.guest_name,
        room: r.room_number || "",
        type: r.room_type,
        status: r.status,
        vip: r.vip,
        source: r.source,
      })),
    [reservations]
  );

  const departures = useMemo(() =>
    (reservations || [])
      .filter((r) => r.check_out === today && r.status === "checked_in")
      .map((r) => ({
        id: r.id,
        guest: r.guest_name,
        room: r.room_number || "",
        type: r.room_type,
        status: r.status,
        balance: r.balance,
        vip: r.vip,
      })),
    [reservations]
  );

  if (loading) {
    return (
      <div className="flex h-96 items-center justify-center">
        <RefreshCw className="h-8 w-8 animate-spin text-nexus-400" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex h-96 flex-col items-center justify-center gap-4">
        <AlertTriangle className="h-10 w-10 text-rose-400" />
        <p className="text-rose-400">{error}</p>
        <button onClick={fetchData} className="btn-primary">Retry</button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Dashboard</h1>
          <p className="text-sm text-slate-400">{new Date(today + "T00:00:00").toLocaleDateString("en-US", { weekday: "long", year: "numeric", month: "long", day: "numeric" })}</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={fetchData} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white" title="Refresh">
            <RefreshCw className="h-4 w-4" />
          </button>
          <Link href="/reservations/new" className="btn-primary">
            <Plus className="h-4 w-4" />
            New Reservation
          </Link>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard label="Occupancy" value={`${stats.occupancyRate}%`} sub={`${stats.occupied}/${stats.totalRooms} rooms`} icon={BedDouble} color="sky" />
        <StatCard label="In-House" value={stats.inHouse} sub={`${stats.arrivalsToday} arriving · ${stats.departuresToday} departing`} icon={Users} color="emerald" />
        <StatCard label="Available" value={stats.available} sub={`${stats.blocked} blocked · ${stats.maintenance} maintenance`} icon={DoorOpen} color="violet" />
        <StatCard label="Revenue Today" value={`$${stats.revenueToday.toLocaleString()}`} sub={stats.revenueToday > stats.revenueYesterday ? "↑ from yesterday" : "↓ from yesterday"} icon={DollarSign} color="amber" trend={stats.revenueToday > stats.revenueYesterday ? "up" : "down"} />
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Arrivals */}
        <div className="card">
          <div className="mb-4 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <ArrowRightLeft className="h-5 w-5 text-nexus-400" />
              <h3 className="font-bold text-white">Arrivals Today</h3>
              <span className="badge badge-blue">{arrivals.length}</span>
            </div>
            <Link href="/reservations" className="text-xs text-nexus-400 hover:underline">View all</Link>
          </div>
          <div className="space-y-3">
            {arrivals.length === 0 && <p className="text-sm text-slate-500">No arrivals today.</p>}
            {arrivals.map((a) => (
              <div key={a.id} className="flex items-center justify-between rounded-lg border border-slate-700/50 bg-slate-800/50 p-3">
                <div className="flex items-center gap-3">
                  <div className="flex h-8 w-8 items-center justify-center rounded-full bg-nexus-500/10">
                    {a.vip ? <Star className="h-4 w-4 text-amber-400" /> : <User className="h-4 w-4 text-nexus-400" />}
                  </div>
                  <div>
                    <p className="text-sm font-medium text-white">{a.guest}</p>
                    <p className="text-xs text-slate-400">{a.room ? `Room ${a.room} · ${a.type}` : a.type}</p>
                  </div>
                </div>
                <span className={`badge ${a.source === "ota" ? "badge-purple" : a.source === "walk_in" ? "badge-amber" : "badge-blue"}`}>{a.source}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Departures */}
        <div className="card">
          <div className="mb-4 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <DoorOpen className="h-5 w-5 text-rose-400" />
              <h3 className="font-bold text-white">Departures Today</h3>
              <span className="badge badge-amber">{departures.length}</span>
            </div>
            <Link href="/reservations" className="text-xs text-nexus-400 hover:underline">View all</Link>
          </div>
          <div className="space-y-3">
            {departures.length === 0 && <p className="text-sm text-slate-500">No departures today.</p>}
            {departures.map((d) => (
              <div key={d.id} className="flex items-center justify-between rounded-lg border border-slate-700/50 bg-slate-800/50 p-3">
                <div className="flex items-center gap-3">
                  <div className="flex h-8 w-8 items-center justify-center rounded-full bg-rose-500/10">
                    {d.vip ? <Star className="h-4 w-4 text-amber-400" /> : <DoorOpen className="h-4 w-4 text-rose-400" />}
                  </div>
                  <div>
                    <p className="text-sm font-medium text-white">{d.guest}</p>
                    <p className="text-xs text-slate-400">Room {d.room} · {d.type}</p>
                  </div>
                </div>
                <span className={`badge ${d.balance > 0 ? "badge-red" : "badge-green"}`}>{d.balance > 0 ? `$${d.balance} due` : "Paid"}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Activity + Quick Actions */}
        <div className="space-y-6">
          <div className="card">
            <div className="mb-4 flex items-center gap-2">
              <Activity className="h-5 w-5 text-nexus-400" />
              <h3 className="font-bold text-white">Recent Activity</h3>
            </div>
            <div className="space-y-3">
              {mockRecentActivity.map((a, i) => (
                <div key={i} className="flex items-start gap-3">
                  <div className={`mt-0.5 rounded-full p-1.5 bg-${a.color}-500/10`}>
                    <a.icon className={`h-3.5 w-3.5 text-${a.color}-400`} />
                  </div>
                  <div>
                    <p className="text-sm text-white">{a.event}</p>
                    <p className="text-xs text-slate-400">{a.detail}</p>
                    <p className="text-[10px] text-slate-500">{a.time}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Revenue Chart */}
      <div className="card">
        <div className="mb-4 flex items-center gap-2">
          <DollarSign className="h-5 w-5 text-emerald-400" />
          <h3 className="font-bold text-white">Weekly Revenue</h3>
        </div>
        <div className="flex items-end gap-2">
          {mockWeeklyRevenue.map((d) => (
            <div key={d.day} className="flex flex-1 flex-col items-center gap-1">
              <div className="w-full rounded-t bg-emerald-500/20" style={{ height: `${(d.revenue / 10000) * 160}px` }}>
                <div className="h-full w-full rounded-t bg-emerald-500/40" />
              </div>
              <span className="text-[10px] text-slate-400">{d.day}</span>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}

function StatCard({ label, value, sub, icon: Icon, color, trend }: { label: string; value: string | number; sub: string; icon: any; color: string; trend?: "up" | "down" }) {
  const colorMap: Record<string, string> = {
    sky: "bg-sky-500/10 text-sky-400",
    emerald: "bg-emerald-500/10 text-emerald-400",
    violet: "bg-violet-500/10 text-violet-400",
    amber: "bg-amber-500/10 text-amber-400",
    rose: "bg-rose-500/10 text-rose-400",
  };
  return (
    <div className="rounded-lg border border-slate-700 bg-slate-800 p-4">
      <div className="flex items-center justify-between">
        <div className={`rounded-md p-1.5 ${colorMap[color] || ""}`}>
          <Icon className="h-5 w-5" />
        </div>
        {trend && (
          trend === "up" ? <TrendingUp className="h-4 w-4 text-emerald-400" /> : <TrendingDown className="h-4 w-4 text-rose-400" />
        )}
      </div>
      <div className="mt-2">
        <div className="text-2xl font-bold text-white">{value}</div>
        <div className="text-xs text-slate-400">{sub}</div>
      </div>
    </div>
  );
}
