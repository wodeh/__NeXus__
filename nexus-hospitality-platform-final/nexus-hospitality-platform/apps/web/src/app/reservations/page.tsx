"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useApi } from "@/hooks/useApi";
import { Sidebar } from "@/components/sidebar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { CalendarDays, Plus, Search, Filter } from "lucide-react";

interface Reservation {
  id: string;
  confirmation_number: string;
  guest_name: string;
  room_number: string;
  check_in_date: string;
  check_out_date: string;
  status: string;
  total_amount: number;
}

export default function ReservationsPage() {
  const { get } = useApi();
  const [statusFilter, setStatusFilter] = useState("all");
  const [search, setSearch] = useState("");

  const { data: reservations = [], isLoading } = useQuery<Reservation[]>({
    queryKey: ["reservations", statusFilter],
    queryFn: () => get("/reservations?limit=100"),
  });

  const filtered = reservations.filter((r) => {
    const matchStatus = statusFilter === "all" || r.status === statusFilter;
    const matchSearch =
      !search ||
      r.guest_name.toLowerCase().includes(search.toLowerCase()) ||
      r.confirmation_number.toLowerCase().includes(search.toLowerCase()) ||
      r.room_number.includes(search);
    return matchStatus && matchSearch;
  });

  const statusColor: Record<string, string> = {
    confirmed: "bg-blue-100 text-blue-700",
    checked_in: "bg-green-100 text-green-700",
    checked_out: "bg-gray-100 text-gray-700",
    cancelled: "bg-red-100 text-red-700",
    no_show: "bg-amber-100 text-amber-700",
  };

  return (
    <div className="flex min-h-screen bg-background">
      <Sidebar />
      <main className="flex-1 p-6">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Reservations</h1>
            <p className="text-sm text-muted-foreground">
              Manage bookings and guest stays
            </p>
          </div>
          <Button>
            <Plus className="mr-2 h-4 w-4" />
            New Reservation
          </Button>
        </div>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm font-medium">All Bookings</CardTitle>
            <div className="flex items-center gap-2">
              <div className="relative">
                <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                <input
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  placeholder="Search guest, room, confirmation..."
                  className="h-9 rounded-md border border-input bg-background pl-8 pr-3 text-sm outline-none ring-offset-background focus-visible:ring-2 focus-visible:ring-ring"
                />
              </div>
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="h-9 rounded-md border border-input bg-background px-2 text-sm outline-none"
              >
                <option value="all">All Status</option>
                <option value="confirmed">Confirmed</option>
                <option value="checked_in">Checked In</option>
                <option value="checked_out">Checked Out</option>
                <option value="cancelled">Cancelled</option>
              </select>
            </div>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="py-8 text-center text-muted-foreground">Loading reservations...</div>
            ) : filtered.length === 0 ? (
              <div className="py-8 text-center text-muted-foreground">No reservations found</div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="border-b text-left text-muted-foreground">
                      <th className="pb-2 pr-4 font-medium">Confirmation</th>
                      <th className="pb-2 pr-4 font-medium">Guest</th>
                      <th className="pb-2 pr-4 font-medium">Room</th>
                      <th className="pb-2 pr-4 font-medium">Check In</th>
                      <th className="pb-2 pr-4 font-medium">Check Out</th>
                      <th className="pb-2 pr-4 font-medium">Status</th>
                      <th className="pb-2 pr-4 font-medium text-right">Total</th>
                    </tr>
                  </thead>
                  <tbody>
                    {filtered.map((r) => (
                      <tr key={r.id} className="border-b last:border-0">
                        <td className="py-3 pr-4 font-mono text-xs text-muted-foreground">
                          {r.confirmation_number}
                        </td>
                        <td className="py-3 pr-4 font-medium">{r.guest_name}</td>
                        <td className="py-3 pr-4">{r.room_number}</td>
                        <td className="py-3 pr-4">{r.check_in_date}</td>
                        <td className="py-3 pr-4">{r.check_out_date}</td>
                        <td className="py-3 pr-4">
                          <Badge className={statusColor[r.status] || "bg-gray-100 text-gray-700"}>
                            {r.status.replace("_", " ")}
                          </Badge>
                        </td>
                        <td className="py-3 pr-4 text-right font-medium">
                          ${r.total_amount?.toLocaleString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
