"use client";

import { useState, useEffect } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import {
  Tv,
  Plus,
  Play,
  Pause,
  Pencil,
  Trash2,
  Search,
  Monitor,
  Wifi,
  WifiOff,
  LayoutGrid,
  List,
  Eye,
  Settings,
} from "lucide-react";
import { getIPTVChannels, getIPTVContent, getIPTVRooms, IPTVChannel, IPTVContent, IPTVRoomStatus } from "@/lib/api";

function ToggleSwitch({ checked }: { checked: boolean }) {
  return (
    <div className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors ${checked ? "bg-nexus-500" : "bg-slate-700"}`}>
      <span className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform ${checked ? "translate-x-4" : "translate-x-1"}`} />
    </div>
  );
}

type WelcomeConfig = {
  headline: string;
  subheadline: string;
  wifiName: string;
  wifiPassword: string;
  emergencyPhone: string;
  checkInTime: string;
  checkOutTime: string;
};

export default function IPTVPage() {
  const { config } = useTenant();
  const [tab, setTab] = useState<"channels" | "content" | "rooms" | "analytics" | "welcome">("channels");
  const [search, setSearch] = useState("");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");
  const [channels, setChannels] = useState<IPTVChannel[]>([]);
  const [content, setContent] = useState<IPTVContent[]>([]);
  const [roomStatus, setRoomStatus] = useState<IPTVRoomStatus[]>([]);
  const [loading, setLoading] = useState(false);
  const [welcomeConfig, setWelcomeConfig] = useState<WelcomeConfig>({
    headline: "Welcome to Your Villa",
    subheadline: "Your luxury retreat awaits",
    wifiName: "VillaGuest",
    wifiPassword: "Welcome2024",
    emergencyPhone: "+970-599-123-456",
    checkInTime: "15:00",
    checkOutTime: "11:00",
  });

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    setLoading(true);
    try {
      const [ch, ct, rs] = await Promise.all([
        getIPTVChannels(),
        getIPTVContent(),
        getIPTVRooms(),
      ]);
      setChannels(ch);
      setContent(ct);
      setRoomStatus(rs);
    } catch (e) {
      console.error("Failed to load IPTV data", e);
    } finally {
      setLoading(false);
    }
  }

  const hasIPTV = hasCapability(config, CAPABILITIES.OPERATIONS.IPTV_BASIC);

  if (!hasIPTV) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-800">
          <Tv className="h-8 w-8 text-slate-500" />
        </div>
        <h2 className="mt-6 text-xl font-bold text-white">IPTV Module</h2>
        <p className="mt-2 max-w-md text-sm text-slate-400">
          Upgrade to Operations or Enterprise tier to manage in-room entertainment.
        </p>
      </div>
    );
  }

  const filteredChannels = channels.filter((c) =>
    c.name.toLowerCase().includes(search.toLowerCase()) ||
    c.category.toLowerCase().includes(search.toLowerCase())
  );

  const filteredContent = content.filter((c) =>
    c.title.toLowerCase().includes(search.toLowerCase()) ||
    c.category.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">IPTV Management</h1>
          <p className="text-sm text-slate-400">Channels, content library, and room bindings</p>
        </div>
        <div className="flex items-center gap-2">
          <button className="btn-secondary gap-2 text-xs">
            <Settings className="h-4 w-4" />
            Settings
          </button>
          <button className="btn-primary gap-2 text-xs">
            <Plus className="h-4 w-4" />
            Add Channel
          </button>
        </div>
      </div>

      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        {(["channels", "content", "rooms", "welcome", "analytics"] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              tab === t ? "bg-slate-800 text-white" : "text-slate-400 hover:bg-slate-800 hover:text-white"
            }`}
          >
            {t === "welcome" ? "Villa Welcome" : t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>

      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
          <input
            placeholder={`Search ${tab}...`}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="input w-full pl-10"
          />
        </div>
        {tab !== "analytics" && (
          <div className="flex gap-1">
            <button
              onClick={() => setViewMode("grid")}
              className={`rounded p-2 ${viewMode === "grid" ? "bg-slate-700 text-white" : "text-slate-500 hover:text-white"}`}
            >
              <LayoutGrid className="h-4 w-4" />
            </button>
            <button
              onClick={() => setViewMode("list")}
              className={`rounded p-2 ${viewMode === "list" ? "bg-slate-700 text-white" : "text-slate-500 hover:text-white"}`}
            >
              <List className="h-4 w-4" />
            </button>
          </div>
        )}
      </div>

      {loading ? (
        <p className="text-sm text-slate-500">Loading IPTV data...</p>
      ) : (
        <>
          {tab === "channels" && (
            <div className={viewMode === "grid" ? "grid grid-cols-4 gap-4" : "space-y-2"}>
              {filteredChannels.map((ch) => (
                <ChannelCard key={ch.id} channel={ch} viewMode={viewMode} />
              ))}
            </div>
          )}

          {tab === "content" && (
            <div className={viewMode === "grid" ? "grid grid-cols-3 gap-4" : "space-y-2"}>
              {filteredContent.map((item) => (
                <ContentCard key={item.id} item={item} viewMode={viewMode} />
              ))}
            </div>
          )}

          {tab === "rooms" && <RoomBindingsTab rooms={roomStatus} />}

          {tab === "welcome" && (
            <VillaWelcomeTab config={welcomeConfig} onChange={setWelcomeConfig} />
          )}

          {tab === "analytics" && <AnalyticsTab />}
        </>
      )}
    </div>
  );
}

function ChannelCard({ channel, viewMode }: { channel: IPTVChannel; viewMode: string }) {
  if (viewMode === "list") {
    return (
      <div className="flex items-center justify-between rounded-lg border border-slate-700 bg-slate-800/50 p-3">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded bg-slate-700 text-sm font-bold text-white">
            {channel.number}
          </div>
          <div>
            <p className="text-sm font-medium text-white">{channel.name}</p>
            <p className="text-xs text-slate-400">{channel.category} · {channel.language}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          {channel.is_premium && (
            <span className="rounded-full bg-amber-500/10 px-2 py-0.5 text-[10px] font-medium text-amber-400">Premium</span>
          )}
          <ToggleSwitch checked={channel.is_active} />
          <button className="rounded p-1 text-slate-400 hover:text-white"><Pencil className="h-3 w-3" /></button>
          <button className="rounded p-1 text-rose-400 hover:text-rose-300"><Trash2 className="h-3 w-3" /></button>
        </div>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-4">
      <div className="flex items-center justify-between">
        <div className="flex h-8 w-8 items-center justify-center rounded bg-nexus-500/10 text-sm font-bold text-nexus-400">
          {channel.number}
        </div>
        {channel.is_premium && (
          <span className="rounded-full bg-amber-500/10 px-2 py-0.5 text-[10px] font-medium text-amber-400">Premium</span>
        )}
      </div>
      <h3 className="mt-2 text-sm font-medium text-white">{channel.name}</h3>
      <p className="text-xs text-slate-400">{channel.category} · {channel.language}</p>
      <div className="mt-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <ToggleSwitch checked={channel.is_active} />
          <span className="text-xs text-slate-400">{channel.is_active ? "Live" : "Offline"}</span>
        </div>
        <div className="flex gap-1">
          <button className="rounded p-1 text-slate-400 hover:text-white"><Pencil className="h-3 w-3" /></button>
          <button className="rounded p-1 text-rose-400 hover:text-rose-300"><Trash2 className="h-3 w-3" /></button>
        </div>
      </div>
    </div>
  );
}

function ContentCard({ item, viewMode }: { item: IPTVContent; viewMode: string }) {
  if (viewMode === "list") {
    return (
      <div className="flex items-center justify-between rounded-lg border border-slate-700 bg-slate-800/50 p-3">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded bg-slate-700">
            {item.type === "movie" ? <Play className="h-4 w-4 text-nexus-400" /> : item.type === "series" ? <Tv className="h-4 w-4 text-violet-400" /> : <Eye className="h-4 w-4 text-sky-400" />}
          </div>
          <div>
            <p className="text-sm font-medium text-white">{item.title}</p>
            <p className="text-xs text-slate-400">{item.type} · {item.category} {item.duration && `· ${item.duration}min`}</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <ToggleSwitch checked={item.is_active} />
          <button className="rounded p-1 text-slate-400 hover:text-white"><Pencil className="h-3 w-3" /></button>
          <button className="rounded p-1 text-rose-400 hover:text-rose-300"><Trash2 className="h-3 w-3" /></button>
        </div>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-4">
      <div className="flex h-24 items-center justify-center rounded bg-slate-700/50">
        {item.type === "movie" ? <Play className="h-8 w-8 text-nexus-400" /> : item.type === "series" ? <Tv className="h-8 w-8 text-violet-400" /> : <Eye className="h-8 w-8 text-sky-400" />}
      </div>
      <h3 className="mt-2 text-sm font-medium text-white">{item.title}</h3>
      <p className="text-xs text-slate-400">{item.type} · {item.category}</p>
      <div className="mt-3 flex items-center justify-between">
        <ToggleSwitch checked={item.is_active} />
        <div className="flex gap-1">
          <button className="rounded p-1 text-slate-400 hover:text-white"><Pencil className="h-3 w-3" /></button>
          <button className="rounded p-1 text-rose-400 hover:text-rose-300"><Trash2 className="h-3 w-3" /></button>
        </div>
      </div>
    </div>
  );
}

function RoomBindingsTab({ rooms }: { rooms: IPTVRoomStatus[] }) {
  return (
    <div className="space-y-3">
      {rooms.length === 0 ? (
        <p className="text-sm text-slate-500">No IPTV room data available.</p>
      ) : (
        rooms.map((r) => (
          <div key={r.id} className="flex items-center justify-between rounded-lg border border-slate-700 bg-slate-800/50 p-4">
            <div className="flex items-center gap-4">
              <div className="flex h-10 w-10 items-center justify-center rounded bg-slate-700 text-sm font-bold text-white">
                {r.room_number}
              </div>
              <div>
                <p className="text-sm font-medium text-white">
                  {r.is_online ? <span className="text-emerald-400">Online</span> : <span className="text-rose-400">Offline</span>}
                </p>
                <p className="text-xs text-slate-400">
                  {r.current_channel ? `Channel ${r.current_channel}` : "No active channel"}
                  {r.last_activity_at && ` · Last active: ${r.last_activity_at.split("T")[1]?.substring(0,5) || ""}`}
                </p>
              </div>
            </div>
            <div className="flex items-center gap-3">
              <button className="btn-secondary text-xs">Edit</button>
              <ToggleSwitch checked={r.is_online} />
            </div>
          </div>
        ))
      )}
    </div>
  );
}

function AnalyticsTab() {
  return (
    <div className="grid grid-cols-3 gap-4">
      <div className="card space-y-2">
        <p className="text-sm text-slate-400">Active Viewers</p>
        <p className="text-3xl font-bold text-white">24</p>
        <p className="text-xs text-emerald-400">+3 from yesterday</p>
      </div>
      <div className="card space-y-2">
        <p className="text-sm text-slate-400">Most Watched</p>
        <p className="text-lg font-bold text-white">HBO</p>
        <p className="text-xs text-slate-400">12 viewers · 4.2 hrs avg</p>
      </div>
      <div className="card space-y-2">
        <p className="text-sm text-slate-400">Content Requests</p>
        <p className="text-3xl font-bold text-white">8</p>
        <p className="text-xs text-slate-400">This week</p>
      </div>
      <div className="card col-span-3 space-y-2">
        <p className="text-sm text-slate-400">Hourly Viewership</p>
        <div className="flex items-end gap-1 h-32">
          {[2,4,3,6,8,12,15,18,22,20,16,14,10,8,6,5,7,9,14,18,20,16,10,5].map((v, i) => (
            <div key={i} className="flex-1 rounded-t bg-nexus-500/20 hover:bg-nexus-500/40 transition-colors" style={{ height: `${(v/22)*100}%` }} title={`${i}:00 - ${v} viewers`} />
          ))}
        </div>
        <div className="mt-2 flex justify-between text-[10px] text-slate-500">
          <span>00:00</span><span>06:00</span><span>12:00</span><span>18:00</span><span>23:00</span>
        </div>
      </div>
    </div>
  );
}

function VillaWelcomeTab({ config, onChange }: { config: WelcomeConfig; onChange: (c: WelcomeConfig) => void }) {
  return (
    <div className="grid grid-cols-2 gap-6">
      <div className="space-y-4">
        <h3 className="text-lg font-bold text-white">Villa Welcome Screen</h3>
        <p className="text-sm text-slate-400">Customize what guests see when they turn on the TV.</p>

        <div className="space-y-3">
          <div>
            <label className="mb-1 block text-xs text-slate-400">Headline</label>
            <input
              className="input w-full"
              value={config.headline}
              onChange={(e) => onChange({ ...config, headline: e.target.value })}
            />
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Subheadline</label>
            <input
              className="input w-full"
              value={config.subheadline}
              onChange={(e) => onChange({ ...config, subheadline: e.target.value })}
            />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">WiFi Name</label>
              <input
                className="input w-full"
                value={config.wifiName}
                onChange={(e) => onChange({ ...config, wifiName: e.target.value })}
              />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">WiFi Password</label>
              <input
                className="input w-full"
                value={config.wifiPassword}
                onChange={(e) => onChange({ ...config, wifiPassword: e.target.value })}
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Check-in Time</label>
              <input
                className="input w-full"
                value={config.checkInTime}
                onChange={(e) => onChange({ ...config, checkInTime: e.target.value })}
              />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Check-out Time</label>
              <input
                className="input w-full"
                value={config.checkOutTime}
                onChange={(e) => onChange({ ...config, checkOutTime: e.target.value })}
              />
            </div>
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Emergency Phone</label>
            <input
              className="input w-full"
              value={config.emergencyPhone}
              onChange={(e) => onChange({ ...config, emergencyPhone: e.target.value })}
            />
          </div>
          <button className="btn-primary w-full">Save Welcome Screen</button>
        </div>
      </div>

      <div className="rounded-xl border border-slate-700 bg-slate-900 p-6 space-y-4">
        <p className="text-xs font-medium text-slate-500 uppercase">Preview</p>
        <div className="rounded-lg bg-slate-800 p-6 text-center space-y-4">
          <h2 className="text-2xl font-bold text-white">{config.headline}</h2>
          <p className="text-sm text-slate-400">{config.subheadline}</p>
          <div className="rounded-lg bg-slate-700/50 p-4 space-y-2">
            <p className="text-xs font-medium text-slate-400">WiFi Access</p>
            <div className="flex items-center justify-center gap-4">
              <div>
                <p className="text-[10px] text-slate-500">Network</p>
                <p className="text-sm font-medium text-white">{config.wifiName}</p>
              </div>
              <div>
                <p className="text-[10px] text-slate-500">Password</p>
                <p className="text-sm font-medium text-white">{config.wifiPassword}</p>
              </div>
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div className="rounded bg-slate-700/50 p-2">
              <p className="text-[10px] text-slate-500">Check-in</p>
              <p className="text-sm font-medium text-white">{config.checkInTime}</p>
            </div>
            <div className="rounded bg-slate-700/50 p-2">
              <p className="text-[10px] text-slate-500">Check-out</p>
              <p className="text-sm font-medium text-white">{config.checkOutTime}</p>
            </div>
          </div>
          <div className="rounded bg-emerald-500/10 p-2">
            <p className="text-[10px] text-emerald-400">Emergency: {config.emergencyPhone}</p>
          </div>
        </div>
      </div>
    </div>
  );
}
