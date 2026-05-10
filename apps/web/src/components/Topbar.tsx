"use client";

import { useState } from "react";
import { Building2, ChevronDown } from "lucide-react";

const tenants = [
  { id: "demo", name: "Demo Hotel" },
  { id: "grand-plaza", name: "Grand Plaza Resort" },
  { id: "sunset-inn", name: "Sunset Inn" },
];

export default function Topbar() {
  const [tenant, setTenant] = useState(tenants[0]);
  const [open, setOpen] = useState(false);

  return (
    <header className="sticky top-0 z-30 flex h-14 items-center justify-between border-b border-slate-800 bg-slate-950/80 px-6 backdrop-blur">
      <h1 className="text-sm font-semibold text-slate-300">Property Management System</h1>
      <div className="relative">
        <button
          onClick={() => setOpen(!open)}
          className="flex items-center gap-2 rounded-lg border border-slate-700 bg-slate-800 px-3 py-1.5 text-sm text-white hover:bg-slate-700"
        >
          <Building2 className="h-4 w-4 text-nexus-400" />
          {tenant.name}
          <ChevronDown className="h-3 w-3 text-slate-400" />
        </button>
        {open && (
          <div className="absolute right-0 mt-1 w-56 rounded-lg border border-slate-700 bg-slate-800 shadow-xl">
            {tenants.map((t) => (
              <button
                key={t.id}
                onClick={() => { setTenant(t); setOpen(false); }}
                className="flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-slate-200 hover:bg-slate-700"
              >
                <Building2 className="h-4 w-4 text-slate-500" />
                {t.name}
              </button>
            ))}
          </div>
        )}
      </div>
    </header>
  );
}
