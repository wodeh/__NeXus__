'use client';

import { useState, useEffect, useMemo, useCallback, useRef } from 'react';
import Link from 'next/link';
import {
  BedDouble, Users, User, ArrowRightLeft, DoorOpen, DollarSign, TrendingUp, TrendingDown,
  AlertTriangle, Clock, CheckCircle2, CalendarDays, ChevronRight, Wifi, WifiOff,
  BatteryWarning, KeyRound, Plus, Star, Activity, RefreshCw, GripVertical, Minus,
  Wine, GlassWater, ShoppingBag, Wrench, Trash2, Phone, Move
} from 'lucide-react';
import { Reservation, Room, getReservations, getRooms } from '@/lib/api';

const today = new Date().toISOString().split('T')[0];

/* ─── Activity Types ─── */
interface ActivityItem {
  id: string;
  time: string;
  event: string;
  detail: string;
  icon: any;
  color: string;
  type: 'checkin' | 'checkout' | 'housekeeping' | 'lock' | 'reservation' | 'minibar' | 'roomservice' | 'wifi' | 'alert';
  room?: string;
}

/* ─── Mock data ─── */
const mockRecentActivity: ActivityItem[] = [
  { id: '1', time: '09:15', event: 'Check-in', detail: 'Room 305 — James Wilson', icon: DoorOpen, color: 'emerald', type: 'checkin', room: '305' },
  { id: '2', time: '09:02', event: 'Lock offline', detail: 'Room 302 — OT-302-2026', icon: WifiOff, color: 'rose', type: 'lock', room: '302' },
  { id: '3', time: '08:45', event: 'Housekeeping done', detail: 'Room 104 — Vacant Clean', icon: CheckCircle2, color: 'sky', type: 'housekeeping', room: '104' },
  { id: '4', time: '08:30', event: 'New reservation', detail: 'Booking.com — Room 305', icon: Plus, color: 'nexus', type: 'reservation', room: '305' },
  { id: '5', time: '08:15', event: 'Low battery alert', detail: 'Room 103 — 12% remaining', icon: BatteryWarning, color: 'amber', type: 'alert', room: '103' },
  { id: '6', time: '07:50', event: 'Check-out', detail: 'Room 201 — Sarah Chen · Bill: $234', icon: DoorOpen, color: 'rose', type: 'checkout', room: '201' },
  { id: '7', time: '07:30', event: 'Minibar refilled', detail: 'Room 405 — 12 items restocked', icon: Wine, color: 'violet', type: 'minibar', room: '405' },
  { id: '8', time: '07:15', event: 'Room service delivered', detail: 'Room 501 — Breakfast tray', icon: ShoppingBag, color: 'amber', type: 'roomservice', room: '501' },
];

/* ─── WiFi Floor Status ─── */
interface WiFiFloorStatus {
  floor: string;
  apCount: number;
  onlineAPs: number;
  avgSignal: number; // dBm (negative, closer to 0 is better)
  clientCount: number;
  issues: string[];
}

const mockWiFiFloors: WiFiFloorStatus[] = [
  { floor: 'Lobby', apCount: 4, onlineAPs: 4, avgSignal: -45, clientCount: 87, issues: [] },
  { floor: 'Floor 1', apCount: 6, onlineAPs: 6, avgSignal: -52, clientCount: 43, issues: [] },
  { floor: 'Floor 2', apCount: 6, onlineAPs: 5, avgSignal: -58, clientCount: 38, issues: ['AP-2B offline'] },
  { floor: 'Floor 3', apCount: 6, onlineAPs: 6, avgSignal: -48, clientCount: 41, issues: [] },
  { floor: 'Floor 4', apCount: 6, onlineAPs: 6, avgSignal: -55, clientCount: 35, issues: [] },
  { floor: 'Floor 5', apCount: 6, onlineAPs: 4, avgSignal: -67, clientCount: 29, issues: ['AP-5A weak signal', 'AP-5C channel overlap'] },
];

/* ─── Draggable Widget System ─── */
interface WidgetConfig {
  id: string;
  title: string;
  x: number;
  y: number;
  w: number;
  h: number;
  minW: number;
  minH: number;
}

const defaultLayout: WidgetConfig[] = [
  { id: 'arrivals', title: 'Arrivals Today', x: 0, y: 0, w: 4, h: 6, minW: 3, minH: 4 },
  { id: 'departures', title: 'Departures Today', x: 4, y: 0, w: 4, h: 6, minW: 3, minH: 4 },
  { id: 'activity', title: 'Recent Activity', x: 8, y: 0, w: 4, h: 6, minW: 3, minH: 4 },
  { id: 'wifi', title: 'WiFi Status by Floor', x: 0, y: 6, w: 6, h: 4, minW: 4, minH: 3 },
  { id: 'revenue', title: 'Weekly Revenue', x: 6, y: 6, w: 6, h: 4, minW: 4, minH: 3 },
];

function signalQuality(dbm: number): { label: string; color: string; bars: number } {
  if (dbm >= -50) return { label: 'Excellent', color: 'text-emerald-400', bars: 4 };
  if (dbm >= -60) return { label: 'Good', color: 'text-sky-400', bars: 3 };
  if (dbm >= -70) return { label: 'Fair', color: 'text-amber-400', bars: 2 };
  return { label: 'Poor', color: 'text-rose-400', bars: 1 };
}

export default function HotelDashboardPage() {
  const [reservations, setReservations] = useState<Reservation[]>([]);
  const [rooms, setRooms] = useState<Room[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [layout, setLayout] = useState<WidgetConfig[]>(defaultLayout);
  const [dragging, setDragging] = useState<string | null>(null);
  const [resizing, setResizing] = useState<string | null>(null);
  const [hoverActivity, setHoverActivity] = useState(false);
  const [activityScroll, setActivityScroll] = useState(0);
  const dragOffset = useRef({ x: 0, y: 0 });
  const containerRef = useRef<HTMLDivElement>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [res, rms] = await Promise.all([getReservations(), getRooms()]);
      setReservations(res);
      setRooms(rms);
    } catch (e: any) {
      setError(e.message || 'Failed to load dashboard data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  /* ─── Drag logic ─── */
  const handleMouseDown = (e: React.MouseEvent, widgetId: string) => {
    const widget = layout.find(w => w.id === widgetId);
    if (!widget || !containerRef.current) return;
    const rect = containerRef.current.getBoundingClientRect();
    const cellW = rect.width / 12;
    const cellH = 80; // px per grid row
    dragOffset.current = {
      x: e.clientX - (widget.x * cellW),
      y: e.clientY - (widget.y * cellH),
    };
    setDragging(widgetId);
  };

  const handleMouseMove = useCallback((e: MouseEvent) => {
    if (!dragging || !containerRef.current) return;
    const rect = containerRef.current.getBoundingClientRect();
    const cellW = rect.width / 12;
    const cellH = 80;
    const newX = Math.max(0, Math.min(11, Math.round((e.clientX - dragOffset.current.x) / cellW)));
    const newY = Math.max(0, Math.round((e.clientY - dragOffset.current.y) / cellH));
    setLayout(prev => prev.map(w => w.id === dragging ? { ...w, x: newX, y: newY } : w));
  }, [dragging]);

  const handleMouseUp = useCallback(() => {
    setDragging(null);
    setResizing(null);
  }, []);

  useEffect(() => {
    if (dragging || resizing) {
      window.addEventListener('mousemove', handleMouseMove);
      window.addEventListener('mouseup', handleMouseUp);
      return () => {
        window.removeEventListener('mousemove', handleMouseMove);
        window.removeEventListener('mouseup', handleMouseUp);
      };
    }
  }, [dragging, resizing, handleMouseMove, handleMouseUp]);

  /* ─── Activity scroll on hover ─── */
  useEffect(() => {
    if (!hoverActivity) return;
    const interval = setInterval(() => {
      setActivityScroll(prev => (prev + 1) % mockRecentActivity.length);
    }, 2000);
    return () => clearInterval(interval);
  }, [hoverActivity]);

  const stats = useMemo(() => {
    const totalRooms = rooms.length;
    const occupied = rooms.filter(r => r.status === 'occupied').length;
    const available = rooms.filter(r => r.status === 'vacant_clean').length;
    const blocked = rooms.filter(r => r.status === 'blocked').length;
    const maintenance = rooms.filter(r => r.status === 'maintenance').length;
    const arrivalsToday = reservations.filter(r => r.check_in === today && r.status !== 'checked_in').length;
    const departuresToday = reservations.filter(r => r.check_out === today && r.status === 'checked_in').length;
    const inHouse = reservations.filter(r => r.status === 'checked_in').length;
    const vipArrivals = reservations.filter(r => r.check_in === today && r.vip).length;
    const occupancyRate = totalRooms ? Math.round((occupied / totalRooms) * 100) : 0;
    return { totalRooms, occupied, available, blocked, maintenance, arrivalsToday, departuresToday, inHouse, vipArrivals, occupancyRate, revenueToday: 8432, revenueYesterday: 7650, pendingBalance: reservations.reduce((sum, r) => sum + (r.balance || 0), 0) };
  }, [rooms, reservations]);

  const arrivals = useMemo(() => reservations.filter(r => r.check_in === today && r.status !== 'checked_in').map(r => ({ id: r.id, guest: r.guest_name, room: r.room_number || '', type: r.room_type, status: r.status, vip: r.vip, source: r.source })), [reservations]);
  const departures = useMemo(() => reservations.filter(r => r.check_out === today && r.status === 'checked_in').map(r => ({ id: r.id, guest: r.guest_name, room: r.room_number || '', type: r.room_type, status: r.status, balance: r.balance, vip: r.vip })), [reservations]);

  if (loading) return <div className="flex h-96 items-center justify-center"><RefreshCw className="h-8 w-8 animate-spin text-nexus-400" /></div>;
  if (error) return <div className="flex h-96 flex-col items-center justify-center gap-4"><AlertTriangle className="h-10 w-10 text-rose-400" /><p className="text-rose-400">{error}</p><button onClick={fetchData} className="btn-primary">Retry</button></div>;

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Dashboard</h1>
          <p className="text-sm text-slate-400">{new Date(today + 'T00:00:00').toLocaleDateString('en-US', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })}</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={fetchData} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white" title="Refresh"><RefreshCw className="h-4 w-4" /></button>
          <Link href="/reservations/new" className="btn-primary"><Plus className="h-4 w-4" />New Reservation</Link>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <StatCard label="Occupancy" value={`${stats.occupancyRate}%`} sub={`${stats.occupied}/${stats.totalRooms} rooms`} icon={BedDouble} color="sky" />
        <StatCard label="In-House" value={stats.inHouse} sub={`${stats.arrivalsToday} arriving · ${stats.departuresToday} departing`} icon={Users} color="emerald" />
        <StatCard label="Available" value={stats.available} sub={`${stats.blocked} blocked · ${stats.maintenance} maintenance`} icon={DoorOpen} color="violet" />
        <StatCard label="Revenue Today" value={`$${stats.revenueToday.toLocaleString()}`} sub={stats.revenueToday > stats.revenueYesterday ? '↑ from yesterday' : '↓ from yesterday'} icon={DollarSign} color="amber" trend={stats.revenueToday > stats.revenueYesterday ? 'up' : 'down'} />
      </div>

      {/* Draggable Grid Dashboard */}
      <div ref={containerRef} className="relative" style={{ minHeight: '800px' }}>
        {layout.map(widget => {
          const style: React.CSSProperties = {
            position: 'absolute',
            left: `${(widget.x / 12) * 100}%`,
            top: `${widget.y * 80}px`,
            width: `${(widget.w / 12) * 100}%`,
            height: `${widget.h * 80}px`,
            transition: dragging === widget.id ? 'none' : 'all 0.2s ease',
            zIndex: dragging === widget.id ? 50 : 10,
          };

          return (
            <div key={widget.id} style={style} className={`p-2 ${dragging === widget.id ? 'cursor-grabbing' : 'cursor-grab'}`}>
              <div className="card h-full flex flex-col">
                {/* Widget Header with Drag Handle */}
                <div
                  className="flex items-center justify-between mb-3 pb-2 border-b border-slate-700/50 select-none"
                  onMouseDown={e => handleMouseDown(e, widget.id)}
                >
                  <div className="flex items-center gap-2 cursor-grab">
                    <GripVertical className="h-4 w-4 text-slate-600" />
                    <h3 className="font-bold text-sm text-white">{widget.title}</h3>
                  </div>
                  {widget.id === 'activity' && (
                    <div className="flex gap-1">
                      <button className="text-[10px] rounded px-2 py-0.5 bg-slate-700 text-slate-400 hover:text-white">All</button>
                      <button className="text-[10px] rounded px-2 py-0.5 bg-nexus-500/20 text-nexus-400">Alerts</button>
                    </div>
                  )}
                </div>

                {/* Widget Content */}
                <div className="flex-1 overflow-hidden">
                  {widget.id === 'arrivals' && <ArrivalsWidget arrivals={arrivals} />}
                  {widget.id === 'departures' && <DeparturesWidget departures={departures} />}
                  {widget.id === 'activity' && <ActivityWidget scrollIndex={activityScroll} onHover={setHoverActivity} />}
                  {widget.id === 'wifi' && <WiFiWidget floors={mockWiFiFloors} />}
                  {widget.id === 'revenue' && <RevenueWidget />}
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

/* ─── Widget Components ─── */
function ArrivalsWidget({ arrivals }: { arrivals: any[] }) {
  return (
    <div className="space-y-2 overflow-y-auto h-full pr-1" style={{ scrollbarWidth: 'thin' }}>
      {arrivals.length === 0 && <p className="text-sm text-slate-500">No arrivals today.</p>}
      {arrivals.map(a => (
        <div key={a.id} className="flex items-center justify-between rounded-lg border border-slate-700/50 bg-slate-800/50 p-2.5">
          <div className="flex items-center gap-2.5">
            <div className="flex h-7 w-7 items-center justify-center rounded-full bg-nexus-500/10">
              {a.vip ? <Star className="h-3.5 w-3.5 text-amber-400" /> : <User className="h-3.5 w-3.5 text-nexus-400" />}
            </div>
            <div>
              <p className="text-sm font-medium text-white">{a.guest}</p>
              <p className="text-[11px] text-slate-400">{a.room ? `Room ${a.room} · ${a.type}` : a.type}</p>
            </div>
          </div>
          <span className={`text-[10px] px-1.5 py-0.5 rounded ${a.source === 'ota' ? 'bg-violet-500/10 text-violet-400' : a.source === 'walk_in' ? 'bg-amber-500/10 text-amber-400' : 'bg-sky-500/10 text-sky-400'}`}>{a.source}</span>
        </div>
      ))}
    </div>
  );
}

function DeparturesWidget({ departures }: { departures: any[] }) {
  return (
    <div className="space-y-2 overflow-y-auto h-full pr-1" style={{ scrollbarWidth: 'thin' }}>
      {departures.length === 0 && <p className="text-sm text-slate-500">No departures today.</p>}
      {departures.map(d => (
        <div key={d.id} className="flex items-center justify-between rounded-lg border border-slate-700/50 bg-slate-800/50 p-2.5">
          <div className="flex items-center gap-2.5">
            <div className="flex h-7 w-7 items-center justify-center rounded-full bg-rose-500/10">
              {d.vip ? <Star className="h-3.5 w-3.5 text-amber-400" /> : <DoorOpen className="h-3.5 w-3.5 text-rose-400" />}
            </div>
            <div>
              <p className="text-sm font-medium text-white">{d.guest}</p>
              <p className="text-[11px] text-slate-400">Room {d.room} · {d.type}</p>
            </div>
          </div>
          <span className={`text-[10px] px-1.5 py-0.5 rounded ${d.balance > 0 ? 'bg-rose-500/10 text-rose-400' : 'bg-emerald-500/10 text-emerald-400'}`}>{d.balance > 0 ? `$${d.balance} due` : 'Paid'}</span>
        </div>
      ))}
    </div>
  );
}

function ActivityWidget({ scrollIndex, onHover }: { scrollIndex: number; onHover: (v: boolean) => void }) {
  const [filter, setFilter] = useState<string>('all');
  const scrollRef = useRef<HTMLDivElement>(null);

  const filtered = useMemo(() => {
    if (filter === 'all') return mockRecentActivity;
    return mockRecentActivity.filter(a => {
      if (filter === 'housekeeping') return a.type === 'housekeeping' || a.type === 'minibar';
      if (filter === 'roomservice') return a.type === 'roomservice';
      if (filter === 'checkout') return a.type === 'checkout';
      return a.type === filter;
    });
  }, [filter]);

  return (
    <div className="flex flex-col h-full" onMouseEnter={() => onHover(true)} onMouseLeave={() => onHover(false)}>
      {/* Activity Filters */}
      <div className="flex gap-1 mb-2 flex-wrap">
        {['all', 'checkin', 'checkout', 'housekeeping', 'roomservice', 'alert'].map(f => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`text-[10px] rounded px-2 py-0.5 capitalize transition-colors ${
              filter === f ? 'bg-nexus-500/20 text-nexus-400' : 'bg-slate-700/50 text-slate-500 hover:text-slate-300'
            }`}
          >
            {f.replace('_', ' ')}
          </button>
        ))}
      </div>

      {/* Scrollable Activity Feed */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto pr-1 space-y-2" style={{ scrollbarWidth: 'thin' }}>
        {filtered.map((a, i) => (
          <Link
            key={a.id}
            href={
              a.type === 'housekeeping' || a.type === 'minibar' ? '/housekeeping' :
              a.type === 'roomservice' ? '/iptv' :
              a.type === 'checkout' ? `/reservations` :
              '#'
            }
            className="flex items-start gap-2.5 p-2 rounded-lg hover:bg-slate-700/30 transition-colors group"
          >
            <div className={`mt-0.5 rounded-full p-1 bg-${a.color}-500/10`}>
              <a.icon className={`h-3 w-3 text-${a.color}-400`} />
            </div>
            <div className="flex-1 min-w-0">
              <div className="flex items-center justify-between">
                <p className="text-sm text-white">{a.event}</p>
                <p className="text-[10px] text-slate-500">{a.time}</p>
              </div>
              <p className="text-xs text-slate-400 truncate">{a.detail}</p>
              {a.room && (
                <div className="flex items-center gap-2 mt-1">
                  <span className="text-[10px] rounded bg-slate-700 px-1.5 py-0.5 text-slate-400">Room {a.room}</span>
                  {a.type === 'minibar' && <span className="text-[10px] text-violet-400">→ Housekeeping</span>}
                  {a.type === 'roomservice' && <span className="text-[10px] text-amber-400">→ IPTV Room Service</span>}
                </div>
              )}
            </div>
            <ChevronRight className="h-3 w-3 text-slate-600 opacity-0 group-hover:opacity-100 transition-opacity" />
          </Link>
        ))}
      </div>
    </div>
  );
}

function WiFiWidget({ floors }: { floors: WiFiFloorStatus[] }) {
  return (
    <div className="space-y-2 overflow-y-auto h-full pr-1" style={{ scrollbarWidth: 'thin' }}>
      {floors.map(floor => {
        const quality = signalQuality(floor.avgSignal);
        return (
          <div key={floor.floor} className="rounded-lg border border-slate-700/50 bg-slate-800/50 p-2.5">
            <div className="flex items-center justify-between mb-1.5">
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium text-white">{floor.floor}</span>
                {floor.issues.length > 0 && (
                  <span className="text-[10px] rounded-full bg-rose-500/10 px-1.5 py-0.5 text-rose-400">
                    {floor.issues.length} issue{floor.issues.length > 1 ? 's' : ''}
                  </span>
                )}
              </div>
              <div className="flex items-center gap-1">
                {[1, 2, 3, 4].map(bar => (
                  <div key={bar} className={`w-1 rounded-sm ${bar <= quality.bars ? 'bg-emerald-400' : 'bg-slate-700'}`} style={{ height: `${bar * 3}px` }} />
                ))}
              </div>
            </div>
            <div className="grid grid-cols-3 gap-2 text-[11px]">
              <div>
                <p className="text-slate-500">Signal</p>
                <p className={quality.color}>{floor.avgSignal} dBm</p>
              </div>
              <div>
                <p className="text-slate-500">APs</p>
                <p className={floor.onlineAPs === floor.apCount ? 'text-emerald-400' : 'text-amber-400'}>
                  {floor.onlineAPs}/{floor.apCount} online
                </p>
              </div>
              <div>
                <p className="text-slate-500">Clients</p>
                <p className="text-sky-400">{floor.clientCount}</p>
              </div>
            </div>
            {floor.issues.length > 0 && (
              <div className="mt-1.5 space-y-0.5">
                {floor.issues.map((issue, i) => (
                  <p key={i} className="text-[10px] text-rose-400 flex items-center gap-1">
                    <AlertTriangle className="h-2.5 w-2.5" />{issue}
                  </p>
                ))}
              </div>
            )}
          </div>
        );
      })}
    </div>
  );
}

function RevenueWidget() {
  const mockWeeklyRevenue = [
    { day: 'Mon', revenue: 6200, occupancy: 65 },
    { day: 'Tue', revenue: 7100, occupancy: 72 },
    { day: 'Wed', revenue: 6800, occupancy: 68 },
    { day: 'Thu', revenue: 7500, occupancy: 75 },
    { day: 'Fri', revenue: 8900, occupancy: 85 },
    { day: 'Sat', revenue: 9200, occupancy: 90 },
    { day: 'Sun', revenue: 8432, occupancy: 73 },
  ];
  return (
    <div className="flex items-end gap-2 h-full pb-2">
      {mockWeeklyRevenue.map(d => (
        <div key={d.day} className="flex flex-1 flex-col items-center gap-1">
          <div className="w-full rounded-t bg-emerald-500/20 relative" style={{ height: `${(d.revenue / 10000) * 120}px` }}>
            <div className="absolute inset-0 rounded-t bg-emerald-500/40" />
            <div className="absolute -top-4 left-1/2 -translate-x-1/2 text-[9px] text-emerald-400 whitespace-nowrap opacity-0 hover:opacity-100 transition-opacity">
              ${d.revenue.toLocaleString()}
            </div>
          </div>
          <span className="text-[10px] text-slate-400">{d.day}</span>
        </div>
      ))}
    </div>
  );
}

function StatCard({ label, value, sub, icon: Icon, color, trend }: { label: string; value: string | number; sub: string; icon: any; color: string; trend?: 'up' | 'down' }) {
  const colorMap: Record<string, string> = { sky: 'bg-sky-500/10 text-sky-400', emerald: 'bg-emerald-500/10 text-emerald-400', violet: 'bg-violet-500/10 text-violet-400', amber: 'bg-amber-500/10 text-amber-400', rose: 'bg-rose-500/10 text-rose-400' };
  return (
    <div className="rounded-lg border border-slate-700 bg-slate-800 p-4">
      <div className="flex items-center justify-between">
        <div className={`rounded-md p-1.5 ${colorMap[color] || ''}`}><Icon className="h-5 w-5" /></div>
        {trend && (trend === 'up' ? <TrendingUp className="h-4 w-4 text-emerald-400" /> : <TrendingDown className="h-4 w-4 text-rose-400" />)}
      </div>
      <div className="mt-2">
        <div className="text-2xl font-bold text-white">{value}</div>
        <div className="text-xs text-slate-400">{sub}</div>
      </div>
    </div>
  );
}
