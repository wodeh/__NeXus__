"use client";

import { useState, useEffect, useCallback } from "react";
import {
  Calendar, ChevronLeft, ChevronRight, RefreshCw, AlertTriangle,
  BedDouble, DoorOpen, DoorClosed, Wrench, Ban, Sparkles,
  ArrowRight, Users,
} from "lucide-react";
import { getRoomStatusView, RoomStatusView, RoomDailyStatus } from "@/lib/api";

export default function RoomStatusPage() {
  const [view, setView] = useState<"day" | "week" | "month">("day");
  const [date, setDate] = useState(new Date().toISOString().split("T")[0]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [data, setData] = useState<RoomStatusDay[]>([]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await getRoomStatusView(date, view);
      setData(result.dates);
    } catch (e: any) {
      setError(e.message || "Failed to load room status");
    } finally {
      setLoading(false);
    }
  }, [date, view]);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const navigateDate = (direction: number) => {
    const d = new Date(date);
    if (view === "day") d.setDate(d.getDate() + direction);
    else if (view === "week") d.setDate(d.getDate() + direction * 7);
    else if (view === "month") d.setMonth(d.getMonth() + direction);
    setDate(d.toISOString().split("T")[0]);
  };

  if (loading) {
    return (
      <div className="flex h-96 items-center justify-center">
        <RefreshCw className="h-8 w-8 animate-spin text-nexus-400" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex h-96 flex-col items-center justify-center gap-4">
        <AlertTriangle className="h-10 w-10 text-rose-400" />
        <p className="text-rose-400">{error}</p>
        <button onClick={fetchData} className="btn-primary">Retry</button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Room Status</h1>
          <p className="text-sm text-slate-400">Live room availability · Occupancy tracking · Housekeeping status</p>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex items-center gap-1 bg-slate-800 rounded-lg p-1">
            {(["day", "week", "month"] as const).map((v) => (
              <button
                key={v}
                onClick={() => setView(v)}
                className={`px-3 py-1.5 text-xs font-medium rounded-md transition-colors ${
                  view === v ? "bg-nexus-500 text-white" : "text-slate-400 hover:text-white"
                }`}
              >
                {v.charAt(0).toUpperCase() + v.slice(1)}
              </button>
            ))}
          </div>
          <button onClick={() => navigateDate(-1)} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white">
            <ChevronLeft className="h-4 w-4" />
          </button>
          <div className="flex items-center gap-2 bg-slate-800 rounded-lg px-3 py-2">
            <Calendar className="h-4 w-4 text-slate-400" />
            <span className="text-sm text-white">{date}</span>
          </div>
          <button onClick={() => navigateDate(1)} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white">
            <ChevronRight className="h-4 w-4" />
          </button>
          <button onClick={fetchData} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white">
            <RefreshCw className="h-4 w-4" />
          </button>
        </div>
      </div>

      {/* Summary Cards */}
      {data[0] && (
        <div className="grid grid-cols-6 gap-4">
          <SummaryCard label="Occupancy" value={`${Math.round(data[0].occupancy)}%`} icon={BedDouble} color="nexus" />
          <SummaryCard label="Arrivals" value={data[0].arrivals.toString()} icon={DoorOpen} color="emerald" />
          <SummaryCard label="Departures" value={data[0].departures.toString()} icon={DoorClosed} color="amber" />
          <SummaryCard label="Stayovers" value={data[0].stayovers.toString()} icon={Users} color="violet" />
          <SummaryCard label="Revenue" value={`$${Math.round(data[0].revenue).toLocaleString()}`} icon={Sparkles} color="sky" />
          <SummaryCard label="Rooms" value={data[0].rooms.length.toString()} icon={BedDouble} color="slate" />
        </div>
      )}

      {/* Room Grid */}
      <div className="card space-y-4">
        <div className="flex items-center gap-4 text-xs">
          <div className="flex items-center gap-1"><div className="h-3 w-3 rounded bg-emerald-500" /> Occupied</div>
          <div className="flex items-center gap-1"><div className="h-3 w-3 rounded bg-sky-500" /> Vacant Clean</div>
          <div className="flex items-center gap-1"><div className="h-3 w-3 rounded bg-amber-500" /> Vacant Dirty</div>
          <div className="flex items-center gap-1"><div className="h-3 w-3 rounded bg-rose-500" /> Maintenance</div>
          <div className="flex items-center gap-1"><div className="h-3 w-3 rounded bg-slate-500" /> Blocked</div>
        </div>

        {view === "day" && data[0] && (
          <div className="grid grid-cols-5 gap-3">
            {data[0].rooms.map((room) => (
              <RoomCard key={room.room_number} room={room} />
            ))}
          </div>
        )}

        {view === "week" && (
          <div className="space-y-4">
            {data.map((day) => (
              <div key={day.date} className="space-y-2">
                <h4 className="text-sm font-medium text-white">{day.date}</h4>
                <div className="grid grid-cols-5 gap-3">
                  {day.rooms.slice(0, 10).map((room) => (
                    <RoomCard key={`${day.date}-${room.room_number}`} room={room} compact />
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}

        {view === "month" && (
          <div className="space-y-4">
            {data.slice(0, 7).map((day) => (
              <div key={day.date} className="space-y-2">
                <h4 className="text-sm font-medium text-white">{day.date}</h4>
                <div className="grid grid-cols-5 gap-3">
                  {day.rooms.slice(0, 5).map((room) => (
                    <RoomCard key={`${day.date}-${room.room_number}`} room={room} compact />
                  ))}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

function RoomCard({ room, compact }: { room: RoomDailyStatus; compact?: boolean }) {
  const statusColors: Record<string, string> = {
    occupied: "bg-emerald-500/20 border-emerald-500/50 text-emerald-400",
    vacant_clean: "bg-sky-500/20 border-sky-500/50 text-sky-400",
    vacant_dirty: "bg-amber-500/20 border-amber-500/50 text-amber-400",
    maintenance: "bg-rose-500/20 border-rose-500/50 text-rose-400",
    blocked: "bg-slate-500/20 border-slate-500/50 text-slate-400",
  };

  if (compact) {
    return (
      <div className={`rounded-lg border p-2 ${statusColors[room.status] || statusColors.blocked}`}>
        <p className="text-xs font-bold">{room.room_number}</p>
        <p className="text-[10px]">{room.room_type}</p>
      </div>
    );
  }

  return (
    <div className={`rounded-lg border p-4 space-y-2 ${statusColors[room.status] || statusColors.blocked}`}>
      <div className="flex items-center justify-between">
        <span className="text-lg font-bold">{room.room_number}</span>
        {room.check_in && <ArrowRight className="h-4 w-4" />}
        {room.check_out && <DoorOpen className="h-4 w-4" />}
      </div>
      <p className="text-sm">{room.room_type}</p>
      {room.guest_name && <p className="text-xs opacity-75">{room.guest_name}</p>}
      {room.rate && <p className="text-xs font-medium">${room.rate}/night</p>}
      <p className="text-[10px] uppercase tracking-wider opacity-60">{room.status.replace("_", " ")}</p>
    </div>
  );
}

function SummaryCard({ label, value, icon: Icon, color }: { label: string; value: string; icon: React.ElementType; color: string }) {
  const colorMap: Record<string, string> = {
    nexus: "text-nexus-400 bg-nexus-500/10",
    emerald: "text-emerald-400 bg-emerald-500/10",
    amber: "text-amber-400 bg-amber-500/10",
    violet: "text-violet-400 bg-violet-500/10",
    sky: "text-sky-400 bg-sky-500/10",
    slate: "text-slate-400 bg-slate-500/10",
  };
  return (
    <div className="card space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-xs text-slate-400">{label}</span>
        <div className={`rounded-lg p-1.5 ${colorMap[color]}`}><Icon className="h-3 w-3" /></div>
      </div>
      <p className="text-xl font-bold text-white">{value}</p>
    </div>
  );
}
