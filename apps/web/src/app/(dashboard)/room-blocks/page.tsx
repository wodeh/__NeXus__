"use client";

import { useState, useEffect, useCallback, useMemo } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import { Lock, Plus, Ban, Calendar, User, ArrowUpRight, RefreshCw, AlertTriangle } from "lucide-react";
import { Room, getRooms } from "@/lib/api";

interface RoomBlock {
  id: string;
  room_ids: string[];
  start_date: string;
  end_date: string;
  reason: string;
  type: string;
  status: string;
}

export default function RoomBlocksPage() {
  const { config } = useTenant();
  const [showForm, setShowForm] = useState(false);
  const [rooms, setRooms] = useState<Room[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const hasCap = hasCapability(config, CAPABILITIES.OPERATIONS.ROOM_BLOCKS);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const rms = await getRooms();
      setRooms(rms);
    } catch (e: any) {
      setError(e.message || "Failed to load room blocks");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const blockedRooms = useMemo(() => {
    return rooms.filter((r) => ["blocked", "maintenance", "out_of_order"].includes(r.status));
  }, [rooms]);

  if (!hasCap) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-800">
          <Lock className="h-8 w-8 text-slate-500" />
        </div>
        <h2 className="mt-6 text-xl font-bold text-white">Room Blocks Locked</h2>
        <p className="mt-2 max-w-md text-sm text-slate-400">Block rooms for VIPs, groups, or maintenance with the Operations or Enterprise tier.</p>
        <button className="btn-primary mt-6">
          <ArrowUpRight className="h-4 w-4" /> Upgrade to Operations
        </button>
      </div>
    );
  }

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
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Room Blocks</h2>
          <p className="text-sm text-slate-400">Hold rooms out of inventory</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={fetchData} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white" title="Refresh">
            <RefreshCw className="h-4 w-4" />
          </button>
          <button onClick={() => setShowForm(!showForm)} className="btn-primary">
            <Plus className="h-4 w-4" /> New Block
          </button>
        </div>
      </div>

      {showForm && (
        <div className="card">
          <h3 className="text-sm font-semibold text-slate-200">Create Room Block</h3>
          <p className="mt-2 text-xs text-slate-500">Form will connect to API when backend is running.</p>
          <div className="mt-4 flex gap-2">
            <button className="btn-secondary text-xs" onClick={() => setShowForm(false)}>Cancel</button>
            <button className="btn-primary text-xs">Create Block</button>
          </div>
        </div>
      )}

      <div className="space-y-3">
        {blockedRooms.map((r) => (
          <div key={r.id} className="card flex items-center justify-between">
            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <span className={`rounded-full px-2.5 py-0.5 text-xs font-medium uppercase ${
                  r.status === "blocked" ? "bg-purple-500/10 text-purple-400" :
                  r.status === "maintenance" ? "bg-amber-500/10 text-amber-400" :
                  "bg-slate-500/10 text-slate-400"
                }`}>{r.status}</span>
                <span className="text-xs text-slate-500">{r.type}</span>
              </div>
              <p className="text-sm font-medium text-white">Room {r.number} — Floor {r.floor}</p>
              <div className="flex items-center gap-4 text-xs text-slate-500">
                <span>{r.bed_type || "Standard"} · ${r.rate_night}/night</span>
                {r.rate_night > 0 && <span className="flex items-center gap-1"><Calendar className="h-3 w-3" /> ${r.rate_night}/night</span>}
              </div>
            </div>
            <button className="rounded-lg p-2 text-slate-400 hover:bg-red-500/10 hover:text-red-400">
              <Ban className="h-4 w-4" />
            </button>
          </div>
        ))}
        {blockedRooms.length === 0 && (
          <div className="py-12 text-center text-sm text-slate-500">
            No blocked rooms.
          </div>
        )}
      </div>
    </div>
  );
}
