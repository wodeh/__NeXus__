"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useApi } from "@/hooks/useApi";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import {
  Hotel,
  Users,
  BedDouble,
  Wrench,
  TrendingUp,
  Calendar,
  ArrowUpRight,
  ArrowDownRight,
} from "lucide-react";

interface DashboardStats {
  totalProperties: number;
  totalRooms: number;
  occupiedRooms: number;
  availableRooms: number;
  todayCheckIns: number;
  todayCheckOuts: number;
  pendingHousekeeping: number;
  openWorkOrders: number;
  revenueToday: number;
  occupancyRate: number;
}

export default function DashboardPage() {
  const { get } = useApi();
  const [activeTab, setActiveTab] = useState("overview");

  const { data: stats, isLoading } = useQuery<DashboardStats>({
    queryKey: ["dashboard-stats"],
    queryFn: () => get("/dashboard/stats"),
  });

  const statCards = [
    {
      title: "Total Properties",
      value: stats?.totalProperties || 0,
      icon: Hotel,
      trend: "+12%",
      trendUp: true,
    },
    {
      title: "Occupancy Rate",
      value: `${stats?.occupancyRate || 0}%`,
      icon: BedDouble,
      trend: "+5%",
      trendUp: true,
    },
    {
      title: "Today's Revenue",
      value: `$${(stats?.revenueToday || 0).toLocaleString()}`,
      icon: TrendingUp,
      trend: "+18%",
      trendUp: true,
    },
    {
      title: "Pending Tasks",
      value: (stats?.pendingHousekeeping || 0) + (stats?.openWorkOrders || 0),
      icon: Wrench,
      trend: "-3%",
      trendUp: false,
    },
  ];

  return (
    <div className="min-h-screen bg-background">
      {/* Header */}
      <header className="border-b bg-card px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-nexus-600">
              <Hotel className="h-5 w-5 text-white" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-foreground">Nexus PMS</h1>
              <p className="text-sm text-muted-foreground">Guru Hospitality Platform</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <Button variant="outline" size="sm">
              <Calendar className="mr-2 h-4 w-4" />
              {new Date().toLocaleDateString()}
            </Button>
            <div className="h-8 w-8 rounded-full bg-nexus-100 flex items-center justify-center">
              <Users className="h-4 w-4 text-nexus-700" />
            </div>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="p-6">
        <Tabs value={activeTab} onValueChange={setActiveTab} className="space-y-6">
          <TabsList className="bg-muted">
            <TabsTrigger value="overview">Overview</TabsTrigger>
            <TabsTrigger value="reservations">Reservations</TabsTrigger>
            <TabsTrigger value="rooms">Rooms</TabsTrigger>
            <TabsTrigger value="housekeeping">Housekeeping</TabsTrigger>
            <TabsTrigger value="rates">Rates</TabsTrigger>
          </TabsList>

          <TabsContent value="overview" className="space-y-6">
            {/* Stats Grid */}
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4">
              {statCards.map((stat) => (
                <Card key={stat.title}>
                  <CardHeader className="flex flex-row items-center justify-between pb-2">
                    <CardTitle className="text-sm font-medium text-muted-foreground">
                      {stat.title}
                    </CardTitle>
                    <stat.icon className="h-4 w-4 text-muted-foreground" />
                  </CardHeader>
                  <CardContent>
                    <div className="text-2xl font-bold">{isLoading ? "—" : stat.value}</div>
                    <div className="flex items-center text-xs">
                      {stat.trendUp ? (
                        <ArrowUpRight className="mr-1 h-3 w-3 text-green-500" />
                      ) : (
                        <ArrowDownRight className="mr-1 h-3 w-3 text-red-500" />
                      )}
                      <span className={stat.trendUp ? "text-green-500" : "text-red-500"}>
                        {stat.trend}
                      </span>
                      <span className="ml-1 text-muted-foreground">vs last month</span>
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>

            {/* Activity & Quick Actions */}
            <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
              <Card className="lg:col-span-2">
                <CardHeader>
                  <CardTitle>Recent Activity</CardTitle>
                </CardHeader>
                <CardContent>
                  <div className="space-y-3">
                    {[
                      { action: "Check-in", guest: "John Smith", room: "101", time: "10:30 AM" },
                      { action: "New Reservation", guest: "Sarah Lee", room: "205", time: "09:15 AM" },
                      { action: "Housekeeping Complete", guest: "—", room: "304", time: "08:45 AM" },
                      { action: "Payment Received", guest: "Mike Brown", room: "112", time: "08:00 AM" },
                    ].map((item, i) => (
                      <div
                        key={i}
                        className="flex items-center justify-between rounded-lg border p-3"
                      >
                        <div>
                          <p className="font-medium">{item.action}</p>
                          <p className="text-sm text-muted-foreground">
                            {item.guest} {item.room && `· Room ${item.room}`}
                          </p>
                        </div>
                        <span className="text-sm text-muted-foreground">{item.time}</span>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Quick Actions</CardTitle>
                </CardHeader>
                <CardContent className="space-y-2">
                  <Button className="w-full justify-start" variant="outline">
                    <BedDouble className="mr-2 h-4 w-4" />
                    New Reservation
                  </Button>
                  <Button className="w-full justify-start" variant="outline">
                    <Users className="mr-2 h-4 w-4" />
                    Check-in Guest
                  </Button>
                  <Button className="w-full justify-start" variant="outline">
                    <Wrench className="mr-2 h-4 w-4" />
                    Create Work Order
                  </Button>
                  <Button className="w-full justify-start" variant="outline">
                    <TrendingUp className="mr-2 h-4 w-4" />
                    Adjust Rates
                  </Button>
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          <TabsContent value="reservations">
            <Card>
              <CardHeader>
                <CardTitle>Reservations</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">Reservation grid coming in next build...</p>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="rooms">
            <Card>
              <CardHeader>
                <CardTitle>Room Status Board</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">Room status board coming in next build...</p>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="housekeeping">
            <Card>
              <CardHeader>
                <CardTitle>Housekeeping Board</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">Housekeeping board coming in next build...</p>
              </CardContent>
            </Card>
          </TabsContent>

          <TabsContent value="rates">
            <Card>
              <CardHeader>
                <CardTitle>Rate Calendar</CardTitle>
              </CardHeader>
              <CardContent>
                <p className="text-muted-foreground">Rate calendar coming in next build...</p>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </main>
    </div>
  );
}
