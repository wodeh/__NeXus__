'use client';

import { useState } from 'react';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';

// Billing & Folio Viewer — charges, payments, invoices, refunds

export default function BillingPage() {
  const [activeTab, setActiveTab] = useState('folios');

  return (
    <div className="min-h-screen bg-slate-950 text-white p-6">
      <div className="max-w-7xl mx-auto">
        <header className="mb-8">
          <h1 className="text-3xl font-bold">Billing & Folios</h1>
          <p className="text-slate-400 mt-1">Manage charges, payments, and invoices</p>
        </header>

        {/* Stats */}
        <div className="grid grid-cols-4 gap-4 mb-6">
          {[
            { label: 'Open Folios', value: '24', color: 'text-amber-400' },
            { label: 'Total Revenue (Today)', value: '$8,450', color: 'text-emerald-400' },
            { label: 'Outstanding Balance', value: '$12,300', color: 'text-rose-400' },
            { label: 'Pending Refunds', value: '3', color: 'text-blue-400' },
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
            <TabsTrigger value="folios">Folios</TabsTrigger>
            <TabsTrigger value="charges">Charges</TabsTrigger>
            <TabsTrigger value="payments">Payments</TabsTrigger>
            <TabsTrigger value="invoices">Invoices</TabsTrigger>
          </TabsList>

          <TabsContent value="folios" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Active Folios</CardTitle>
                <Button size="sm" className="bg-emerald-500 hover:bg-emerald-600">New Folio</Button>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[
                    { guest: 'John Smith', room: '201', balance: 450.00, status: 'open', nights: 3 },
                    { guest: 'Maria Garcia', room: '305', balance: 720.00, status: 'open', nights: 4 },
                    { guest: 'James Wilson', room: '112', balance: -50.00, status: 'credit', nights: 2 },
                    { guest: 'Lisa Chen', room: '401', balance: 0.00, status: 'closed', nights: 5 },
                  ].map((folio, i) => (
                    <div key={i} className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
                      <div className="flex items-center gap-4">
                        <div className="w-10 h-10 rounded-full bg-slate-600 flex items-center justify-center font-medium">
                          {folio.guest[0]}
                        </div>
                        <div>
                          <p className="font-medium text-white">{folio.guest}</p>
                          <p className="text-sm text-slate-400">Room {folio.room} • {folio.nights} nights</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-4">
                        <div className="text-right">
                          <p className={`font-mono font-medium ${folio.balance < 0 ? 'text-emerald-400' : folio.balance === 0 ? 'text-slate-400' : 'text-white'}`}>
                            {folio.balance < 0 ? `-$${Math.abs(folio.balance).toFixed(2)}` : `$${folio.balance.toFixed(2)}`}
                          </p>
                        </div>
                        <Badge variant={folio.status === 'open' ? 'default' : folio.status === 'credit' ? 'secondary' : 'outline'}>
                          {folio.status}
                        </Badge>
                        <Button size="sm" variant="outline" className="border-slate-600">View</Button>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="charges" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Recent Charges</CardTitle>
                <Button size="sm" className="bg-emerald-500 hover:bg-emerald-600">Add Charge</Button>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[
                    { item: 'Room Charge — Night 1', amount: 150.00, type: 'room', folio: 'John Smith', time: '2 hours ago' },
                    { item: 'Room Service', amount: 45.50, type: 'fb', folio: 'Maria Garcia', time: '3 hours ago' },
                    { item: 'Mini Bar', amount: 28.00, type: 'misc', folio: 'James Wilson', time: '4 hours ago' },
                    { item: 'Spa Treatment', amount: 120.00, type: 'amenity', folio: 'Lisa Chen', time: 'Yesterday' },
                  ].map((charge, i) => (
                    <div key={i} className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg">
                      <div>
                        <p className="font-medium text-white">{charge.item}</p>
                        <p className="text-sm text-slate-400">{charge.folio} • {charge.time}</p>
                      </div>
                      <div className="flex items-center gap-3">
                        <Badge variant="outline">{charge.type}</Badge>
                        <span className="font-mono text-white">${charge.amount.toFixed(2)}</span>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="payments" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Payment History</CardTitle>
                <Button size="sm" className="bg-emerald-500 hover:bg-emerald-600">Record Payment</Button>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[
                    { method: 'Credit Card', amount: 450.00, folio: 'John Smith', status: 'completed', time: '1 hour ago' },
                    { method: 'Cash', amount: 200.00, folio: 'Maria Garcia', status: 'completed', time: '2 hours ago' },
                    { method: 'Credit Card', amount: 150.00, folio: 'James Wilson', status: 'refunded', time: 'Yesterday' },
                    { method: 'Bank Transfer', amount: 750.00, folio: 'Lisa Chen', status: 'completed', time: '2 days ago' },
                  ].map((payment, i) => (
                    <div key={i} className="flex items-center justify-between p-3 bg-slate-700/50 rounded-lg">
                      <div>
                        <p className="font-medium text-white">{payment.method}</p>
                        <p className="text-sm text-slate-400">{payment.folio} • {payment.time}</p>
                      </div>
                      <div className="flex items-center gap-3">
                        <span className="font-mono text-white">${payment.amount.toFixed(2)}</span>
                        <Badge variant={payment.status === 'completed' ? 'default' : 'secondary'}>
                          {payment.status}
                        </Badge>
                      </div>
                    </div>
                  ))}
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="invoices" className="mt-6">
            <Card className="bg-slate-800/50 border-slate-700">
              <CardHeader className="flex flex-row items-center justify-between">
                <CardTitle>Invoices</CardTitle>
                <Button size="sm" className="bg-emerald-500 hover:bg-emerald-600">Generate Invoice</Button>
              </CardHeader>
              <CardContent>
                <div className="space-y-3">
                  {[
                    { number: 'INV-2026-001', customer: 'John Smith', amount: 450.00, status: 'paid', date: 'May 10, 2026' },
                    { number: 'INV-2026-002', customer: 'Maria Garcia', amount: 720.00, status: 'pending', date: 'May 10, 2026' },
                    { number: 'INV-2026-003', customer: 'Corporate LLC', amount: 2500.00, status: 'overdue', date: 'May 1, 2026' },
                  ].map((invoice) => (
                    <div key={invoice.number} className="flex items-center justify-between p-4 bg-slate-700/50 rounded-lg">
                      <div>
                        <p className="font-medium text-white">{invoice.number}</p>
                        <p className="text-sm text-slate-400">{invoice.customer} • {invoice.date}</p>
                      </div>
                      <div className="flex items-center gap-4">
                        <span className="font-mono text-white">${invoice.amount.toFixed(2)}</span>
                        <Badge variant={invoice.status === 'paid' ? 'default' : invoice.status === 'overdue' ? 'destructive' : 'secondary'}>
                          {invoice.status}
                        </Badge>
                        <Button size="sm" variant="outline" className="border-slate-600">Download</Button>
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
