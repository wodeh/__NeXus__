"use client";

import { useState } from "react";
import { Plus, Search, Building2, MapPin, ChevronRight } from "lucide-react";

const mockProperties = [
  { id: "p-001", name: "Grand Plaza Resort", address: "123 Ocean Drive, Miami, FL", rooms: 120, status: "active" },
  { id: "p-002", name: "Sunset Inn", address: "456 Mountain Road, Denver, CO", rooms: 45, status: "active" },
  { id: "p-003", name: "Harbor View Hotel", address: "789 Bay Street, San Francisco, CA", rooms: 200, status: "maintenance" },
];

export default function PropertiesPage() {
  const [search, setSearch] = useState("");
  const filtered = mockProperties.filter((p) =>
    p.name.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Properties</h2>
          <p className="text-sm text-slate-400">Manage hotels and locations</p>
        </div>
        <button className="btn-primary">
          <Plus className="h-4 w-4" />
          Add Property
        </button>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
        <input
          className="input w-full max-w-md pl-9"
          placeholder="Search properties..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {filtered.map((p) => (
          <div key={p.id} className="card-hover">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-nexus-500/10">
                  <Building2 className="h-5 w-5 text-nexus-400" />
                </div>
                <div>
                  <p className="font-medium text-white">{p.name}</p>
                  <span className={`text-xs ${p.status === "active" ? "badge-green" : "badge-amber"}`}>
                    {p.status}
                  </span>
                </div>
              </div>
              <ChevronRight className="h-4 w-4 text-slate-600" />
            </div>
            <div className="mt-4 space-y-1 text-sm text-slate-400">
              <div className="flex items-center gap-2"><MapPin className="h-3.5 w-3.5" />{p.address}</div>
              <p>{p.rooms} rooms</p>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}
