"use client";

import { useState, useMemo, useCallback, useRef, useEffect } from "react";
import Link from "next/link";
import {
  ChevronLeft,
  ChevronRight,
  CalendarDays,
  Plus,
  BedDouble,
  ArrowRightLeft,
  User,
  X,
  Search,
  Filter,
  Grid3X3,
  List,
  MapPin,
  AlertTriangle,
  CheckCircle2,
  Clock,
  GripVertical,
  RefreshCw,
} from "lucide-react";
import { Reservation, Room, getReservations, getRooms } from "@/lib/api";

/* ─── Types ─── */
interface TapechartReservation extends Reservation {
  color?: string;
}

/* ─── Helpers ─── */
function addDays(dateStr: string, days: number): string {
  const d = new Date(dateStr + "T00:00:00");
  d.setDate(d.getDate() + days);
  return d.toISOString().split("T")[0];
}

function isDateInRange(date: string, start: string, end: string): boolean {
  const t = new Date(date + "T00:00:00").getTime();
  const s = new Date(start + "T00:00:00").getTime();
  const e = new Date(end + "T00:00:00").getTime();
  return t >= s && t < e;
}

const resColors = [
  "bg-sky-500", "bg-violet-500", "bg-amber-500", "bg-rose-500",
  "bg-emerald-500", "bg-orange-500", "bg-cyan-500", "bg-pink-500",
];

const statusColors: Record<string, string> = {
  occupied: "bg-rose-500",
  vacant_clean: "bg-emerald-500",
  vacant_dirty: "bg-amber-500",
  blocked: "bg-slate-500",
  maintenance: "bg-orange-500",
};

const statusBadge: Record<string, string> = {
  confirmed: "badge-blue",
  checked_in: "badge-green",
  checked_out: "badge-amber",
  cancelled: "badge-red",
  no_show: "badge-red",
};

/* ─── Component ─── */
export default function TapechartPage() {
  const today = useMemo(() => new Date().toISOString().split("T")[0], []);
  const [startDate, setStartDate] = useState(today);
  const [dayCount, setDayCount] = useState(14);
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [filterFloor, setFilterFloor] = useState<string>("all");
  const [filterType, setFilterType] = useState<string>("all");
  const [searchRoom, setSearchRoom] = useState("");
  const [draggingRes, setDraggingRes] = useState<TapechartReservation | null>(null);
  const [hoveredCell, setHoveredCell] = useState<{ room: string; date: string } | null>(null);
  const [showNewRes, setShowNewRes] = useState<{ room: string; date: string } | null>(null);
  const [selectedRes, setSelectedRes] = useState<TapechartReservation | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [reservations, setReservations] = useState<TapechartReservation[]>([]);
  const [rooms, setRooms] = useState<Room[]>([]);
  const scrollRef = useRef<HTMLDivElement>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [res, rms] = await Promise.all([getReservations(), getRooms()]);
      const colored = res.map((r, i) => ({ ...r, color: r.color || resColors[i % resColors.length] }));
      setReservations(colored);
      setRooms(rms);
    } catch (e: any) {
      setError(e.message || "Failed to load data");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const dates = useMemo(() => {
    return Array.from({ length: dayCount }, (_, i) => addDays(startDate, i));
  }, [startDate, dayCount]);

  const filteredRooms = useMemo(() => {
    return rooms.filter((r) => {
      const matchFloor = filterFloor === "all" || r.floor === filterFloor;
      const matchType = filterType === "all" || r.type === filterType;
      const matchSearch = !searchRoom || r.number.includes(searchRoom) || r.type.toLowerCase().includes(searchRoom.toLowerCase());
      return matchFloor && matchType && matchSearch;
    });
  }, [rooms, filterFloor, filterType, searchRoom]);

  const floors = useMemo(() => [...new Set(rooms.map((r) => r.floor))].sort(), [rooms]);
  const types = useMemo(() => [...new Set(rooms.map((r) => r.type))].sort(), [rooms]);

  const getReservationsForRoomDate = useCallback(
    (roomNumber: string, date: string) => {
      return reservations.filter(
        (r) => r.room_number === roomNumber && isDateInRange(date, r.check_in, r.check_out)
      );
    },
    [reservations]
  );

  const getSpanForRes = useCallback(
    (res: TapechartReservation, roomNumber: string) => {
      if (res.room_number !== roomNumber) return 0;
      const overlapStart = new Date(Math.max(new Date(startDate + "T00:00:00").getTime(), new Date(res.check_in + "T00:00:00").getTime()));
      const overlapEnd = new Date(Math.min(new Date(addDays(startDate, dayCount) + "T00:00:00").getTime(), new Date(res.check_out + "T00:00:00").getTime()));
      const days = Math.round((overlapEnd.getTime() - overlapStart.getTime()) / (1000 * 60 * 60 * 24));
      return Math.max(0, days);
    },
    [startDate, dayCount]
  );

  const handlePrev = () => setStartDate((d) => addDays(d, -7));
  const handleNext = () => setStartDate((d) => addDays(d, 7));
  const handleToday = () => setStartDate(today);

  const handleCellClick = (room: Room, date: string) => {
    const existing = getReservationsForRoomDate(room.number, date);
    if (existing.length > 0) {
      setSelectedRes(existing[0]);
    } else if (!draggingRes) {
      setShowNewRes({ room: room.number, date });
    }
  };

  const handleDrop = (roomNumber: string, date: string) => {
    if (!draggingRes) return;
    console.log("Move", draggingRes.id, "to", roomNumber, "starting", date);
    setDraggingRes(null);
  };

  const occupancyStats = useMemo(() => {
    const totalRooms = filteredRooms.length;
    const occupiedToday = filteredRooms.filter((r) => r.status === "occupied").length;
    const availableToday = filteredRooms.filter((r) => r.status === "vacant_clean").length;
    const departingToday = reservations.filter((r) => r.check_out === today && r.status === "checked_in").length;
    const arrivingToday = reservations.filter((r) => r.check_in === today && r.status !== "checked_in").length;
    return { totalRooms, occupiedToday, availableToday, departingToday, arrivingToday, occupancyRate: totalRooms ? Math.round((occupiedToday / totalRooms) * 100) : 0 };
  }, [filteredRooms, reservations, today]);

  if (loading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <RefreshCw className="h-8 w-8 animate-spin text-nexus-400" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex h-screen flex-col items-center justify-center gap-4">
        <AlertTriangle className="h-10 w-10 text-rose-400" />
        <p className="text-rose-400">{error}</p>
        <button onClick={fetchData} className="btn-primary">Retry</button>
      </div>
    );
  }

  return (
    <div className="flex h-screen flex-col space-y-4 overflow-hidden p-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Grid3X3 className="h-6 w-6 text-nexus-400" />
            Tapechart
          </h1>
          <p className="text-sm text-slate-400">Visual room availability grid — drag to assign, click to book</p>
        </div>
        <div className="flex items-center gap-2">
          <div className="flex items-center rounded-lg border border-slate-700 bg-slate-800">
            <button onClick={handlePrev} className="p-2 text-slate-400 hover:text-white">
              <ChevronLeft className="h-4 w-4" />
            </button>
            <button onClick={handleToday} className="px-3 py-2 text-xs font-medium text-white hover:bg-slate-700">
              Today
            </button>
            <button onClick={handleNext} className="p-2 text-slate-400 hover:text-white">
              <ChevronRight className="h-4 w-4" />
            </button>
          </div>
          <select
            value={dayCount}
            onChange={(e) => setDayCount(Number(e.target.value))}
            className="input text-xs"
          >
            <option value={7}>7 days</option>
            <option value={14}>14 days</option>
            <option value={21}>21 days</option>
            <option value={30}>30 days</option>
          </select>
          <div className="flex rounded-lg border border-slate-700 bg-slate-800">
            <button
              onClick={() => setViewMode("grid")}
              className={`p-2 ${viewMode === "grid" ? "text-white bg-slate-700" : "text-slate-400 hover:text-white"}`}
            >
              <Grid3X3 className="h-4 w-4" />
            </button>
            <button
              onClick={() => setViewMode("list")}
              className={`p-2 ${viewMode === "list" ? "text-white bg-slate-700" : "text-slate-400 hover:text-white"}`}
            >
              <List className="h-4 w-4" />
            </button>
          </div>
          <button onClick={fetchData} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white" title="Refresh">
            <RefreshCw className={`h-4 w-4 ${loading ? "animate-spin" : ""}`} />
          </button>
          <Link href="/reservations/new" className="btn-primary">
            <Plus className="h-4 w-4" />
            New Booking
          </Link>
        </div>
      </div>

      {/* Stats Row */}
      <div className="grid grid-cols-5 gap-3">
        <StatCard label="Total Rooms" value={occupancyStats.totalRooms} icon={BedDouble} color="sky" />
        <StatCard label="Occupied" value={occupancyStats.occupiedToday} icon={CheckCircle2} color="rose" />
        <StatCard label="Available" value={occupancyStats.availableToday} icon={Grid3X3} color="emerald" />
        <StatCard label="Arriving Today" value={occupancyStats.arrivingToday} icon={ArrowRightLeft} color="violet" />
        <StatCard label="Departing Today" value={occupancyStats.departingToday} icon={Clock} color="amber" />
      </div>

      {/* Filters */}
      <div className="flex items-center gap-3">
        <select value={filterFloor} onChange={(e) => setFilterFloor(e.target.value)} className="input text-xs">
          <option value="all">All Floors</option>
          {floors.map((f) => (
            <option key={f} value={f}>Floor {f}</option>
          ))}
        </select>
        <select value={filterType} onChange={(e) => setFilterType(e.target.value)} className="input text-xs">
          <option value="all">All Types</option>
          {types.map((t) => (
            <option key={t} value={t}>{t}</option>
          ))}
        </select>
        <div className="relative">
          <Search className="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-slate-500" />
          <input
            type="text"
            placeholder="Search room..."
            value={searchRoom}
            onChange={(e) => setSearchRoom(e.target.value)}
            className="input w-40 pl-8 text-xs"
          />
        </div>
      </div>

      {/* Tapechart Grid */}
      <div ref={scrollRef} className="flex-1 overflow-auto rounded-lg border border-slate-700">
        <div className="min-w-max">
          {/* Header Row */}
          <div className="flex border-b border-slate-700 bg-slate-800">
            <div className="sticky left-0 z-10 w-32 shrink-0 border-r border-slate-700/50 bg-slate-800 p-2">
              <span className="text-[10px] font-bold uppercase text-slate-500">Room</span>
            </div>
            {dates.map((date) => {
              const isToday = date === today;
              const isWeekend = new Date(date + "T00:00:00").getDay() % 6 === 0;
              return (
                <div key={date} className={`w-20 shrink-0 border-r border-slate-700/50 p-2 text-center ${isToday ? "bg-nexus-500/10" : isWeekend ? "bg-slate-800/50" : ""}`}>
                  <div className={`text-[10px] font-bold ${isToday ? "text-nexus-400" : "text-slate-400"}`}>
                    {isToday ? "TODAY" : new Date(date + "T00:00:00").toLocaleDateString("en-US", { weekday: "short" }).toUpperCase()}
                  </div>
                  <div className={`text-xs ${isToday ? "font-bold text-white" : "text-slate-300"}`}>{new Date(date + "T00:00:00").getDate()}</div>
                  <div className="text-[9px] text-slate-500">{new Date(date + "T00:00:00").toLocaleDateString("en-US", { month: "short" })}</div>
                </div>
              );
            })}
          </div>

          {/* Room Rows */}
          {floors.map((floor) => {
            const floorRooms = filteredRooms.filter((r) => r.floor === floor);
            if (floorRooms.length === 0) return null;
            return (
              <div key={floor}>
                <div className="flex border-b border-slate-700/50 bg-slate-800/30">
                  <div className="sticky left-0 z-10 w-32 shrink-0 border-r border-slate-700/50 p-2">
                    <span className="text-[10px] font-bold uppercase text-slate-500">Floor {floor}</span>
                  </div>
                  <div className="flex">{dates.map((date) => <div key={date} className="w-20 shrink-0 border-r border-slate-700/30" />)}</div>
                </div>
                {floorRooms.map((room) => (
                  <div key={room.number} className="flex border-b border-slate-700/30 hover:bg-slate-800/20">
                    <div className="sticky left-0 z-10 w-32 shrink-0 border-r border-slate-700/50 bg-slate-900 p-2">
                      <div className="flex items-center gap-1.5">
                        <div className={`h-2 w-2 rounded-full ${statusColors[room.status]}`} />
                        <span className="text-xs font-bold text-white">{room.number}</span>
                      </div>
                      <div className="mt-0.5 text-[9px] text-slate-500">{room.type}</div>
                      <div className="text-[9px] text-slate-600">{room.bed_type}</div>
                    </div>
                    {dates.map((date) => {
                      const resList = getReservationsForRoomDate(room.number, date);
                      const isToday = date === today;
                      const isHovered = hoveredCell?.room === room.number && hoveredCell?.date === date;
                      const isDropTarget = draggingRes && isHovered && resList.length === 0;
                      const spanInfo = resList.map((r) => getSpanForRes(r, room.number)).find((s) => s > 0);
                      const spanRes = spanInfo ? resList.find((r) => getSpanForRes(r, room.number) > 0) : null;
                      return (
                        <div
                          key={`${room.number}-${date}`}
                          className={`relative w-20 shrink-0 border-r border-slate-700/30 min-h-[48px] ${isToday ? "bg-nexus-500/5" : ""} ${isDropTarget ? "bg-emerald-500/10" : ""} ${!resList.length && !draggingRes ? "cursor-pointer hover:bg-slate-800/40" : ""}`}
                          onMouseEnter={() => setHoveredCell({ room: room.number, date })}
                          onMouseLeave={() => setHoveredCell(null)}
                          onClick={() => handleCellClick(room, date)}
                          onDragOver={(e) => { e.preventDefault(); setHoveredCell({ room: room.number, date }); }}
                          onDrop={(e) => { e.preventDefault(); handleDrop(room.number, date); }}
                        >
                          {spanRes && (
                            <div
                              draggable
                              onDragStart={() => setDraggingRes(spanRes)}
                              onDragEnd={() => setDraggingRes(null)}
                              className={`absolute inset-y-0.5 left-0.5 z-10 rounded cursor-move ${spanRes.color || "bg-sky-500"} hover:brightness-110`}
                              style={{ width: `${Math.max(spanInfo || 1, 1) * 5 - 0.25}rem`, minWidth: "4.5rem" }}
                              onClick={(e) => { e.stopPropagation(); setSelectedRes(spanRes); }}
                            >
                              <div className="flex h-full items-center px-1.5 overflow-hidden">
                                <span className="text-[9px] font-medium text-white truncate">{spanRes.guest_name}</span>
                                {spanRes.vip && <span className="ml-1 text-[7px] text-amber-300">★</span>}
                              </div>
                            </div>
                          )}
                          {!resList.length && !draggingRes && isHovered && (
                            <div className="absolute inset-0 flex items-center justify-center">
                              <Plus className="h-3 w-3 text-slate-600 opacity-0 hover:opacity-100" />
                            </div>
                          )}
                        </div>
                      );
                    })}
                  </div>
                ))}
              </div>
            );
          })}
        </div>
      </div>

      {/* Legend */}
      <div className="flex items-center gap-4 text-[10px] text-slate-500">
        <span className="font-bold uppercase">Room Status:</span>
        {Object.entries(statusColors).map(([status, color]) => (
          <span key={status} className="flex items-center gap-1"><div className={`h-2 w-2 rounded-full ${color}`} /> {status.replace("_", " ")}</span>
        ))}
        <span className="ml-auto">Drag reservation blocks to reassign rooms</span>
      </div>

      {/* New Reservation Modal */}
      {showNewRes && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-lg rounded-xl border border-slate-700 bg-slate-800 p-6 shadow-2xl">
            <div className="mb-4 flex items-center justify-between">
              <h3 className="text-lg font-bold text-white">New Reservation</h3>
              <button onClick={() => setShowNewRes(null)} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
            </div>
            <div className="space-y-3">
              <div>
                <label className="mb-1 block text-xs text-slate-400">Room</label>
                <input type="text" value={showNewRes.room} disabled className="input w-full text-sm" />
              </div>
              <div>
                <label className="mb-1 block text-xs text-slate-400">Date</label>
                <input type="text" value={showNewRes.date} disabled className="input w-full text-sm" />
              </div>
              <p className="text-xs text-slate-500">Use <Link href="/reservations/new" className="text-nexus-400 hover:underline">full booking form</Link> for complete reservation details.</p>
            </div>
            <div className="mt-4 flex justify-end gap-2">
              <button onClick={() => setShowNewRes(null)} className="btn-secondary">Cancel</button>
              <Link href={`/reservations/new?room=${showNewRes.room}&date=${showNewRes.date}`} className="btn-primary">Continue</Link>
            </div>
          </div>
        </div>
      )}

      {/* Reservation Detail Modal */}
      {selectedRes && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-800 p-6 shadow-2xl">
            <div className="mb-4 flex items-center justify-between">
              <h3 className="text-lg font-bold text-white">Reservation Detail</h3>
              <button onClick={() => setSelectedRes(null)} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
            </div>
            <div className="space-y-3 text-sm">
              <div className="flex items-center gap-2"><User className="h-4 w-4 text-slate-500" /><span className="font-medium text-white">{selectedRes.guest_name}</span>{selectedRes.vip && <span className="text-amber-400 text-xs">★ VIP</span>}</div>
              <div className="flex items-center gap-2"><BedDouble className="h-4 w-4 text-slate-500" /><span className="text-slate-300">Room {selectedRes.room_number} · {selectedRes.room_type}</span></div>
              <div className="flex items-center gap-2"><CalendarDays className="h-4 w-4 text-slate-500" /><span className="text-slate-300">{selectedRes.check_in} → {selectedRes.check_out}</span></div>
              <div className="flex items-center gap-2"><MapPin className="h-4 w-4 text-slate-500" /><span className={`badge ${statusBadge[selectedRes.status] || "badge-gray"}`}>{selectedRes.status}</span></div>
              <div className="flex items-center gap-2"><Clock className="h-4 w-4 text-slate-500" /><span className="text-slate-300">{selectedRes.adults} adults, {selectedRes.children} children</span></div>
            </div>
            <div className="mt-4 flex justify-end gap-2">
              <button onClick={() => setSelectedRes(null)} className="btn-secondary">Close</button>
              <Link href={`/reservations/${selectedRes.id}`} className="btn-primary">View Full Details</Link>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

/* ─── Stat Card ─── */
function StatCard({ label, value, icon: Icon, color }: { label: string; value: number; icon: any; color: string }) {
  const colorMap: Record<string, string> = {
    sky: "bg-sky-500/10 text-sky-400",
    rose: "bg-rose-500/10 text-rose-400",
    emerald: "bg-emerald-500/10 text-emerald-400",
    violet: "bg-violet-500/10 text-violet-400",
    amber: "bg-amber-500/10 text-amber-400",
  };
  return (
    <div className="rounded-lg border border-slate-700 bg-slate-800 p-3">
      <div className="flex items-center gap-2">
        <div className={`rounded-md p-1.5 ${colorMap[color] || ""}`}>
          <Icon className="h-4 w-4" />
        </div>
        <div>
          <div className="text-lg font-bold text-white">{value}</div>
          <div className="text-[10px] text-slate-500">{label}</div>
        </div>
      </div>
    </div>
  );
}
