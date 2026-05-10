"use client";

import { useState } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import { Lock, Plus, Users, Calendar, Mail, Phone, ArrowUpRight } from "lucide-react";

interface GroupReservation {
  id: string;
  group_name: string;
  contact_name: string;
  contact_email: string;
  contact_phone: string;
  num_rooms: number;
  num_guests: number;
  check_in: string;
  check_out: string;
  status: string;
  rate_code: string;
}

const mockGroups: GroupReservation[] = [
  { id: "grp-501", group_name: "Wedding Party - Johnson", contact_name: "Sarah Johnson", contact_email: "sarah.j@example.com", contact_phone: "+1-555-0100", num_rooms: 5, num_guests: 12, check_in: "2026-05-24", check_out: "2026-05-26", status: "definite", rate_code: "GROUP-WEDDING" },
  { id: "grp-502", group_name: "Corporate Retreat - TechCorp", contact_name: "Mike Chen", contact_email: "mike@techcorp.com", contact_phone: "+1-555-0200", num_rooms: 8, num_guests: 15, check_in: "2026-05-17", check_out: "2026-05-19", status: "tentative", rate_code: "CORP-RETREAT" },
];

export default function GroupsPage() {
  const { config } = useTenant();
  const [showForm, setShowForm] = useState(false);

  const hasCap = hasCapability(config, CAPABILITIES.OPERATIONS.GROUP_RESERVATIONS);

  if (!hasCap) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-800">
          <Lock className="h-8 w-8 text-slate-500" />
        </div>
        <h2 className="mt-6 text-xl font-bold text-white">Group Reservations Locked</h2>
        <p className="mt-2 max-w-md text-sm text-slate-400">Manage wedding parties, corporate retreats, and tour groups with the Operations or Enterprise tier.</p>
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
          <h2 className="text-2xl font-bold text-white">Group Reservations</h2>
          <p className="text-sm text-slate-400">Weddings, corporate blocks, tour groups</p>
        </div>
        <button onClick={() => setShowForm(!showForm)} className="btn-primary">
          <Plus className="h-4 w-4" /> New Group
        </button>
      </div>

      {showForm && (
        <div className="card">
          <h3 className="text-sm font-semibold text-slate-200">Create Group Reservation</h3>
          <p className="mt-2 text-xs text-slate-500">Form will connect to API when backend is running.</p>
          <div className="mt-4 flex gap-2">
            <button className="btn-secondary text-xs" onClick={() => setShowForm(false)}>Cancel</button>
            <button className="btn-primary text-xs">Create Group</button>
          </div>
        </div>
      )}

      <div className="space-y-3">
        {mockGroups.map((g) => (
          <div key={g.id} className="card">
            <div className="flex items-center justify-between">
              <div className="space-y-1">
                <div className="flex items-center gap-2">
                  <span className="rounded-full bg-nexus-500/10 px-2.5 py-0.5 text-xs font-medium text-nexus-400 uppercase">{g.status}</span>
                  <span className="text-xs text-slate-500">{g.rate_code}</span>
                </div>
                <p className="text-sm font-medium text-white">{g.group_name}</p>
                <div className="flex items-center gap-4 text-xs text-slate-500">
                  <span className="flex items-center gap-1"><Users className="h-3 w-3" /> {g.num_rooms} rooms / {g.num_guests} guests</span>
                  <span className="flex items-center gap-1"><Calendar className="h-3 w-3" /> {g.check_in} → {g.check_out}</span>
                  <span className="flex items-center gap-1"><Mail className="h-3 w-3" /> {g.contact_email}</span>
                  <span className="flex items-center gap-1"><Phone className="h-3 w-3" /> {g.contact_phone}</span>
                </div>
              </div>
              <div className="flex gap-2">
                <button className="btn-secondary text-xs">Edit</button>
                <button className="rounded-lg p-2 text-slate-400 hover:bg-red-500/10 hover:text-red-400">
                  <ArrowUpRight className="h-4 w-4" />
                </button>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
