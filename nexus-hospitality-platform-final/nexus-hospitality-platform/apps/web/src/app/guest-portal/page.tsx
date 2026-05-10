'use client';

import { useState } from 'react';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useAuth } from '@/hooks/useAuth';

// Guest Self-Service Portal — web-based check-in, services, billing

export default function GuestPortalPage() {
  const { user } = useAuth();
  const [reservationCode, setReservationCode] = useState('');
  const [activeTab, setActiveTab] = useState('services');

  const { data: reservation, isLoading } = useQuery({
    queryKey: ['guest-reservation', reservationCode],
    queryFn: async () => {
      if (!reservationCode) return null;
      const res = await fetch(`/api/tenants/demo/reservations?code=${reservationCode}`);
      return res.json();
    },
    enabled: !!reservationCode,
  });

  const orderMutation = useMutation({
    mutationFn: async (orderData: any) => {
      const res = await fetch('/api/tenants/demo/guest-orders', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(orderData),
      });
      return res.json();
    },
  });

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 text-white p-6">
      <div className="max-w-5xl mx-auto">
        <header className="mb-8 text-center">
          <h1 className="text-4xl font-bold mb-2">Guest Portal</h1>
          <p className="text-slate-400">Your stay, simplified</p>
        </header>

        {/* Reservation Lookup */}
        <Card className="mb-6 bg-slate-800/50 border-slate-700">
          <CardContent className="pt-6">
            <div className="flex gap-4">
              <Input
                placeholder="Enter reservation code"
                value={reservationCode}
                onChange={(e) => setReservationCode(e.target.value)}
                className="bg-slate-700 border-slate-600 text-white"
              />
              <Button className="bg-amber-500 hover:bg-amber-600 text-slate-900 font-semibold">
                Look Up
              </Button>
            </div>
          </CardContent>
        </Card>

        {reservation && (
          <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
            <TabsList className="grid w-full grid-cols-4 bg-slate-800">
              <TabsTrigger value="services">Services</TabsTrigger>
              <TabsTrigger value="billing">Billing</TabsTrigger>
              <TabsTrigger value="room">Room Control</TabsTrigger>
              <TabsTrigger value="concierge">Concierge</TabsTrigger>
            </TabsList>

            <TabsContent value="services" className="mt-6">
              <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                {[
                  { name: 'Room Service', icon: '🍽️', desc: 'Order food & drinks' },
                  { name: 'Housekeeping', icon: '🧹', desc: 'Request cleaning' },
                  { name: 'Maintenance', icon: '🔧', desc: 'Report an issue' },
                  { name: 'Spa & Wellness', icon: '💆', desc: 'Book treatments' },
                  { name: 'Transport', icon: '🚗', desc: 'Airport transfer' },
                  { name: 'Late Checkout', icon: '🕐', desc: 'Extend your stay' },
                ].map((service) => (
                  <Card key={service.name} className="bg-slate-800/50 border-slate-700 hover:border-amber-500/50 cursor-pointer transition-all">
                    <CardContent className="pt-6 text-center">
                      <div className="text-4xl mb-3">{service.icon}</div>
                      <h3 className="font-semibold text-white">{service.name}</h3>
                      <p className="text-sm text-slate-400 mt-1">{service.desc}</p>
                    </CardContent>
                  </Card>
                ))}
              </div>
            </TabsContent>

            <TabsContent value="billing" className="mt-6">
              <Card className="bg-slate-800/50 border-slate-700">
                <CardHeader>
                  <CardTitle>Current Charges</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {[
                      { item: 'Room Charge — 3 nights', amount: 450.00 },
                      { item: 'Room Service', amount: 67.50 },
                      { item: 'Resort Fee', amount: 45.00 },
                    ].map((charge, i) => (
                      <div key={i} className="flex justify-between py-2 border-b border-slate-700">
                        <span className="text-slate-300">{charge.item}</span>
                        <span className="font-mono text-white">${charge.amount.toFixed(2)}</span>
                      </div>
                    ))}
                    <div className="flex justify-between pt-4 font-bold text-lg">
                      <span>Total Balance</span>
                      <span className="text-amber-400">$562.50</span>
                    </div>
                  </div>
                  <Button className="w-full mt-6 bg-emerald-500 hover:bg-emerald-600 text-white">
                    Pay Now
                  </Button>
                </CardContent>
              </Card>
            </TabsContent>

            <TabsContent value="room" className="mt-6">
              <div className="grid grid-cols-2 gap-4">
                <Card className="bg-slate-800/50 border-slate-700">
                  <CardContent className="pt-6">
                    <h3 className="font-semibold mb-4">Smart Lock</h3>
                    <div className="flex items-center gap-4">
                      <div className="w-16 h-16 rounded-full bg-emerald-500/20 flex items-center justify-center text-2xl">
                        🔓
                      </div>
                      <div>
                        <p className="text-white font-medium">Unlocked</p>
                        <p className="text-sm text-slate-400">Battery: 78%</p>
                      </div>
                    </div>
                    <div className="flex gap-2 mt-4">
                      <Button variant="outline" className="flex-1 border-slate-600">Lock</Button>
                      <Button variant="outline" className="flex-1 border-slate-600">Unlock</Button>
                    </div>
                  </CardContent>
                </Card>

                <Card className="bg-slate-800/50 border-slate-700">
                  <CardContent className="pt-6">
                    <h3 className="font-semibold mb-4">Climate</h3>
                    <div className="text-center">
                      <div className="text-5xl font-bold text-white">72°F</div>
                      <p className="text-slate-400 mt-2">Humidity: 45%</p>
                    </div>
                    <div className="flex gap-2 mt-4">
                      <Button variant="outline" className="flex-1 border-slate-600">-</Button>
                      <Button variant="outline" className="flex-1 border-slate-600">+</Button>
                    </div>
                  </CardContent>
                </Card>
              </div>
            </TabsContent>

            <TabsContent value="concierge" className="mt-6">
              <Card className="bg-slate-800/50 border-slate-700">
                <CardContent className="pt-6">
                  <h3 className="font-semibold mb-4">Send a Request</h3>
                  <textarea
                    className="w-full h-32 bg-slate-700 border-slate-600 rounded-lg p-3 text-white placeholder-slate-400"
                    placeholder="What can we help you with?"
                  />
                  <Button className="mt-4 w-full bg-amber-500 hover:bg-amber-600 text-slate-900 font-semibold">
                    Submit Request
                  </Button>
                </CardContent>
              </Card>
            </TabsContent>
          </Tabs>
        )}

        {!reservation && !isLoading && (
          <div className="text-center py-20 text-slate-500">
            <div className="text-6xl mb-4">🏨</div>
            <p className="text-lg">Enter your reservation code to access your portal</p>
          </div>
        )}
      </div>
    </div>
  );
}
