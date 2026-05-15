"use client";

import { useEffect, useState } from "react";
import { AlertTriangle, ShieldCheck, TrendingUp } from "lucide-react";
import { getOverbookingConfidence, OverbookingConfidence } from "@/lib/api";

export default function OverbookingWidget({ date }: { date?: string }) {
  const [scores, setScores] = useState<OverbookingConfidence[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    async function fetch() {
      try {
        const data = await getOverbookingConfidence(date);
        setScores(data);
      } catch (e) {
        console.error(e);
      } finally {
        setLoading(false);
      }
    }
    fetch();
  }, [date]);

  if (loading) return <div className="text-xs text-slate-400">Loading...</div>;
  if (scores.length === 0) return null;

  return (
    <div className="rounded-lg border border-slate-700 bg-slate-800 p-3">
      <div className="mb-2 flex items-center gap-2">
        <TrendingUp className="h-4 w-4 text-nexus-400" />
        <span className="text-sm font-semibold text-white">Overbooking Confidence</span>
      </div>
      <div className="space-y-2">
        {scores.map((s) => (
          <div key={s.room_type} className="flex items-center justify-between">
            <div className="flex items-center gap-2">
              {s.confidence >= 80 ? (
                <ShieldCheck className="h-3.5 w-3.5 text-emerald-400" />
              ) : s.confidence >= 50 ? (
                <AlertTriangle className="h-3.5 w-3.5 text-amber-400" />
              ) : (
                <AlertTriangle className="h-3.5 w-3.5 text-rose-400" />
              )}
              <span className="text-xs text-slate-300">{s.room_type}</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="h-2 w-16 overflow-hidden rounded-full bg-slate-700">
                <div
                  className={`h-full rounded-full ${
                    s.confidence >= 80 ? "bg-emerald-500" : s.confidence >= 50 ? "bg-amber-500" : "bg-rose-500"
                  }`}
                  style={{ width: `${s.confidence}%` }}
                />
              </div>
              <span className={`text-xs font-semibold ${
                s.confidence >= 80 ? "text-emerald-400" : s.confidence >= 50 ? "text-amber-400" : "text-rose-400"
              }`}>
                {s.confidence.toFixed(0)}%
              </span>
              {s.suggested_overbook > 0 && (
                <span className="rounded bg-nexus-500/10 px-1.5 py-0.5 text-[10px] text-nexus-400">+{s.suggested_overbook}</span>
              )}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
