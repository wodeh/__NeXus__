'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Badge } from '@/components/ui/badge';

// Admin Settings — tenant config, users, properties, integrations

export default function AdminPage() {
  const [activeTab, setActiveTab] = useState('general');

  return (
    <div className="min-h-screen bg-slate-950 text-white p-6">
      <div className="max-w-6xl mx-auto">
        <header className="mb-8">
          <h1 className="text-3xl font-bold">Admin Settings</h1>
          <p className="text-slate-400 mt-1">Configure your property management system</p>
        </header>

        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-5 bg-slate-800">
            <TabsTrigger value="general">General</TabsTrigger>
            <TabsTrigger value="users">Users</TabsTrigger>
            <TabsTrigger value="properties">Properties</TabsTrigger>
            <TabsTrigger value="billing">Billing</TabsTrigger>
            <TabsTrigger value="integrations">Integrations</TabsTrigger>
          </TabsList>

          <TabsContent value="general" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader>
                <CardTitle>Tenant Configuration</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-sm text-slate-400">Property Name</label>
                    <Input defaultValue="Grand Hotel" className="bg-slate-700 border-slate-600 text-white mt-1" />
                  </div>
                  <div>
                    <label className="text-sm text-slate-400">Timezone</label>
                    <Input defaultValue="America/New_York" className="bg-slate-700 border-slate-600 text-white mt-1" />
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="text-sm text-slate-400">Currency</label>
                    <Input defaultValue="USD" className="bg-slate-700 border-slate-600 text-white mt-1" />
                  </div>
                  <div>
                    <label className="text-sm text-slate-400">Check-in Time</label>
                    <Input defaultValue="15:00" className="bg-slate-700 border-slate-600 text-white mt-1" />
                  </div>
                </div>
                <div>
                  <label className="text-sm text-slate-400">Address</label>
                  <Input defaultValue="123 Main St, New York, NY 10001" className="bg-slate-700 border-slate-600 text-white mt-1" />
                </div>
                <Button className="bg-amber-500 hover:bg-amber-600 text-slate-900 font-semibold">Save Changes</Button>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="users" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>User Management</CardTitle>
                <Button size="sm" className="bg-emerald-500 hover:bg-emerald-600">Add User</Button>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[
                    { name: 'Alice Johnson', email: 'alice@hotel.com', role: 'Admin', status: 'active' },
                    { name: 'Bob Smith', email: 'bob@hotel.com', role: 'Manager', status: 'active' },
                    { name: 'Carol White', email: 'carol@hotel.com', role: 'Housekeeper', status: 'active' },
                    { name: 'Dave Brown', email: 'dave@hotel.com', role: 'Receptionist', status: 'inactive' },
                  ].map((user) => (
                    <div key={user.email} className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg">
                      <div>
                        <p className="font-medium text-white">{user.name}</p>
                        <p className="text-sm text-slate-400">{user.email}</p>
                      </div>
                      <div className="flex items-center gap-3">
                        <Badge variant={user.role === 'Admin' ? 'default' : 'secondary'}>{user.role}</Badge>
                        <Badge variant={user.status === 'active' ? 'outline' : 'destructive'}>{user.status}</Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="properties" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Properties</CardTitle>
                <Button size="sm" className="bg-emerald-500 hover:bg-emerald-600">Add Property</Button>
              </CardHeader>
              <CardContent>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {[
                    { name: 'Grand Hotel Downtown', rooms: 120, city: 'New York', status: 'active' },
                    { name: 'Seaside Resort', rooms: 85, city: 'Miami', status: 'active' },
                    { name: 'Mountain Lodge', rooms: 45, city: 'Denver', status: 'maintenance' },
                  ].map((prop) => (
                    <Card key={prop.name} className="bg-slate-700/50 border-slate-600">
                      <CardContent className="pt-4">
                        <div className="flex justify-between items-start">
                          <div>
                            <h3 className="font-semibold text-white">{prop.name}</h3>
                            <p className="text-sm text-slate-400">{prop.city} • {prop.rooms} rooms</p>
                          </div>
                          <Badge variant={prop.status === 'active' ? 'outline' : 'destructive'}>{prop.status}</Badge>
                        </div>
                      </CardContent>
                    </Card>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="billing" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader>
                <CardTitle>Subscription & Billing</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="p-4 bg-slate-700/50 rounded-lg">
                  <div className="flex justify-between items-center">
                    <div>
                      <p className="font-semibold text-white">Growth Plan</p>
                      <p className="text-sm text-slate-400">$299/month • Renews May 15, 2026</p>
                    </div>
                    <Button variant="outline" className="border-amber-500 text-amber-400">Upgrade</Button>
                  </div>
                </div>
                <div className="space-y-2">
                  <h4 className="font-medium text-white">Usage</h4>
                  <div className="space-y-2">
                    <div className="flex justify-between text-sm">
                      <span className="text-slate-400">Rooms</span>
                      <span className="text-white">120 / 200</span>
                    </div>
                    <div className="w-full bg-slate-700 rounded-full h-2">
                      <div className="bg-amber-500 h-2 rounded-full" style={{ width: '60%' }} />
                    </div>
                  </div>
                  <div className="space-y-2">
                    <div className="flex justify-between text-sm">
                      <span className="text-slate-400">Users</span>
                      <span className="text-white">8 / 15</span>
                    </div>
                    <div className="w-full bg-slate-700 rounded-full h-2">
                      <div className="bg-emerald-500 h-2 rounded-full" style={{ width: '53%' }} />
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="integrations" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader>
                <CardTitle>Connected Services</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[
                    { name: 'Stripe', status: 'connected', desc: 'Payment processing' },
                    { name: 'Booking.com', status: 'connected', desc: 'Channel manager' },
                    { name: 'Expedia', status: 'connected', desc: 'Channel manager' },
                    { name: 'SendGrid', status: 'disconnected', desc: 'Email notifications' },
                    { name: 'Twilio', status: 'connected', desc: 'SMS notifications' },
                  ].map((integration) => (
                    <div key={integration.name} className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg">
                      <div>
                        <p className="font-medium text-white">{integration.name}</p>
                        <p className="text-sm text-slate-400">{integration.desc}</p>
                      </div>
                      <Badge variant={integration.status === 'connected' ? 'default' : 'destructive'}>
                        {integration.status}
                      </Badge>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>
    </div>
  );
}
