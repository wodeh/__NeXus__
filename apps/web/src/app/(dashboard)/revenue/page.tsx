"use client";

import { useState } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import { TrendingUp, DollarSign, Zap, Lock, ArrowUpRight } from "lucide-react";

export default function RevenuePage() {
  const { config } = useTenant();
  const [tab, setTab] = useState<"forecasts" | "pricing" | "recommendations">("forecasts");

  const hasRevenue = hasCapability(config, CAPABILITIES.REVENUE.DYNAMIC_PRICING);

  if (!hasRevenue) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-800">
          <Lock className="h-8 w-8 text-slate-500" />
        </div>
        <h2 className="mt-6 text-xl font-bold text-white">Revenue Management Locked</h2>
        <p className="mt-2 max-w-md text-sm text-slate-400">
          Revenue management features require the Revenue or Enterprise license tier.
          Upgrade to unlock dynamic pricing, OTA integration, and revenue forecasting.
        </p>
        <button className="btn-primary mt-6">
          <ArrowUpRight className="h-4 w-4" />
          Upgrade to Revenue
        </button>
      </div>
    );
  }

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
          <p className="mt-4 text-sm text-slate-500">Forecast data will appear here when backend is connected.</p>
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
          <p className="text-sm text-slate-500">No price recommendations pending.</p>
        </div>
      )}
    </div>
  );
}
