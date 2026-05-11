"use client";

import { useState } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import {
  Globe,
  RefreshCw,
  Plus,
  Trash2,
  CheckCircle2,
  AlertTriangle,
  XCircle,
  TrendingUp,
  TrendingDown,
  DollarSign,
  Percent,
  Search,
  Filter,
  Download,
  ExternalLink,
  CalendarDays,
  Users,
  ArrowUpRight,
} from "lucide-react";

const mockConnections = [
  { id: "c1", source: "booking.com", displayName: "Booking.com", isActive: true, commissionPct: 15, lastSyncAt: "2026-05-10 14:30", lastSyncStatus: "success", bookings30d: 42, revenue30d: 28400, netRevenue30d: 24140 },
  { id: "c2", source: "expedia", displayName: "Expedia", isActive: true, commissionPct: 18, lastSyncAt: "2026-05-10 13:15", lastSyncStatus: "success", bookings30d: 28, revenue30d: 19600, netRevenue30d: 16072 },
  { id: "c3", source: "airbnb", displayName: "Airbnb", isActive: true, commissionPct: 3, lastSyncAt: "2026-05-10 12:00", lastSyncStatus: "warning", bookings30d: 15, revenue30d: 12400, netRevenue30d: 12028 },
  { id: "c4", source: "direct", displayName: "Direct Bookings", isActive: true, commissionPct: 0, lastSyncAt: null, lastSyncStatus: "n/a", bookings30d: 35, revenue30d: 31500, netRevenue30d: 31500 },
  { id: "c5", source: "whatsapp", displayName: "WhatsApp Bot", isActive: true, commissionPct: 0, lastSyncAt: null, lastSyncStatus: "n/a", bookings30d: 8, revenue30d: 7200, netRevenue30d: 7200 },
];

const mockReservations = [
  { id: "cr1", source: "booking.com", externalRef: "BDC-88421", guestName: "Alice Chen", roomType: "Deluxe King", checkIn: "2026-05-10", checkOut: "2026-05-14", nights: 4, total: 920, commission: 138, netAmount: 782, status: "confirmed" },
  { id: "cr2", source: "expedia", externalRef: "EXP-55219", guestName: "Bob Jones", roomType: "Standard", checkIn: "2026-05-11", checkOut: "2026-05-13", nights: 2, total: 340, commission: 61, netAmount: 279, status: "confirmed" },
  { id: "cr3", source: "airbnb", externalRef: "ABN-11209", guestName: "Carol White", roomType: "Suite", checkIn: "2026-05-12", checkOut: "2026-05-16", nights: 4, total: 1520, commission: 46, netAmount: 1474, status: "confirmed" },
  { id: "cr4", source: "direct", externalRef: "DIR-001", guestName: "David Kim", roomType: "Deluxe", checkIn: "2026-05-15", checkOut: "2026-05-18", nights: 3, total: 720, commission: 0, netAmount: 720, status: "confirmed" },
];

const mockSyncLogs = [
  { id: "sl1", channel: "Booking.com", direction: "pull", status: "success", records: 12, startedAt: "2026-05-10 14:30", duration: "2.3s" },
  { id: "sl2", channel: "Expedia", direction: "pull", status: "success", records: 8, startedAt: "2026-05-10 13:15", duration: "1.8s" },
  { id: "sl3", channel: "Airbnb", direction: "pull", status: "partial", records: 3, startedAt: "2026-05-10 12:00", duration: "4.1s", error: "Rate limit hit" },
];

export default function ChannelManagerPage() {
  const { tenantConfig } = useTenant();
  const [activeTab, setActiveTab] = useState<"overview" | "reservations" | "logs">("overview");
  const [search, setSearch] = useState("");

  if (!hasCapability(tenantConfig, CAPABILITIES.REVENUE.CHANNEL_MANAGER)) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <Globe className="h-16 w-16 text-slate-600" />
        <h2 className="mt-4 text-xl font-bold text-white">Channel Manager</h2>
        <p className="mt-2 text-slate-400">Upgrade to Revenue tier to manage OTA channels.</p>
      </div>
    );
  }

  const totalBookings = mockConnections.reduce((s, c) => s + c.bookings30d, 0);
  const totalRevenue = mockConnections.reduce((s, c) => s + c.revenue30d, 0);
  const totalNet = mockConnections.reduce((s, c) => s + c.netRevenue30d, 0);
  const totalCommission = totalRevenue - totalNet;

  const filteredReservations = mockReservations.filter((r) =>
    r.guestName.toLowerCase().includes(search.toLowerCase()) ||
    r.externalRef.toLowerCase().includes(search.toLowerCase())
  );

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

      {/* KPI Cards */}
      <div className="grid grid-cols-4 gap-4">
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><Users className="h-4 w-4" /> Total Bookings</div>
          <p className="text-2xl font-bold text-white">{totalBookings}</p>
          <div className="flex items-center gap-1 text-xs text-emerald-400"><ArrowUpRight className="h-3 w-3" /> +12% vs last month</div>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><DollarSign className="h-4 w-4" /> Gross Revenue</div>
          <p className="text-2xl font-bold text-white">${totalRevenue.toLocaleString()}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><DollarSign className="h-4 w-4" /> Net Revenue</div>
          <p className="text-2xl font-bold text-white">${totalNet.toLocaleString()}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><Percent className="h-4 w-4" /> Commission Paid</div>
          <p className="text-2xl font-bold text-white">${totalCommission.toLocaleString()}</p>
          <div className="text-xs text-slate-500">{((totalCommission / totalRevenue) * 100).toFixed(1)}% of gross</div>
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
              {mockConnections.map((conn) => (
                <tr key={conn.id} className="hover:bg-slate-800/50">
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <Globe className="h-4 w-4 text-nexus-400" />
                      <span className="font-medium text-white">{conn.displayName}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`rounded px-2 py-0.5 text-xs ${conn.isActive ? "bg-emerald-500/10 text-emerald-400" : "bg-slate-500/10 text-slate-400"}`}>
                      {conn.isActive ? "Active" : "Inactive"}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-400">{conn.commissionPct}%</td>
                  <td className="px-4 py-3 text-white">{conn.bookings30d}</td>
                  <td className="px-4 py-3 text-slate-300">${conn.revenue30d.toLocaleString()}</td>
                  <td className="px-4 py-3 text-emerald-400">${conn.netRevenue30d.toLocaleString()}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-1">
                      {conn.lastSyncStatus === "success" && <CheckCircle2 className="h-3 w-3 text-emerald-400" />}
                      {conn.lastSyncStatus === "warning" && <AlertTriangle className="h-3 w-3 text-amber-400" />}
                      {conn.lastSyncStatus === "error" && <XCircle className="h-3 w-3 text-rose-400" />}
                      <span className="text-xs text-slate-400">{conn.lastSyncAt || "N/A"}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3 text-right">
                    <div className="flex items-center justify-end gap-2">
                      <button className="rounded p-1 text-slate-400 hover:text-nexus-400"><RefreshCw className="h-4 w-4" /></button>
                      <button className="rounded p-1 text-slate-400 hover:text-rose-400"><Trash2 className="h-4 w-4" /></button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
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
              <option>Booking.com</option>
              <option>Expedia</option>
              <option>Airbnb</option>
              <option>Direct</option>
            </select>
          </div>
          <div className="card space-y-4">
            <table className="w-full text-left text-sm">
              <thead className="bg-slate-800 text-xs uppercase text-slate-400">
                <tr>
                  <th className="px-4 py-3">Ref</th>
                  <th className="px-4 py-3">Channel</th>
                  <th className="px-4 py-3">Guest</th>
                  <th className="px-4 py-3">Room Type</th>
                  <th className="px-4 py-3">Dates</th>
                  <th className="px-4 py-3">Total</th>
                  <th className="px-4 py-3">Commission</th>
                  <th className="px-4 py-3">Net</th>
                  <th className="px-4 py-3">Status</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800">
                {filteredReservations.map((res) => (
                  <tr key={res.id} className="hover:bg-slate-800/50">
                    <td className="px-4 py-3 font-mono text-xs text-slate-400">{res.externalRef}</td>
                    <td className="px-4 py-3">
                      <span className="rounded bg-slate-800 px-2 py-0.5 text-xs text-slate-300">{res.source}</span>
                    </td>
                    <td className="px-4 py-3 text-white">{res.guestName}</td>
                    <td className="px-4 py-3 text-slate-400">{res.roomType}</td>
                    <td className="px-4 py-3 text-slate-400">{res.checkIn} → {res.checkOut}</td>
                    <td className="px-4 py-3 text-white">${res.total}</td>
                    <td className="px-4 py-3 text-amber-400">${res.commission}</td>
                    <td className="px-4 py-3 text-emerald-400">${res.netAmount}</td>
                    <td className="px-4 py-3">
                      <span className="rounded bg-emerald-500/10 px-2 py-0.5 text-xs text-emerald-400">{res.status}</span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Logs Tab */}
      {activeTab === "logs" && (
        <div className="card space-y-4">
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-800 text-xs uppercase text-slate-400">
              <tr>
                <th className="px-4 py-3">Channel</th>
                <th className="px-4 py-3">Direction</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Records</th>
                <th className="px-4 py-3">Duration</th>
                <th className="px-4 py-3">Error</th>
                <th className="px-4 py-3">Time</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {mockSyncLogs.map((log) => (
                <tr key={log.id} className="hover:bg-slate-800/50">
                  <td className="px-4 py-3 text-white">{log.channel}</td>
                  <td className="px-4 py-3 text-slate-400">{log.direction}</td>
                  <td className="px-4 py-3">
                    <span className={`rounded px-2 py-0.5 text-xs ${log.status === "success" ? "bg-emerald-500/10 text-emerald-400" : log.status === "partial" ? "bg-amber-500/10 text-amber-400" : "bg-rose-500/10 text-rose-400"}`}>
                      {log.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-white">{log.records}</td>
                  <td className="px-4 py-3 text-slate-400">{log.duration}</td>
                  <td className="px-4 py-3 text-rose-400">{log.error || "—"}</td>
                  <td className="px-4 py-3 text-slate-500">{log.startedAt}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
