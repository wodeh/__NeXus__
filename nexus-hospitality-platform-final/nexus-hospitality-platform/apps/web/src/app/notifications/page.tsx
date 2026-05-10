'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

// Notification Center — email/SMS logs, guest messages, alerts

export default function NotificationsPage() {
  const [activeTab, setActiveTab] = useState('messages');

  return (
    <div className="min-h-screen bg-slate-950 text-white p-6">
      <div className="max-w-7xl mx-auto">
        <header className="mb-8">
          <h1 className="text-3xl font-bold">Notification Center</h1>
          <p className="text-slate-400 mt-1">Email, SMS, and guest messaging logs</p>
        </header>

        {/* Stats */}
        <div className="grid grid-cols-4 gap-4 mb-6">
          {[
            { label: 'Unread Messages', value: '8', color: 'text-amber-400' },
            { label: 'Sent Today', value: '124', color: 'text-emerald-400' },
            { label: 'Failed', value: '3', color: 'text-rose-400' },
            { label: 'Queued', value: '12', color: 'text-blue-400' },
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
          <TabsList className="grid w-full grid-cols-4 bg-slate-800">
            <TabsTrigger value="messages">Guest Messages</TabsTrigger>
            <TabsTrigger value="email">Email Log</TabsTrigger>
            <TabsTrigger value="sms">SMS Log</TabsTrigger>
            <TabsTrigger value="templates">Templates</TabsTrigger>
          </TabsList>

          <TabsContent value="messages" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Guest Conversations</CardTitle>
                <Button size="sm" className="bg-emerald-500 hover:bg-emerald-600">New Message</Button>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[
                    { guest: 'John Smith', room: '201', message: 'Can I get a late checkout?', time: '10 min ago', unread: true },
                    { guest: 'Maria Garcia', room: '305', message: 'Room service order #45 status?', time: '25 min ago', unread: true },
                    { guest: 'James Wilson', room: '112', message: 'WiFi password not working', time: '1 hour ago', unread: false },
                    { guest: 'Lisa Chen', room: '401', message: 'Thank you for the upgrade!', time: '2 hours ago', unread: false },
                  ].map((msg, i) => (
                    <div key={i} className={`flex items-start gap-4 p-4 rounded-lg ${msg.unread ? 'bg-amber-500/10 border border-amber-500/20' : 'bg-slate-700/50'}`}>
                      <div className="w-10 h-10 rounded-full bg-slate-600 flex items-center justify-center font-medium">
                        {msg.guest[0]}
                      </div>
                      <div className="flex-1">
                        <div className="flex items-center gap-2">
                          <p className="font-medium text-white">{msg.guest}</p>
                          <Badge variant="outline">Room {msg.room}</Badge>
                          {msg.unread && <Badge className="bg-amber-500/20 text-amber-400">New</Badge>}
                        </div>
                        <p className="text-white mt-1">{msg.message}</p>
                        <p className="text-sm text-slate-400 mt-1">{msg.time}</p>
                      </div>
                      <Button size="sm" variant="outline" className="border-slate-600">Reply</Button>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="email" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader>
                <CardTitle>Email Delivery Log</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[
                    { recipient: 'john.smith@email.com', subject: 'Booking Confirmation — Room 201', status: 'delivered', time: '5 min ago', type: 'confirmation' },
                    { recipient: 'maria.garcia@email.com', subject: 'Check-in Reminder', status: 'delivered', time: '15 min ago', type: 'reminder' },
                    { recipient: 'james.wilson@email.com', subject: 'Invoice #2026-001', status: 'bounced', time: '30 min ago', type: 'invoice' },
                    { recipient: 'lisa.chen@email.com', subject: 'Thank you for your stay', status: 'delivered', time: '1 hour ago', type: 'post_stay' },
                  ].map((email, i) => (
                    <div key={i} className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg">
                      <div>
                        <p className="font-medium text-white">{email.subject}</p>
                        <p className="text-sm text-slate-400">{email.recipient} • {email.time}</p>
                      </div>
                      <div className="flex items-center gap-3">
                        <Badge variant="outline">{email.type}</Badge>
                        <Badge variant={email.status === 'delivered' ? 'default' : 'destructive'}>
                          {email.status}
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="sms" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader>
                <CardTitle>SMS Delivery Log</CardTitle>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[
                    { phone: '+1-555-0123', message: 'Your room 201 is ready. Check-in anytime after 3 PM.', status: 'delivered', time: '10 min ago' },
                    { phone: '+1-555-0456', message: 'Room service order #45 is on its way!', status: 'delivered', time: '20 min ago' },
                    { phone: '+1-555-0789', message: 'Checkout reminder: Departure is at 11 AM tomorrow.', status: 'failed', time: '45 min ago' },
                  ].map((sms, i) => (
                    <div key={i} className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg">
                      <div className="flex-1">
                        <p className="text-white">{sms.message}</p>
                        <p className="text-sm text-slate-400">{sms.phone} • {sms.time}</p>
                      </div>
                      <Badge variant={sms.status === 'delivered' ? 'default' : 'destructive'}>
                        {sms.status}
                      </Badge>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="templates" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Message Templates</CardTitle>
                <Button size="sm" className="bg-emerald-500 hover:bg-emerald-600">New Template</Button>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[
                    { name: 'Check-in Welcome', subject: 'Welcome to {property_name}', channel: 'email', usage: 145 },
                    { name: 'Booking Confirmation', subject: 'Booking Confirmed — {reservation_code}', channel: 'email', usage: 89 },
                    { name: 'Checkout Reminder', subject: 'Checkout Tomorrow at {checkout_time}', channel: 'sms', usage: 67 },
                    { name: 'Late Checkout Offer', subject: 'Extend Your Stay?', channel: 'email', usage: 34 },
                  ].map((template, i) => (
                    <div key={i} className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg">
                      <div>
                        <p className="font-medium text-white">{template.name}</p>
                        <p className="text-sm text-slate-400">{template.subject} • Used {template.usage} times</p>
                      </div>
                      <div className="flex items-center gap-3">
                        <Badge variant="outline">{template.channel}</Badge>
                        <Button size="sm" variant="outline" className="border-slate-600">Edit</Button>
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
