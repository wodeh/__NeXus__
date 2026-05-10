"use client";

import { useState } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import { Lock, Plus, Star, Phone, Mail, ArrowUpRight, MessageSquare } from "lucide-react";

interface GuestProfile {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  phone: string;
  vip_status: string;
  loyalty_tier: string;
  total_stays: number;
  total_nights: number;
  total_revenue: number;
  notes: string;
}

const mockProfiles: GuestProfile[] = [
  { id: "gp-2001", first_name: "Alice", last_name: "Chen", email: "alice.chen@example.com", phone: "+1-555-0101", vip_status: "gold", loyalty_tier: "Gold", total_stays: 12, total_nights: 45, total_revenue: 12500, notes: "Prefers quiet rooms. Allergic to down feathers." },
  { id: "gp-2002", first_name: "Bob", last_name: "Jones", email: "bob.jones@example.com", phone: "+1-555-0102", vip_status: "silver", loyalty_tier: "Silver", total_stays: 5, total_nights: 15, total_revenue: 3500, notes: "Business traveler. Early riser." },
];

export default function CRMPage() {
  const { config } = useTenant();
  const [selectedProfile, setSelectedProfile] = useState<GuestProfile | null>(null);

  const hasCap = hasCapability(config, CAPABILITIES.ENTERPRISE.ADVANCED_CRM);

  if (!hasCap) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-800">
          <Lock className="h-8 w-8 text-slate-500" />
        </div>
        <h2 className="mt-6 text-xl font-bold text-white">Guest CRM Locked</h2>
        <p className="mt-2 max-w-md text-sm text-slate-400">VIP tracking, loyalty tiers, communication history, and guest preferences require the Enterprise tier.</p>
        <button className="btn-primary mt-6">
          <ArrowUpRight className="h-4 w-4" /> Upgrade to Enterprise
        </button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Guest CRM</h2>
          <p className="text-sm text-slate-400">Profiles, VIP status, communication log</p>
        </div>
        <button className="btn-primary">
          <Plus className="h-4 w-4" /> Add Profile
        </button>
      </div>

      <div className="grid grid-cols-2 gap-4">
        {mockProfiles.map((p) => (
          <button key={p.id} onClick={() => setSelectedProfile(p)} className="card text-left transition-all hover:border-nexus-500/30">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-slate-800 text-sm font-bold text-white">
                  {p.first_name[0]}{p.last_name[0]}
                </div>
                <div>
                  <p className="text-sm font-medium text-white">{p.first_name} {p.last_name}</p>
                  <p className="text-xs text-slate-500">{p.email}</p>
                </div>
              </div>
              <div className="flex items-center gap-1">
                <Star className="h-3 w-3 text-amber-400" />
                <span className="text-xs font-medium text-amber-400 uppercase">{p.vip_status}</span>
              </div>
            </div>
            <div className="mt-3 flex items-center gap-4 text-xs text-slate-500">
              <span>{p.loyalty_tier} Tier</span>
              <span>{p.total_stays} stays</span>
              <span>{p.total_nights} nights</span>
              <span>${p.total_revenue.toLocaleString()} revenue</span>
            </div>
            <p className="mt-2 text-xs text-slate-400 italic">{p.notes}</p>
          </button>
        ))}
      </div>

      {selectedProfile && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={() => setSelectedProfile(null)}>
          <div className="w-full max-w-lg rounded-xl border border-slate-700 bg-slate-900 p-6" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between">
              <h3 className="text-xl font-bold text-white">{selectedProfile.first_name} {selectedProfile.last_name}</h3>
              <button onClick={() => setSelectedProfile(null)} className="text-slate-400 hover:text-white">×</button>
            </div>
            <div className="mt-4 space-y-2 text-sm">
              <div className="flex justify-between"><span className="text-slate-400">Email</span><span className="text-white">{selectedProfile.email}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Phone</span><span className="text-white">{selectedProfile.phone}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">VIP Status</span><span className="text-amber-400 uppercase">{selectedProfile.vip_status}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Loyalty</span><span className="text-white">{selectedProfile.loyalty_tier}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Total Revenue</span><span className="text-white">${selectedProfile.total_revenue.toLocaleString()}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Notes</span><span className="text-slate-300">{selectedProfile.notes}</span></div>
            </div>
            <div className="mt-6 flex gap-2">
              <button className="btn-primary flex-1"><MessageSquare className="h-4 w-4" /> Log Communication</button>
              <button className="btn-secondary flex-1">Edit Profile</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
