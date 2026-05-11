"use client";

import { useState } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import {
  ExternalLink,
  Plus,
  Trash2,
  Copy,
  CheckCircle2,
  XCircle,
  Tag,
  Gift,
  Coffee,
  Wifi,
  Car,
  Utensils,
  Sparkles,
  Palette,
  Link as LinkIcon,
  Eye,
  ToggleRight,
  ToggleLeft,
} from "lucide-react";

const mockPromoCodes = [
  { id: "p1", code: "SUMMER25", discountType: "percentage", discountValue: 25, maxUses: 100, usesCount: 42, minNights: 2, validFrom: "2026-05-01", validUntil: "2026-08-31", isActive: true },
  { id: "p2", code: "WELCOME10", discountType: "percentage", discountValue: 10, maxUses: 0, usesCount: 156, minNights: 1, validFrom: "2026-01-01", validUntil: "2026-12-31", isActive: true },
  { id: "p3", code: "FLASH50", discountType: "percentage", discountValue: 50, maxUses: 20, usesCount: 20, minNights: 3, validFrom: "2026-04-01", validUntil: "2026-04-30", isActive: false },
  { id: "p4", code: "WEEKEND20", discountType: "fixed", discountValue: 20, maxUses: 50, usesCount: 12, minNights: 2, validFrom: "2026-05-01", validUntil: "2026-07-31", isActive: true },
];

const mockUpsells = [
  { id: "u1", name: "Breakfast Buffet", description: "Full continental + hot buffet", price: 25, perNight: true, isActive: true },
  { id: "u2", name: "Airport Transfer", description: "Private car pickup", price: 65, perNight: false, isActive: true },
  { id: "u3", name: "Late Checkout", description: "Extend until 3PM", price: 30, perNight: false, isActive: true },
  { id: "u4", name: "Spa Package", description: "60min massage + sauna access", price: 120, perNight: false, isActive: true },
  { id: "u5", name: "Premium WiFi", description: "High-speed fiber for streaming", price: 10, perNight: true, isActive: false },
];

export default function BookingEnginePage() {
  const { config } = useTenant();
  const [activeTab, setActiveTab] = useState<"widget" | "promos" | "upsells">("widget");
  const [widgetEnabled, setWidgetEnabled] = useState(true);
  const [showPromoCode, setShowPromoCode] = useState(true);
  const [upsellEnabled, setUpsellEnabled] = useState(true);
  const [requireDeposit, setRequireDeposit] = useState(false);
  const [depositPct, setDepositPct] = useState(25);
  const [themeColor, setThemeColor] = useState("#6366f1");

  if (!hasCapability(config, CAPABILITIES.REVENUE.DIRECT_BOOKING)) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <ExternalLink className="h-16 w-16 text-slate-600" />
        <h2 className="mt-4 text-xl font-bold text-white">Direct Booking Engine</h2>
        <p className="mt-2 text-slate-400">Upgrade to Revenue tier to enable direct bookings.</p>
      </div>
    );
  }

  const embedCode = `<script src="https://nexus.com/widget.js" data-tenant="demo" data-color="${themeColor}"></script>`;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Booking Engine</h1>
          <p className="text-sm text-slate-400">Direct booking widget · Promo codes · Upsells</p>
        </div>
        <button className="btn-primary gap-2 text-sm">
          <Eye className="h-4 w-4" /> Preview Widget
        </button>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><ExternalLink className="h-4 w-4" /> Direct Bookings</div>
          <p className="text-2xl font-bold text-white">43</p>
          <div className="text-xs text-emerald-400">+18% vs last month</div>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><Tag className="h-4 w-4" /> Promo Code Uses</div>
          <p className="text-2xl font-bold text-white">210</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><Gift className="h-4 w-4" /> Upsell Revenue</div>
          <p className="text-2xl font-bold text-white">$4,280</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><CheckCircle2 className="h-4 w-4" /> Conversion Rate</div>
          <p className="text-2xl font-bold text-white">4.2%</p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        <button onClick={() => setActiveTab("widget")} className={`rounded-lg px-4 py-2 text-sm font-medium ${activeTab === "widget" ? "bg-slate-800 text-white" : "text-slate-400"}`}>Widget Config</button>
        <button onClick={() => setActiveTab("promos")} className={`rounded-lg px-4 py-2 text-sm font-medium ${activeTab === "promos" ? "bg-slate-800 text-white" : "text-slate-400"}`}>Promo Codes</button>
        <button onClick={() => setActiveTab("upsells")} className={`rounded-lg px-4 py-2 text-sm font-medium ${activeTab === "upsells" ? "bg-slate-800 text-white" : "text-slate-400"}`}>Upsells</button>
      </div>

      {activeTab === "widget" && (
        <div className="grid grid-cols-2 gap-6">
          <div className="card space-y-6">
            <h3 className="text-lg font-semibold text-white">Widget Settings</h3>

            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-white">Enable Booking Widget</p>
                <p className="text-xs text-slate-400">Show embeddable widget on your website</p>
              </div>
              <button onClick={() => setWidgetEnabled(!widgetEnabled)} className="text-nexus-400">
                {widgetEnabled ? <ToggleRight className="h-6 w-6" /> : <ToggleLeft className="h-6 w-6 text-slate-600" />}
              </button>
            </div>

            <div className="space-y-2">
              <label className="text-sm text-slate-400">Widget Title</label>
              <input className="input text-sm" defaultValue="Book Your Stay" />
            </div>

            <div className="space-y-2">
              <label className="text-sm text-slate-400">Subtitle</label>
              <input className="input text-sm" defaultValue="Best rates guaranteed when you book direct" />
            </div>

            <div className="space-y-2">
              <label className="text-sm text-slate-400">Theme Color</label>
              <div className="flex items-center gap-2">
                <input type="color" value={themeColor} onChange={(e) => setThemeColor(e.target.value)} className="h-8 w-8 rounded cursor-pointer" />
                <input className="input text-sm w-24" value={themeColor} onChange={(e) => setThemeColor(e.target.value)} />
              </div>
            </div>

            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-white">Show Promo Code Field</p>
              </div>
              <button onClick={() => setShowPromoCode(!showPromoCode)} className="text-nexus-400">
                {showPromoCode ? <ToggleRight className="h-6 w-6" /> : <ToggleLeft className="h-6 w-6 text-slate-600" />}
              </button>
            </div>

            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-white">Enable Upsells</p>
              </div>
              <button onClick={() => setUpsellEnabled(!upsellEnabled)} className="text-nexus-400">
                {upsellEnabled ? <ToggleRight className="h-6 w-6" /> : <ToggleLeft className="h-6 w-6 text-slate-600" />}
              </button>
            </div>

            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-white">Require Deposit</p>
              </div>
              <button onClick={() => setRequireDeposit(!requireDeposit)} className="text-nexus-400">
                {requireDeposit ? <ToggleRight className="h-6 w-6" /> : <ToggleLeft className="h-6 w-6 text-slate-600" />}
              </button>
            </div>

            {requireDeposit && (
              <div className="space-y-2">
                <label className="text-sm text-slate-400">Deposit Percentage</label>
                <input type="range" min="10" max="100" value={depositPct} onChange={(e) => setDepositPct(parseInt(e.target.value))} className="w-full" />
                <p className="text-sm text-white">{depositPct}%</p>
              </div>
            )}

            <button className="btn-primary w-full">Save Widget Settings</button>
          </div>

          <div className="card space-y-6">
            <h3 className="text-lg font-semibold text-white">Embed Code</h3>
            <div className="rounded-lg bg-slate-800/50 p-4 space-y-3">
              <p className="text-sm text-slate-400">Add this script to your website HTML:</p>
              <div className="flex items-center gap-2 rounded bg-slate-900 p-3">
                <code className="flex-1 text-xs text-nexus-400 break-all">{embedCode}</code>
                <button className="rounded bg-slate-800 px-2 py-1 text-xs text-white hover:bg-slate-700"><Copy className="h-3 w-3" /></button>
              </div>
            </div>

            <div className="space-y-3">
              <h4 className="text-sm font-medium text-white">Direct Booking Link</h4>
              <div className="flex items-center gap-2 rounded bg-slate-800/50 p-3">
                <input className="input flex-1 text-sm" value="https://book.nexus.com/demo" readOnly />
                <button className="btn-secondary text-xs"><LinkIcon className="h-3 w-3" /> Open</button>
              </div>
            </div>

            <div className="rounded-lg border border-slate-700 p-4 space-y-3">
              <h4 className="text-sm font-medium text-white">Widget Preview</h4>
              <div className="rounded bg-white p-4 space-y-3">
                <p className="text-lg font-bold text-slate-900">Book Your Stay</p>
                <p className="text-xs text-slate-500">Best rates guaranteed when you book direct</p>
                <div className="grid grid-cols-2 gap-2">
                  <input className="rounded border p-2 text-xs" placeholder="Check-in" />
                  <input className="rounded border p-2 text-xs" placeholder="Check-out" />
                </div>
                <select className="w-full rounded border p-2 text-xs">
                  <option>2 Adults, 0 Children</option>
                </select>
                {showPromoCode && <input className="w-full rounded border p-2 text-xs" placeholder="Promo code" />}
                <button className="w-full rounded py-2 text-sm text-white" style={{ backgroundColor: themeColor }}>Check Availability</button>
              </div>
            </div>
          </div>
        </div>
      )}

      {activeTab === "promos" && (
        <div className="space-y-4">
          <div className="flex justify-end">
            <button className="btn-primary gap-2 text-sm"><Plus className="h-4 w-4" /> New Promo Code</button>
          </div>
          <div className="card space-y-4">
            <table className="w-full text-left text-sm">
              <thead className="bg-slate-800 text-xs uppercase text-slate-400">
                <tr>
                  <th className="px-4 py-3">Code</th>
                  <th className="px-4 py-3">Discount</th>
                  <th className="px-4 py-3">Uses</th>
                  <th className="px-4 py-3">Min Nights</th>
                  <th className="px-4 py-3">Valid Period</th>
                  <th className="px-4 py-3">Status</th>
                  <th className="px-4 py-3 text-right">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800">
                {mockPromoCodes.map((p) => (
                  <tr key={p.id} className="hover:bg-slate-800/50">
                    <td className="px-4 py-3">
                      <div className="flex items-center gap-2">
                        <span className="font-mono text-sm font-medium text-nexus-400">{p.code}</span>
                        <button className="text-slate-500 hover:text-white"><Copy className="h-3 w-3" /></button>
                      </div>
                    </td>
                    <td className="px-4 py-3">
                      <span className={`font-medium ${p.discountType === "percentage" ? "text-nexus-400" : "text-emerald-400"}`}>
                        {p.discountType === "percentage" ? `${p.discountValue}%` : `$${p.discountValue}`}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-slate-400">
                      {p.maxUses > 0 ? `${p.usesCount}/${p.maxUses}` : p.usesCount}
                    </td>
                    <td className="px-4 py-3 text-slate-400">{p.minNights}</td>
                    <td className="px-4 py-3 text-slate-400">{p.validFrom} → {p.validUntil}</td>
                    <td className="px-4 py-3">
                      <span className={`rounded px-2 py-0.5 text-xs ${p.isActive ? "bg-emerald-500/10 text-emerald-400" : "bg-slate-500/10 text-slate-400"}`}>
                        {p.isActive ? "Active" : "Inactive"}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button className="text-slate-400 hover:text-rose-400"><Trash2 className="h-4 w-4" /></button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === "upsells" && (
        <div className="space-y-4">
          <div className="flex justify-end">
            <button className="btn-primary gap-2 text-sm"><Plus className="h-4 w-4" /> New Upsell</button>
          </div>
          <div className="grid grid-cols-3 gap-4">
            {mockUpsells.map((u) => (
              <div key={u.id} className="card space-y-3">
                <div className="flex items-center justify-between">
                  <h3 className="font-medium text-white">{u.name}</h3>
                  <span className={`rounded px-2 py-0.5 text-xs ${u.isActive ? "bg-emerald-500/10 text-emerald-400" : "bg-slate-500/10 text-slate-400"}`}>
                    {u.isActive ? "Active" : "Inactive"}
                  </span>
                </div>
                <p className="text-sm text-slate-400">{u.description}</p>
                <div className="flex items-center justify-between">
                  <span className="text-lg font-bold text-white">${u.price}</span>
                  {u.perNight && <span className="text-xs text-slate-500">per night</span>}
                </div>
                <div className="flex gap-2">
                  <button className="btn-secondary flex-1 text-xs">Edit</button>
                  <button className="rounded bg-rose-500/10 p-2 text-rose-400 hover:bg-rose-500/20"><Trash2 className="h-4 w-4" /></button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
