"use client";

import { useState } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import { Lock, Plus, Building2, Percent, Mail, Phone, ArrowUpRight, ToggleLeft, ToggleRight } from "lucide-react";

interface Agent {
  id: string;
  name: string;
  type: string;
  commission_pct: number;
  contact_name: string;
  contact_email: string;
  contact_phone: string;
  contract_ref: string;
  is_active: boolean;
  source_code: string;
}

const mockAgents: Agent[] = [
  { id: "ag-101", name: "Booking.com", type: "ota", commission_pct: 15, contact_name: "Partner Support", contact_email: "partners@booking.com", contact_phone: "+1-800-BOOKING", contract_ref: "CNT-2026-001", is_active: true, source_code: "BKG" },
  { id: "ag-102", name: "Expedia", type: "ota", commission_pct: 18, contact_name: "Account Manager", contact_email: "hotel@expedia.com", contact_phone: "+1-800-EXPEDIA", contract_ref: "CNT-2026-002", is_active: true, source_code: "EXP" },
  { id: "ag-103", name: "Virtuoso Travel", type: "travel_agent", commission_pct: 10, contact_name: "Jane Smith", contact_email: "jane@virtuoso.com", contact_phone: "+1-555-0300", contract_ref: "CNT-2026-003", is_active: true, source_code: "VRT" },
  { id: "ag-104", name: "TechCorp Corporate", type: "corporate", commission_pct: 0, contact_name: "Mike Chen", contact_email: "mike@techcorp.com", contact_phone: "+1-555-0200", contract_ref: "CNT-2026-004", is_active: true, source_code: "TCP" },
];

export default function AgentsPage() {
  const { config } = useTenant();

  const hasCap = hasCapability(config, CAPABILITIES.REVENUE.AGENT_MANAGEMENT);

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

      <div className="space-y-3">
        {mockAgents.map((a) => (
          <div key={a.id} className="card">
            <div className="flex items-center justify-between">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="rounded-full bg-nexus-500/10 px-2.5 py-0.5 text-xs font-medium text-nexus-400 uppercase">{a.type.replace("_", " ")}</span>
                  <span className="rounded-full bg-emerald-500/10 px-2.5 py-0.5 text-xs font-medium text-emerald-400">{a.is_active ? "Active" : "Inactive"}</span>
                </div>
                <p className="text-sm font-medium text-white">{a.name}</p>
                <div className="flex items-center gap-4 text-xs text-slate-500">
                  <span className="flex items-center gap-1"><Percent className="h-3 w-3" /> {a.commission_pct}% commission</span>
                  <span className="flex items-center gap-1"><Building2 className="h-3 w-3" /> {a.contract_ref}</span>
                  <span className="flex items-center gap-1"><Mail className="h-3 w-3" /> {a.contact_email}</span>
                  <span className="flex items-center gap-1"><Phone className="h-3 w-3" /> {a.contact_phone}</span>
                </div>
              </div>
              <div className="flex items-center gap-2">
                <button className="btn-secondary text-xs">Edit</button>
                {a.is_active ? (
                  <button className="rounded-lg p-2 text-emerald-400 hover:bg-emerald-500/10"><ToggleRight className="h-5 w-5" /></button>
                ) : (
                  <button className="rounded-lg p-2 text-slate-500 hover:bg-slate-800"><ToggleLeft className="h-5 w-5" /></button>
                )}
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
