"use client";

import { useAuth } from "@/lib/auth";
import VillaDashboardPage from "./villa-dashboard/page";
import HotelDashboardPage from "./hotel-dashboard/page";

export default function DashboardPage() {
  const { isVillaOwner, isSuperAdmin } = useAuth();
  if (isSuperAdmin) return <HotelDashboardPage />; // Super admin sees hotel/system dashboard
  return isVillaOwner ? <VillaDashboardPage /> : <HotelDashboardPage />;
}
