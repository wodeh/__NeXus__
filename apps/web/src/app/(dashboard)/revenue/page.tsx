"use client";

import { useState } from "react";
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
} from "lucide-react";

const today = "2026-05-10";

const mockDailyRevenue = Array.from({ length: 30 }, (_, i) => {
  const base = 6000 + Math.random() * 4000;
  const weekend = i % 7 === 5 || i % 7 === 6 ? 1.3 : 1;
  return {
    day: `May ${i + 1}`,
    revenue: Math.round(base * weekend),
    rooms: Math.round(65 + Math.random() * 25),
    adr: Math.round(95 + Math.random() * 40),
  };
});

const mockRoomTypeRevenue = [
  { type: "Standard", revenue: 28400, nights: 320, color: "sky" },
  { type: "Deluxe", revenue: 41200, nights: 280, color: "nexus" },
  { type: "Deluxe King", revenue: 38900, nights: 260, color: "violet" },
  { type: "Suite", revenue: 52600, nights: 180, color: "amber" },
  { type: "Accessible", revenue: 12400, nights: 120, color: "emerald" },
];

const mockPricingRules = [
  {
    id: "pr-1",
    name: "Weekend Premium",
    condition: "Friday + Saturday",
    adjustment: "+15%",
    active: true,
    impact: "$4,200/mo",
  },
  {
    id: "pr-2",
    name: "Last Minute Discount",
    condition: "Booking within 24h",
    adjustment: "-10%",
    active: true,
    impact: "$1,800/mo",
  },
  {
    id: "pr-3",
    name: "Long Stay Discount",
    condition: "7+ nights",
    adjustment: "-12%",
    active: true,
    impact: "$2,400/mo",
  },
  {
    id: "pr-4",
    name: "Corporate Rate",
    condition: "Company booking",
    adjustment: "-8%",
    active: true,
    impact: "$3,600/mo",
  },
  {
    id: "pr-5",
    name: "High Occupancy Boost",
    condition: "Occupancy > 85%",
    adjustment: "+20%",
    active: false,
    impact: "$0 (disabled)",
  },
];

const mockForecast = [
  { period: "Next 7 Days", predicted: 58400, confidence: 92 },
  { period: "Next 14 Days", predicted: 112000, confidence: 87 },
  { period: "Next 30 Days", predicted: 235000, confidence: 78 },
  { period: "Next 90 Days", predicted: 680000, confidence: 65 },
];

export default function RevenuePage() {
  const [view, setView] = useState<"overview" | "forecast" | "pricing" | "analytics">("overview");
  const [dateRange, setDateRange] = useState("30d");

  const totalRevenue = mockDailyRevenue.reduce((s, d) => s + d.revenue, 0);
  const avgDaily = Math.round(totalRevenue / mockDailyRevenue.length);
  const totalNights = mockRoomTypeRevenue.reduce((s, r) => s + r.nights, 0);
  const overallADR = Math.round(totalRevenue / totalNights);
  const revPAR = Math.round(totalRevenue / (120 * 30)); // 120 rooms * 30 days

  const maxRevenue = Math.max(...mockDailyRevenue.map((d) => d.revenue));
  const totalRoomRevenue = mockRoomTypeRevenue.reduce((s, r) => s + r.revenue, 0);

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
          <button className="btn-secondary text-xs gap-2">
            <Download className="h-4 w-4" /> Export
          </button>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-4 gap-4">
        <KpiCard
          label="Total Revenue"
          value={`$${totalRevenue.toLocaleString()}`}
          change="+12.3%"
          positive
          icon={DollarSign}
          color="nexus"
        />
        <KpiCard
          label="Average Daily Rate"
          value={`$${overallADR}`}
          change="+5.2%"
          positive
          icon={BedDouble}
          color="emerald"
        />
        <KpiCard
          label="RevPAR"
          value={`$${revPAR}`}
          change="-2.1%"
          positive={false}
          icon={TrendingUp}
          color="amber"
        />
        <KpiCard
          label="Room Nights Sold"
          value={totalNights.toString()}
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
              {mockDailyRevenue.map((d, i) => (
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
            <div className="flex justify-between text-[10px] text-slate-500">
              <span>May 1</span>
              <span>May 10</span>
              <span>May 20</span>
              <span>May 30</span>
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
                {mockRoomTypeRevenue.map((rt) => {
                  const pct = Math.round((rt.revenue / totalRoomRevenue) * 100);
                  const colorMap: Record<string, string> = {
                    sky: "bg-sky-400",
                    nexus: "bg-nexus-400",
                    violet: "bg-violet-400",
                    amber: "bg-amber-400",
                    emerald: "bg-emerald-400",
                  };
                  return (
                    <div key={rt.type} className="space-y-1">
                      <div className="flex justify-between text-sm">
                        <span className="text-white">{rt.type}</span>
                        <span className="text-slate-400">${rt.revenue.toLocaleString()} ({pct}%)</span>
                      </div>
                      <div className="h-2 rounded-full bg-slate-800">
                        <div className={`h-full rounded-full ${colorMap[rt.color]} transition-all`} style={{ width: `${pct}%` }} />
                      </div>
                      <div className="flex justify-between text-xs text-slate-500">
                        <span>{rt.nights} nights</span>
                        <span>ADR: ${Math.round(rt.revenue / rt.nights)}</span>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>

            <div className="card space-y-4">
              <h3 className="text-lg font-semibold text-white flex items-center gap-2">
                <LineChart className="h-5 w-5 text-emerald-400" />
                Occupancy vs Revenue Correlation
              </h3>
              <div className="h-48 flex items-end gap-2">
                {mockDailyRevenue.slice(0, 14).map((d, i) => (
                  <div key={i} className="flex-1 flex flex-col items-center gap-1">
                    <div className="text-[10px] text-slate-500">{Math.round(d.rooms)}%</div>
                    <div className="flex gap-0.5 w-full">
                      <div
                        className="flex-1 rounded-t bg-emerald-500/40"
                        style={{ height: `${(d.rooms / 100) * 80}px` }}
                      />
                      <div
                        className="flex-1 rounded-t bg-nexus-500/40"
                        style={{ height: `${(d.revenue / maxRevenue) * 80}px` }}
                      />
                    </div>
                    <div className="text-[10px] text-slate-500">{d.day.replace("May ", "")}</div>
                  </div>
                ))}
              </div>
              <div className="flex justify-center gap-4 text-xs text-slate-400">
                <div className="flex items-center gap-1">
                  <div className="h-3 w-3 rounded bg-emerald-500/40" /> Occupancy %
                </div>
                <div className="flex items-center gap-1">
                  <div className="h-3 w-3 rounded bg-nexus-500/40" /> Revenue
                </div>
              </div>
            </div>
          </div>
        </>
      )}

      {/* Forecast Tab */}
      {view === "forecast" && (
        <div className="space-y-4">
          <div className="card space-y-4">
            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
              <Target className="h-5 w-5 text-nexus-400" />
              Revenue Forecast
            </h3>
            <div className="grid grid-cols-4 gap-4">
              {mockForecast.map((f) => (
                <div key={f.period} className="rounded-lg border border-slate-700 bg-slate-800/50 p-4 space-y-2">
                  <p className="text-xs text-slate-400">{f.period}</p>
                  <p className="text-2xl font-bold text-white">${(f.predicted / 1000).toFixed(0)}k</p>
                  <div className="flex items-center gap-2">
                    <div className="flex-1 h-2 rounded-full bg-slate-700">
                      <div className="h-full rounded-full bg-emerald-400" style={{ width: `${f.confidence}%` }} />
                    </div>
                    <span className="text-xs text-slate-400">{f.confidence}% confidence</span>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="card space-y-4">
            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
              <Zap className="h-5 w-5 text-amber-400" />
              Demand Forecasting
            </h3>
            <div className="space-y-3">
              {[
                { period: "May 11-17", demand: "High", reason: "Conference in town", suggestedRate: "+$25", occupancy: "92%" },
                { period: "May 18-24", demand: "Medium", reason: "Regular season", suggestedRate: "Base rate", occupancy: "78%" },
                { period: "May 25-31", demand: "Very High", reason: "Memorial Day weekend", suggestedRate: "+$45", occupancy: "98%" },
                { period: "Jun 1-7", demand: "Low", reason: "Post-holiday dip", suggestedRate: "-$15", occupancy: "62%" },
              ].map((row, i) => (
                <div key={i} className="flex items-center justify-between rounded-lg border border-slate-700 bg-slate-800/50 p-3">
                  <div className="flex items-center gap-4">
                    <div className="w-24">
                      <p className="text-sm font-medium text-white">{row.period}</p>
                    </div>
                    <div className="w-24">
                      <span className={`rounded px-2 py-0.5 text-xs font-medium ${
                        row.demand === "Very High" ? "bg-rose-500/10 text-rose-400" :
                        row.demand === "High" ? "bg-amber-500/10 text-amber-400" :
                        row.demand === "Medium" ? "bg-sky-500/10 text-sky-400" :
                        "bg-slate-500/10 text-slate-400"
                      }`}>
                        {row.demand}
                      </span>
                    </div>
                    <p className="text-sm text-slate-400 w-48">{row.reason}</p>
                  </div>
                  <div className="flex items-center gap-4">
                    <div className="text-right">
                      <p className="text-xs text-slate-400">Suggested Rate</p>
                      <p className="text-sm font-medium text-white">{row.suggestedRate}</p>
                    </div>
                    <div className="text-right">
                      <p className="text-xs text-slate-400">Est. Occupancy</p>
                      <p className="text-sm font-medium text-white">{row.occupancy}</p>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Pricing Rules Tab */}
      {view === "pricing" && (
        <div className="card space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
              <Layers className="h-5 w-5 text-nexus-400" />
              Active Pricing Rules
            </h3>
            <button className="btn-primary text-xs gap-2">
              <Zap className="h-4 w-4" /> New Rule
            </button>
          </div>
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-800 text-xs uppercase text-slate-400">
              <tr>
                <th className="px-4 py-3">Rule Name</th>
                <th className="px-4 py-3">Condition</th>
                <th className="px-4 py-3">Adjustment</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Impact</th>
                <th className="px-4 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {mockPricingRules.map((rule) => (
                <tr key={rule.id} className="hover:bg-slate-800/50">
                  <td className="px-4 py-3">
                    <p className="font-medium text-white">{rule.name}</p>
                  </td>
                  <td className="px-4 py-3 text-slate-400">{rule.condition}</td>
                  <td className="px-4 py-3">
                    <span className={`font-medium ${rule.adjustment.startsWith("+") ? "text-emerald-400" : "text-amber-400"}`}>
                      {rule.adjustment}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`rounded px-2 py-0.5 text-xs ${rule.active ? "bg-emerald-500/10 text-emerald-400" : "bg-slate-500/10 text-slate-400"}`}>
                      {rule.active ? "Active" : "Disabled"}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-400">{rule.impact}</td>
                  <td className="px-4 py-3 text-right">
                    <button className="text-xs text-nexus-400 hover:text-nexus-300">Edit</button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Analytics Tab */}
      {view === "analytics" && (
        <div className="grid grid-cols-2 gap-4">
          <div className="card space-y-4">
            <h3 className="text-lg font-semibold text-white">Revenue Segments</h3>
            <div className="space-y-3">
              {[
                { segment: "Direct Bookings", revenue: 42000, pct: 28, trend: "+8%" },
                { segment: "OTA (Booking.com)", revenue: 38000, pct: 25, trend: "+12%" },
                { segment: "OTA (Expedia)", revenue: 24000, pct: 16, trend: "-3%" },
                { segment: "Corporate", revenue: 21000, pct: 14, trend: "+5%" },
                { segment: "Walk-in", revenue: 12000, pct: 8, trend: "-2%" },
                { segment: "Travel Agents", revenue: 13500, pct: 9, trend: "+1%" },
              ].map((seg) => (
                <div key={seg.segment} className="flex items-center justify-between rounded bg-slate-800/50 p-3">
                  <div className="flex-1">
                    <p className="text-sm text-white">{seg.segment}</p>
                    <div className="mt-1 h-1.5 rounded-full bg-slate-700 w-full">
                      <div className="h-full rounded-full bg-nexus-400" style={{ width: `${seg.pct}%` }} />
                    </div>
                  </div>
                  <div className="text-right ml-4">
                    <p className="text-sm font-medium text-white">${seg.revenue.toLocaleString()}</p>
                    <p className={`text-xs ${seg.trend.startsWith("+") ? "text-emerald-400" : "text-rose-400"}`}>{seg.trend}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>

          <div className="card space-y-4">
            <h3 className="text-lg font-semibold text-white">Key Metrics</h3>
            <div className="space-y-3">
              {[
                { label: "GOP (Gross Operating Profit)", value: "$68,420", change: "+14.2%", positive: true },
                { label: "GOP Margin", value: "45.6%", change: "+2.1%", positive: true },
                { label: "Cost per Occupied Room", value: "$42", change: "-5%", positive: true },
                { label: "Labor Cost Ratio", value: "28%", change: "+1.2%", positive: false },
                { label: "Cancellation Rate", value: "4.2%", change: "-0.8%", positive: true },
                { label: "No-Show Rate", value: "1.8%", change: "+0.3%", positive: false },
              ].map((m) => (
                <div key={m.label} className="flex items-center justify-between rounded bg-slate-800/50 p-3">
                  <p className="text-sm text-slate-300">{m.label}</p>
                  <div className="text-right">
                    <p className="text-sm font-medium text-white">{m.value}</p>
                    <p className={`text-xs ${m.positive ? "text-emerald-400" : "text-rose-400"}`}>{m.change}</p>
                  </div>
                </div>
              ))}
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
