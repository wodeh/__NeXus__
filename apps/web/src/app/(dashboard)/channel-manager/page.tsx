// @ts-nocheck
"use client";

import { useState, useEffect, useCallback, useMemo } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import { Channel, getChannels, updateChannel, deleteChannel, Reservation, getReservations } from "@/lib/api";
import {
  Globe, RefreshCw, Plus, Trash2, CheckCircle2, AlertTriangle, XCircle,
  DollarSign, Percent, Search, Download, CalendarDays, Users, ArrowUpRight, Loader2
} from "lucide-react";

export default function ChannelManagerPage() {
  const { config } = useTenant();
  const [channels, setChannels] = useState<Channel[]>([]);
  const [reservations, setReservations] = useState<Reservation[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<"overview" | "reservations" | "logs">("overview");
  const [search, setSearch] = useState("");

  const hasCap = hasCapability(config, CAPABILITIES.REVENUE.CHANNEL_MANAGER);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [chData, resData] = await Promise.all([
        getChannels(),
        getReservations()
      ]);
      setChannels(chData);
      setReservations(resData);
    } catch (e: any) {
      setError(e.message || "Failed to load data");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // Compute 30-day metrics from real reservations grouped by source
  const channelMetrics = useMemo(() => {
    const now = new Date();
    const thirtyDaysAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
    
    return channels.map(ch => {
      const chReservations = reservations.filter(r => {
        const rDate = new Date(r.check_in);
        return r.source === ch.source && rDate >= thirtyDaysAgo && rDate <= now;
      });
      
      const bookings30d = chReservations.length;
      const revenue30d = chReservations.reduce((s, r) => s + r.total, 0);
      const commission = Math.round(revenue30d * (ch.commission_pct / 100));
      const netRevenue30d = revenue30d - commission;
      
      return {
        ...ch,
        bookings30d,
        revenue30d,
        netRevenue30d,
        commission
      };
    });
  }, [channels, reservations]);

  const totalBookings = channelMetrics.reduce((s, c) => s + c.bookings30d, 0);
  const totalRevenue = channelMetrics.reduce((s, c) => s + c.revenue30d, 0);
  const totalNet = channelMetrics.reduce((s, c) => s + c.netRevenue30d, 0);
  const totalCommission = channelMetrics.reduce((s, c) => s + c.commission, 0);

  const filteredReservations = useMemo(() => {
    return reservations.filter((r) =>
      r.guest_name.toLowerCase().includes(search.toLowerCase()) ||
      r.id.toLowerCase().includes(search.toLowerCase())
    );
  }, [reservations, search]);

  const toggleChannel = async (id: string, current: boolean) => {
    try {
      await updateChannel(id, { is_active: !current });
      setChannels(prev => prev.map(c => c.id === id ? { ...c, is_active: !current } : c));
    } catch (e: any) {
      setError(e.message || "Failed to update channel");
    }
  };

  const removeChannel = async (id: string) => {
    if (!confirm("Delete this channel?")) return;
    try {
      await deleteChannel(id);
      setChannels(prev => prev.filter(c => c.id !== id));
    } catch (e: any) {
      setError(e.message || "Failed to delete channel");
    }
  };

  if (!hasCap) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <Globe className="h-16 w-16 text-slate-600" />
        <h2 className="mt-4 text-xl font-bold text-white">Channel Manager</h2>
        <p className="mt-2 text-slate-400">Upgrade to Revenue tier to manage OTA channels.</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Channel Manager</h1>
          <p className="text-sm text-slate-400">OTA connections · Commission tracking · Auto-sync</p>
        </div>
        <button className="btn-primary gap-2 text-sm">
          <Plus className="h-4 w-4" /> Add Channel
        </button>
      </div>

      {error && (
        <div className="rounded-lg border border-rose-500/20 bg-rose-500/10 p-4 text-sm text-rose-400">
          {error}
        </div>
      )}

      {/* KPI Cards */}
      <div className="grid grid-cols-4 gap-4">
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><Users className="h-4 w-4" /> Total Bookings (30d)</div>
          <p className="text-2xl font-bold text-white">{loading ? "—" : totalBookings}</p>
          <div className="flex items-center gap-1 text-xs text-emerald-400"><ArrowUpRight className="h-3 w-3" /> Live from reservations</div>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><DollarSign className="h-4 w-4" /> Gross Revenue (30d)</div>
          <p className="text-2xl font-bold text-white">{loading ? "—" : `$${totalRevenue.toLocaleString()}`}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><DollarSign className="h-4 w-4" /> Net Revenue (30d)</div>
          <p className="text-2xl font-bold text-white">{loading ? "—" : `$${totalNet.toLocaleString()}`}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><Percent className="h-4 w-4" /> Commission (30d)</div>
          <p className="text-2xl font-bold text-white">{loading ? "—" : `$${totalCommission.toLocaleString()}`}</p>
          <div className="text-xs text-slate-500">{totalRevenue > 0 ? ((totalCommission / totalRevenue) * 100).toFixed(1) : "0.0"}% of gross</div>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        {(["overview", "reservations", "logs"] as const).map((tab) => (
          <button key={tab} onClick={() => setActiveTab(tab)} className={`rounded-lg px-4 py-2 text-sm font-medium ${activeTab === tab ? "bg-slate-800 text-white" : "text-slate-400"}`}>
            {tab === "overview" ? "Channels Overview" : tab === "reservations" ? "Reservations" : "Sync Logs"}
          </button>
        ))}
      </div>

      {/* Overview Tab */}
      {activeTab === "overview" && (
        <div className="card space-y-4">
          {loading ? (
            <div className="flex items-center justify-center py-12"><Loader2 className="h-8 w-8 animate-spin text-nexus-400" /></div>
          ) : channels.length === 0 ? (
            <div className="py-12 text-center text-slate-500">No channels configured yet. Add your first OTA connection.</div>
          ) : (
            <table className="w-full text-left text-sm">
              <thead className="bg-slate-800 text-xs uppercase text-slate-400">
                <tr>
                  <th className="px-4 py-3">Channel</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3">Commission</th>
                  <th className="px-4 py-3">Bookings (30d)</th>
                  <th className="px-4 py-3">Revenue</th>
                  <th className="px-4 py-3">Net Revenue</th>
                  <th className="px-4 py-3">Last Sync</th>
                  <th className="px-4 py-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800">
                {channelMetrics.map((conn) => (
                  <tr key={conn.id} className="hover:bg-slate-800/50">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <Globe className="h-4 w-4 text-nexus-400" />
                        <span className="font-medium text-white">{conn.display_name}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`rounded px-2 py-0.5 text-xs ${conn.is_active ? "bg-emerald-500/10 text-emerald-400" : "bg-slate-500/10 text-slate-400"}`}>
                        {conn.is_active ? "Active" : "Inactive"}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-slate-400">{conn.commission_pct}%</td>
                    <td className="px-4 py-3 text-white">{conn.bookings30d}</td>
                    <td className="px-4 py-3 text-slate-300">${conn.revenue30d.toLocaleString()}</td>
                    <td className="px-4 py-3 text-emerald-400">${conn.netRevenue30d.toLocaleString()}</td>
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-1">
                        {conn.last_sync_status === "success" && <CheckCircle2 className="h-3 w-3 text-emerald-400" />}
                        {conn.last_sync_status === "warning" && <AlertTriangle className="h-3 w-3 text-amber-400" />}
                        {conn.last_sync_status === "error" && <XCircle className="h-3 w-3 text-rose-400" />}
                        {conn.last_sync_status === "n/a" && <span className="text-xs text-slate-500">—</span>}
                        <span className="text-xs text-slate-400">{conn.last_sync_at ? new Date(conn.last_sync_at).toLocaleString() : "N/A"}</span>
                      </div>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-2">
                        <button onClick={() => toggleChannel(conn.id, conn.is_active)} className="rounded p-1 text-slate-400 hover:text-nexus-400">
                          <RefreshCw className="h-4 w-4" />
                        </button>
                        <button onClick={() => removeChannel(conn.id)} className="rounded p-1 text-slate-400 hover:text-rose-400">
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>
      )}

      {/* Reservations Tab */}
      {activeTab === "reservations" && (
        <div className="space-y-4">
          <div className="flex items-center gap-3">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
              <input className="input w-full pl-9" placeholder="Search guest or reference..." value={search} onChange={(e) => setSearch(e.target.value)} />
            </div>
            <select className="input text-xs">
              <option>All Channels</option>
              {channels.map(c => <option key={c.id} value={c.source}>{c.display_name}</option>)}
            </select>
          </div>
          <div className="card space-y-4">
            {loading ? (
              <div className="flex items-center justify-center py-12"><Loader2 className="h-8 w-8 animate-spin text-nexus-400" /></div>
            ) : (
              <table className="w-full text-left text-sm">
                <thead className="bg-slate-800 text-xs uppercase text-slate-400">
                  <tr>
                    <th className="px-4 py-3">Ref</th>
                    <th className="px-4 py-3">Channel</th>
                    <th className="px-4 py-3">Guest</th>
                    <th className="px-4 py-3">Room Type</th>
                    <th className="px-4 py-3">Dates</th>
                    <th className="px-4 py-3">Total</th>
                    <th className="px-4 py-3">Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-800">
                  {filteredReservations.map((res) => {
                    const ch = channels.find(c => c.source === res.source);
                    const commission = ch ? Math.round(res.total * (ch.commission_pct / 100)) : 0;
                    return (
                      <tr key={res.id} className="hover:bg-slate-800/50">
                        <td className="px-4 py-3 font-mono text-xs text-slate-400">{res.id.slice(0, 8)}</td>
                        <td className="px-4 py-3">
                          <span className="rounded bg-slate-800 px-2 py-0.5 text-xs text-slate-300">{ch?.display_name || res.source}</span>
                        </td>
                        <td className="px-4 py-3 text-white">{res.guest_name}</td>
                        <td className="px-4 py-3 text-slate-400">{res.room_type}</td>
                        <td className="px-4 py-3 text-slate-400">{res.check_in} → {res.check_out}</td>
                        <td className="px-4 py-3 text-white">${res.total}</td>
                        <td className="px-4 py-3">
                          <span className={`rounded px-2 py-0.5 text-xs ${res.status === "confirmed" ? "bg-emerald-500/10 text-emerald-400" : res.status === "cancelled" ? "bg-rose-500/10 text-rose-400" : "bg-amber-500/10 text-amber-400"}`}>
                            {res.status}
                          </span>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            )}
          </div>
        </div>
      )}

      {/* Logs Tab */}
      {activeTab === "logs" && (
        <div className="card space-y-4">
          <div className="py-12 text-center text-slate-500">
            Sync logs will appear here after OTA integrations are configured.
          </div>
        </div>
      )}
    </div>
  );
}
