"use client";

import { useState } from "react";
import { Bell, Search, User, LogOut } from "lucide-react";
import { TenantConfig, tierName } from "@/lib/tenant";
import { useAuth } from "@/lib/auth";

export default function Topbar({ tenantConfig }: { tenantConfig: TenantConfig | null }) {
  const [search, setSearch] = useState("");
  const { user, logout } = useAuth();

  return (
    <header className="flex h-16 items-center justify-between border-b border-slate-800 bg-slate-900 px-6">
      <div className="flex items-center gap-4">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
          <input
            className="w-64 rounded-lg border border-slate-700 bg-slate-800 pl-9 pr-3 py-2 text-sm text-white placeholder-slate-500 outline-none focus:border-nexus-500 focus:ring-1 focus:ring-nexus-500"
            placeholder="Search anything..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
      </div>

      <div className="flex items-center gap-4">
        {tenantConfig && (
          <div className="flex items-center gap-2 text-sm">
            <span className="font-medium text-white">{tenantConfig.name}</span>
            <span className="rounded-full bg-nexus-500/10 px-2.5 py-0.5 text-xs font-medium text-nexus-400">
              {tierName(tenantConfig.license_tier)}
            </span>
          </div>
        )}

        {user && (
          <div className="flex items-center gap-2 text-sm text-slate-300">
            <User className="h-4 w-4" />
            <span className="hidden sm:inline">{user.name}</span>
            <span className="rounded-full bg-slate-700 px-2 py-0.5 text-[10px] uppercase text-slate-400">
              {user.role}
            </span>
          </div>
        )}

        <button className="relative rounded-lg p-2 text-slate-400 hover:bg-slate-800 hover:text-white">
          <Bell className="h-5 w-5" />
          <span className="absolute right-1.5 top-1.5 h-2 w-2 rounded-full bg-red-500" />
        </button>

        <button
          onClick={() => {
            logout();
            window.location.href = "/login";
          }}
          className="rounded-lg p-2 text-slate-400 hover:bg-slate-800 hover:text-red-400"
          title="Sign Out"
        >
          <LogOut className="h-5 w-5" />
        </button>
      </div>
    </header>
  );
}
