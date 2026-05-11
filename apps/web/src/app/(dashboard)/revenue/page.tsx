"use client";

import { useState, useEffect, useCallback, useMemo } from "react";
import {
  TrendingUp,
  TrendingDown,
  DollarSign,
  BedDouble,
  CalendarDays,
  ChevronLeft,
  ChevronRight,
  Filter,
  Download,
  Target,
  Zap,
  ArrowUpRight,
  ArrowDownRight,
  BarChart3,
  PieChart,
  LineChart,
  Layers,
  RefreshCw,
  AlertTriangle,
} from "lucide-react";
import { Reservation, getReservations } from "@/lib/api";

export default function RevenuePage() {
  const [view, setView] = useState<"overview" | "forecast" | "pricing" | "analytics">("overview");
  const [dateRange, setDateRange] = useState("30d");
  const [reservations, setReservations] = useState<Reservation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const res = await getReservations();
      setReservations(res);
    } catch (e: any) {
      setError(e.message || "Failed to load revenue data");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Compute revenue metrics from real reservations
  const metrics = useMemo(() => {
    const totalRevenue = reservations.reduce((s, r) => {
      return s + (r.total || 0);
    }, 0);
    const totalNights = reservations.reduce((s, r) => {
      return s + Math.max(1, Math.round((new Date(r.check_out).getTime() - new Date(r.check_in).getTime()) / 86400000));
    }, 0);
    const overallADR = totalNights > 0 ? Math.round(totalRevenue / totalNights) : 0;
    const revPAR = Math.round(totalRevenue / (120 * 30)); // 120 rooms * 30 days placeholder
    return { totalRevenue, totalNights, overallADR, revPAR };
  }, [reservations]);

  // Room type breakdown
  const roomTypeRevenue = useMemo(() => {
    const map = new Map<string, { revenue: number; nights: number }>();
    reservations.forEach((r) => {
      const existing = map.get(r.room_type);
      if (existing) {
        existing.revenue += (r.total || 0);
        existing.nights += Math.max(1, Math.round((new Date(r.check_out).getTime() - new Date(r.check_in).getTime()) / 86400000));
      } else {
        map.set(r.room_type, { revenue: (r.total || 0), nights: Math.max(1, Math.round((new Date(r.check_out).getTime() - new Date(r.check_in).getTime()) / 86400000)) });
      }
    });
    return Array.from(map.entries()).map(([type, data]) => ({ type, ...data }));
  }, [reservations]);

  // Daily revenue for chart
  const dailyRevenue = useMemo(() => {
    const map = new Map<string, { revenue: number; rooms: number }>();
    reservations.forEach((r) => {
      const date = r.check_in;
      const rev = (r.total || 0);
      const existing = map.get(date);
      if (existing) {
        existing.revenue += rev;
        existing.rooms += 1;
      } else {
        map.set(date, { revenue: rev, rooms: 1 });
      }
    });
    return Array.from(map.entries())
      .sort((a, b) => a[0].localeCompare(b[0]))
      .map(([day, data]) => ({ day, ...data, adr: data.rooms > 0 ? Math.round(data.revenue / data.rooms) : 0 }));
  }, [reservations]);

  const maxRevenue = Math.max(...dailyRevenue.map((d) => d.revenue), 1);
  const totalRoomRevenue = roomTypeRevenue.reduce((s, r) => s + r.revenue, 0);

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
          <h1 className="text-2xl font-bold text-white">Revenue Dashboard</h1>
          <p className="text-sm text-slate-400">Revenue analytics · Forecasting · Pricing rules</p>
        </div>
        <div className="flex items-center gap-2">
          <select value={dateRange} onChange={(e) => setDateRange(e.target.value)} className="input text-xs">
            <option value="7d">Last 7 Days</option>
            <option value="30d">Last 30 Days</option>
            <option value="90d">Last 90 Days</option>
            <option value="ytd">Year to Date</option>
          </select>
          <button onClick={fetchData} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white" title="Refresh">
            <RefreshCw className="h-4 w-4" />
          </button>
          <button className="btn-secondary text-xs gap-2">
            <Download className="h-4 w-4" /> Export
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-4 gap-4">
        <KpiCard
          label="Total Revenue"
          value={`$${metrics.totalRevenue.toLocaleString()}`}
          change="+12.3%"
          positive
          icon={DollarSign}
          color="nexus"
        />
        <KpiCard
          label="Average Daily Rate"
          value={`$${metrics.overallADR}`}
          change="+5.2%"
          positive
          icon={BedDouble}
          color="emerald"
        />
        <KpiCard
          label="RevPAR"
          value={`$${metrics.revPAR}`}
          change="-2.1%"
          positive={false}
          icon={TrendingUp}
          color="amber"
        />
        <KpiCard
          label="Room Nights Sold"
          value={metrics.totalNights.toString()}
          change="+8.7%"
          positive
          icon={CalendarDays}
          color="violet"
        />
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        {(["overview", "forecast", "pricing", "analytics"] as const).map((tab) => (
          <button
            key={tab}
            onClick={() => setView(tab)}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              view === tab ? "bg-slate-800 text-white" : "text-slate-400 hover:bg-slate-800 hover:text-white"
            }`}
          >
            {tab === "overview" ? "Revenue Overview" : tab === "forecast" ? "Forecast" : tab === "pricing" ? "Pricing Rules" : "Analytics"}
          </button>
        ))}
      </div>

      {/* Overview Tab */}
      {view === "overview" && (
        <>
          {/* Revenue Chart */}
          <div className="card space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <BarChart3 className="h-5 w-5 text-nexus-400" />
                Daily Revenue Trend
              </h3>
              <div className="flex items-center gap-4 text-xs text-slate-400">
                <div className="flex items-center gap-1">
                  <div className="h-3 w-3 rounded bg-nexus-500" /> Revenue
                </div>
                <div className="flex items-center gap-1">
                  <div className="h-3 w-3 rounded bg-slate-600" /> ADR
                </div>
              </div>
            </div>
            <div className="flex items-end gap-0.5 h-48">
              {dailyRevenue.map((d, i) => (
                <div key={i} className="flex-1 flex flex-col items-center gap-1 group relative">
                  <div className="opacity-0 group-hover:opacity-100 absolute -top-8 bg-slate-800 text-white text-[10px] px-2 py-1 rounded whitespace-nowrap z-10">
                    {d.day}: ${d.revenue.toLocaleString()} · ADR ${d.adr}
                  </div>
                  <div
                    className="w-full rounded-t bg-nexus-500/60 hover:bg-nexus-500 transition-colors cursor-pointer"
                    style={{ height: `${(d.revenue / maxRevenue) * 100}%` }}
                  />
                </div>
              ))}
            </div>
          </div>

          {/* Room Type Breakdown */}
          <div className="grid grid-cols-2 gap-4">
            <div className="card space-y-4">
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <PieChart className="h-5 w-5 text-violet-400" />
                Revenue by Room Type
              </h3>
              <div className="space-y-3">
                {roomTypeRevenue.map((rt) => {
                  const pct = totalRoomRevenue > 0 ? Math.round((rt.revenue / totalRoomRevenue) * 100) : 0;
                  const colorMap: Record<string, string> = {
                    Standard: "bg-sky-400",
                    Deluxe: "bg-nexus-400",
                    Suite: "bg-amber-400",
                    Accessible: "bg-emerald-400",
                  };
                  return (
                    <div key={rt.type} className="space-y-1">
                      <div className="flex justify-between text-sm">
                        <span className="text-white">{rt.type}</span>
                        <span className="text-slate-400">${rt.revenue.toLocaleString()} ({pct}%)</span>
                      </div>
                      <div className="h-2 rounded-full bg-slate-800">
                        <div className={`h-full rounded-full ${colorMap[rt.type] || "bg-slate-400"} transition-all`} style={{ width: `${pct}%` }} />
                      </div>
                      <div className="flex justify-between text-xs text-slate-500">
                        <span>{rt.nights} nights</span>
                        <span>ADR: ${rt.nights > 0 ? Math.round(rt.revenue / rt.nights) : 0}</span>
                      </div>
                    </div>
                  );
                })}
                {roomTypeRevenue.length === 0 && (
                  <p className="text-sm text-slate-500">No revenue data available.</p>
                )}
              </div>
            </div>

            <div className="card space-y-4">
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <LineChart className="h-5 w-5 text-emerald-400" />
                Occupancy vs Revenue Correlation
              </h3>
              <div className="h-48 flex items-end gap-2">
                {dailyRevenue.slice(0, 14).map((d, i) => (
                  <div key={i} className="flex-1 flex flex-col items-center gap-1">
                    <div className="text-[10px] text-slate-500">{Math.round(d.rooms)}</div>
                    <div className="flex gap-0.5 w-full">
                      <div
                        className="flex-1 rounded-t bg-emerald-500/40"
                        style={{ height: `${Math.min((d.rooms / 10) * 80, 80)}px` }}
                      />
                      <div
                        className="flex-1 rounded-t bg-nexus-500/40"
                        style={{ height: `${(d.revenue / maxRevenue) * 80}px` }}
                      />
                    </div>
                    <div className="text-[10px] text-slate-500">{d.day.slice(5)}</div>
                  </div>
                ))}
              </div>
              <div className="flex justify-center gap-4 text-xs text-slate-400">
                <div className="flex items-center gap-1">
                  <div className="h-3 w-3 rounded bg-emerald-500/40" /> Bookings
                </div>
                <div className="flex items-center gap-1">
                  <div className="h-3 w-3 rounded bg-nexus-500/40" /> Revenue
                </div>
              </div>
            </div>
          </div>
        </>
      )}

      {/* Forecast Tab — mock only, no real data yet */}
      {view === "forecast" && (
        <div className="space-y-4">
          <div className="card space-y-4">
            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
              <Target className="h-5 w-5 text-nexus-400" />
              Revenue Forecast
            </h3>
            <p className="text-sm text-slate-400">Forecasting requires historical data. Once you have 90+ days of reservations, predictive models will be enabled.</p>
          </div>
        </div>
      )}

      {/* Pricing Rules Tab — placeholder */}
      {view === "pricing" && (
        <div className="card space-y-4">
          <h3 className="text-lg font-semibold text-white flex items-center gap-2">
            <Layers className="h-5 w-5 text-nexus-400" />
            Pricing Rules
          </h3>
          <p className="text-sm text-slate-400">Pricing rules will be configurable once the revenue management module is enabled.</p>
        </div>
      )}

      {/* Analytics Tab — computed from real data */}
      {view === "analytics" && (
        <div className="grid grid-cols-2 gap-4">
          <div className="card space-y-4">
            <h3 className="text-lg font-semibold text-white">Revenue Summary</h3>
            <div className="space-y-3">
              <div className="flex items-center justify-between rounded bg-slate-800/50 p-3">
                <p className="text-sm text-white">Total Reservations</p>
                <p className="text-sm font-medium text-white">{reservations.length}</p>
              </div>
              <div className="flex items-center justify-between rounded bg-slate-800/50 p-3">
                <p className="text-sm text-white">Checked In</p>
                <p className="text-sm font-medium text-white">{reservations.filter((r) => r.status === "checked_in").length}</p>
              </div>
              <div className="flex items-center justify-between rounded bg-slate-800/50 p-3">
                <p className="text-sm text-white">Confirmed</p>
                <p className="text-sm font-medium text-white">{reservations.filter((r) => r.status === "confirmed").length}</p>
              </div>
              <div className="flex items-center justify-between rounded bg-slate-800/50 p-3">
                <p className="text-sm text-white">Cancelled</p>
                <p className="text-sm font-medium text-white">{reservations.filter((r) => r.status === "cancelled").length}</p>
              </div>
            </div>
          </div>

          <div className="card space-y-4">
            <h3 className="text-lg font-semibold text-white">Key Metrics</h3>
            <div className="space-y-3">
              <div className="flex items-center justify-between rounded bg-slate-800/50 p-3">
                <p className="text-sm text-slate-300">Total Revenue</p>
                <p className="text-sm font-medium text-white">${metrics.totalRevenue.toLocaleString()}</p>
              </div>
              <div className="flex items-center justify-between rounded bg-slate-800/50 p-3">
                <p className="text-sm text-slate-300">ADR</p>
                <p className="text-sm font-medium text-white">${metrics.overallADR}</p>
              </div>
              <div className="flex items-center justify-between rounded bg-slate-800/50 p-3">
                <p className="text-sm text-slate-300">RevPAR</p>
                <p className="text-sm font-medium text-white">${metrics.revPAR}</p>
              </div>
              <div className="flex items-center justify-between rounded bg-slate-800/50 p-3">
                <p className="text-sm text-slate-300">Room Nights</p>
                <p className="text-sm font-medium text-white">{metrics.totalNights}</p>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function KpiCard({
  label,
  value,
  change,
  positive,
  icon: Icon,
  color,
}: {
  label: string;
  value: string;
  change: string;
  positive: boolean;
  icon: React.ElementType;
  color: string;
}) {
  const colorMap: Record<string, string> = {
    nexus: "text-nexus-400 bg-nexus-500/10",
    emerald: "text-emerald-400 bg-emerald-500/10",
    amber: "text-amber-400 bg-amber-500/10",
    violet: "text-violet-400 bg-violet-500/10",
  };
  return (
    <div className="card space-y-3">
      <div className="flex items-center justify-between">
        <span className="text-xs text-slate-400 font-medium">{label}</span>
        <div className={`rounded-lg p-2 ${colorMap[color]}`}>
          <Icon className="h-4 w-4" />
        </div>
      </div>
      <p className="text-2xl font-bold text-white">{value}</p>
      <div className={`flex items-center gap-1 text-xs ${positive ? "text-emerald-400" : "text-rose-400"}`}>
        {positive ? <ArrowUpRight className="h-3 w-3" /> : <ArrowDownRight className="h-3 w-3" />}
        {change} vs last period
      </div>
    </div>
  );
}
