"use client";

import { useState } from "react";
import { Plus, Search, User, Phone, Mail, ChevronRight } from "lucide-react";

const mockGuests = [
  { id: "g-001", first_name: "Alice", last_name: "Chen", email: "alice@example.com", phone: "+1-555-0101", status: "checked_in" },
  { id: "g-002", first_name: "Bob", last_name: "Smith", email: "bob@example.com", phone: "+1-555-0102", status: "confirmed" },
  { id: "g-003", first_name: "Carol", last_name: "Jones", email: "carol@example.com", phone: "+1-555-0103", status: "checked_out" },
];

export default function GuestsPage() {
  const [search, setSearch] = useState("");
  const filtered = mockGuests.filter((g) =>
    `${g.first_name} ${g.last_name}`.toLowerCase().includes(search.toLowerCase()) ||
    g.email.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Guests</h2>
          <p className="text-sm text-slate-400">Manage guest profiles and preferences</p>
        </div>
        <button className="btn-primary">
          <Plus className="h-4 w-4" />
          Add Guest
        </button>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
        <input
          className="input w-full max-w-md pl-9"
          placeholder="Search guests..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {filtered.map((g) => (
          <div key={g.id} className="card-hover">
            <div className="flex items-center gap-3">
              <div className="flex h-10 w-10 items-center justify-center rounded-full bg-nexus-500/10">
                <User className="h-5 w-5 text-nexus-400" />
              </div>
              <div>
                <p className="font-medium text-white">{g.first_name} {g.last_name}</p>
                <span className={`text-xs ${g.status === "checked_in" ? "badge-green" : g.status === "checked_out" ? "badge-amber" : "badge-blue"}`}>
                  {g.status.replace("_", " ")}
                </span>
              </div>
            </div>
            <div className="mt-4 space-y-1 text-sm text-slate-400">
              <div className="flex items-center gap-2"><Mail className="h-3.5 w-3.5" />{g.email}</div>
              <div className="flex items-center gap-2"><Phone className="h-3.5 w-3.5" />{g.phone}</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
