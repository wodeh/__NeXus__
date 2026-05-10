"use client";

import { useState } from "react";
import { Settings, Bell, Shield, Globe, Database, Save } from "lucide-react";

export default function SettingsPage() {
  const [tenantName, setTenantName] = useState("Demo Hotel");
  const [timezone, setTimezone] = useState("America/New_York");
  const [currency, setCurrency] = useState("USD");

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">Settings</h2>
        <p className="text-sm text-slate-400">Configure your property and system preferences</p>
      </div>

      <div className="grid gap-6 lg:grid-cols-2">
        <div className="card space-y-4">
          <div className="flex items-center gap-2">
            <Settings className="h-5 w-5 text-nexus-400" />
            <h3 className="font-semibold text-white">General</h3>
          </div>
          <div>
            <label className="block text-sm text-slate-400">Property Name</label>
            <input className="input mt-1 w-full" value={tenantName} onChange={(e) => setTenantName(e.target.value)} />
          </div>
          <div>
            <label className="block text-sm text-slate-400">Timezone</label>
            <select className="input mt-1 w-full" value={timezone} onChange={(e) => setTimezone(e.target.value)}>
              <option>America/New_York</option>
              <option>America/Los_Angeles</option>
              <option>Europe/London</option>
              <option>Asia/Dubai</option>
              <option>Asia/Singapore</option>
            </select>
          </div>
          <div>
            <label className="block text-sm text-slate-400">Currency</label>
            <select className="input mt-1 w-full" value={currency} onChange={(e) => setCurrency(e.target.value)}>
              <option>USD</option>
              <option>EUR</option>
              <option>GBP</option>
              <option>AED</option>
              <option>SGD</option>
            </select>
          </div>
        </div>

        <div className="card space-y-4">
          <div className="flex items-center gap-2">
            <Bell className="h-5 w-5 text-nexus-400" />
            <h3 className="font-semibold text-white">Notifications</h3>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-300">Email alerts</span>
            <input type="checkbox" defaultChecked className="h-4 w-4 accent-nexus-500" />
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-300">SMS alerts</span>
            <input type="checkbox" className="h-4 w-4 accent-nexus-500" />
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-300">Push notifications</span>
            <input type="checkbox" defaultChecked className="h-4 w-4 accent-nexus-500" />
          </div>
        </div>

        <div className="card space-y-4">
          <div className="flex items-center gap-2">
            <Shield className="h-5 w-5 text-nexus-400" />
            <h3 className="font-semibold text-white">Security</h3>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-300">Two-factor authentication</span>
            <input type="checkbox" className="h-4 w-4 accent-nexus-500" />
          </div>
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-300">Session timeout (30 min)</span>
            <input type="checkbox" defaultChecked className="h-4 w-4 accent-nexus-500" />
          </div>
        </div>

        <div className="card space-y-4">
          <div className="flex items-center gap-2">
            <Database className="h-5 w-5 text-nexus-400" />
            <h3 className="font-semibold text-white">System</h3>
          </div>
          <div>
            <label className="block text-sm text-slate-400">API URL</label>
            <input className="input mt-1 w-full" defaultValue="http://localhost:8080" readOnly />
          </div>
          <div>
            <label className="block text-sm text-slate-400">Version</label>
            <p className="mt-1 text-sm text-slate-300">NeXus PMS v1.0.0</p>
          </div>
        </div>
      </div>

      <div className="flex justify-end">
        <button className="btn-primary">
          <Save className="h-4 w-4" />
          Save Changes
        </button>
      </div>
    </div>
  );
}
