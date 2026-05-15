// @ts-nocheck
"use client";

import { useState, useEffect, useMemo } from "react";
import Link from "next/link";
import {
  Home,
  CalendarDays,
  DollarSign,
  TrendingUp,
  TrendingDown,
  Users,
  CheckCircle2,
  Clock,
  AlertTriangle,
  Plus,
  RefreshCw,
  Star,
  MapPin,
  Bed,
  ArrowRightLeft,
  Smartphone,
} from "lucide-react";
import { useAuth } from "@/lib/auth";
import {
  getVillas,
  getVillaReservations,
  getVillaRevenue,
  VillaProperty,
  VillaReservation,
} from "@/lib/api";

export default function VillaDashboardPage() {
  const { user } = useAuth();
  const [villas, setVillas] = useState<VillaProperty[]>([]);
  const [reservations, setReservations] = useState<VillaReservation[]>([]);
  const [revenue, setRevenue] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      const [v, r, rev] = await Promise.all([
        getVillas(),
        getVillaReservations(),
        getVillaRevenue("2024-01-01", "2024-12-31"),
      ]);
      setVillas(v);
      setReservations(r);
      setRevenue(rev);
    } catch (e: any) {
      setError(e.message || "Failed to load villa data");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, []);

  const stats = useMemo(() => {
    const totalVillas = villas.length;
    const activeVillas = villas.filter((v) => v.is_active).length;
    const totalReservations = reservations.length;
    const pendingReservations = reservations.filter((r) => r.status === "pending").length;
    const confirmedReservations = reservations.filter((r) => r.status === "confirmed" || r.status === "reserved").length;
    const checkedIn = reservations.filter((r) => r.status === "checked_in").length;
    const totalRevenue = revenue?.total_revenue || reservations.reduce((sum, r) => sum + (r.total_amount || 0), 0);
    const occupancyRate = totalVillas ? Math.round((checkedIn / totalVillas) * 100) : 0;

    return {
      totalVillas,
      activeVillas,
      totalReservations,
      pendingReservations,
      confirmedReservations,
      checkedIn,
      totalRevenue,
      occupancyRate,
    };
  }, [villas, reservations, revenue]);

  const recentReservations = useMemo(() => {
    return reservations.slice(0, 5).map((r) => ({
      id: r.id,
      guest: r.guest_name,
      villa: r.villa_name,
      checkIn: r.check_in_date,
      checkOut: r.check_out_date,
      status: r.status,
      amount: r.total_amount,
      nights: r.nights,
      source: r.source,
    }));
  }, [reservations]);

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
          <h1 className="text-2xl font-bold text-white">Villa Owner Dashboard</h1>
          <p className="text-sm text-slate-400">Welcome back, {user?.name || "Owner"} · Manage your villa portfolio</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={fetchData} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white" title="Refresh">
            <RefreshCw className="h-4 w-4" />
          </button>
          <Link href="/villa-reservations/new" className="btn-primary gap-2">
            <Plus className="h-4 w-4" />
            New Booking
          </Link>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard label="Total Villas" value={stats.totalVillas} sub={`${stats.activeVillas} active`} icon={Home} color="sky" />
        <StatCard label="Total Bookings" value={stats.totalReservations} sub={`${stats.confirmedReservations} confirmed`} icon={CalendarDays} color="emerald" />
        <StatCard label="Occupancy" value={`${stats.occupancyRate}%`} sub={`${stats.checkedIn} checked in`} icon={Bed} color="violet" />
        <StatCard label="Total Revenue" value={`$${Math.round(stats.totalRevenue).toLocaleString()}`} sub="All time" icon={DollarSign} color="amber" trend="up" />
      </div>

      <div className="grid gap-6 lg:grid-cols-3">
        {/* Recent Bookings */}
        <div className="card lg:col-span-2">
          <div className="mb-4 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <ArrowRightLeft className="h-5 w-5 text-nexus-400" />
              <h3 className="font-bold text-white">Recent Bookings</h3>
              <span className="badge badge-blue">{reservations.length}</span>
            </div>
            <Link href="/villa-reservations" className="text-xs text-nexus-400 hover:underline">View all</Link>
          </div>
          <div className="space-y-3">
            {recentReservations.length === 0 && <p className="text-sm text-slate-500">No bookings yet.</p>}
            {recentReservations.map((r) => (
              <div key={r.id} className="flex items-center justify-between rounded-lg border border-slate-700/50 bg-slate-800/50 p-3">
                <div className="flex items-center gap-3">
                  <div className="flex h-8 w-8 items-center justify-center rounded-full bg-nexus-500/10">
                    <Users className="h-4 w-4 text-nexus-400" />
                  </div>
                  <div>
                    <p className="text-sm font-medium text-white">{r.guest}</p>
                    <p className="text-xs text-slate-400">{r.villa} · {r.checkIn} → {r.checkOut} · {r.nights} nights</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <span className={`badge ${r.status === "pending" ? "badge-amber" : r.status === "reserved" ? "badge-purple" : "badge-green"}`}>
                    {r.status}
                  </span>
                  <span className="text-xs text-slate-400">${r.amount}</span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Quick Actions + Status */}
        <div className="space-y-6">
          <div className="card">
            <h3 className="mb-4 font-bold text-white">Quick Actions</h3>
            <div className="space-y-2">
              <Link href="/villas" className="flex items-center gap-2 rounded-lg bg-slate-800/50 p-3 text-sm text-white hover:bg-slate-800">
                <Home className="h-4 w-4 text-sky-400" /> Manage Villas
              </Link>
              <Link href="/villa-reservations" className="flex items-center gap-2 rounded-lg bg-slate-800/50 p-3 text-sm text-white hover:bg-slate-800">
                <CalendarDays className="h-4 w-4 text-emerald-400" /> View Bookings
              </Link>
              <Link href="/villa-revenue" className="flex items-center gap-2 rounded-lg bg-slate-800/50 p-3 text-sm text-white hover:bg-slate-800">
                <DollarSign className="h-4 w-4 text-amber-400" /> Revenue Report
              </Link>
              <Link href="/cleaner-tracking" className="flex items-center gap-2 rounded-lg bg-slate-800/50 p-3 text-sm text-white hover:bg-slate-800">
                <Smartphone className="h-4 w-4 text-violet-400" /> Cleaner Tracking
              </Link>
            </div>
          </div>

          <div className="card">
            <h3 className="mb-4 font-bold text-white">Villa Status</h3>
            <div className="space-y-2">
              {villas.slice(0, 4).map((v) => (
                <div key={v.id} className="flex items-center justify-between text-sm">
                  <div className="flex items-center gap-2">
                    <Home className="h-4 w-4 text-slate-400" />
                    <span className="text-white">{v.name}</span>
                  </div>
                  <span className={`rounded px-2 py-0.5 text-xs ${v.status === "available" ? "bg-emerald-500/10 text-emerald-400" : v.status === "reserved" ? "bg-amber-500/10 text-amber-400" : "bg-rose-500/10 text-rose-400"}`}>
                    {v.status}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>

      {/* Revenue Breakdown */}
      {revenue?.villa_breakdown && (
        <div className="card">
          <div className="mb-4 flex items-center gap-2">
            <TrendingUp className="h-5 w-5 text-emerald-400" />
            <h3 className="font-bold text-white">Revenue by Villa</h3>
          </div>
          <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
            {revenue.villa_breakdown.map((v: any) => (
              <div key={v.villa_id} className="rounded-lg border border-slate-700 bg-slate-800/50 p-4">
                <p className="text-sm font-medium text-white">{v.villa_name}</p>
                <p className="mt-1 text-2xl font-bold text-emerald-400">${v.revenue.toLocaleString()}</p>
                <p className="text-xs text-slate-400">{v.nights_booked} nights · {v.occupancy_pct}% occupancy</p>
              </div>
            ))}
          </div>
        </div>
      )}
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
        {trend && (trend === "up" ? <TrendingUp className="h-4 w-4 text-emerald-400" /> : <TrendingDown className="h-4 w-4 text-rose-400" />)}
      </div>
      <div className="mt-2">
        <div className="text-2xl font-bold text-white">{value}</div>
        <div className="text-xs text-slate-400">{sub}</div>
      </div>
    </div>
  );
}
