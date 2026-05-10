"use client";

import { useState } from "react";
import { Crown, Star, Gift, TrendingUp } from "lucide-react";

const mockTiers = [
  { name: "Silver", min_nights: 0, discount: 5, color: "text-slate-300" },
  { name: "Gold", min_nights: 10, discount: 10, color: "text-amber-400" },
  { name: "Platinum", min_nights: 25, discount: 15, color: "text-nexus-300" },
];

const mockMembers = [
  { id: "m-001", name: "Alice Chen", tier: "Gold", points: 2450, nights: 14 },
  { id: "m-002", name: "Bob Smith", tier: "Silver", points: 820, nights: 4 },
  { id: "m-003", name: "Carol Jones", tier: "Platinum", points: 5100, nights: 32 },
];

export default function LoyaltyPage() {
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">Loyalty Program</h2>
        <p className="text-sm text-slate-400">Guest rewards and tier management</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-3">
        {mockTiers.map((tier) => (
          <div key={tier.name} className="card text-center">
            <Crown className={`mx-auto h-8 w-8 ${tier.color}`} />
            <p className="mt-3 text-lg font-bold text-white">{tier.name}</p>
            <p className="text-sm text-slate-400">{tier.min_nights}+ nights</p>
            <p className="mt-1 text-sm text-nexus-400">{tier.discount}% discount</p>
          </div>
        ))}
      </div>

      <div className="card">
        <h3 className="text-sm font-semibold text-slate-200">Members</h3>
        <div className="mt-4 space-y-3">
          {mockMembers.map((m) => (
            <div key={m.id} className="flex items-center justify-between rounded-lg bg-slate-800/50 px-4 py-3">
              <div className="flex items-center gap-3">
                <div className="flex h-8 w-8 items-center justify-center rounded-full bg-nexus-500/10">
                  <Star className="h-4 w-4 text-nexus-400" />
                </div>
                <div>
                  <p className="font-medium text-white">{m.name}</p>
                  <p className="text-xs text-slate-400">{m.tier} Tier</p>
                </div>
              </div>
              <div className="text-right text-sm">
                <p className="text-white">{m.points} pts</p>
                <p className="text-slate-400">{m.nights} nights</p>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
