// @ts-nocheck
"use client";

import { useState, useEffect, useCallback } from "react";
import { Plus, Search, Building2, MapPin, ChevronRight, RefreshCw, Globe, Clock, DollarSign } from "lucide-react";
import { Property, getProperties, createProperty } from "@/lib/api";

export default function PropertiesPage() {
  const [properties, setProperties] = useState<Property[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [showAdd, setShowAdd] = useState(false);
  const [saving, setSaving] = useState(false);
  const [form, setForm] = useState({ name: "", timezone: "UTC", locale: "en", currency: "USD" });

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const props = await getProperties();
      setProperties(props);
    } catch (e: any) {
      setError(e.message || "Failed to load properties");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const filtered = properties.filter((p) =>
    p.name.toLowerCase().includes(search.toLowerCase())
  );

  async function handleSave() {
    if (!form.name.trim()) return;
    setSaving(true);
    try {
      await createProperty(form);
      setShowAdd(false);
      setForm({ name: "", timezone: "UTC", locale: "en", currency: "USD" });
      await fetchData();
    } catch (e: any) {
      setError(e.message);
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Properties</h2>
          <p className="text-sm text-slate-400">Manage hotels and locations</p>
        </div>
        <div className="flex items-center gap-2">
          <button
            onClick={fetchData}
            className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white"
            title="Refresh"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
          </button>
          <button onClick={() => setShowAdd(true)} className="btn-primary">
            <Plus className="h-4 w-4" />
            Add Property
          </button>
        </div>
      </div>

      {error && (
        <div className="rounded-lg border border-rose-500/20 bg-rose-500/5 px-4 py-3 text-sm text-rose-400">
          {error}
        </div>
      )}

      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
        <input
          className="input w-full max-w-md pl-9"
          placeholder="Search properties..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      {loading ? (
        <div className="flex items-center justify-center py-20 text-slate-500">
          <RefreshCw className="mr-2 h-5 w-5 animate-spin" /> Loading properties...
        </div>
      ) : (
        <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {filtered.map((p) => (
            <div key={p.id} className="card-hover">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-nexus-500/10">
                    <Building2 className="h-5 w-5 text-nexus-400" />
                  </div>
                  <div>
                    <p className="font-medium text-white">{p.name}</p>
                    <span className="text-xs badge-green">active</span>
                  </div>
                </div>
                <ChevronRight className="h-4 w-4 text-slate-600" />
              </div>
              <div className="mt-4 space-y-2 text-sm text-slate-400">
                <div className="flex items-center gap-2">
                  <MapPin className="h-3.5 w-3.5" />
                  {p.locale?.toUpperCase() || "—"}
                </div>
                <div className="flex items-center gap-4">
                  <span className="flex items-center gap-1">
                    <Globe className="h-3.5 w-3.5" /> {p.timezone}
                  </span>
                  <span className="flex items-center gap-1">
                    <DollarSign className="h-3.5 w-3.5" /> {p.currency}
                  </span>
                </div>
                <p className="text-xs text-slate-500">
                  Created {new Date(p.created_at).toLocaleDateString()}
                </p>
              </div>
            </div>
          ))}
          {filtered.length === 0 && (
            <div className="col-span-full rounded-lg border border-dashed border-slate-700 py-12 text-center text-slate-500">
              No properties found. {search ? "Try a different search." : "Add your first property."}
            </div>
          )}
        </div>
      )}

      {/* Add Property Modal */}
      {showAdd && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm px-4">
          <div className="w-full max-w-md rounded-2xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
            <h3 className="text-lg font-bold text-white mb-4">Add Property</h3>
            <div className="space-y-3">
              <div>
                <label className="text-xs text-slate-500">Property Name</label>
                <input
                  className="input w-full mt-1"
                  value={form.name}
                  onChange={(e) => setForm({ ...form, name: e.target.value })}
                  placeholder="e.g. Grand Plaza Resort"
                />
              </div>
              <div className="grid grid-cols-3 gap-3">
                <div>
                  <label className="text-xs text-slate-500">Timezone</label>
                  <input
                    className="input w-full mt-1"
                    value={form.timezone}
                    onChange={(e) => setForm({ ...form, timezone: e.target.value })}
                  />
                </div>
                <div>
                  <label className="text-xs text-slate-500">Locale</label>
                  <input
                    className="input w-full mt-1"
                    value={form.locale}
                    onChange={(e) => setForm({ ...form, locale: e.target.value })}
                  />
                </div>
                <div>
                  <label className="text-xs text-slate-500">Currency</label>
                  <input
                    className="input w-full mt-1"
                    value={form.currency}
                    onChange={(e) => setForm({ ...form, currency: e.target.value })}
                  />
                </div>
              </div>
            </div>
            <div className="mt-6 flex items-center justify-end gap-2">
              <button onClick={() => setShowAdd(false)} className="btn-secondary">Cancel</button>
              <button onClick={handleSave} disabled={saving || !form.name.trim()} className="btn-primary disabled:opacity-50">
                {saving ? "Saving..." : "Save Property"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
