"use client";

import { useState, useEffect, useCallback } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import { Agent, getAgents, updateAgent, deleteAgent } from "@/lib/api";
import { Lock, Plus, Building2, Percent, Mail, Phone, ArrowUpRight, ToggleLeft, ToggleRight, Loader2, Trash2 } from "lucide-react";

export default function AgentsPage() {
  const { config } = useTenant();
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const hasCap = hasCapability(config, CAPABILITIES.REVENUE.AGENT_MANAGEMENT);

  const fetchAgents = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await getAgents();
      setAgents(data);
    } catch (e: any) {
      setError(e.message || "Failed to load agents");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchAgents();
  }, [fetchAgents]);

  const toggleAgent = async (id: string, current: boolean) => {
    try {
      await updateAgent(id, { is_active: !current });
      setAgents(prev => prev.map(a => a.id === id ? { ...a, is_active: !current } : a));
    } catch (e: any) {
      setError(e.message || "Failed to update agent");
    }
  };

  const removeAgent = async (id: string) => {
    if (!confirm("Delete this agent?")) return;
    try {
      await deleteAgent(id);
      setAgents(prev => prev.filter(a => a.id !== id));
    } catch (e: any) {
      setError(e.message || "Failed to delete agent");
    }
  };

  if (!hasCap) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-800">
          <Lock className="h-8 w-8 text-slate-500" />
        </div>
        <h2 className="mt-6 text-xl font-bold text-white">Agent Management Locked</h2>
        <p className="mt-2 max-w-md text-sm text-slate-400">Manage OTAs, travel agents, and corporate partners with commission tracking. Requires Revenue or Enterprise tier.</p>
        <button className="btn-primary mt-6">
          <ArrowUpRight className="h-4 w-4" /> Upgrade to Revenue
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Agents & Partners</h2>
          <p className="text-sm text-slate-400">OTAs, travel agents, corporate accounts</p>
        </div>
        <button className="btn-primary">
          <Plus className="h-4 w-4" /> Add Agent
        </button>
      </div>

      {loading && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-nexus-400" />
        </div>
      )}

      {error && (
        <div className="rounded-lg border border-rose-500/20 bg-rose-500/10 p-4 text-sm text-rose-400">
          {error}
        </div>
      )}

      {!loading && agents.length === 0 && (
        <div className="py-12 text-center text-slate-500">
          No agents yet. Add your first OTA or corporate partner.
        </div>
      )}

      <div className="space-y-3">
        {agents.map((a) => (
          <div key={a.id} className="card">
            <div className="flex items-center justify-between">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="rounded-full bg-nexus-500/10 px-2.5 py-0.5 text-xs font-medium text-nexus-400 uppercase">{a.type.replace("_", " ")}</span>
                  <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${a.is_active ? "bg-emerald-500/10 text-emerald-400" : "bg-slate-500/10 text-slate-400"}`}>
                    {a.is_active ? "Active" : "Inactive"}
                  </span>
                </div>
                <p className="text-sm font-medium text-white">{a.name}</p>
                <div className="flex items-center gap-4 text-xs text-slate-500">
                  <span className="flex items-center gap-1"><Percent className="h-3 w-3" /> {a.commission_pct}% commission</span>
                  <span className="flex items-center gap-1"><Building2 className="h-3 w-3" /> {a.contract_ref || "—"}</span>
                  <span className="flex items-center gap-1"><Mail className="h-3 w-3" /> {a.contact_email || "—"}</span>
                  <span className="flex items-center gap-1"><Phone className="h-3 w-3" /> {a.contact_phone || "—"}</span>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <button className="btn-secondary text-xs">Edit</button>
                <button onClick={() => toggleAgent(a.id, a.is_active)} className={`rounded-lg p-2 ${a.is_active ? "text-emerald-400 hover:bg-emerald-500/10" : "text-slate-500 hover:bg-slate-800"}`}>
                  {a.is_active ? <ToggleRight className="h-5 w-5" /> : <ToggleLeft className="h-5 w-5" />}
                </button>
                <button onClick={() => removeAgent(a.id)} className="rounded-lg p-2 text-slate-500 hover:text-rose-400 hover:bg-rose-500/10">
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
