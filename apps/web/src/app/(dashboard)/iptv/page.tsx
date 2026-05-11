"use client";

import { useState } from "react";
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

interface Channel {
  id: string;
  name: string;
  number: number;
  streamUrl: string;
  logoUrl?: string;
  category: string;
  language: string;
  isActive: boolean;
  isPremium: boolean;
}

interface ContentItem {
  id: string;
  title: string;
  type: "movie" | "series" | "music" | "info";
  description: string;
  duration?: number;
  thumbnailUrl?: string;
  category: string;
  isActive: boolean;
}

const mockChannels: Channel[] = [
  { id: "c1", name: "CNN International", number: 1, streamUrl: "https://cnn-intl/stream", category: "news", language: "en", isActive: true, isPremium: false },
  { id: "c2", name: "BBC World", number: 2, streamUrl: "https://bbc-world/stream", category: "news", language: "en", isActive: true, isPremium: false },
  { id: "c3", name: "Al Jazeera", number: 3, streamUrl: "https://aljazeera/stream", category: "news", language: "en", isActive: true, isPremium: false },
  { id: "c4", name: "ESPN", number: 10, streamUrl: "https://espn/stream", category: "sports", language: "en", isActive: true, isPremium: true },
  { id: "c5", name: "BeIN Sports", number: 11, streamUrl: "https://bein/stream", category: "sports", language: "en", isActive: true, isPremium: true },
  { id: "c6", name: "HBO", number: 20, streamUrl: "https://hbo/stream", category: "movies", language: "en", isActive: true, isPremium: true },
  { id: "c7", name: "Netflix Channel", number: 21, streamUrl: "https://netflix-channel/stream", category: "movies", language: "en", isActive: true, isPremium: true },
  { id: "c8", name: "Kids TV", number: 50, streamUrl: "https://kids-tv/stream", category: "kids", language: "en", isActive: true, isPremium: false },
  { id: "c9", name: "Cartoon Network", number: 51, streamUrl: "https://cn/stream", category: "kids", language: "en", isActive: true, isPremium: false },
  { id: "c10", name: "Music FM", number: 80, streamUrl: "https://music-fm/stream", category: "music", language: "en", isActive: true, isPremium: false },
];

const mockContent: ContentItem[] = [
  { id: "m1", title: "Welcome to Grand Plaza", type: "info", description: "Hotel overview, amenities guide, and local attractions", category: "welcome", isActive: true },
  { id: "m2", title: "Room Service Menu", type: "info", description: "Full dining menu with ordering instructions", category: "dining", isActive: true },
  { id: "m3", title: "Spa & Wellness Guide", type: "info", description: "Spa treatments, gym hours, and booking info", category: "wellness", isActive: true },
  { id: "m4", title: "The Dark Knight", type: "movie", description: "Batman faces the Joker in Gotham City", duration: 152, category: "action", isActive: true },
  { id: "m5", title: "Inception", type: "movie", description: "A thief who steals corporate secrets through dream-sharing technology", duration: 148, category: "sci-fi", isActive: true },
  { id: "m6", title: "Breaking Bad S1", type: "series", description: "A high school chemistry teacher turned methamphetamine manufacturer", category: "drama", isActive: true },
  { id: "m7", title: "Local Attractions", type: "info", description: "Top 10 places to visit within 5km of the hotel", category: "tourism", isActive: true },
];

function ToggleSwitch({ checked }: { checked: boolean }) {
  return (
    <div className={`relative inline-flex h-5 w-9 items-center rounded-full transition-colors ${checked ? "bg-nexus-500" : "bg-slate-700"}`}>
      <span className={`inline-block h-3.5 w-3.5 transform rounded-full bg-white transition-transform ${checked ? "translate-x-4" : "translate-x-1"}`} />
    </div>
  );
}

export default function IPTVPage() {
  const { config } = useTenant();
  const [tab, setTab] = useState<"channels" | "content" | "rooms" | "analytics">("channels");
  const [search, setSearch] = useState("");
  const [viewMode, setViewMode] = useState<"grid" | "list">("grid");

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

  const filteredChannels = mockChannels.filter((c) =>
    c.name.toLowerCase().includes(search.toLowerCase()) ||
    c.category.toLowerCase().includes(search.toLowerCase())
  );

  const filteredContent = mockContent.filter((c) =>
    c.title.toLowerCase().includes(search.toLowerCase()) ||
    c.category.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-6">
      {/* Header */}
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

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        {(["channels", "content", "rooms", "analytics"] as const).map((t) => (
          <button
            key={t}
            onClick={() => setTab(t)}
            className={`rounded-lg px-4 py-2 text-sm font-medium transition-colors ${
              tab === t ? "bg-slate-800 text-white" : "text-slate-400 hover:bg-slate-800 hover:text-white"
            }`}
          >
            {t.charAt(0).toUpperCase() + t.slice(1)}
          </button>
        ))}
      </div>

      {/* Search bar */}
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

      {/* Channels Tab */}
      {tab === "channels" && (
        <div className={viewMode === "grid" ? "grid grid-cols-4 gap-4" : "space-y-2"}>
          {filteredChannels.map((ch) => (
            <ChannelCard key={ch.id} channel={ch} viewMode={viewMode} />
          ))}
        </div>
      )}

      {/* Content Tab */}
      {tab === "content" && (
        <div className={viewMode === "grid" ? "grid grid-cols-3 gap-4" : "space-y-2"}>
          {filteredContent.map((item) => (
            <ContentCard key={item.id} item={item} viewMode={viewMode} />
          ))}
        </div>
      )}

      {/* Rooms Tab */}
      {tab === "rooms" && <RoomBindingsTab />}

      {/* Analytics Tab */}
      {tab === "analytics" && <AnalyticsTab />}
    </div>
  );
}

function ChannelCard({ channel, viewMode }: { channel: Channel; viewMode: string }) {
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
          {channel.isPremium && (
            <span className="rounded-full bg-amber-500/10 px-2 py-0.5 text-[10px] font-medium text-amber-400">Premium</span>
          )}
          <ToggleSwitch checked={channel.isActive} />
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
        {channel.isPremium && (
          <span className="rounded-full bg-amber-500/10 px-2 py-0.5 text-[10px] font-medium text-amber-400">Premium</span>
        )}
      </div>
      <h3 className="mt-2 text-sm font-medium text-white">{channel.name}</h3>
      <p className="text-xs text-slate-400">{channel.category} · {channel.language}</p>
      <div className="mt-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <ToggleSwitch checked={channel.isActive} />
          <span className="text-xs text-slate-400">{channel.isActive ? "Live" : "Offline"}</span>
        </div>
        <div className="flex gap-1">
          <button className="rounded p-1 text-slate-400 hover:text-white"><Pencil className="h-3 w-3" /></button>
          <button className="rounded p-1 text-rose-400 hover:text-rose-300"><Trash2 className="h-3 w-3" /></button>
        </div>
      </div>
    </div>
  );
}

function ContentCard({ item, viewMode }: { item: ContentItem; viewMode: string }) {
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
          <ToggleSwitch checked={item.isActive} />
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
        <ToggleSwitch checked={item.isActive} />
        <div className="flex gap-1">
          <button className="rounded p-1 text-slate-400 hover:text-white"><Pencil className="h-3 w-3" /></button>
          <button className="rounded p-1 text-rose-400 hover:text-rose-300"><Trash2 className="h-3 w-3" /></button>
        </div>
      </div>
    </div>
  );
}

function RoomBindingsTab() {
  const rooms = [
    { room: "101", guest: "Alice Chen", channel: "HBO, ESPN, CNN", language: "en", welcome: "Welcome Alice!" },
    { room: "102", guest: null, channel: "Basic Package", language: "en", welcome: "Welcome to Grand Plaza" },
    { room: "103", guest: null, channel: "Basic Package", language: "en", welcome: "Welcome to Grand Plaza" },
    { room: "201", guest: "Bob Jones", channel: "Premium Package", language: "en", welcome: "Welcome back, Bob!" },
    { room: "202", guest: null, channel: "Basic Package", language: "en", welcome: "Welcome to Grand Plaza" },
    { room: "203", guest: "Carol White", channel: "Premium Package", language: "en", welcome: "Welcome Carol!" },
  ];

  return (
    <div className="space-y-3">
      {rooms.map((r) => (
        <div key={r.room} className="flex items-center justify-between rounded-lg border border-slate-700 bg-slate-800/50 p-4">
          <div className="flex items-center gap-4">
            <div className="flex h-10 w-10 items-center justify-center rounded bg-slate-700 text-sm font-bold text-white">
              {r.room}
            </div>
            <div>
              <p className="text-sm font-medium text-white">
                {r.guest ? `Guest: ${r.guest}` : "Vacant"}
              </p>
              <p className="text-xs text-slate-400">{r.channel} · Lang: {r.language}</p>
            </div>
          </div>
          <div className="flex items-center gap-3">
            <div className="rounded bg-slate-700/50 px-3 py-1.5 text-xs text-slate-300">
              {r.welcome}
            </div>
            <button className="btn-secondary text-xs">Edit</button>
            <ToggleSwitch checked={true} />
          </div>
        </div>
      ))}
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
