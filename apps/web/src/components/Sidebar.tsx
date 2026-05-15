"use client";

import { useState } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  LayoutGrid,
  CalendarDays,
  Users,
  Building2,
  DoorOpen,
  Sparkles,
  TrendingUp,
  ShieldCheck,
  Settings,
  Crown,
  ChevronLeft,
  ChevronRight,
  Globe,
  MessageCircle,
  ExternalLink,
  Star,
  Mail,
  Tv,
  Lock,
  Home,
  Smartphone,
  Wifi,
} from "lucide-react";
import { TenantConfig, hasCapability, CAPABILITIES } from "@/lib/tenant";
import { useAuth } from "@/lib/auth";

interface NavItem {
  label: string;
  href: string;
  icon: React.ElementType;
  cap: string;
  adminOnly?: boolean;
}

const navItems: NavItem[] = [
  { label: "Dashboard", href: "/", icon: LayoutDashboard, cap: CAPABILITIES.CORE.RESERVATIONS },
  { label: "Command Center", href: "/command", icon: LayoutGrid, cap: CAPABILITIES.OPERATIONS.FRONT_DESK },
  { label: "Floor Plan", href: "/floor", icon: LayoutGrid, cap: CAPABILITIES.OPERATIONS.FLOOR_DASHBOARD },
  { label: "Reservations", href: "/reservations", icon: CalendarDays, cap: CAPABILITIES.CORE.RESERVATIONS },
  { label: "Guests", href: "/guests", icon: Users, cap: CAPABILITIES.CORE.GUESTS },
  { label: "Room Blocks", href: "/room-blocks", icon: DoorOpen, cap: CAPABILITIES.OPERATIONS.ROOM_BLOCKS },
  { label: "Groups", href: "/groups", icon: Users, cap: CAPABILITIES.OPERATIONS.GROUP_RESERVATIONS },
  { label: "CRM", href: "/crm", icon: Users, cap: CAPABILITIES.ENTERPRISE.ADVANCED_CRM },
  { label: "Agents", href: "/agents", icon: Building2, cap: CAPABILITIES.REVENUE.AGENT_MANAGEMENT },
  { label: "Properties", href: "/properties", icon: Building2, cap: CAPABILITIES.CORE.PROPERTIES },
  { label: "Rooms", href: "/rooms", icon: DoorOpen, cap: CAPABILITIES.CORE.ROOMS },
  { label: "Housekeeping", href: "/housekeeping", icon: Sparkles, cap: CAPABILITIES.CORE.HOUSEKEEPING },
  { label: "IPTV", href: "/iptv", icon: Tv, cap: CAPABILITIES.OPERATIONS.IPTV_BASIC },
  { label: "Smart Locks", href: "/locks", icon: Lock, cap: CAPABILITIES.OPERATIONS.SMART_LOCKS },
  { label: "WiFi Monitor", href: "/wifi", icon: Wifi, cap: CAPABILITIES.OPERATIONS.SMART_LOCKS },
  { label: "Channel Manager", href: "/channel-manager", icon: Globe, cap: CAPABILITIES.REVENUE.CHANNEL_MANAGER },
  { label: "WhatsApp Bot", href: "/whatsapp", icon: MessageCircle, cap: CAPABILITIES.REVENUE.WHATSAPP_BOT },
  { label: "Booking Engine", href: "/booking-engine", icon: ExternalLink, cap: CAPABILITIES.REVENUE.DIRECT_BOOKING },
  { label: "Reviews", href: "/reviews", icon: Star, cap: CAPABILITIES.REVENUE.GUEST_REVIEWS },
  { label: "Communications", href: "/communications", icon: Mail, cap: CAPABILITIES.REVENUE.COMMUNICATIONS },
  { label: "Revenue", href: "/revenue", icon: TrendingUp, cap: CAPABILITIES.REVENUE.DYNAMIC_PRICING },
  { label: "Rate Rules", href: "/rate-rules", icon: TrendingUp, cap: CAPABILITIES.REVENUE.DYNAMIC_PRICING },
  { label: "Guest Journey", href: "/guest-journey", icon: Globe, cap: CAPABILITIES.REVENUE.REVENUE_FORECAST },
  { label: "Upsells", href: "/upsells", icon: Star, cap: CAPABILITIES.REVENUE.DYNAMIC_PRICING },
  { label: "Competitors", href: "/competitors", icon: Globe, cap: CAPABILITIES.REVENUE.DYNAMIC_PRICING },
  { label: "Audit", href: "/audit", icon: ShieldCheck, cap: CAPABILITIES.CORE.AUDIT_LOGS },
  { label: "Settings", href: "/settings", icon: Settings, cap: CAPABILITIES.CORE.SETTINGS },
  { label: "My Villas", href: "/villas", icon: Home, cap: CAPABILITIES.VILLA.PROPERTIES },
  { label: "Bookings", href: "/villa-reservations", icon: CalendarDays, cap: CAPABILITIES.VILLA.RESERVATIONS },
  { label: "Tapechart", href: "/villa-tapechart", icon: CalendarDays, cap: CAPABILITIES.VILLA.RESERVATIONS },
  { label: "Villa Revenue", href: "/villa-revenue", icon: TrendingUp, cap: CAPABILITIES.VILLA.REVENUE },
  { label: "Cleaner Tracking", href: "/cleaner-tracking", icon: Smartphone, cap: CAPABILITIES.VILLA.CLEANER_TRACKING },
  // Admin link — only visible to super_admin
  { label: "Admin", href: "/admin", icon: ShieldCheck, cap: "admin:full", adminOnly: true },
];

export default function Sidebar({
  tenantConfig,
  propertyCount,
  isSuperAdmin,
}: {
  tenantConfig: TenantConfig | null;
  propertyCount?: number | null;
  isSuperAdmin?: boolean;
}) {
  const pathname = usePathname();
  const [collapsed, setCollapsed] = useState(false);
  const { isVillaOwner } = useAuth();

  const filteredNav = navItems.filter((item) => {
    // Admin-only link — only for super_admin
    if (item.adminOnly) return isSuperAdmin;

    if (isSuperAdmin) return true;
    if (!tenantConfig) return true;
    if (isVillaOwner) {
      return item.cap.startsWith("villa:") ||
             item.href === "/" ||
             item.href === "/settings" ||
             item.cap === CAPABILITIES.OPERATIONS.IPTV_BASIC;
    }

    // Hide Properties menu if hotel has only 1 property
    if (item.href === "/properties" && propertyCount !== null && propertyCount <= 1) {
      return false;
    }

    return hasCapability(tenantConfig, item.cap);
  });

  return (
    <aside
      className={`fixed left-0 top-0 z-40 flex h-screen flex-col border-r border-slate-800 bg-slate-900 transition-all duration-300 ${
        collapsed ? "w-16" : "w-60"
      }`}
    >
      <div className="flex h-16 items-center justify-between border-b border-slate-800 px-4">
        {!collapsed && (
          <Link href="/" className="flex items-center gap-2">
            <Crown className="h-6 w-6 text-nexus-400" />
            <span className="text-lg font-bold text-white">NeXus</span>
          </Link>
        )}
        {collapsed && <Crown className="h-6 w-6 text-nexus-400" />}
        <button
          onClick={() => setCollapsed(!collapsed)}
          className="rounded p-1 text-slate-500 hover:bg-slate-800 hover:text-white"
        >
          {collapsed ? (
            <ChevronRight className="h-4 w-4" />
          ) : (
            <ChevronLeft className="h-4 w-4" />
          )}
        </button>
      </div>

      <nav className="flex-1 space-y-1 overflow-y-auto p-3 scrollbar-hide">
        {filteredNav.map((item) => {
          const Icon = item.icon;
          const active = pathname === item.href;
          return (
            <Link
              key={item.href}
              href={item.href}
              className={`flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                active
                  ? "bg-slate-800 text-white"
                  : "text-slate-400 hover:bg-slate-800 hover:text-white"
              } ${collapsed ? "justify-center" : ""}`}
              title={collapsed ? item.label : undefined}
            >
              <Icon className="h-5 w-5 flex-shrink-0" />
              {!collapsed && <span>{item.label}</span>}
            </Link>
          );
        })}
      </nav>

      {!collapsed && tenantConfig && (
        <div className="border-t border-slate-800 p-3">
          <p className="text-[10px] uppercase tracking-wider text-slate-500">License</p>
          <p className="mt-0.5 text-xs font-semibold text-nexus-400 capitalize">
            {isVillaOwner ? "Villa Rental" : tenantConfig.license_tier}
          </p>
          <p className="mt-0.5 text-[10px] text-slate-500 leading-tight">
            {isVillaOwner ? "unlimited villas" : `${tenantConfig.max_rooms} rooms max`}
          </p>
        </div>
      )}
    </aside>
  );
}
