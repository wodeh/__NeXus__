"use client";

import { useState, useEffect, useCallback } from "react";
import {
  Route, Zap, Mail, MessageSquare, Clock, ChevronRight, Plus, X,
  Edit3, Trash2, Play, Pause, RefreshCw, AlertTriangle, Target,
  Users, Smartphone, CheckCircle2,
} from "lucide-react";
import {
  getJourneys, getJourneyExecutions, createJourney, updateJourney, deleteJourney,
  GuestJourney, GuestJourneyExecution,
} from "@/lib/api";

export default function GuestJourneyPage() {
  const [tab, setTab] = useState<"journeys" | "executions">("journeys");
  const [journeys, setJourneys] = useState<GuestJourney[]>([]);
  const [executions, setExecutions] = useState<GuestJourneyExecution[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [editing, setEditing] = useState<GuestJourney | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [j, e] = await Promise.all([getJourneys(), getJourneyExecutions()]);
      setJourneys(j);
      setExecutions(e);
    } catch (e: any) {
      setError(e.message || "Failed to load");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleCreate = async (journey: Partial<GuestJourney>) => {
    try {
      await createJourney(journey as any);
      setShowCreate(false);
      fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleUpdate = async (id: string, updates: Partial<GuestJourney>) => {
    try {
      await updateJourney(id, updates);
      setEditing(null);
      fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this journey?")) return;
    try {
      await deleteJourney(id);
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
            <Route className="h-6 w-6 text-nexus-400" /> Guest Journey
          </h1>
          <p className="text-sm text-slate-400">AI-powered automation sequences · Pre-arrival upsells · Post-stay rebooking</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={() => setShowCreate(true)} className="btn-primary gap-2">
            <Plus className="h-4 w-4" /> New Journey
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        <StatCard label="Active Journeys" value={journeys.filter(j => j.is_active).length} icon={Route} color="nexus" />
        <StatCard label="Running Executions" value={executions.filter(e => e.status === "running").length} icon={Play} color="emerald" />
        <StatCard label="Completed" value={executions.filter(e => e.status === "completed").length} icon={CheckCircle2} color="amber" />
        <StatCard label="Total Touchpoints" value={journeys.reduce((s, j) => s + (j.steps?.length || 0), 0)} icon={Zap} color="violet" />
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        {(["journeys", "executions"] as const).map((t) => (
          <button key={t} onClick={() => setTab(t)}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${tab === t ? "bg-slate-800 text-white" : "text-slate-400 hover:bg-slate-800 hover:text-white"}`}
          >
            {t === "journeys" ? "Journey Builder" : "Live Executions"}
          </button>
        ))}
      </div>

      {/* Journeys Tab */}
      {tab === "journeys" && (
        <div className="space-y-4">
          {journeys.map((journey) => (
            <div key={journey.id} className="card space-y-4">
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <div className={`flex h-10 w-10 items-center justify-center rounded-lg ${journey.is_active ? "bg-emerald-500/10" : "bg-slate-800"}`}>
                    <Route className={`h-5 w-5 ${journey.is_active ? "text-emerald-400" : "text-slate-500"}`} />
                  </div>
                  <div>
                    <p className="text-lg font-semibold text-white">{journey.name}</p>
                    <div className="flex items-center gap-2 text-xs text-slate-400">
                      <span className="rounded bg-slate-800 px-2 py-0.5 capitalize">{journey.trigger?.replace("_", " ") || "-"}</span>
                      <span>{journey.steps?.length || 0} steps</span>
                      {journey.is_active && <span className="text-emerald-400">Active</span>}
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <button onClick={() => setEditing(journey)} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white">
                    <Edit3 className="h-4 w-4" />
                  </button>
                  <button onClick={() => handleDelete(journey.id)} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-rose-400">
                    <Trash2 className="h-4 w-4" />
                  </button>
                </div>
              </div>

              {/* Steps */}
              <div className="space-y-2">
                {journey.steps?.map((step, i) => (
                  <div key={step.id} className="flex items-center gap-3 rounded bg-slate-800/50 p-3">
                    <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-slate-700 text-xs font-bold text-white">
                      {i + 1}
                    </div>
                    <div className="flex-1">
                      <div className="flex items-center gap-2">
                        {step.channel === "email" && <Mail className="h-4 w-4 text-sky-400" />}
                        {step.channel === "whatsapp" && <MessageSquare className="h-4 w-4 text-emerald-400" />}
                        {step.channel === "sms" && <Smartphone className="h-4 w-4 text-violet-400" />}
                        <span className="text-sm text-white capitalize">{step.channel}</span>
                        <ChevronRight className="h-3 w-3 text-slate-500" />
                        <span className="text-sm text-slate-300">{step.template_id.replace("_", " ")}</span>
                      </div>
                    </div>
                    <div className="flex items-center gap-3 text-xs text-slate-400">
                      <span className="flex items-center gap-1"><Clock className="h-3 w-3" /> +{step.delay_hours}h</span>
                      <span className="rounded bg-slate-700 px-2 py-0.5 capitalize">{step.condition.replace("_", " ")}</span>
                      {step.upsell_offer_id && <span className="rounded bg-nexus-500/10 px-2 py-0.5 text-nexus-400">Upsell</span>}
                    </div>
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Executions Tab */}
      {tab === "executions" && (
        <div className="card space-y-4">
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-800 text-xs uppercase text-slate-400">
              <tr>
                <th className="px-4 py-2">Reservation</th>
                <th className="px-4 py-2">Journey</th>
                <th className="px-4 py-2">Progress</th>
                <th className="px-4 py-2">Status</th>
                <th className="px-4 py-2">Started</th>
                <th className="px-4 py-2">Next Step</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {executions.map((e) => (
                <tr key={e.id} className="hover:bg-slate-800/50">
                  <td className="px-4 py-3">
                    <p className="text-white">{e.reservation_id.slice(0, 8)}</p>
                    <p className="text-xs text-slate-500">{e.guest_phone}</p>
                  </td>
                  <td className="px-4 py-3 text-slate-300">{e.journey_id.slice(0, 8)}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <div className="h-2 w-24 rounded-full bg-slate-800">
                        <div className="h-full rounded-full bg-nexus-500" style={{ width: `${(e.current_step / Math.max(e.total_steps, 1)) * 100}%` }} />
                      </div>
                      <span className="text-xs text-slate-400">{e.current_step}/{e.total_steps}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span className={`rounded px-2 py-0.5 text-xs capitalize ${
                      e.status === "running" ? "bg-emerald-500/10 text-emerald-400" :
                      e.status === "completed" ? "bg-sky-500/10 text-sky-400" :
                      e.status === "failed" ? "bg-rose-500/10 text-rose-400" :
                      "bg-amber-500/10 text-amber-400"
                    }`}>
                      {e.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-xs text-slate-400">{new Date(e.started_at).toLocaleDateString()}</td>
                  <td className="px-4 py-3 text-xs text-slate-400">{e.next_trigger_at ? new Date(e.next_trigger_at).toLocaleString() : "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Create Modal */}
      {showCreate && <JourneyModal onSave={handleCreate} onClose={() => setShowCreate(false)} />}
      {editing && <JourneyModal journey={editing} onSave={(j) => handleUpdate(editing.id, j)} onClose={() => setEditing(null)} />}
    </div>
  );
}

function StatCard({ label, value, icon: Icon, color }: { label: string; value: number; icon: React.ElementType; color: string }) {
  const colorMap: Record<string, string> = {
    nexus: "text-nexus-400 bg-nexus-500/10",
    emerald: "text-emerald-400 bg-emerald-500/10",
    amber: "text-amber-400 bg-amber-500/10",
    violet: "text-violet-400 bg-violet-500/10",
  };
  return (
    <div className="card space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-xs text-slate-400">{label}</span>
        <div className={`rounded-lg p-1.5 ${colorMap[color]}`}><Icon className="h-3 w-3" /></div>
      </div>
      <p className="text-xl font-bold text-white">{value}</p>
    </div>
  );
}

function JourneyModal({ journey, onSave, onClose }: { journey?: GuestJourney; onSave: (j: Partial<GuestJourney>) => void; onClose: () => void }) {
  const [form, setForm] = useState({
    name: journey?.name || "",
    trigger: journey?.trigger || "booking_confirmed",
    is_active: journey?.is_active ?? true,
    steps: journey?.steps || [],
  });

  const addStep = () => {
    setForm({
      ...form,
      steps: [...form.steps, { id: crypto.randomUUID(), delay_hours: 24, channel: "whatsapp", template_id: "", condition: "always", is_active: true }],
    });
  };

  const updateStep = (index: number, field: string, value: any) => {
    const steps = [...form.steps];
    steps[index] = { ...steps[index], [field]: value };
    setForm({ ...form, steps });
  };

  const removeStep = (index: number) => {
    setForm({ ...form, steps: form.steps.filter((_, i) => i !== index) });
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-2xl rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl space-y-4 max-h-[90vh] overflow-auto">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">{journey ? "Edit Journey" : "New Journey"}</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>

        <div className="space-y-3">
          <div>
            <label className="mb-1 block text-xs text-slate-400">Journey Name</label>
            <input className="input w-full" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Trigger</label>
              <select className="input w-full" value={form.trigger} onChange={(e) => setForm({ ...form, trigger: e.target.value })}>
                <option value="booking_confirmed">Booking Confirmed</option>
                <option value="check_in">Check In</option>
                <option value="check_out">Check Out</option>
                <option value="no_show">No Show</option>
              </select>
            </div>
            <div className="flex items-center gap-2 pt-6">
              <input type="checkbox" id="is_active" checked={form.is_active} onChange={(e) => setForm({ ...form, is_active: e.target.checked })} />
              <label htmlFor="is_active" className="text-sm text-white">Active</label>
            </div>
          </div>
        </div>

        <div className="space-y-2">
          <div className="flex items-center justify-between">
            <p className="text-sm font-medium text-white">Steps</p>
            <button onClick={addStep} className="btn-primary text-xs gap-1"><Plus className="h-3 w-3" /> Add Step</button>
          </div>
          {form.steps.map((step, i) => (
            <div key={step.id} className="rounded bg-slate-800/50 p-3 space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-xs font-bold text-slate-400">Step {i + 1}</span>
                <button onClick={() => removeStep(i)} className="text-rose-400 hover:text-rose-300"><Trash2 className="h-4 w-4" /></button>
              </div>
              <div className="grid grid-cols-4 gap-2">
                <select className="input text-xs" value={step.channel} onChange={(e) => updateStep(i, "channel", e.target.value)}>
                  <option value="email">Email</option>
                  <option value="whatsapp">WhatsApp</option>
                  <option value="sms">SMS</option>
                </select>
                <input type="number" className="input text-xs" placeholder="Delay (hrs)" value={step.delay_hours} onChange={(e) => updateStep(i, "delay_hours", parseInt(e.target.value) || 0)} />
                <input className="input text-xs" placeholder="Template ID" value={step.template_id} onChange={(e) => updateStep(i, "template_id", e.target.value)} />
                <select className="input text-xs" value={step.condition} onChange={(e) => updateStep(i, "condition", e.target.value)}>
                  <option value="always">Always</option>
                  <option value="vip_only">VIP Only</option>
                  <option value="first_time">First Time</option>
                  <option value="returning">Returning</option>
                </select>
              </div>
            </div>
          ))}
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <button onClick={onClose} className="btn-secondary">Cancel</button>
          <button onClick={() => onSave(form)} className="btn-primary gap-2"><CheckCircle2 className="h-4 w-4" /> Save</button>
        </div>
      </div>
    </div>
  );
}
