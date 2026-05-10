'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Progress } from '@/components/ui/progress';

// Loyalty Program Dashboard — guest tiers, points, rewards

export default function LoyaltyPage() {
  const [selectedTier, setSelectedTier] = useState('gold');

  const tiers = [
    { id: 'silver', name: 'Silver', level: 1, min: 0, max: 4999, discount: 5, perks: ['Late checkout (1pm)', '10% spa discount'] },
    { id: 'gold', name: 'Gold', level: 2, min: 5000, max: 14999, discount: 10, perks: ['Late checkout (3pm)', 'Free breakfast', 'Room upgrades', 'Priority support'] },
    { id: 'platinum', name: 'Platinum', level: 3, min: 15000, max: 50000, discount: 15, perks: ['Late checkout (6pm)', 'Free breakfast', 'Guaranteed upgrades', 'Priority support', 'Free airport transfer'] },
  ];

  const guests = [
    { name: 'John Smith', tier: 'platinum', points: 18450, stays: 45, nights: 128, revenue: 45200 },
    { name: 'Maria Garcia', tier: 'gold', points: 8200, stays: 22, nights: 67, revenue: 18500 },
    { name: 'James Wilson', tier: 'gold', points: 6100, stays: 18, nights: 52, revenue: 14200 },
    { name: 'Lisa Chen', tier: 'silver', points: 3200, stays: 12, nights: 35, revenue: 8900 },
    { name: 'Robert Taylor', tier: 'silver', points: 1800, stays: 8, nights: 22, revenue: 5200 },
  ];

  const rewards = [
    { name: 'Room Upgrade', points: 2000, type: 'room_upgrade', available: true },
    { name: 'Late Checkout', points: 1000, type: 'late_checkout', available: true },
    { name: 'Free Breakfast', points: 1500, type: 'free_breakfast', available: true },
    { name: 'Spa Treatment', points: 3000, type: 'service', available: false },
    { name: 'Free Night', points: 10000, type: 'free_night', available: true },
  ];

  const currentTier = tiers.find(t => t.id === selectedTier);

  return (
    <div className="min-h-screen bg-slate-950 text-white p-6">
      <div className="max-w-7xl mx-auto">
        <header className="mb-8">
          <h1 className="text-3xl font-bold">Loyalty Program</h1>
          <p className="text-slate-400 mt-1">Guest rewards and tier management</p>
        </header>

        {/* Tier Cards */}
        <div className="grid grid-cols-3 gap-4 mb-6">
          {tiers.map((tier) => (
            <Card
              key={tier.id}
              className={`cursor-pointer transition-all ${selectedTier === tier.id ? 'bg-amber-500/20 border-amber-500' : 'bg-slate-800/50 border-slate-700'}`}
              onClick={() => setSelectedTier(tier.id)}
            >
              <CardContent className="pt-6">
                <div className="flex items-center justify-between mb-3">
                  <h3 className="text-xl font-bold">{tier.name}</h3>
                  <Badge className={tier.id === 'platinum' ? 'bg-purple-500' : tier.id === 'gold' ? 'bg-amber-500' : 'bg-slate-500'}>
                    Level {tier.level}
                  </Badge>
                </div>
                <p className="text-sm text-slate-400 mb-3">{tier.min.toLocaleString()} - {tier.max === 50000 ? '∞' : tier.max.toLocaleString()} points</p>
                <p className="text-2xl font-bold mb-3">{tier.discount}% Discount</p>
                <div className="space-y-1">
                  {tier.perks.map((perk, i) => (
                    <p key={i} className="text-sm text-slate-300">✓ {perk}</p>
                  ))}
                </div>
              </CardContent>
            </Card>
          ))}
        </div>

        <div className="grid grid-cols-2 gap-6">
          {/* Top Guests */}
          <Card className="bg-slate-800/50 border-slate-700">
            <CardHeader>
              <CardTitle>Top Guests</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {guests.map((guest, i) => (
                  <div key={i} className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg">
                    <div className="flex items-center gap-3">
                      <div className="w-10 h-10 rounded-full bg-slate-600 flex items-center justify-center font-medium">
                        {guest.name[0]}
                      </div>
                      <div>
                        <p className="font-medium text-white">{guest.name}</p>
                        <p className="text-sm text-slate-400">{guest.stays} stays • {guest.nights} nights</p>
                      </div>
                    </div>
                    <div className="text-right">
                      <p className="font-mono text-amber-400">{guest.points.toLocaleString()} pts</p>
                      <Badge variant="outline">{guest.tier}</Badge>
                    </div>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>

          {/* Rewards Catalog */}
          <Card className="bg-slate-800/50 border-slate-700">
            <CardHeader className="flex flex-row items-center justify-between">
              <CardTitle>Rewards Catalog</CardTitle>
              <Button size="sm" className="bg-emerald-500 hover:bg-emerald-600">Add Reward</Button>
            </CardHeader>
            <CardContent>
              <div className="space-y-3">
                {rewards.map((reward, i) => (
                  <div key={i} className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg">
                    <div>
                      <p className="font-medium text-white">{reward.name}</p>
                      <p className="text-sm text-slate-400">{reward.points.toLocaleString()} points</p>
                    </div>
                    <Badge variant={reward.available ? 'default' : 'secondary'}>
                      {reward.available ? 'Available' : 'Out of Stock'}
                    </Badge>
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
}
