"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  CalendarDays,
  Users,
  Building2,
  DoorOpen,
  ClipboardList,
  TrendingUp,
  ShieldCheck,
  Settings,
  LogOut,
  Hotel,
} from "lucide-react";

const nav = [
  { href: "/", label: "Dashboard", icon: LayoutDashboard },
  { href: "/reservations", label: "Reservations", icon: CalendarDays },
  { href: "/guests", label: "Guests", icon: Users },
  { href: "/properties", label: "Properties", icon: Building2 },
  { href: "/rooms", label: "Rooms", icon: DoorOpen },
  { href: "/housekeeping", label: "Housekeeping", icon: ClipboardList },
  { href: "/revenue", label: "Revenue", icon: TrendingUp },
  { href: "/audit", label: "Audit", icon: ShieldCheck },
  { href: "/settings", label: "Settings", icon: Settings },
];

export default function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="fixed left-0 top-0 z-40 flex h-screen w-60 flex-col border-r border-slate-800 bg-slate-950">
      <div className="flex items-center gap-3 px-5 py-5">
        <Hotel className="h-7 w-7 text-nexus-400" />
        <span className="text-lg font-bold tracking-tight text-white">NeXus</span>
      </div>

      <nav className="flex-1 space-y-1 px-3 py-2">
        {nav.map((item) => {
          const Icon = item.icon;
          const isActive = pathname === item.href || pathname.startsWith(item.href + "/");
          return (
            <Link
              key={item.href}
              href={item.href}
              className={isActive ? "nav-link-active" : "nav-link"}
            >
              <Icon className="h-4 w-4" />
              {item.label}
            </Link>
          );
        })}
      </nav>

      <div className="border-t border-slate-800 p-3">
        <button className="nav-link w-full">
          <LogOut className="h-4 w-4" />
          Sign Out
        </button>
      </div>
    </aside>
  );
}
