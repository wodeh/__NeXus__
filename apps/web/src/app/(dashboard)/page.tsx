"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import {
  CalendarDays,
  Users,
  Building2,
  DoorOpen,
  TrendingUp,
  ArrowUpRight,
  ArrowDownRight,
} from "lucide-react";
import { apiClient } from "@/lib/api";

const tenant = "demo";

interface Stat {
  label: string;
  value: string;
  change: string;
  up: boolean;
  icon: React.ElementType;
  href: string;
}

export default function DashboardPage() {
  const [stats, setStats] = useState<Stat[]>([
    { label: "Active Reservations", value: "—", change: "—", up: true, icon: CalendarDays, href: "/reservations" },
    { label: "Checked-in Guests", value: "—", change: "—", up: true, icon: Users, href: "/guests" },
    { label: "Properties", value: "—", change: "—", up: true, icon: Building2, href: "/properties" },
    { label: "Rooms Available", value: "—", change: "—", up: false, icon: DoorOpen, href: "/rooms" },
    { label: "Revenue Today", value: "—", change: "—", up: true, icon: TrendingUp, href: "/revenue" },
  ]);

  useEffect(() => {
    async function load() {
      try {
        const [resData, guestData, propData, roomData] = await Promise.all([
          apiClient.get(`/tenants/${tenant}/reservations`).catch(() => ({ guests: [] })),
          apiClient.get(`/tenants/${tenant}/guests`).catch(() => ({ guests: [] })),
          apiClient.get(`/tenants/${tenant}/properties`).catch(() => ({ properties: [] })),
          apiClient.get(`/tenants/${tenant}/properties`).catch(() => ({ properties: [] })),
        ]);

        setStats((prev) =>
          prev.map((s) => {
            if (s.label === "Active Reservations") return { ...s, value: String(resData.guests?.length ?? 0) };
            if (s.label === "Checked-in Guests") return { ...s, value: String(guestData.guests?.length ?? 0) };
            if (s.label === "Properties") return { ...s, value: String(propData.properties?.length ?? 0) };
            if (s.label === "Rooms Available") return { ...s, value: "12" };
            if (s.label === "Revenue Today") return { ...s, value: "$0" };
            return s;
          })
        );
      } catch {
        // backend may be offline — keep placeholders
      }
    }
    load();
  }, []);

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-2xl font-bold text-white">Dashboard</h2>
        <p className="text-sm text-slate-400">Overview of your property operations</p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-5">
        {stats.map((stat) => {
          const Icon = stat.icon;
          const Arrow = stat.up ? ArrowUpRight : ArrowDownRight;
          return (
            <Link
              key={stat.label}
              href={stat.href}
              className="card-hover group"
            >
              <div className="flex items-center justify-between">
                <Icon className="h-5 w-5 text-nexus-400" />
                <Arrow className="h-4 w-4 text-slate-600 group-hover:text-nexus-400" />
              </div>
              <div className="mt-3">
                <p className="text-2xl font-bold text-white">{stat.value}</p>
                <p className="mt-1 text-xs text-slate-400">{stat.label}</p>
              </div>
            </Link>
          );
        })}
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <div className="card">
          <h3 className="text-sm font-semibold text-slate-200">Recent Reservations</h3>
          <p className="mt-4 text-sm text-slate-500">No recent reservations to display.</p>
        </div>
        <div className="card">
          <h3 className="text-sm font-semibold text-slate-200">Occupancy Trend</h3>
          <p className="mt-4 text-sm text-slate-500">Chart will appear once data is available.</p>
        </div>
      </div>
    </div>
  );
}
