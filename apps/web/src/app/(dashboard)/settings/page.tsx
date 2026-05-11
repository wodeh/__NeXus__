"use client";

import { useState, useEffect } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES, tierName } from "@/lib/tenant";
import { updateTenantConfig } from "@/lib/api";
import {
  Settings,
  Globe,
  DollarSign,
  Bell,
  Shield,
  Plug,
  Check,
  X,
  Building2,
  Save,
  RefreshCw,
} from "lucide-react";

export default function SettingsPage() {
  const { config, loading: configLoading } = useTenant();
  const [saving, setSaving] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saveSuccess, setSaveSuccess] = useState(false);

  // Controlled form state synced from config
  const [form, setForm] = useState({
    name: "",
    timezone: "UTC",
    currency_code: "USD",
    language: "en",
    date_format: "YYYY-MM-DD",
  });

  const [notifEmail, setNotifEmail] = useState(true);
  const [notifSMS, setNotifSMS] = useState(false);
  const [notifPush, setNotifPush] = useState(true);

  useEffect(() => {
    if (config) {
      setForm({
        name: config.name || "",
        timezone: config.settings?.timezone || "UTC",
        currency_code: config.settings?.currency_code || "USD",
        language: config.settings?.language || "en",
        date_format: config.settings?.date_format || "YYYY-MM-DD",
      });
    }
  }, [config]);

  async function handleSave() {
    setSaving(true);
    setSaveError(null);
    setSaveSuccess(false);
    try {
      const tenant = localStorage.getItem("nexus-tenant") || "demo";
      await updateTenantConfig(tenant, {
        name: form.name,
        settings: {
          timezone: form.timezone,
          currency_code: form.currency_code,
          date_format: form.date_format || "YYYY-MM-DD",
          language: form.language,
        },
      });
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (e: any) {
      setSaveError(e.message || "Failed to save settings");
    } finally {
      setSaving(false);
    }
  }

  if (configLoading) {
    return (
      <div className="flex items-center justify-center py-20 text-slate-500">
        <RefreshCw className="mr-2 h-5 w-5 animate-spin" /> Loading settings...
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Settings</h2>
          <p className="text-sm text-slate-400">Property configuration and preferences</p>
        </div>
        <button
          onClick={handleSave}
          disabled={saving}
          className="btn-primary disabled:opacity-50"
        >
          {saving ? (
            <RefreshCw className="mr-2 h-4 w-4 animate-spin" />
          ) : (
            <Save className="mr-2 h-4 w-4" />
          )}
          {saving ? "Saving..." : "Save Changes"}
        </button>
      </div>

      {saveSuccess && (
        <div className="rounded-lg border border-emerald-500/20 bg-emerald-500/5 px-4 py-3 text-sm text-emerald-400">
          Settings saved successfully.
        </div>
      )}
      {saveError && (
        <div className="rounded-lg border border-rose-500/20 bg-rose-500/5 px-4 py-3 text-sm text-rose-400">
          {saveError}
        </div>
      )}

      {/* Property Info */}
      <div className="card space-y-4">
        <div className="flex items-center gap-2">
          <Building2 className="h-5 w-5 text-nexus-400" />
          <h3 className="text-sm font-semibold text-slate-200">Property Information</h3>
        </div>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <label className="text-xs text-slate-500">Property Name</label>
            <input
              className="input mt-1 w-full"
              value={form.name}
              onChange={(e) => setForm({ ...form, name: e.target.value })}
            />
          </div>
          <div>
            <label className="text-xs text-slate-500">Timezone</label>
            <input
              className="input mt-1 w-full"
              value={form.timezone}
              onChange={(e) => setForm({ ...form, timezone: e.target.value })}
            />
          </div>
          <div>
            <label className="text-xs text-slate-500">Currency</label>
            <input
              className="input mt-1 w-full"
              value={form.currency_code}
              onChange={(e) => setForm({ ...form, currency_code: e.target.value })}
            />
          </div>
          <div>
            <label className="text-xs text-slate-500">Language</label>
            <input
              className="input mt-1 w-full"
              value={form.language}
              onChange={(e) => setForm({ ...form, language: e.target.value })}
            />
          </div>
        </div>
      </div>

      {/* License */}
      <div className="card space-y-4">
        <div className="flex items-center gap-2">
          <Shield className="h-5 w-5 text-nexus-400" />
          <h3 className="text-sm font-semibold text-slate-200">License</h3>
        </div>
        <div className="grid grid-cols-4 gap-4 text-sm">
          <div>
            <p className="text-xs text-slate-500">Tier</p>
            <p className="font-medium text-nexus-400 capitalize">{tierName(config?.license_tier || "enterprise")}</p>
          </div>
          <div>
            <p className="text-xs text-slate-500">Status</p>
            <p className="font-medium text-emerald-400 capitalize">{config?.license_status || "active"}</p>
          </div>
          <div>
            <p className="text-xs text-slate-500">Max Rooms</p>
            <p className="font-medium text-white">{config?.max_rooms || 5000}</p>
          </div>
          <div>
            <p className="text-xs text-slate-500">Max Users</p>
            <p className="font-medium text-white">{config?.max_users || 200}</p>
          </div>
        </div>
        <div className="flex gap-2">
          {["core", "operations", "revenue", "enterprise"].map((tier) => (
            <button
              key={tier}
              className={`rounded-lg px-3 py-2 text-xs font-medium transition-colors ${
                config?.license_tier === tier
                  ? "bg-nexus-500 text-white"
                  : "bg-slate-800 text-slate-400 hover:bg-slate-700"
              }`}
            >
              {tierName(tier)}
            </button>
          ))}
        </div>
      </div>

      {/* Notifications */}
      <div className="card space-y-4">
        <div className="flex items-center gap-2">
          <Bell className="h-5 w-5 text-nexus-400" />
          <h3 className="text-sm font-semibold text-slate-200">Notifications</h3>
        </div>
        {[
          { label: "Email Alerts", desc: "Reservation confirmations, cancellations", state: notifEmail, set: setNotifEmail },
          { label: "SMS Alerts", desc: "Critical alerts only", state: notifSMS, set: setNotifSMS },
          { label: "Push Notifications", desc: "Real-time updates in browser", state: notifPush, set: setNotifPush },
        ].map((n) => (
          <div key={n.label} className="flex items-center justify-between">
            <div>
              <p className="text-sm text-white">{n.label}</p>
              <p className="text-xs text-slate-500">{n.desc}</p>
            </div>
            <button
              onClick={() => n.set(!n.state)}
              className={`rounded-full p-1 transition-colors ${n.state ? "bg-nexus-500 text-white" : "bg-slate-700 text-slate-500"}`}
            >
              {n.state ? <Check className="h-4 w-4" /> : <X className="h-4 w-4" />}
            </button>
          </div>
        ))}
      </div>

      {/* Integrations */}
      <div className="card space-y-4">
        <div className="flex items-center gap-2">
          <Plug className="h-5 w-5 text-nexus-400" />
          <h3 className="text-sm font-semibold text-slate-200">Integrations</h3>
        </div>
        {[
          { name: "Booking.com", status: "connected", icon: Globe },
          { name: "Expedia", status: "connected", icon: Globe },
          { name: "Airbnb", status: "disconnected", icon: Globe },
          { name: "Stripe", status: "connected", icon: DollarSign },
        ].map((intg) => (
          <div key={intg.name} className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <intg.icon className="h-4 w-4 text-slate-400" />
              <div>
                <p className="text-sm text-white">{intg.name}</p>
                <span className={`text-xs ${intg.status === "connected" ? "text-emerald-400" : "text-red-400"}`}>
                  {intg.status}
                </span>
              </div>
            </div>
            <button className="btn-secondary text-xs">{intg.status === "connected" ? "Configure" : "Connect"}</button>
          </div>
        ))}
      </div>
    </div>
  );
}
