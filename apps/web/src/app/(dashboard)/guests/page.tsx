"use client";

import { useState, useEffect, useCallback, useMemo } from "react";
import { Plus, Search, User, Phone, Mail, ChevronRight, RefreshCw, AlertTriangle } from "lucide-react";
import { Reservation, getReservations } from "@/lib/api";

interface Guest {
  id: string;
  name: string;
  email: string;
  phone: string;
  status: string;
  vip: boolean;
  visits: number;
  lastStay: string;
}

export default function GuestsPage() {
  const [search, setSearch] = useState("");
  const [guests, setGuests] = useState<Guest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const reservations = await getReservations();
      // Derive unique guests from reservations
      const guestMap = new Map<string, Guest>();
      reservations.forEach((r) => {
        const key = r.email || r.guest_name;
        const existing = guestMap.get(key);
        if (existing) {
          existing.visits += 1;
          if (r.check_out > existing.lastStay) existing.lastStay = r.check_out;
          if (r.status === "checked_in") existing.status = "checked_in";
        } else {
          guestMap.set(key, {
            id: r.id,
            name: r.guest_name,
            email: r.email || "",
            phone: r.phone || "",
            status: r.status,
            vip: r.vip,
            visits: 1,
            lastStay: r.check_out,
          });
        }
      });
      setGuests(Array.from(guestMap.values()));
    } catch (e: any) {
      setError(e.message || "Failed to load guests");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const filtered = useMemo(() => {
    const term = search.toLowerCase();
    return guests.filter((g) =>
      g.name.toLowerCase().includes(term) ||
      g.email.toLowerCase().includes(term) ||
      g.phone.toLowerCase().includes(term)
    );
  }, [guests, search]);

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
          <h2 className="text-2xl font-bold text-white">Guests</h2>
          <p className="text-sm text-slate-400">Manage guest profiles and preferences</p>
        </div>
        <div className="flex items-center gap-2">
          <button onClick={fetchData} className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white" title="Refresh">
            <RefreshCw className="h-4 w-4" />
          </button>
          <button className="btn-primary">
            <Plus className="h-4 w-4" />
            Add Guest
          </button>
        </div>
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
                <p className="font-medium text-white">{g.name}</p>
                <div className="flex items-center gap-1">
                  <span className={`text-xs ${g.status === "checked_in" ? "badge-green" : g.status === "checked_out" ? "badge-amber" : "badge-blue"}`}>
                    {g.status.replace("_", " ")}
                  </span>
                  {g.vip && <span className="text-[10px] text-amber-400">★ VIP</span>}
                </div>
              </div>
            </div>
            <div className="mt-4 space-y-1 text-sm text-slate-400">
              {g.email && <div className="flex items-center gap-2"><Mail className="h-3.5 w-3.5" />{g.email}</div>}
              {g.phone && <div className="flex items-center gap-2"><Phone className="h-3.5 w-3.5" />{g.phone}</div>}
              <div className="flex items-center justify-between pt-1">
                <span className="text-[10px] text-slate-500">{g.visits} stay{g.visits > 1 ? "s" : ""}</span>
                <span className="text-[10px] text-slate-500">Last: {g.lastStay}</span>
              </div>
            </div>
          </div>
        ))}
        {filtered.length === 0 && (
          <div className="col-span-full py-12 text-center text-sm text-slate-500">
            No guests found.
          </div>
        )}
      </div>
    </div>
  );
}
