"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useApi } from "@/hooks/useApi";
import { Sidebar } from "@/components/sidebar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { BedDouble, CheckCircle2, AlertCircle, Wrench } from "lucide-react";

interface Room {
  id: string;
  room_number: string;
  room_type_name: string;
  floor: string;
  status: string;
  is_ooo: boolean;
}

export default function RoomsPage() {
  const { get } = useApi();
  const [floorFilter, setFloorFilter] = useState("all");

  const { data: rooms = [], isLoading } = useQuery<Room[]>({
    queryKey: ["rooms"],
    queryFn: () => get("/rooms?limit=500"),
  });

  const floors = Array.from(new Set(rooms.map((r) => r.floor))).sort();
  const filtered = floorFilter === "all" ? rooms : rooms.filter((r) => r.floor === floorFilter);

  const statusConfig: Record<string, { color: string; icon: any }> = {
    available: { color: "bg-green-100 text-green-700 border-green-200", icon: CheckCircle2 },
    occupied: { color: "bg-blue-100 text-blue-700 border-blue-200", icon: BedDouble },
    cleaning: { color: "bg-amber-100 text-amber-700 border-amber-200", icon: AlertCircle },
    maintenance: { color: "bg-red-100 text-red-700 border-red-200", icon: Wrench },
    inspected: { color: "bg-purple-100 text-purple-700 border-purple-200", icon: CheckCircle2 },
  };

  const counts = {
    available: rooms.filter((r) => r.status === "available" && !r.is_ooo).length,
    occupied: rooms.filter((r) => r.status === "occupied").length,
    cleaning: rooms.filter((r) => r.status === "cleaning").length,
    maintenance: rooms.filter((r) => r.is_ooo || r.status === "maintenance").length,
    inspected: rooms.filter((r) => r.status === "inspected").length,
  };

  return (
    <div className="flex min-h-screen bg-background">
      <Sidebar />
      <main className="flex-1 p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-bold">Room Status Board</h1>
          <p className="text-sm text-muted-foreground">Real-time room occupancy and status</p>
        </div>

        <div className="mb-4 flex flex-wrap gap-2">
          {Object.entries(counts).map(([status, count]) => (
            <Badge
              key={status}
              variant="outline"
              className={`cursor-pointer px-3 py-1 capitalize ${statusConfig[status]?.color}`}
            >
              {status}: {count}
            </Badge>
          ))}
          <div className="ml-auto flex items-center gap-2">
            <label className="text-sm text-muted-foreground">Floor:</label>
            <select
              value={floorFilter}
              onChange={(e) => setFloorFilter(e.target.value)}
              className="h-8 rounded-md border border-input bg-background px-2 text-sm outline-none"
            >
              <option value="all">All</option>
              {floors.map((f) => (
                <option key={f} value={f}>Floor {f}</option>
              ))}
            </select>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3 md:grid-cols-4 lg:grid-cols-6">
          {isLoading ? (
            <div className="col-span-full py-8 text-center text-muted-foreground">Loading rooms...</div>
          ) : (
            filtered.map((room) => {
              const config = statusConfig[room.status] || statusConfig.available;
              const Icon = config.icon;
              return (
                <Card
                  key={room.id}
                  className={`border-2 transition-shadow hover:shadow-md ${config.color}`}
                >
                  <CardContent className="p-3">
                    <div className="flex items-center justify-between">
                      <span className="text-lg font-bold">{room.room_number}</span>
                      <Icon className="h-4 w-4" />
                    </div>
                    <p className="mt-1 text-xs opacity-80">{room.room_type_name}</p>
                    <p className="text-xs opacity-60">Floor {room.floor}</p>
                  </CardContent>
                </Card>
              );
            })
          )}
        </div>
      </main>
    </div>
  );
}
