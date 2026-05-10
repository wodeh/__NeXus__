"use client";

import { useState } from "react";
import { TrendingUp, DollarSign, BarChart3, Zap, ChevronRight } from "lucide-react";

const mockForecasts = [
  { date: "2026-05-10", occupancy: 78, adr: 185, revpar: 144 },
  { date: "2026-05-11", occupancy: 82, adr: 190, revpar: 156 },
  { date: "2026-05-12", occupancy: 75, adr: 180, revpar: 135 },
  { date: "2026-05-13", occupancy: 88, adr: 210, revpar: 185 },
  { date: "2026-05-14", occupancy: 92, adr: 225, revpar: 207 },
];

const mockRecommendations = [
  { id: "rec-001", room_type: "Deluxe King", date: "2026-05-12", current: 185, suggested: 210, reason: "High demand forecast" },
  { id: "rec-002", room_type: "Suite", date: "2026-05-13", current: 320, suggested: 350, reason: "Weekend premium" },
];

export default function RevenuePage() {
  const [tab, setTab] = useState<"forecasts" | "pricing" | "recommendations">("forecasts");

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Revenue Management</h2>
          <p className="text-sm text-slate-400">Forecasts, pricing rules, and recommendations</p>
        </div>
      </div>

      <div className="flex gap-2">
        <button onClick={() => setTab("forecasts")} className={tab === "forecasts" ? "btn-primary" : "btn-secondary"}>Forecasts</button>
        <button onClick={() => setTab("pricing")} className={tab === "pricing" ? "btn-primary" : "btn-secondary"}>Pricing Rules</button>
        <button onClick={() => setTab("recommendations")} className={tab === "recommendations" ? "btn-primary" : "btn-secondary"}>Recommendations</button>
      </div>

      {tab === "forecasts" && (
        <div className="card">
          <h3 className="text-sm font-semibold text-slate-200">5-Day Forecast</h3>
          <div className="mt-4 overflow-x-auto">
            <table className="w-full text-left text-sm">
              <thead className="text-xs uppercase text-slate-400">
                <tr><th className="px-3 py-2">Date</th><th className="px-3 py-2">Occupancy %</th><th className="px-3 py-2">ADR ($)</th><th className="px-3 py-2">RevPAR ($)</th></tr>
              </thead>
              <tbody className="divide-y divide-slate-800">
                {mockForecasts.map((f) => (
                  <tr key={f.date} className="hover:bg-slate-800/50">
                    <td className="px-3 py-2 text-white">{f.date}</td>
                    <td className="px-3 py-2 text-slate-300">{f.occupancy}%</td>
                    <td className="px-3 py-2 text-slate-300">${f.adr}</td>
                    <td className="px-3 py-2 text-slate-300">${f.revpar}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {tab === "pricing" && (
        <div className="card">
          <h3 className="text-sm font-semibold text-slate-200">Dynamic Pricing Rules</h3>
          <p className="mt-4 text-sm text-slate-500">No pricing rules configured yet.</p>
          <button className="btn-primary mt-4">
            <Zap className="h-4 w-4" />
            Create Rule
          </button>
        </div>
      )}

      {tab === "recommendations" && (
        <div className="space-y-4">
          {mockRecommendations.map((rec) => (
            <div key={rec.id} className="card-hover flex items-center justify-between">
              <div>
                <p className="font-medium text-white">{rec.room_type} — {rec.date}</p>
                <p className="text-sm text-slate-400">{rec.reason}</p>
                <div className="mt-1 flex items-center gap-3 text-sm">
                  <span className="text-slate-500">Current: <span className="text-slate-300">${rec.current}</span></span>
                  <span className="text-nexus-400">Suggested: <span className="font-medium">${rec.suggested}</span></span>
                  <span className="badge-green">+{Math.round((rec.suggested - rec.current) / rec.current * 100)}%</span>
                </div>
              </div>
              <button className="btn-primary">
                <DollarSign className="h-4 w-4" />
                Apply
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
