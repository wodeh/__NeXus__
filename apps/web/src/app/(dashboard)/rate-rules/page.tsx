// @ts-nocheck
"use client";

import { useEffect, useState } from "react";
import { Plus, Trash2, Calculator } from "lucide-react";
import { getRateRules, createRateRule, deleteRateRule, calculateRate, RateRule, RateRuleCreateRequest } from "@/lib/api";

export default function RateRulesPage() {
  const [rules, setRules] = useState<RateRule[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAdd, setShowAdd] = useState(false);
  const [calcResult, setCalcResult] = useState<any>(null);

  const [form, setForm] = useState<RateRuleCreateRequest>({
    name: "",
    room_type: "",
    condition_type: "season",
    condition_value: "",
    rate_adjustment_type: "percentage",
    rate_adjustment_value: 0,
    min_nights: 1,
    priority: 0,
  });

  async function fetchRules() {
    setLoading(true);
    try {
      const data = await getRateRules();
      setRules(data);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    fetchRules();
  }, []);

  async function handleCreate() {
    try {
      await createRateRule(form);
      setShowAdd(false);
      fetchRules();
    } catch (e: any) {
      alert(e.message);
    }
  }

  async function handleDelete(id: string) {
    if (!confirm("Delete this rule?")) return;
    try {
      await deleteRateRule(id);
      fetchRules();
    } catch (e: any) {
      alert(e.message);
    }
  }

  async function handleCalculate() {
    try {
      const result = await calculateRate({
        room_type: form.room_type || "standard",
        check_in: new Date().toISOString().split("T")[0],
        check_out: new Date(Date.now() + 86400000).toISOString().split("T")[0],
      });
      setCalcResult(result);
    } catch (e: any) {
      alert(e.message);
    }
  }

  return (
    <div className="space-y-4 p-4">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-bold text-white">Rate Rules</h1>
          <p className="text-sm text-slate-400">Dynamic pricing based on conditions.</p>
        </div>
        <button onClick={() => setShowAdd(true)} className="btn-primary flex items-center gap-2">
          <Plus className="h-4 w-4" /> Add Rule
        </button>
      </div>

      {loading ? (
        <div className="text-sm text-slate-400">Loading...</div>
      ) : rules.length === 0 ? (
        <div className="rounded-lg border border-slate-700 bg-slate-800 p-8 text-center text-sm text-slate-400">No rate rules yet.</div>
      ) : (
        <div className="space-y-2">
          {rules.map((rule) => (
            <div key={rule.id} className="flex items-center justify-between rounded-lg border border-slate-700 bg-slate-800 p-3">
              <div>
                <div className="text-sm font-medium text-white">{rule.name}</div>
                <div className="text-xs text-slate-400">
                  {rule.condition_type} · {rule.rate_adjustment_type} {rule.rate_adjustment_value}
                  {rule.room_type && ` · ${rule.room_type}`}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <span className={`rounded px-2 py-1 text-xs ${rule.active ? "bg-emerald-500/10 text-emerald-400" : "bg-slate-700 text-slate-400"}`}>
                  {rule.active ? "Active" : "Inactive"}
                </span>
                <button onClick={() => handleDelete(rule.id)} className="text-rose-400 hover:text-rose-300">
                  <Trash2 className="h-4 w-4" />
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {showAdd && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
            <h3 className="mb-4 text-lg font-bold text-white">Add Rate Rule</h3>
            <div className="space-y-3">
              <input className="input w-full" placeholder="Rule name *" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
              <select className="input w-full" value={form.condition_type} onChange={(e) => setForm({ ...form, condition_type: e.target.value })}>
                <option value="season">Season</option>
                <option value="day_of_week">Day of Week</option>
                <option value="length_of_stay">Length of Stay</option>
                <option value="advance_booking">Advance Booking</option>
                <option value="occupancy">Occupancy</option>
              </select>
              <input className="input w-full" placeholder="Condition value" value={form.condition_value} onChange={(e) => setForm({ ...form, condition_value: e.target.value })} />
              <select className="input w-full" value={form.rate_adjustment_type} onChange={(e) => setForm({ ...form, rate_adjustment_type: e.target.value })}>
                <option value="percentage">Percentage</option>
                <option value="fixed_amount">Fixed Amount</option>
                <option value="fixed_rate">Fixed Rate</option>
              </select>
              <input className="input w-full" type="number" placeholder="Adjustment value (cents or %)" value={form.rate_adjustment_value} onChange={(e) => setForm({ ...form, rate_adjustment_value: parseInt(e.target.value) || 0 })} />
              <input className="input w-full" placeholder="Room type (optional)" value={form.room_type} onChange={(e) => setForm({ ...form, room_type: e.target.value })} />
              <input className="input w-full" type="number" placeholder="Min nights" value={form.min_nights} onChange={(e) => setForm({ ...form, min_nights: parseInt(e.target.value) || 1 })} />
              <input className="input w-full" type="number" placeholder="Priority" value={form.priority} onChange={(e) => setForm({ ...form, priority: parseInt(e.target.value) || 0 })} />
            </div>
            <div className="mt-4 flex items-center justify-between">
              <button onClick={handleCalculate} className="btn-secondary flex items-center gap-2 text-sm">
                <Calculator className="h-4 w-4" /> Test Calc
              </button>
              <div className="flex gap-2">
                <button onClick={() => setShowAdd(false)} className="btn-secondary text-sm">Cancel</button>
                <button onClick={handleCreate} className="btn-primary text-sm">Save</button>
              </div>
            </div>
            {calcResult && (
              <div className="mt-3 rounded bg-slate-800 p-3 text-xs text-slate-300">
                <div>Base: ${(calcResult.base_rate / 100).toFixed(2)} · Nights: {calcResult.nights}</div>
                {calcResult.adjustments.map((a: any) => (
                  <div key={a.rule_name} className="text-slate-400">{a.rule_name}: {a.amount > 0 ? "+" : ""}${(a.amount / 100).toFixed(2)}</div>
                ))}
                <div className="font-semibold text-white">Final: ${(calcResult.final_rate / 100).toFixed(2)}</div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
