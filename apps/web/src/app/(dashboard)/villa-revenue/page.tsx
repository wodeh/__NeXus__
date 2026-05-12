"use client";

import { useState, useEffect, useCallback } from "react";
import {
  TrendingUp, DollarSign, Home, Users, Calendar, RefreshCw, AlertTriangle,
  ArrowUpRight, BarChart3, PieChart,
} from "lucide-react";
import {
  getVillaRevenue, getVillas,
  VillaRevenueStats, VillaProperty,
} from "@/lib/api";

export default function VillaRevenuePage() {
  const [stats, setStats] = useState<VillaRevenueStats | null>(null);
  const [villas, setVillas] = useState<VillaProperty[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [period, setPeriod] = useState<"7d" | "30d" | "90d" | "year">("30d");

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const end = new Date().toISOString().split("T")[0];
      const start = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split("T")[0];
      const [s, v] = await Promise.all([
        getVillaRevenue(start, end),
        getVillas(),
      ]);
      setStats(s);
      setVillas(v);
    } catch (e: any) {
      setError(e.message || "Failed to load");
    } finally {
      setLoading(false);
    }
  }, [period]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <TrendingUp className="h-6 w-6 text-nexus-400" /> Revenue
          </h1>
          <p className="text-sm text-slate-400">Villa rental performance · Occupancy · Down payments</p>
        </div>
        <div className="flex items-center gap-2">
          {(["7d", "30d", "90d", "year"] as const).map((p) => (
            <button
              key={p}
              onClick={() => setPeriod(p)}
              className={`rounded-lg px-3 py-1.5 text-xs font-medium ${
                period === p ? "bg-slate-700 text-white" : "text-slate-400 hover:bg-slate-800 hover:text-white"
              }`}
            >
              {p}
            </button>
          ))}
        </div>
      </div>

      {/* KPIs */}
      <div className="grid grid-cols-4 gap-4">
        <KpiCard label="Total Revenue" value={`$${Math.round(stats?.total_revenue || 0).toLocaleString()}`} change="+12.3%" icon={DollarSign} color="nexus" />
        <KpiCard label="Reservations" value={(stats?.total_reservations || 0).toString()} change="+8" icon={Calendar} color="emerald" />
        <KpiCard label="Occupancy" value={`${Math.round(stats?.occupancy_rate || 0)}%`} change="+4.2%" icon={Home} color="amber" />
        <KpiCard label="Avg Booking" value={`$${Math.round(stats?.avg_booking_value || 0).toLocaleString()}`} change="+5.1%" icon={Users} color="violet" />
      </div>

      {/* Revenue Breakdown */}
      <div className="grid grid-cols-2 gap-4">
        <div className="card space-y-4">
          <h3 className="text-lg font-semibold text-white">By Villa</h3>
          <div className="space-y-3">
            {stats?.villa_breakdown?.map((vb) => (
              <div key={vb.villa_id} className="space-y-1">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-white">{vb.villa_name}</span>
                  <span className="text-slate-400">${Math.round(vb.revenue).toLocaleString()}</span>
                </div>
                <div className="h-2 w-full rounded-full bg-slate-800">
                  <div
                    className="h-full rounded-full bg-nexus-500"
                    style={{ width: `${Math.min(vb.occupancy_pct, 100)}%` }}
                  />
                </div>
                <div className="flex items-center justify-between text-xs text-slate-500">
                  <span>{vb.nights_booked} nights</span>
                  <span>{Math.round(vb.occupancy_pct)}% occupancy</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="card space-y-4">
          <h3 className="text-lg font-semibold text-white">Financial Summary</h3>
          <div className="space-y-3">
            <div className="flex items-center justify-between rounded bg-slate-800/50 p-3">
              <span className="text-sm text-slate-400">Down Payments Received</span>
              <span className="text-lg font-bold text-emerald-400">
                ${Math.round(stats?.down_payment_total || 0).toLocaleString()}
              </span>
            </div>
            <div className="flex items-center justify-between rounded bg-slate-800/50 p-3">
              <span className="text-sm text-slate-400">Pending Balance</span>
              <span className="text-lg font-bold text-amber-400">
                ${Math.round(stats?.pending_balance || 0).toLocaleString()}
              </span>
            </div>
            <div className="flex items-center justify-between rounded bg-slate-800/50 p-3">
              <span className="text-sm text-slate-400">Collected Ratio</span>
              <span className="text-lg font-bold text-white">
                {stats && stats.total_revenue && stats.total_revenue > 0
                  ? `${Math.round(((stats.down_payment_total || 0) / stats.total_revenue) * 100)}%`
                  : "0%"}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

function KpiCard({ label, value, change, icon: Icon, color }: { label: string; value: string; change: string; icon: React.ElementType; color: string }) {
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
      <div className="flex items-center gap-1 text-xs text-emerald-400">
        <ArrowUpRight className="h-3 w-3" />{change} vs last period
      </div>
    </div>
  );
}
