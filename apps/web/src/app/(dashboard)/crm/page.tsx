"use client";

import { useState, useEffect, useCallback, useMemo } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import { Lock, Plus, Star, Phone, Mail, ArrowUpRight, MessageSquare, RefreshCw, AlertTriangle } from "lucide-react";
import { Reservation, getReservations } from "@/lib/api";

interface GuestProfile {
  id: string;
  name: string;
  email: string;
  phone: string;
  vip: boolean;
  total_stays: number;
  total_nights: number;
  total_revenue: number;
  notes: string;
  lastStay: string;
}

export default function CRMPage() {
  const { config } = useTenant();
  const [selectedProfile, setSelectedProfile] = useState<GuestProfile | null>(null);
  const [profiles, setProfiles] = useState<GuestProfile[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const hasCap = hasCapability(config, CAPABILITIES.ENTERPRISE.ADVANCED_CRM);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const reservations = await getReservations();
      const map = new Map<string, GuestProfile>();
      reservations.forEach((r) => {
        const key = r.email || r.guest_name;
        const nights = Math.max(1, Math.round((new Date(r.check_out).getTime() - new Date(r.check_in).getTime()) / 86400000));
        const revenue = (r.total || 0);
        const existing = map.get(key);
        if (existing) {
          existing.total_stays += 1;
          existing.total_nights += nights;
          existing.total_revenue += (r.total || 0);
          if (r.check_out > existing.lastStay) existing.lastStay = r.check_out;
        } else {
          map.set(key, {
            id: r.id,
            name: r.guest_name,
            email: r.email || "",
            phone: r.phone || "",
            vip: r.vip,
            total_stays: 1,
            total_nights: nights,
            total_revenue: (r.total || 0),
            notes: "",
            lastStay: r.check_out,
          });
        }
      });
      setProfiles(Array.from(map.values()).sort((a, b) => b.total_revenue - a.total_revenue));
    } catch (e: any) {
      setError(e.message || "Failed to load CRM data");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

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
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Guest CRM</h2>
          <p className="text-sm text-slate-400">Profiles, VIP status, communication log</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={fetchData} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white" title="Refresh">
            <RefreshCw className="h-4 w-4" />
          </button>
          <button className="btn-primary">
            <Plus className="h-4 w-4" /> Add Profile
          </button>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        {profiles.map((p) => (
          <button key={p.id} onClick={() => setSelectedProfile(p)} className="card text-left transition-all hover:border-nexus-500/30">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-slate-800 text-sm font-bold text-white">
                  {p.name.split(" ").map(n => n[0]).join("").slice(0,2)}
                </div>
                <div>
                  <p className="text-sm font-medium text-white">{p.name}</p>
                  <p className="text-xs text-slate-500">{p.email || "No email"}</p>
                </div>
              </div>
              <div className="flex items-center gap-1">
                {p.vip && <Star className="h-3 w-3 text-amber-400" />}
                <span className="text-xs font-medium text-amber-400 uppercase">{p.vip ? "VIP" : "Guest"}</span>
              </div>
            </div>
            <div className="mt-3 flex items-center gap-4 text-xs text-slate-500">
              <span>{p.total_stays} stays</span>
              <span>{p.total_nights} nights</span>
              <span>${p.total_revenue.toLocaleString()} revenue</span>
              <span>Last: {p.lastStay}</span>
            </div>
          </button>
        ))}
        {profiles.length === 0 && (
          <div className="col-span-2 py-12 text-center text-sm text-slate-500">
            No guest profiles found.
          </div>
        )}
      </div>

      {selectedProfile && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={() => setSelectedProfile(null)}>
          <div className="w-full max-w-lg rounded-xl border border-slate-700 bg-slate-900 p-6" onClick={(e) => e.stopPropagation()}>
            <div className="flex items-center justify-between">
              <h3 className="text-xl font-bold text-white">{selectedProfile.name}</h3>
              <button onClick={() => setSelectedProfile(null)} className="text-slate-400 hover:text-white">×</button>
            </div>
            <div className="mt-4 space-y-2 text-sm">
              <div className="flex justify-between"><span className="text-slate-400">Email</span><span className="text-white">{selectedProfile.email || "—"}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Phone</span><span className="text-white">{selectedProfile.phone || "—"}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">VIP Status</span><span className="text-amber-400 uppercase">{selectedProfile.vip ? "VIP" : "Standard"}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Total Stays</span><span className="text-white">{selectedProfile.total_stays}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Total Nights</span><span className="text-white">{selectedProfile.total_nights}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Total Revenue</span><span className="text-white">${selectedProfile.total_revenue.toLocaleString()}</span></div>
              <div className="flex justify-between"><span className="text-slate-400">Last Stay</span><span className="text-white">{selectedProfile.lastStay}</span></div>
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
