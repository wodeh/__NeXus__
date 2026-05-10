'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

// Channel Manager Dashboard — OTA connections, pending reservations, sync status

export default function ChannelsPage() {
  const [activeTab, setActiveTab] = useState('pending');

  return (
    <div className="min-h-screen bg-slate-950 text-white p-6">
      <div className="max-w-7xl mx-auto">
        <header className="mb-8">
          <h1 className="text-3xl font-bold">Channel Manager</h1>
          <p className="text-slate-400 mt-1">Manage OTA connections and incoming reservations</p>
        </header>

        {/* Stats Row */}
        <div className="grid grid-cols-4 gap-4 mb-6">
          {[
            { label: 'Pending', value: '12', color: 'text-amber-400' },
            { label: 'Accepted Today', value: '8', color: 'text-emerald-400' },
            { label: 'Rejected', value: '2', color: 'text-rose-400' },
            { label: 'Commission (MTD)', value: '$4,250', color: 'text-blue-400' },
          ].map((stat) => (
            <Card key={stat.label} className="bg-slate-800/50 border-slate-700">
              <CardContent className="pt-6">
                <p className="text-sm text-slate-400">{stat.label}</p>
                <p className={`text-3xl font-bold ${stat.color}`}>{stat.value}</p>
              </CardContent>
            </Card>
          ))}
        </div>

        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-3 bg-slate-800">
            <TabsTrigger value="pending">Pending Reservations</TabsTrigger>
            <TabsTrigger value="channels">Connected Channels</TabsTrigger>
            <TabsTrigger value="history">History</TabsTrigger>
          </TabsList>

          <TabsContent value="pending" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader>
                <CardTitle>Incoming Reservations</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[
                    { guest: 'John Smith', channel: 'Booking.com', dates: 'May 12-15', amount: 450, status: 'pending', rooms: 1 },
                    { guest: 'Maria Garcia', channel: 'Expedia', dates: 'May 14-18', amount: 720, status: 'pending', rooms: 2 },
                    { guest: 'James Wilson', channel: 'Airbnb', dates: 'May 16-20', amount: 580, status: 'pending', rooms: 1 },
                    { guest: 'Lisa Chen', channel: 'Booking.com', dates: 'May 18-22', amount: 900, status: 'pending', rooms: 2 },
                  ].map((res, i) => (
                    <div key={i} className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
                      <div className="flex items-center gap-4">
                        <div className="w-12 h-12 rounded-full bg-slate-600 flex items-center justify-center text-xl">
                          {res.channel === 'Booking.com' ? 'B' : res.channel === 'Expedia' ? 'E' : 'A'}
                        </div>
                        <div>
                          <p className="font-medium text-white">{res.guest}</p>
                          <p className="text-sm text-slate-400">{res.dates} • {res.rooms} room{res.rooms > 1 ? 's' : ''}</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-4">
                        <div className="text-right">
                          <p className="font-mono text-white">${res.amount}</p>
                          <p className="text-xs text-slate-400">{res.channel}</p>
                        </div>
                        <div className="flex gap-2">
                          <Button size="sm" className="bg-emerald-500 hover:bg-emerald-600">Accept</Button>
                          <Button size="sm" variant="outline" className="border-rose-500 text-rose-400">Decline</Button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="channels" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Connected Channels</CardTitle>
                <Button size="sm" className="bg-emerald-500 hover:bg-emerald-600">Add Channel</Button>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[
                    { name: 'Booking.com', hotelId: '12345678', status: 'active', lastSync: '2 min ago', commission: '15%' },
                    { name: 'Expedia', hotelId: 'EXP-98765', status: 'active', lastSync: '5 min ago', commission: '18%' },
                    { name: 'Airbnb', hotelId: 'AB-12345', status: 'warning', lastSync: '1 hour ago', commission: '3%' },
                    { name: 'Agoda', hotelId: 'AGD-54321', status: 'active', lastSync: '10 min ago', commission: '15%' },
                  ].map((channel) => (
                    <div key={channel.name} className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
                      <div className="flex items-center gap-4">
                        <div className="w-10 h-10 rounded-lg bg-blue-500/20 flex items-center justify-center text-blue-400 font-bold">
                          {channel.name[0]}
                        </div>
                        <div>
                          <p className="font-medium text-white">{channel.name}</p>
                          <p className="text-sm text-slate-400">Hotel ID: {channel.hotelId}</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-6">
                        <div className="text-right">
                          <p className="text-sm text-slate-400">Commission</p>
                          <p className="font-medium text-white">{channel.commission}</p>
                        </div>
                        <div className="text-right">
                          <p className="text-sm text-slate-400">Last Sync</p>
                          <p className="text-sm text-white">{channel.lastSync}</p>
                        </div>
                        <Badge variant={channel.status === 'active' ? 'default' : 'secondary'}>
                          {channel.status}
                        </Badge>
                        <Button size="sm" variant="outline" className="border-slate-600">Sync Now</Button>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="history" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader>
                <CardTitle>Reservation History</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[
                    { guest: 'Robert Taylor', channel: 'Booking.com', action: 'Accepted', time: '2 hours ago', amount: 340 },
                    { guest: 'Sarah Lee', channel: 'Expedia', action: 'Accepted', time: '3 hours ago', amount: 520 },
                    { guest: 'Michael Brown', channel: 'Airbnb', action: 'Declined', time: '5 hours ago', amount: 280 },
                    { guest: 'Emily Davis', channel: 'Booking.com', action: 'Accepted', time: 'Yesterday', amount: 750 },
                  ].map((entry, i) => (
                    <div key={i} className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg">
                      <div>
                        <p className="font-medium text-white">{entry.guest}</p>
                        <p className="text-sm text-slate-400">{entry.channel} • {entry.time}</p>
                      </div>
                      <div className="flex items-center gap-4">
                        <span className="font-mono text-white">${entry.amount}</span>
                        <Badge variant={entry.action === 'Accepted' ? 'default' : 'destructive'}>
                          {entry.action}
                        </Badge>
                      </div>
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
