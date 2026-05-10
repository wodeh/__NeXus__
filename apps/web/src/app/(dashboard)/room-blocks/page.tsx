"use client";

import { useState } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import { Lock, Plus, Ban, Calendar, User, ArrowUpRight } from "lucide-react";

interface RoomBlock {
  id: string;
  room_ids: string[];
  start_date: string;
  end_date: string;
  reason: string;
  type: string;
  status: string;
  created_by: string;
}

const mockBlocks: RoomBlock[] = [
  { id: "rb-1001", room_ids: ["r-104"], start_date: "2026-05-10", end_date: "2026-05-17", reason: "VIP Hold - Mr. Smith", type: "vip", status: "active", created_by: "manager@demo.com" },
  { id: "rb-1002", room_ids: ["r-302", "r-303", "r-304"], start_date: "2026-05-24", end_date: "2026-05-26", reason: "Wedding Party - Johnson Group", type: "group", status: "active", created_by: "sales@demo.com" },
];

export default function RoomBlocksPage() {
  const { config } = useTenant();
  const [showForm, setShowForm] = useState(false);

  const hasCap = hasCapability(config, CAPABILITIES.OPERATIONS.ROOM_BLOCKS);

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

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Room Blocks</h2>
          <p className="text-sm text-slate-400">Hold rooms out of inventory</p>
        </div>
        <button onClick={() => setShowForm(!showForm)} className="btn-primary">
          <Plus className="h-4 w-4" /> New Block
        </button>
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
        {mockBlocks.map((b) => (
          <div key={b.id} className="card flex items-center justify-between">
            <div className="space-y-1">
              <div className="flex items-center gap-2">
                <span className="rounded-full bg-nexus-500/10 px-2.5 py-0.5 text-xs font-medium text-nexus-400 uppercase">{b.type}</span>
                <span className="rounded-full bg-emerald-500/10 px-2.5 py-0.5 text-xs font-medium text-emerald-400">{b.status}</span>
              </div>
              <p className="text-sm font-medium text-white">{b.reason}</p>
              <div className="flex items-center gap-4 text-xs text-slate-500">
                <span>Rooms: {b.room_ids.join(", ")}</span>
                <span className="flex items-center gap-1"><Calendar className="h-3 w-3" /> {b.start_date} → {b.end_date}</span>
                <span className="flex items-center gap-1"><User className="h-3 w-3" /> {b.created_by}</span>
              </div>
            </div>
            <button className="rounded-lg p-2 text-slate-400 hover:bg-red-500/10 hover:text-red-400">
              <Ban className="h-4 w-4" />
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
