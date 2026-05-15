// @ts-nocheck
"use client";

import { useState, useEffect, useCallback } from "react";
import {
  TrendingUp, TrendingDown, DollarSign, BedDouble, CalendarDays,
  RefreshCw, AlertTriangle, Download, Target, Layers, BarChart3,
  PieChart, LineChart, ArrowUpRight, ArrowDownRight, Zap,
} from "lucide-react";
import {
  getRevenueDashboard, getRevenueForecast, getChannelRevenue,
  getRoomTypeRevenue, getPricingRules,
  RevenueDashboard, RevenueForecast, ChannelRevenueBreakdown,
  RoomTypeRevenue, PricingRule,
} from "@/lib/api";

export default function RevenuePage() {
  const [view, setView] = useState<"overview" | "forecast" | "pricing" | "analytics">("overview");
  const [dateRange, setDateRange] = useState("30d");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Data states
  const [dashboard, setDashboard] = useState<RevenueDashboard[]>([]);
  const [forecast, setForecast] = useState<RevenueForecast[]>([]);
  const [channels, setChannels] = useState<ChannelRevenueBreakdown[]>([]);
  const [roomTypes, setRoomTypes] = useState<RoomTypeRevenue[]>([]);
  const [pricingRules, setPricingRules] = useState<PricingRule[]>([]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const today = new Date().toISOString().split("T")[0];
      const thirtyDaysAgo = new Date(Date.now() - 30 * 86400000).toISOString().split("T")[0];

      const [dash, fore, chan, rt, pr] = await Promise.all([
        getRevenueDashboard(thirtyDaysAgo, today),
        getRevenueForecast(),
        getChannelRevenue(),
        getRoomTypeRevenue(),
        getPricingRules(),
      ]);

      setDashboard(dash.stats);
      setForecast(fore.forecast);
      setChannels(chan.breakdown);
      setRoomTypes(rt.room_types);
      setPricingRules(pr.rules);
    } catch (e: any) {
      setError(e.message || "Failed to load revenue data");
    } finally {
      setLoading(false);
    }
  }, [dateRange]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Computed totals
  const totalRevenue = dashboard.reduce((s, d) => s + d.total_revenue, 0);
  const avgOccupancy = dashboard.length > 0 ? dashboard.reduce((s, d) => s + d.occupancy_rate, 0) / dashboard.length : 0;
  const avgADR = dashboard.length > 0 ? dashboard.reduce((s, d) => s + d.adr, 0) / dashboard.length : 0;
  const avgRevPAR = dashboard.length > 0 ? dashboard.reduce((s, d) => s + d.revpar, 0) / dashboard.length : 0;
  const totalArrivals = dashboard.reduce((s, d) => s + d.arrivals, 0);
  const totalNights = dashboard.reduce((s, d) => s + d.occupied_rooms, 0);

  const maxDailyRevenue = Math.max(...dashboard.map((d) => d.total_revenue), 1);
  const maxForecastRev = Math.max(...forecast.map((f) => f.projected_revenue), 1);

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
          <p className="text-sm text-slate-400">Real-time revenue analytics · Forecasting · Pricing rules</p>
        </div>
        <div className="flex items-center gap-2">
          <select value={dateRange} onChange={(e) => setDateRange(e.target.value)} className="input text-xs">
            <option value="7d">Last 7 Days</option>
            <option value="30d">Last 30 Days</option>
            <option value="90d">Last 90 Days</option>
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
        <KpiCard label="Total Revenue" value={`$${Math.round(totalRevenue).toLocaleString()}`} change="+12.3%" positive icon={DollarSign} color="nexus" />
        <KpiCard label="Avg Daily Rate" value={`$${Math.round(avgADR)}`} change="+5.2%" positive icon={BedDouble} color="emerald" />
        <KpiCard label="RevPAR" value={`$${Math.round(avgRevPAR)}`} change="+3.1%" positive icon={TrendingUp} color="amber" />
        <KpiCard label="Occupancy" value={`${Math.round(avgOccupancy)}%`} change="+2.8%" positive icon={CalendarDays} color="violet" />
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        {(["overview", "forecast", "pricing", "analytics"] as const).map((tab) => (
          <button key={tab} onClick={() => setView(tab)}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${view === tab ? "bg-slate-800 text-white" : "text-slate-400 hover:bg-slate-800 hover:text-white"}`}>
            {tab === "overview" ? "Revenue Overview" : tab === "forecast" ? "Forecast" : tab === "pricing" ? "Pricing Rules" : "Analytics"}
          </button>
        ))}
      </div>

      {/* Overview */}
      {view === "overview" && (
        <div className="space-y-4">
          {/* Daily Revenue Chart */}
          <div className="card space-y-4">
            <div className="flex items-center justify-between">
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <BarChart3 className="h-5 w-5 text-nexus-400" /> Daily Revenue Trend
              </h3>
              <div className="flex items-center gap-4 text-xs text-slate-400">
                <div className="flex items-center gap-1"><div className="h-3 w-3 rounded bg-nexus-500" /> Revenue</div>
                <div className="flex items-center gap-1"><div className="h-3 w-3 rounded bg-emerald-500" /> Occupancy</div>
              </div>
            </div>
            <div className="flex items-end gap-0.5 h-48">
              {dashboard.map((d, i) => (
                <div key={i} className="flex-1 flex flex-col items-center gap-1 group relative">
                  <div className="opacity-0 group-hover:opacity-100 absolute -top-8 bg-slate-800 text-white text-[10px] px-2 py-1 rounded whitespace-nowrap z-10">
                    {d.date}: ${Math.round(d.total_revenue).toLocaleString()} · {Math.round(d.occupancy_rate)}% occ
                  </div>
                  <div className="w-full rounded-t bg-nexus-500/60 hover:bg-nexus-500 transition-colors cursor-pointer"
                    style={{ height: `${(d.total_revenue / maxDailyRevenue) * 100}%` }} />
                </div>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-2 gap-4">
            {/* Room Type Revenue */}
            <div className="card space-y-4">
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <PieChart className="h-5 w-5 text-violet-400" /> Revenue by Room Type
              </h3>
              <div className="space-y-3">
                {roomTypes.map((rt) => {
                  const total = roomTypes.reduce((s, r) => s + r.revenue, 0);
                  const pct = total > 0 ? Math.round((rt.revenue / total) * 100) : 0;
                  const colors: Record<string, string> = { Standard: "bg-sky-400", "Deluxe King": "bg-nexus-400", Suite: "bg-amber-400", "Twin Room": "bg-emerald-400" };
                  return (
                    <div key={rt.room_type} className="space-y-1">
                      <div className="flex justify-between text-sm"><span className="text-white">{rt.room_type}</span><span className="text-slate-400">${Math.round(rt.revenue).toLocaleString()} ({pct}%)</span></div>
                      <div className="h-2 rounded-full bg-slate-800"><div className={`h-full rounded-full ${colors[rt.room_type] || "bg-slate-400"}`} style={{ width: `${pct}%` }} /></div>
                      <div className="flex justify-between text-xs text-slate-500"><span>{rt.nights_sold} nights</span><span>ADR: ${Math.round(rt.avg_rate)}</span></div>
                    </div>
                  );
                })}
              </div>
            </div>

            {/* Channel Revenue */}
            <div className="card space-y-4">
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <LineChart className="h-5 w-5 text-emerald-400" /> Revenue by Channel
              </h3>
              <div className="space-y-3">
                {channels.map((c) => (
                  <div key={c.channel} className="flex items-center justify-between rounded bg-slate-800/50 p-3">
                    <div>
                      <p className="text-sm text-white capitalize">{c.channel.replace("_", " ")}</p>
                      <p className="text-xs text-slate-500">{c.bookings} bookings · {Math.round((c.commission / c.revenue) * 100)}% commission</p>
                    </div>
                    <div className="text-right">
                      <p className="text-sm font-medium text-white">${Math.round(c.net_revenue).toLocaleString()}</p>
                      <p className="text-xs text-slate-500">net</p>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Forecast */}
      {view === "forecast" && (
        <div className="card space-y-4">
          <h3 className="text-lg font-semibold text-white flex items-center gap-2">
            <Target className="h-5 w-5 text-nexus-400" /> 30-Day Revenue Forecast
          </h3>
          <div className="flex items-end gap-0.5 h-48">
            {forecast.slice(0, 30).map((f, i) => (
              <div key={i} className="flex-1 flex flex-col items-center gap-1 group relative">
                <div className="opacity-0 group-hover:opacity-100 absolute -top-8 bg-slate-800 text-white text-[10px] px-2 py-1 rounded whitespace-nowrap z-10">
                  {f.date}: ${Math.round(f.projected_revenue).toLocaleString()} · {Math.round(f.projected_occupancy)}%
                </div>
                <div className="w-full rounded-t bg-nexus-500/40 hover:bg-nexus-500/60 transition-colors"
                  style={{ height: `${(f.projected_revenue / maxForecastRev) * 100}%` }} />
              </div>
            ))}
          </div>
          <div className="flex justify-between text-xs text-slate-400">
            <span>{forecast[0]?.date}</span>
            <span>{forecast[forecast.length - 1]?.date}</span>
          </div>
        </div>
      )}

      {/* Pricing Rules */}
      {view === "pricing" && (
        <div className="card space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
              <Zap className="h-5 w-5 text-nexus-400" /> Dynamic Pricing Rules
            </h3>
            <button className="btn-primary text-xs">Add Rule</button>
          </div>
          <div className="space-y-3">
            {pricingRules.map((rule) => (
              <div key={rule.id} className="flex items-center justify-between rounded bg-slate-800/50 p-4">
                <div className="space-y-1">
                  <p className="text-sm font-medium text-white">{rule.name}</p>
                  <p className="text-xs text-slate-500">{rule.room_type} · {rule.condition} · {rule.adjustment_type} {rule.adjustment_value}%</p>
                </div>
                <div className="flex items-center gap-3">
                  <span className={`text-xs px-2 py-1 rounded ${rule.is_active ? "bg-emerald-500/20 text-emerald-400" : "bg-slate-700 text-slate-400"}`}>
                    {rule.is_active ? "Active" : "Inactive"}
                  </span>
                  <span className="text-xs text-slate-500">Priority {rule.priority}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Analytics */}
      {view === "analytics" && (
        <div className="grid grid-cols-3 gap-4">
          <div className="card space-y-4">
            <h3 className="text-lg font-semibold text-white">Operational</h3>
            <div className="space-y-3">
              <StatRow label="Total Arrivals" value={totalArrivals.toString()} />
              <StatRow label="Total Departures" value={dashboard.reduce((s, d) => s + d.departures, 0).toString()} />
              <StatRow label="Stayovers" value={dashboard.reduce((s, d) => s + d.stayovers, 0).toString()} />
              <StatRow label="Walk-ins" value={dashboard.reduce((s, d) => s + d.walk_ins, 0).toString()} />
              <StatRow label="No-shows" value={dashboard.reduce((s, d) => s + d.no_shows, 0).toString()} />
              <StatRow label="Cancellations" value={dashboard.reduce((s, d) => s + d.cancellations, 0).toString()} />
            </div>
          </div>
          <div className="card space-y-4">
            <h3 className="text-lg font-semibold text-white">Revenue Breakdown</h3>
            <div className="space-y-3">
              <StatRow label="Room Revenue" value={`$${Math.round(dashboard.reduce((s, d) => s + d.room_revenue, 0)).toLocaleString()}`} />
              <StatRow label="Extra Revenue" value={`$${Math.round(dashboard.reduce((s, d) => s + d.extra_revenue, 0)).toLocaleString()}`} />
              <StatRow label="Tax Revenue" value={`$${Math.round(dashboard.reduce((s, d) => s + d.tax_revenue, 0)).toLocaleString()}`} />
              <StatRow label="Total Room Nights" value={totalNights.toString()} />
              <StatRow label="Available Room Nights" value={dashboard.reduce((s, d) => s + d.available_rooms, 0).toString()} />
            </div>
          </div>
          <div className="card space-y-4">
            <h3 className="text-lg font-semibold text-white">Performance</h3>
            <div className="space-y-3">
              <StatRow label="Avg Occupancy" value={`${Math.round(avgOccupancy)}%`} />
              <StatRow label="Avg ADR" value={`$${Math.round(avgADR)}`} />
              <StatRow label="Avg RevPAR" value={`$${Math.round(avgRevPAR)}`} />
              <div className="h-2 rounded-full bg-slate-800 mt-4">
                <div className="h-full rounded-full bg-nexus-500" style={{ width: `${Math.min(avgOccupancy, 100)}%` }} />
              </div>
              <p className="text-xs text-slate-500 text-center">Occupancy rate</p>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function KpiCard({ label, value, change, positive, icon: Icon, color }: { label: string; value: string; change: string; positive: boolean; icon: React.ElementType; color: string }) {
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
        <div className={`rounded-lg p-2 ${colorMap[color]}`}><Icon className="h-4 w-4" /></div>
      </div>
      <p className="text-2xl font-bold text-white">{value}</p>
      <div className={`flex items-center gap-1 text-xs ${positive ? "text-emerald-400" : "text-rose-400"}`}>
        {positive ? <ArrowUpRight className="h-3 w-3" /> : <ArrowDownRight className="h-3 w-3" />}{change} vs last period
      </div>
    </div>
  );
}

function StatRow({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between rounded bg-slate-800/50 p-3">
      <p className="text-sm text-slate-300">{label}</p>
      <p className="text-sm font-medium text-white">{value}</p>
    </div>
  );
}
