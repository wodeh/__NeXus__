"use client";

import { useState, useMemo, useCallback, useRef } from "react";
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
} from "lucide-react";

/* ─── Types ─── */
interface Reservation {
  id: string;
  guest_name: string;
  email?: string;
  phone?: string;
  room_number: string;
  room_type: string;
  check_in: string; // YYYY-MM-DD
  check_out: string; // YYYY-MM-DD
  adults: number;
  children: number;
  status: "confirmed" | "checked_in" | "checked_out" | "cancelled" | "no_show";
  source: "walk_in" | "ota" | "direct" | "agent";
  total: number;
  balance: number;
  vip?: boolean;
  color?: string;
}

interface Room {
  number: string;
  type: string;
  floor: string;
  status: "occupied" | "vacant_clean" | "vacant_dirty" | "blocked" | "maintenance";
  bed_type?: string;
}

/* ─── Mock Data ─── */
const mockRooms: Room[] = [
  { number: "101", type: "Standard", floor: "1", status: "occupied", bed_type: "Queen" },
  { number: "102", type: "Standard", floor: "1", status: "vacant_clean", bed_type: "Queen" },
  { number: "103", type: "Deluxe", floor: "1", status: "occupied", bed_type: "King" },
  { number: "104", type: "Standard", floor: "1", status: "blocked", bed_type: "Twin" },
  { number: "105", type: "Standard", floor: "1", status: "vacant_dirty", bed_type: "Queen" },
  { number: "106", type: "Accessible", floor: "1", status: "vacant_clean", bed_type: "King" },
  { number: "201", type: "Deluxe King", floor: "2", status: "occupied", bed_type: "King" },
  { number: "202", type: "Deluxe King", floor: "2", status: "vacant_clean", bed_type: "King" },
  { number: "203", type: "Suite", floor: "2", status: "occupied", bed_type: "King+Sofa" },
  { number: "204", type: "Deluxe King", floor: "2", status: "vacant_clean", bed_type: "King" },
  { number: "205", type: "Deluxe", floor: "2", status: "maintenance", bed_type: "Queen" },
  { number: "206", type: "Standard", floor: "2", status: "vacant_clean", bed_type: "Twin" },
  { number: "301", type: "Suite", floor: "3", status: "occupied", bed_type: "King+Sofa" },
  { number: "302", type: "Suite", floor: "3", status: "blocked", bed_type: "King+Sofa" },
  { number: "303", type: "Deluxe King", floor: "3", status: "vacant_clean", bed_type: "King" },
  { number: "304", type: "Deluxe", floor: "3", status: "occupied", bed_type: "Queen" },
  { number: "305", type: "Standard", floor: "3", status: "vacant_clean", bed_type: "Queen" },
  { number: "306", type: "Standard", floor: "3", status: "vacant_dirty", bed_type: "Twin" },
  { number: "401", type: "Suite", floor: "4", status: "vacant_clean", bed_type: "King+Sofa" },
  { number: "402", type: "Deluxe King", floor: "4", status: "vacant_clean", bed_type: "King" },
  { number: "403", type: "Deluxe", floor: "4", status: "occupied", bed_type: "Queen" },
  { number: "404", type: "Standard", floor: "4", status: "vacant_clean", bed_type: "Twin" },
];

const roomColors: Record<string, string> = {
  Standard: "bg-slate-600",
  Deluxe: "bg-sky-600",
  "Deluxe King": "bg-violet-600",
  Suite: "bg-amber-600",
  Accessible: "bg-emerald-600",
};

const statusColors: Record<string, string> = {
  occupied: "bg-rose-500",
  vacant_clean: "bg-emerald-500",
  vacant_dirty: "bg-amber-500",
  blocked: "bg-slate-500",
  maintenance: "bg-orange-500",
};

const resColors = [
  "bg-sky-500",
  "bg-violet-500",
  "bg-amber-500",
  "bg-rose-500",
  "bg-emerald-500",
  "bg-orange-500",
  "bg-cyan-500",
  "bg-pink-500",
];

function generateMockReservations(): Reservation[] {
  const baseDate = new Date("2026-05-10");
  const guests = [
    "Alice Chen", "Bob Smith", "Carol Jones", "David Lee", "Emma Wilson",
    "Frank Brown", "Grace Kim", "Henry Park", "Iris Zhao", "Jack Ma",
    "Kate Liu", "Leo Wang", "Mia Tang", "Noah Zhang", "Olivia Wu",
  ];
  const res: Reservation[] = [];
  
  guests.forEach((name, i) => {
    const offset = Math.floor(Math.random() * 14) - 3;
    const nights = Math.floor(Math.random() * 5) + 1;
    const checkIn = new Date(baseDate);
    checkIn.setDate(checkIn.getDate() + offset);
    const checkOut = new Date(checkIn);
    checkOut.setDate(checkOut.getDate() + nights);
    
    const room = mockRooms[i % mockRooms.length];
    const isOccupied = offset <= 0 && offset + nights > 0;
    
    res.push({
      id: `r-${String(i + 1).padStart(3, "0")}`,
      guest_name: name,
      email: `${name.toLowerCase().replace(" ", ".")}@email.com`,
      room_number: isOccupied ? room.number : "",
      room_type: room.type,
      check_in: checkIn.toISOString().split("T")[0],
      check_out: checkOut.toISOString().split("T")[0],
      adults: Math.floor(Math.random() * 2) + 1,
      children: Math.random() > 0.7 ? 1 : 0,
      status: isOccupied ? (Math.random() > 0.3 ? "checked_in" : "confirmed") : "confirmed",
      source: ["walk_in", "ota", "direct", "agent"][Math.floor(Math.random() * 4)] as any,
      total: nights * (room.type === "Suite" ? 229 : room.type === "Deluxe" || room.type === "Deluxe King" ? 129 : 89),
      balance: Math.random() > 0.5 ? 0 : Math.floor(Math.random() * 300),
      vip: Math.random() > 0.85,
      color: resColors[i % resColors.length],
    });
  });
  
  return res;
}

const mockReservations = generateMockReservations();

/* ─── Helpers ─── */
function addDays(dateStr: string, days: number): string {
  const d = new Date(dateStr + "T00:00:00");
  d.setDate(d.getDate() + days);
  return d.toISOString().split("T")[0];
}

function diffDays(a: string, b: string): number {
  const d1 = new Date(a + "T00:00:00").getTime();
  const d2 = new Date(b + "T00:00:00").getTime();
  return Math.round((d2 - d1) / (1000 * 60 * 60 * 24));
}

function formatDateShort(dateStr: string): string {
  const d = new Date(dateStr + "T00:00:00");
  return d.toLocaleDateString("en-US", { weekday: "short", month: "short", day: "numeric" });
}

function isDateInRange(date: string, start: string, end: string): boolean {
  const t = new Date(date + "T00:00:00").getTime();
  const s = new Date(start + "T00:00:00").getTime();
  const e = new Date(end + "T00:00:00").getTime();
  return t >= s && t < e;
}

/* ─── Component ─── */
export default function TapechartPage() {
  const today = "2026-05-10";
  const [startDate, setStartDate] = useState(today);
  const [dayCount, setDayCount] = useState(14);
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [filterFloor, setFilterFloor] = useState<string>("all");
  const [filterType, setFilterType] = useState<string>("all");
  const [searchRoom, setSearchRoom] = useState("");
  const [draggingRes, setDraggingRes] = useState<Reservation | null>(null);
  const [hoveredCell, setHoveredCell] = useState<{ room: string; date: string } | null>(null);
  const [showNewRes, setShowNewRes] = useState<{ room: string; date: string } | null>(null);
  const [selectedRes, setSelectedRes] = useState<Reservation | null>(null);
  const scrollRef = useRef<HTMLDivElement>(null);

  const dates = useMemo(() => {
    return Array.from({ length: dayCount }, (_, i) => addDays(startDate, i));
  }, [startDate, dayCount]);

  const filteredRooms = useMemo(() => {
    return mockRooms.filter((r) => {
      const matchFloor = filterFloor === "all" || r.floor === filterFloor;
      const matchType = filterType === "all" || r.type === filterType;
      const matchSearch = !searchRoom || r.number.includes(searchRoom) || r.type.toLowerCase().includes(searchRoom.toLowerCase());
      return matchFloor && matchType && matchSearch;
    });
  }, [filterFloor, filterType, searchRoom]);

  const floors = useMemo(() => [...new Set(mockRooms.map((r) => r.floor))].sort(), []);
  const types = useMemo(() => [...new Set(mockRooms.map((r) => r.type))].sort(), []);

  const getReservationsForRoomDate = useCallback(
    (roomNumber: string, date: string) => {
      return mockReservations.filter(
        (r) => r.room_number === roomNumber && isDateInRange(date, r.check_in, r.check_out)
      );
    },
    []
  );

  const getSpanForRes = useCallback(
    (res: Reservation, roomNumber: string) => {
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
    // In production: API call to move reservation
    console.log("Move", draggingRes.id, "to", roomNumber, "starting", date);
    setDraggingRes(null);
  };

  const occupancyStats = useMemo(() => {
    const totalRooms = filteredRooms.length;
    const occupiedToday = filteredRooms.filter((r) => r.status === "occupied").length;
    const availableToday = filteredRooms.filter((r) => r.status === "vacant_clean").length;
    const departingToday = mockReservations.filter((r) => r.check_out === today && r.status === "checked_in").length;
    const arrivingToday = mockReservations.filter((r) => r.check_in === today && r.status !== "checked_in").length;
    return { totalRooms, occupiedToday, availableToday, departingToday, arrivingToday, occupancyRate: totalRooms ? Math.round((occupiedToday / totalRooms) * 100) : 0 };
  }, [filteredRooms, today]);

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
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
          <input
            className="input w-full pl-9 text-xs"
            placeholder="Search room or type..."
            value={searchRoom}
            onChange={(e) => setSearchRoom(e.target.value)}
          />
        </div>
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
        {draggingRes && (
          <div className="flex items-center gap-2 rounded-lg border border-dashed border-nexus-500/30 bg-nexus-500/5 px-3 py-1.5">
            <GripVertical className="h-3 w-3 text-nexus-400" />
            <span className="text-xs text-nexus-400">Dragging: {draggingRes.guest_name}</span>
            <button onClick={() => setDraggingRes(null)} className="text-slate-500 hover:text-white">
              <X className="h-3 w-3" />
            </button>
          </div>
        )}
      </div>

      {/* Tapechart Grid */}
      <div className="flex-1 overflow-hidden rounded-lg border border-slate-700 bg-slate-900">
        <div ref={scrollRef} className="h-full overflow-auto">
          <div className="min-w-max">
            {/* Date Header */}
            <div className="sticky top-0 z-20 flex border-b border-slate-700 bg-slate-800">
              <div className="sticky left-0 z-30 w-32 shrink-0 border-r border-slate-700 bg-slate-800 p-2">
                <span className="text-[10px] font-bold uppercase text-slate-400">Room</span>
              </div>
              {dates.map((date) => {
                const isToday = date === today;
                const isWeekend = new Date(date + "T00:00:00").getDay() % 6 === 0;
                return (
                  <div
                    key={date}
                    className={`w-20 shrink-0 border-r border-slate-700/50 p-2 text-center ${isToday ? "bg-nexus-500/10" : isWeekend ? "bg-slate-800/50" : ""}`}
                  >
                    <div className={`text-[10px] font-bold ${isToday ? "text-nexus-400" : "text-slate-400"}`}>
                      {isToday ? "TODAY" : new Date(date + "T00:00:00").toLocaleDateString("en-US", { weekday: "short" }).toUpperCase()}
                    </div>
                    <div className={`text-xs ${isToday ? "text-white font-bold" : "text-slate-300"}`}>
                      {new Date(date + "T00:00:00").getDate()}
                    </div>
                    <div className="text-[9px] text-slate-500">
                      {new Date(date + "T00:00:00").toLocaleDateString("en-US", { month: "short" })}
                    </div>
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
                  {/* Floor Header */}
                  <div className="flex border-b border-slate-700/50 bg-slate-800/30">
                    <div className="sticky left-0 z-10 w-32 shrink-0 border-r border-slate-700/50 p-2">
                      <span className="text-[10px] font-bold uppercase text-slate-500">Floor {floor}</span>
                    </div>
                    <div className="flex">
                      {dates.map((date) => (
                        <div key={date} className="w-20 shrink-0 border-r border-slate-700/30" />
                      ))}
                    </div>
                  </div>
                  
                  {floorRooms.map((room) => (
                    <div key={room.number} className="flex border-b border-slate-700/30 hover:bg-slate-800/20">
                      {/* Room Info Cell */}
                      <div className="sticky left-0 z-10 w-32 shrink-0 border-r border-slate-700/50 bg-slate-900 p-2">
                        <div className="flex items-center gap-1.5">
                          <div className={`h-2 w-2 rounded-full ${statusColors[room.status]}`} />
                          <span className="text-xs font-bold text-white">{room.number}</span>
                        </div>
                        <div className="mt-0.5 text-[9px] text-slate-500">{room.type}</div>
                        <div className="text-[9px] text-slate-600">{room.bed_type}</div>
                      </div>
                      
                      {/* Date Cells */}
                      {dates.map((date) => {
                        const reservations = getReservationsForRoomDate(room.number, date);
                        const isToday = date === today;
                        const isHovered = hoveredCell?.room === room.number && hoveredCell?.date === date;
                        const isDropTarget = draggingRes && isHovered && reservations.length === 0;
                        
                        // Check if this cell is the start of a reservation span
                        const startsHere = reservations.some((r) => r.check_in === date);
                        const spanRes = reservations.find((r) => r.check_in === date);
                        
                        return (
                          <div
                            key={`${room.number}-${date}`}
                            className={`relative w-20 shrink-0 border-r border-slate-700/30 min-h-[48px] ${
                              isToday ? "bg-nexus-500/5" : ""
                            } ${isDropTarget ? "bg-emerald-500/10 border-emerald-500/30" : ""} ${
                              !reservations.length && !draggingRes ? "cursor-pointer hover:bg-slate-800/40" : ""
                            }`}
                            onMouseEnter={() => setHoveredCell({ room: room.number, date })}
                            onMouseLeave={() => setHoveredCell(null)}
                            onClick={() => handleCellClick(room, date)}
                            onDragOver={(e) => {
                              e.preventDefault();
                              setHoveredCell({ room: room.number, date });
                            }}
                            onDrop={(e) => {
                              e.preventDefault();
                              handleDrop(room.number, date);
                            }}
                          >
                            {/* Render reservation bar if it starts here */}
                            {spanRes && (
                              <div
                                draggable
                                onDragStart={() => setDraggingRes(spanRes)}
                                onDragEnd={() => setDraggingRes(null)}
                                className={`absolute inset-y-0.5 left-0.5 z-10 rounded cursor-move ${spanRes.color || "bg-sky-500"} hover:brightness-110 transition-all`}
                                style={{
                                  width: `${Math.max(getSpanForRes(spanRes, room.number), 1) * 5 - 0.25}rem`,
                                  minWidth: "4.5rem",
                                }}
                                onClick={(e) => {
                                  e.stopPropagation();
                                  setSelectedRes(spanRes);
                                }}
                              >
                                <div className="flex h-full items-center px-1.5 overflow-hidden">
                                  <span className="text-[9px] font-medium text-white truncate leading-tight">
                                    {spanRes.guest_name}
                                  </span>
                                  {spanRes.vip && (
                                    <span className="ml-1 text-[7px] text-amber-300">★</span>
                                  )}
                                </div>
                              </div>
                            )}
                            
                            {/* Continuation indicator (subtle) */}
                            {reservations.some((r) => r.check_in !== date) && !spanRes && (
                              <div className="absolute inset-0 bg-slate-800/10" />
                            )}
                            
                            {/* Empty cell hover state for quick book */}
                            {!reservations.length && !draggingRes && isHovered && (
                              <div className="absolute inset-0 flex items-center justify-center">
                                <Plus className="h-3 w-3 text-slate-600 opacity-0 hover:opacity-100 transition-opacity" />
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
      </div>

      {/* Legend */}
      <div className="flex items-center gap-4 text-[10px] text-slate-500">
        <span className="font-bold uppercase">Room Status:</span>
        <span className="flex items-center gap-1"><div className="h-2 w-2 rounded-full bg-emerald-500" /> Available</span>
        <span className="flex items-center gap-1"><div className="h-2 w-2 rounded-full bg-rose-500" /> Occupied</span>
        <span className="flex items-center gap-1"><div className="h-2 w-2 rounded-full bg-amber-500" /> Dirty</span>
        <span className="flex items-center gap-1"><div className="h-2 w-2 rounded-full bg-slate-500" /> Blocked</span>
        <span className="flex items-center gap-1"><div className="h-2 w-2 rounded-full bg-orange-500" /> Maintenance</span>
        <span className="ml-auto">Drag reservation blocks to reassign rooms</span>
      </div>

      {/* New Reservation Modal */}
      {showNewRes && (
        <NewReservationModal
          room={showNewRes.room}
          date={showNewRes.date}
          rooms={mockRooms}
          onClose={() => setShowNewRes(null)}
        />
      )}

      {/* Reservation Detail Modal */}
      {selectedRes && (
        <ReservationDetailModal
          reservation={selectedRes}
          onClose={() => setSelectedRes(null)}
        />
      )}
    </div>
  );
}

/* ─── Sub-components ─── */

function StatCard({ label, value, icon: Icon, color }: { label: string; value: number; icon: React.ElementType; color: string }) {
  const colorMap: Record<string, string> = {
    sky: "bg-sky-500/10 text-sky-400 border-sky-500/20",
    emerald: "bg-emerald-500/10 text-emerald-400 border-emerald-500/20",
    violet: "bg-violet-500/10 text-violet-400 border-violet-500/20",
    amber: "bg-amber-500/10 text-amber-400 border-amber-500/20",
    rose: "bg-rose-500/10 text-rose-400 border-rose-500/20",
    nexus: "bg-nexus-500/10 text-nexus-400 border-nexus-500/20",
  };
  return (
    <div className={`rounded-lg border p-2.5 ${colorMap[color]}`}>
      <div className="flex items-center justify-between">
        <span className="text-[10px] font-medium opacity-80">{label}</span>
        <Icon className="h-3.5 w-3.5 opacity-60" />
      </div>
      <p className="mt-1 text-xl font-bold">{value}</p>
    </div>
  );
}

function NewReservationModal({ room, date, rooms, onClose }: { room: string; date: string; rooms: Room[]; onClose: () => void }) {
  const roomData = rooms.find((r) => r.number === room);
  const [form, setForm] = useState({
    guest_name: "",
    email: "",
    phone: "",
    nights: 1,
    adults: 2,
    children: 0,
    special_requests: "",
  });

  const rate = roomData?.type === "Suite" ? 229 : roomData?.type === "Deluxe" || roomData?.type === "Deluxe King" ? 129 : 89;
  const total = form.nights * rate;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-bold text-white">New Booking</h3>
            <p className="text-xs text-slate-400">
              Room {room} · {roomData?.type} · {formatDateShort(date)}
            </p>
          </div>
          <button onClick={onClose} className="text-slate-400 hover:text-white">
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="mt-4 space-y-3">
          <div>
            <label className="mb-1 block text-xs text-slate-400">Guest Name *</label>
            <input
              className="input w-full"
              value={form.guest_name}
              onChange={(e) => setForm({ ...form, guest_name: e.target.value })}
              placeholder="Full name"
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Email</label>
              <input
                className="input w-full"
                value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })}
                placeholder="guest@email.com"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Phone</label>
              <input
                className="input w-full"
                value={form.phone}
                onChange={(e) => setForm({ ...form, phone: e.target.value })}
                placeholder="+1-555-0000"
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Nights</label>
              <input
                type="number"
                min={1}
                max={30}
                className="input w-full"
                value={form.nights}
                onChange={(e) => setForm({ ...form, nights: parseInt(e.target.value) || 1 })}
              />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Guests</label>
              <div className="flex items-center gap-2">
                <input
                  type="number"
                  min={1}
                  max={6}
                  className="input w-16"
                  value={form.adults}
                  onChange={(e) => setForm({ ...form, adults: parseInt(e.target.value) || 1 })}
                />
                <span className="text-xs text-slate-500">+</span>
                <input
                  type="number"
                  min={0}
                  max={4}
                  className="input w-16"
                  value={form.children}
                  onChange={(e) => setForm({ ...form, children: parseInt(e.target.value) || 0 })}
                />
              </div>
            </div>
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Special Requests</label>
            <textarea
              className="input w-full"
              rows={2}
              value={form.special_requests}
              onChange={(e) => setForm({ ...form, special_requests: e.target.value })}
              placeholder="Any special needs..."
            />
          </div>
        </div>

        <div className="mt-4 flex items-center justify-between border-t border-slate-700 pt-4">
          <div className="text-sm text-slate-400">
            {form.nights} nights · <span className="font-bold text-white">${total}</span>
            <span className="text-xs text-slate-500"> @ ${rate}/night</span>
          </div>
          <div className="flex gap-2">
            <button onClick={onClose} className="btn-secondary text-xs">Cancel</button>
            <button className="btn-primary text-xs">Create Booking</button>
          </div>
        </div>
      </div>
    </div>
  );
}

function ReservationDetailModal({ reservation, onClose }: { reservation: Reservation; onClose: () => void }) {
  const nights = diffDays(reservation.check_in, reservation.check_out);
  
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <div className={`flex h-10 w-10 items-center justify-center rounded-full ${reservation.vip ? "bg-amber-500/10" : "bg-slate-700"}`}>
              <User className={`h-5 w-5 ${reservation.vip ? "text-amber-400" : "text-slate-400"}`} />
            </div>
            <div>
              <h3 className="text-lg font-bold text-white flex items-center gap-2">
                {reservation.guest_name}
                {reservation.vip && <span className="rounded bg-amber-500/10 px-1.5 py-0.5 text-[10px] text-amber-400">VIP</span>}
              </h3>
              <p className="text-xs text-slate-400">{reservation.email}</p>
            </div>
          </div>
          <button onClick={onClose} className="text-slate-400 hover:text-white">
            <X className="h-5 w-5" />
          </button>
        </div>

        <div className="mt-4 grid grid-cols-2 gap-3 text-sm">
          <div className="rounded-lg bg-slate-800 p-3">
            <div className="text-[10px] uppercase text-slate-500">Room</div>
            <div className="font-medium text-white">{reservation.room_number || "Unassigned"}</div>
            <div className="text-xs text-slate-400">{reservation.room_type}</div>
          </div>
          <div className="rounded-lg bg-slate-800 p-3">
            <div className="text-[10px] uppercase text-slate-500">Status</div>
            <div className="font-medium text-white capitalize">{reservation.status.replace("_", " ")}</div>
            <div className="text-xs text-slate-400 capitalize">{reservation.source.replace("_", " ")}</div>
          </div>
          <div className="rounded-lg bg-slate-800 p-3">
            <div className="text-[10px] uppercase text-slate-500">Check In</div>
            <div className="font-medium text-white">{formatDateShort(reservation.check_in)}</div>
          </div>
          <div className="rounded-lg bg-slate-800 p-3">
            <div className="text-[10px] uppercase text-slate-500">Check Out</div>
            <div className="font-medium text-white">{formatDateShort(reservation.check_out)}</div>
          </div>
          <div className="rounded-lg bg-slate-800 p-3">
            <div className="text-[10px] uppercase text-slate-500">Guests</div>
            <div className="font-medium text-white">{reservation.adults} adults {reservation.children > 0 && `+ ${reservation.children} children`}</div>
          </div>
          <div className="rounded-lg bg-slate-800 p-3">
            <div className="text-[10px] uppercase text-slate-500">Balance</div>
            <div className={`font-medium ${reservation.balance > 0 ? "text-amber-400" : "text-emerald-400"}`}>
              ${reservation.balance}
            </div>
          </div>
        </div>

        <div className="mt-4 flex gap-2">
          <Link href={`/reservations/${reservation.id}`} className="flex-1 btn-primary text-center text-xs">
            Open Full Profile
          </Link>
          <button className="btn-secondary text-xs">Edit</button>
          <button className="rounded-lg border border-rose-500/20 bg-rose-500/10 px-3 py-2 text-xs text-rose-400 hover:bg-rose-500/20">
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
}
