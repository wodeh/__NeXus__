"use client";

import { useState } from "react";
import Link from "next/link";
import {
  BedDouble,
  Users,
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
} from "lucide-react";

const today = "2026-05-10";

const mockStats = {
  totalRooms: 120,
  occupied: 87,
  available: 28,
  blocked: 5,
  arrivalsToday: 12,
  departuresToday: 8,
  inHouse: 87,
  revenueToday: 8432,
  revenueYesterday: 7650,
  pendingBalance: 2140,
  vipArrivals: 3,
  complaints: 1,
  maintenance: 2,
  lowBatteryLocks: 1,
  offlineLocks: 1,
};

const occupancyRate = Math.round((mockStats.occupied / mockStats.totalRooms) * 100);

const mockArrivals = [
  { id: "r-101", guest: "James Wilson", room: "305", type: "Deluxe", status: "confirmed", time: "14:00", vip: true, source: "direct" },
  { id: "r-102", guest: "Maria Garcia", room: "412", type: "Suite", status: "confirmed", time: "15:30", vip: false, source: "ota" },
  { id: "r-103", guest: "Robert Kim", room: "", type: "Standard", status: "confirmed", time: "16:00", vip: false, source: "walk_in" },
  { id: "r-104", guest: "Sarah Johnson", room: "201", type: "Deluxe King", status: "confirmed", time: "13:00", vip: true, source: "direct" },
];

const mockDepartures = [
  { id: "r-201", guest: "Alice Chen", room: "201", type: "Deluxe King", status: "checked_in", time: "11:00", balance: 129, vip: true },
  { id: "r-202", guest: "Carol Jones", room: "412", type: "Suite", status: "checked_in", time: "12:00", balance: 0, vip: false },
  { id: "r-203", guest: "David Lee", room: "103", type: "Deluxe", status: "checked_in", time: "11:30", balance: 129, vip: false },
];

const mockRecentActivity = [
  { time: "09:15", event: "Check-in", detail: "Room 305 — James Wilson", icon: DoorOpen, color: "emerald" },
  { time: "09:02", event: "Lock offline", detail: "Room 302 — OT-302-2026", icon: WifiOff, color: "rose" },
  { time: "08:45", event: "Housekeeping done", detail: "Room 104 — Vacant Clean", icon: CheckCircle2, color: "sky" },
  { time: "08:30", event: "New reservation", detail: "Booking.com — Room 305", icon: Plus, color: "nexus" },
  { time: "08:15", event: "Low battery alert", detail: "Room 103 — 12% remaining", icon: BatteryWarning, color: "amber" },
];

const mockWeeklyRevenue = [
  { day: "Mon", revenue: 6200, occupancy: 65 },
  { day: "Tue", revenue: 7100, occupancy: 72 },
  { day: "Wed", revenue: 6800, occupancy: 68 },
  { day: "Thu", revenue: 7500, occupancy: 75 },
  { day: "Fri", revenue: 8900, occupancy: 85 },
  { day: "Sat", revenue: 9200, occupancy: 90 },
  { day: "Sun", revenue: 8432, occupancy: 73 },
];

export default function DashboardPage() {
  const [dateRange, setDateRange] = useState("today");

  const revenueChange = mockStats.revenueToday - mockStats.revenueYesterday;
  const revenueChangePct = Math.round((revenueChange / mockStats.revenueYesterday) * 100);

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Dashboard</h1>
          <p className="text-sm text-slate-400">{today} · Real-time hotel overview</p>
        </div>
        <div className="flex items-center gap-2">
          <select
            value={dateRange}
            onChange={(e) => setDateRange(e.target.value)}
            className="input text-xs"
          >
            <option value="today">Today</option>
            <option value="week">This Week</option>
            <option value="month">This Month</option>
          </select>
          <Link href="/command" className="btn-primary gap-2 text-xs">
            <Activity className="h-4 w-4" />
            Command Center
          </Link>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-4 gap-4">
        {/* Occupancy */}
        <div className="card space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-medium">Occupancy</span>
            <BedDouble className="h-4 w-4 text-slate-500" />
          </div>
          <div className="flex items-end gap-2">
            <span className="text-3xl font-bold text-white">{occupancyRate}%</span>
            <span className="text-xs text-slate-400 mb-1">{mockStats.occupied}/{mockStats.totalRooms}</span>
          </div>
          <div className="h-2 rounded-full bg-slate-800">
            <div
              className="h-full rounded-full bg-emerald-400 transition-all"
              style={{ width: `${occupancyRate}%` }}
            />
          </div>
          <div className="flex justify-between text-xs text-slate-500">
            <span>Available: {mockStats.available}</span>
            <span>Blocked: {mockStats.blocked}</span>
          </div>
        </div>

        {/* Revenue */}
        <div className="card space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-medium">Revenue Today</span>
            <DollarSign className="h-4 w-4 text-slate-500" />
          </div>
          <div className="flex items-end gap-2">
            <span className="text-3xl font-bold text-white">${mockStats.revenueToday.toLocaleString()}</span>
          </div>
          <div className={`flex items-center gap-1 text-xs ${revenueChange >= 0 ? "text-emerald-400" : "text-rose-400"}`}>
            {revenueChange >= 0 ? <TrendingUp className="h-3 w-3" /> : <TrendingDown className="h-3 w-3" />}
            {revenueChange >= 0 ? "+" : ""}{revenueChangePct}% vs yesterday
          </div>
          <div className="text-xs text-slate-500">Pending balance: ${mockStats.pendingBalance}</div>
        </div>

        {/* Arrivals */}
        <div className="card space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-medium">Today's Arrivals</span>
            <ArrowRightLeft className="h-4 w-4 text-slate-500" />
          </div>
          <div className="flex items-end gap-2">
            <span className="text-3xl font-bold text-white">{mockStats.arrivalsToday}</span>
            {mockStats.vipArrivals > 0 && (
              <span className="flex items-center gap-1 text-xs text-amber-400 mb-1">
                <Star className="h-3 w-3" /> {mockStats.vipArrivals} VIP
              </span>
            )}
          </div>
          <div className="text-xs text-slate-500">{mockArrivals.filter(a => !a.room).length} unassigned</div>
        </div>

        {/* Departures */}
        <div className="card space-y-3">
          <div className="flex items-center justify-between">
            <span className="text-xs text-slate-400 font-medium">Today's Departures</span>
            <DoorOpen className="h-4 w-4 text-slate-500" />
          </div>
          <div className="flex items-end gap-2">
            <span className="text-3xl font-bold text-white">{mockStats.departuresToday}</span>
          </div>
          <div className="text-xs text-slate-500">
            {mockDepartures.filter(d => d.balance > 0).length} with balance due
          </div>
        </div>
      </div>

      {/* Alerts Row */}
      {(mockStats.lowBatteryLocks > 0 || mockStats.offlineLocks > 0 || mockStats.complaints > 0 || mockStats.maintenance > 0) && (
        <div className="flex flex-wrap gap-2">
          {mockStats.lowBatteryLocks > 0 && (
            <div className="flex items-center gap-2 rounded-lg border border-amber-500/20 bg-amber-500/5 px-3 py-2">
              <BatteryWarning className="h-4 w-4 text-amber-400" />
              <span className="text-xs text-amber-400">{mockStats.lowBatteryLocks} lock{mockStats.lowBatteryLocks > 1 ? "s" : ""} low battery</span>
            </div>
          )}
          {mockStats.offlineLocks > 0 && (
            <div className="flex items-center gap-2 rounded-lg border border-rose-500/20 bg-rose-500/5 px-3 py-2">
              <WifiOff className="h-4 w-4 text-rose-400" />
              <span className="text-xs text-rose-400">{mockStats.offlineLocks} lock{mockStats.offlineLocks > 1 ? "s" : ""} offline</span>
            </div>
          )}
          {mockStats.maintenance > 0 && (
            <div className="flex items-center gap-2 rounded-lg border border-sky-500/20 bg-sky-500/5 px-3 py-2">
              <AlertTriangle className="h-4 w-4 text-sky-400" />
              <span className="text-xs text-sky-400">{mockStats.maintenance} maintenance request{mockStats.maintenance > 1 ? "s" : ""}</span>
            </div>
          )}
        </div>
      )}

      {/* Main Grid */}
      <div className="grid grid-cols-3 gap-4">
        {/* Arrivals */}
        <div className="card space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold text-white flex items-center gap-2">
              <CalendarDays className="h-4 w-4 text-emerald-400" />
              Arrivals Today
            </h3>
            <Link href="/reservations" className="text-xs text-nexus-400 hover:text-nexus-300 flex items-center gap-1">
              View All <ChevronRight className="h-3 w-3" />
            </Link>
          </div>
          <div className="space-y-2">
            {mockArrivals.map((a) => (
              <div key={a.id} className="flex items-center gap-3 rounded bg-slate-800/50 p-2">
                <div className={`flex h-8 w-8 items-center justify-center rounded-full ${a.vip ? "bg-amber-500/10" : "bg-slate-700"}`}>
                  <Users className={`h-4 w-4 ${a.vip ? "text-amber-400" : "text-slate-400"}`} />
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-white truncate">{a.guest}{a.vip && <Star className="inline h-3 w-3 text-amber-400 ml-1" />}</p>
                  <p className="text-xs text-slate-400">{a.type} · {a.time}</p>
                </div>
                {a.room ? (
                  <span className="rounded bg-slate-700 px-2 py-0.5 text-xs text-white">{a.room}</span>
                ) : (
                  <span className="rounded bg-rose-500/10 px-2 py-0.5 text-xs text-rose-400">Unassigned</span>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Departures */}
        <div className="card space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold text-white flex items-center gap-2">
              <DoorOpen className="h-4 w-4 text-amber-400" />
              Departures Today
            </h3>
            <Link href="/reservations" className="text-xs text-nexus-400 hover:text-nexus-300 flex items-center gap-1">
              View All <ChevronRight className="h-3 w-3" />
            </Link>
          </div>
          <div className="space-y-2">
            {mockDepartures.map((d) => (
              <div key={d.id} className="flex items-center gap-3 rounded bg-slate-800/50 p-2">
                <div className={`flex h-8 w-8 items-center justify-center rounded-full ${d.balance > 0 ? "bg-amber-500/10" : "bg-emerald-500/10"}`}>
                  {d.balance > 0 ? <DollarSign className="h-4 w-4 text-amber-400" /> : <CheckCircle2 className="h-4 w-4 text-emerald-400" />}
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-white truncate">{d.guest}</p>
                  <p className="text-xs text-slate-400">{d.room} · {d.time}</p>
                </div>
                {d.balance > 0 ? (
                  <span className="rounded bg-amber-500/10 px-2 py-0.5 text-xs text-amber-400">${d.balance}</span>
                ) : (
                  <span className="rounded bg-emerald-500/10 px-2 py-0.5 text-xs text-emerald-400">Paid</span>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Revenue Chart */}
        <div className="card space-y-3">
          <div className="flex items-center justify-between">
            <h3 className="text-sm font-semibold text-white flex items-center gap-2">
              <TrendingUp className="h-4 w-4 text-nexus-400" />
              Weekly Revenue
            </h3>
            <span className="text-xs text-slate-400">${mockWeeklyRevenue.reduce((s, d) => s + d.revenue, 0).toLocaleString()}</span>
          </div>
          <div className="flex items-end gap-1 h-32">
            {mockWeeklyRevenue.map((d, i) => {
              const maxRev = Math.max(...mockWeeklyRevenue.map(x => x.revenue));
              const height = (d.revenue / maxRev) * 100;
              const isToday = d.day === "Sun";
              return (
                <div key={i} className="flex-1 flex flex-col items-center gap-1">
                  <div className="text-[10px] text-slate-400">${(d.revenue / 1000).toFixed(1)}k</div>
                  <div
                    className={`w-full rounded-t ${isToday ? "bg-nexus-500" : "bg-slate-700"} hover:opacity-80 transition-opacity cursor-pointer`}
                    style={{ height: `${height}%` }}
                  />
                  <div className={`text-[10px] ${isToday ? "text-nexus-400 font-medium" : "text-slate-500"}`}>{d.day}</div>
                </div>
              );
            })}
          </div>
        </div>
      </div>

      {/* Bottom Row */}
      <div className="grid grid-cols-2 gap-4">
        {/* Recent Activity */}
        <div className="card space-y-3">
          <h3 className="text-sm font-semibold text-white flex items-center gap-2">
            <Activity className="h-4 w-4 text-slate-400" />
            Recent Activity
          </h3>
          <div className="space-y-2">
            {mockRecentActivity.map((act, i) => (
              <div key={i} className="flex items-center gap-3 rounded bg-slate-800/50 p-2">
                <div className={`flex h-8 w-8 items-center justify-center rounded-full bg-${act.color}-500/10`}>
                  <act.icon className={`h-4 w-4 text-${act.color}-400`} />
                </div>
                <div className="flex-1">
                  <p className="text-sm text-white">{act.event}</p>
                  <p className="text-xs text-slate-400">{act.detail}</p>
                </div>
                <span className="text-xs text-slate-500">{act.time}</span>
              </div>
            ))}
          </div>
        </div>

        {/* Quick Actions */}
        <div className="card space-y-3">
          <h3 className="text-sm font-semibold text-white flex items-center gap-2">
            <Plus className="h-4 w-4 text-nexus-400" />
            Quick Actions
          </h3>
          <div className="grid grid-cols-2 gap-2">
            <Link href="/reservations" className="flex items-center gap-2 rounded-lg border border-slate-700 bg-slate-800 p-3 text-sm text-white hover:border-nexus-500 hover:bg-slate-700 transition-colors">
              <Users className="h-4 w-4 text-nexus-400" />
              View All Reservations
            </Link>
            <Link href="/housekeeping" className="flex items-center gap-2 rounded-lg border border-slate-700 bg-slate-800 p-3 text-sm text-white hover:border-nexus-500 hover:bg-slate-700 transition-colors">
              <BedDouble className="h-4 w-4 text-emerald-400" />
              Housekeeping Board
            </Link>
            <Link href="/locks" className="flex items-center gap-2 rounded-lg border border-slate-700 bg-slate-800 p-3 text-sm text-white hover:border-nexus-500 hover:bg-slate-700 transition-colors">
              <KeyRound className="h-4 w-4 text-violet-400" />
              Smart Locks
            </Link>
            <Link href="/iptv" className="flex items-center gap-2 rounded-lg border border-slate-700 bg-slate-800 p-3 text-sm text-white hover:border-nexus-500 hover:bg-slate-700 transition-colors">
              <Wifi className="h-4 w-4 text-sky-400" />
              IPTV Control
            </Link>
            <Link href="/revenue" className="flex items-center gap-2 rounded-lg border border-slate-700 bg-slate-800 p-3 text-sm text-white hover:border-nexus-500 hover:bg-slate-700 transition-colors">
              <DollarSign className="h-4 w-4 text-amber-400" />
              Revenue Dashboard
            </Link>
            <Link href="/command" className="flex items-center gap-2 rounded-lg border border-slate-700 bg-slate-800 p-3 text-sm text-white hover:border-nexus-500 hover:bg-slate-700 transition-colors">
              <Activity className="h-4 w-4 text-rose-400" />
              Command Center
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
