"use client";

import { useState } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import { ShieldCheck, Search, FileText, Trash2, Download, Lock, ArrowUpRight } from "lucide-react";

export default function AuditPage() {
  const { config } = useTenant();
  const [search, setSearch] = useState("");

  const hasAudit = hasCapability(config, CAPABILITIES.CORE.AUDIT_LOGS);

  if (!hasAudit) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-800">
          <Lock className="h-8 w-8 text-slate-500" />
        </div>
        <h2 className="mt-6 text-xl font-bold text-white">Audit Logs Locked</h2>
        <p className="mt-2 max-w-md text-sm text-slate-400">
          Audit logging requires a higher license tier.
        </p>
        <button className="btn-primary mt-6">
          <ArrowUpRight className="h-4 w-4" />
          Upgrade License
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Audit Logs</h2>
          <p className="text-sm text-slate-400">Activity tracking and compliance</p>
        </div>
        <div className="flex gap-2">
          <button className="btn-secondary">
            <Download className="h-4 w-4" /> Export
          </button>
          <button className="btn-secondary">
            <FileText className="h-4 w-4" /> GDPR Request
          </button>
        </div>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
        <input className="input w-full max-w-md pl-9" placeholder="Search audit logs..." value={search} onChange={(e) => setSearch(e.target.value)} />
      </div>

      <div className="card">
        <p className="text-sm text-slate-500">No audit logs to display.</p>
      </div>
    </div>
  );
}
