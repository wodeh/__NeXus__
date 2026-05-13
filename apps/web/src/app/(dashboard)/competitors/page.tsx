"use client";

import { useState, useEffect, useCallback } from "react";
import {
  Target, Plus, Trash2, CheckCircle2, X, RefreshCw, AlertTriangle,
  TrendingUp, TrendingDown, Building2, DollarSign, Zap, ArrowUpRight,
  ArrowDownRight, Clock, Settings, ExternalLink,
} from "lucide-react";
import {
  getCompetitors, getCompetitorRates, getRateRecommendations, applyRateRecommendation,
  getRateShopConfig, updateRateShopConfig,
  createCompetitor, deleteCompetitor,
  CompetitorHotel, CompetitorRate, RateRecommendation, RateShopConfig,
} from "@/lib/api";

export default function CompetitorsPage() {
  const [competitors, setCompetitors] = useState<CompetitorHotel[]>([]);
  const [rates, setRates] = useState<CompetitorRate[]>([]);
  const [recommendations, setRecommendations] = useState<RateRecommendation[]>([]);
  const [config, setConfig] = useState<RateShopConfig | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddCompetitor, setShowAddCompetitor] = useState(false);
  const [showConfig, setShowConfig] = useState(false);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [c, r, rec, cfg] = await Promise.all([
        getCompetitors(),
        getCompetitorRates(),
        getRateRecommendations(),
        getRateShopConfig(),
      ]);
      setCompetitors(c);
      setRates(r);
      setRecommendations(rec);
      setConfig(cfg);
    } catch (e: any) {
      setError(e.message || "Failed to load");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleApplyRec = async (id: string) => {
    try {
      await applyRateRecommendation(id);
      fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleAddCompetitor = async (hotel: Partial<CompetitorHotel>) => {
    try {
      await createCompetitor(hotel as any);
      setShowAddCompetitor(false);
      fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleDeleteCompetitor = async (id: string) => {
    if (!confirm("Remove this competitor?")) return;
    try {
      await deleteCompetitor(id);
      fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleSaveConfig = async (updates: Partial<RateShopConfig>) => {
    try {
      await updateRateShopConfig(updates);
      setShowConfig(false);
      fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  };

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
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Target className="h-6 w-6 text-nexus-400" /> Rate Intelligence
          </h1>
          <p className="text-sm text-slate-400">Competitor rate shopping · AI pricing recommendations · Auto-adjustment</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => setShowConfig(true)} className="btn-secondary gap-2 text-xs">
            <Settings className="h-4 w-4" /> Config
          </button>
          <button onClick={() => setShowAddCompetitor(true)} className="btn-primary gap-2">
            <Plus className="h-4 w-4" /> Add Competitor
          </button>
        </div>
      </div>

      {/* Competitors */}
      <div className="grid grid-cols-3 gap-4">
        {competitors.map((comp) => (
          <div key={comp.id} className="card space-y-3">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Building2 className="h-5 w-5 text-slate-400" />
                <span className="font-semibold text-white">{comp.name}</span>
              </div>
              <button onClick={() => handleDeleteCompetitor(comp.id)} className="text-slate-400 hover:text-rose-400">
                <Trash2 className="h-4 w-4" />
              </button>
            </div>
            <div className="text-xs text-slate-400 space-y-1">
              <p>{comp.address}, {comp.city}</p>
              <div className="flex items-center gap-2">
                <span className="flex items-center gap-1">{"★".repeat(comp.star_rating || 0)}</span>
                <span>{comp.room_count} rooms</span>
              </div>
              {comp.last_scraped && (
                <p className="flex items-center gap-1 text-emerald-400">
                  <Clock className="h-3 w-3" /> Last scraped: {new Date(comp.last_scraped).toLocaleString()}
                </p>
              )}
            </div>
          </div>
        ))}
      </div>

      {/* Rate Comparison Table */}
      <div className="card space-y-4">
        <h3 className="text-lg font-semibold text-white flex items-center gap-2">
          <DollarSign className="h-5 w-5 text-nexus-400" /> Live Rate Comparison
        </h3>
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-800 text-xs uppercase text-slate-400">
            <tr>
              <th className="px-4 py-2">Competitor</th>
              <th className="px-4 py-2">Room Type</th>
              <th className="px-4 py-2">Date</th>
              <th className="px-4 py-2">Rate</th>
              <th className="px-4 py-2">Availability</th>
              <th className="px-4 py-2">Min Stay</th>
              <th className="px-4 py-2">Source</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            {rates.map((rate) => (
              <tr key={rate.id} className="hover:bg-slate-800/50">
                <td className="px-4 py-2 text-white">{rate.competitor_name}</td>
                <td className="px-4 py-2 text-slate-300">{rate.room_type}</td>
                <td className="px-4 py-2 text-slate-400">{rate.date}</td>
                <td className="px-4 py-2">
                  <span className={`font-bold ${rate.is_promo ? "text-amber-400" : "text-white"}`}>
                    ${Math.round(rate.rate)}
                  </span>
                  {rate.is_promo && <span className="ml-2 rounded bg-amber-500/10 px-1.5 py-0.5 text-[10px] text-amber-400">PROMO</span>}
                </td>
                <td className="px-4 py-2 text-slate-400">{rate.availability}</td>
                <td className="px-4 py-2 text-slate-400">{rate.min_stay}n</td>
                <td className="px-4 py-2 text-xs text-slate-500 capitalize">{rate.source?.replace("_", " ") || "-"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* AI Recommendations */}
      <div className="card space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-semibold text-white flex items-center gap-2">
            <Zap className="h-5 w-5 text-amber-400" /> AI Rate Recommendations
          </h3>
          <span className="text-xs text-slate-400">{recommendations.length} pending</span>
        </div>

        <div className="space-y-3">
          {recommendations.map((rec) => {
            const diff = rec.recommended_rate - rec.current_rate;
            const diffPct = Math.round((diff / rec.current_rate) * 100);
            const isIncrease = diff > 0;

            return (
              <div key={rec.id} className="flex items-center justify-between rounded bg-slate-800/50 p-4">
                <div className="space-y-1">
                  <div className="flex items-center gap-3">
                    <span className="text-sm font-bold text-white">{rec.room_type}</span>
                    <span className="text-xs text-slate-400">{rec.date}</span>
                    <span className={`flex items-center gap-1 text-xs font-bold ${isIncrease ? "text-emerald-400" : "text-rose-400"}`}>
                      {isIncrease ? <ArrowUpRight className="h-3 w-3" /> : <ArrowDownRight className="h-3 w-3" />}
                      {diffPct > 0 ? `+${diffPct}%` : `${diffPct}%`}
                    </span>
                  </div>
                  <div className="flex items-center gap-3 text-sm">
                    <span className="text-slate-400">Current: <span className="text-white">${Math.round(rec.current_rate)}</span></span>
                    <span className="text-slate-400">→</span>
                    <span className="text-nexus-400 font-bold">${Math.round(rec.recommended_rate)}</span>
                  </div>
                  <p className="text-xs text-slate-500">{rec.reason}</p>
                  <div className="flex items-center gap-2">
                    {rec.factors?.map((f, i) => (
                      <span key={i} className="rounded bg-slate-700 px-2 py-0.5 text-[10px] text-slate-400 capitalize">{f.replace("_", " ")}</span>
                    ))}
                    <span className="rounded bg-sky-500/10 px-2 py-0.5 text-[10px] text-sky-400">{(rec.confidence * 100).toFixed(0)}% confidence</span>
                  </div>
                </div>
                <button
                  onClick={() => handleApplyRec(rec.id)}
                  className="btn-primary text-xs gap-2"
                >
                  <CheckCircle2 className="h-4 w-4" /> Apply
                </button>
              </div>
            );
          })}
        </div>
      </div>

      {/* Add Competitor Modal */}
      {showAddCompetitor && <CompetitorModal onSave={handleAddCompetitor} onClose={() => setShowAddCompetitor(false)} />}

      {/* Config Modal */}
      {showConfig && config && <ConfigModal config={config} onSave={handleSaveConfig} onClose={() => setShowConfig(false)} />}
    </div>
  );
}

function CompetitorModal({ onSave, onClose }: { onSave: (c: Partial<CompetitorHotel>) => void; onClose: () => void }) {
  const [form, setForm] = useState({ name: "", address: "", city: "", country: "", star_rating: 4, room_count: 100, website: "", booking_url: "" });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">Add Competitor</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>
        <div className="space-y-3">
          <input className="input w-full" placeholder="Hotel Name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          <input className="input w-full" placeholder="Address" value={form.address} onChange={(e) => setForm({ ...form, address: e.target.value })} />
          <div className="grid grid-cols-2 gap-3">
            <input className="input w-full" placeholder="City" value={form.city} onChange={(e) => setForm({ ...form, city: e.target.value })} />
            <input className="input w-full" placeholder="Country" value={form.country} onChange={(e) => setForm({ ...form, country: e.target.value })} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <input type="number" className="input w-full" placeholder="Star Rating" value={form.star_rating} onChange={(e) => setForm({ ...form, star_rating: parseInt(e.target.value) || 0 })} />
            <input type="number" className="input w-full" placeholder="Room Count" value={form.room_count} onChange={(e) => setForm({ ...form, room_count: parseInt(e.target.value) || 0 })} />
          </div>
          <input className="input w-full" placeholder="Website URL" value={form.website} onChange={(e) => setForm({ ...form, website: e.target.value })} />
          <input className="input w-full" placeholder="Booking.com URL" value={form.booking_url} onChange={(e) => setForm({ ...form, booking_url: e.target.value })} />
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <button onClick={onClose} className="btn-secondary">Cancel</button>
          <button onClick={() => onSave(form)} className="btn-primary gap-2"><CheckCircle2 className="h-4 w-4" /> Add</button>
        </div>
      </div>
    </div>
  );
}

function ConfigModal({ config, onSave, onClose }: { config: RateShopConfig; onSave: (c: Partial<RateShopConfig>) => void; onClose: () => void }) {
  const [form, setForm] = useState({ ...config });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">Rate Shop Configuration</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>
        <div className="space-y-3">
          <label className="flex items-center gap-2 text-sm text-white">
            <input type="checkbox" checked={form.enabled} onChange={(e) => setForm({ ...form, enabled: e.target.checked })} />
            Enabled
          </label>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Frequency (hours)</label>
            <input type="number" className="input w-full" value={form.frequency_hours} onChange={(e) => setForm({ ...form, frequency_hours: parseInt(e.target.value) || 6 })} />
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Lookahead (days)</label>
            <input type="number" className="input w-full" value={form.lookahead_days} onChange={(e) => setForm({ ...form, lookahead_days: parseInt(e.target.value) || 30 })} />
          </div>
          <label className="flex items-center gap-2 text-sm text-white">
            <input type="checkbox" checked={form.auto_adjust} onChange={(e) => setForm({ ...form, auto_adjust: e.target.checked })} />
            Auto-apply recommendations
          </label>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Max adjustment %</label>
            <input type="number" step="0.01" className="input w-full" value={form.max_adjustment_pct} onChange={(e) => setForm({ ...form, max_adjustment_pct: parseFloat(e.target.value) || 0.30 })} />
          </div>
        </div>
        <div className="flex justify-end gap-2 pt-2">
          <button onClick={onClose} className="btn-secondary">Cancel</button>
          <button onClick={() => onSave(form)} className="btn-primary gap-2"><CheckCircle2 className="h-4 w-4" /> Save</button>
        </div>
      </div>
    </div>
  );
}
