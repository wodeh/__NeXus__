"use client";

import { useAuth } from "@/lib/auth";
import VillaDashboardPage from "./villa-dashboard/page";
import HotelDashboardPage from "./hotel-dashboard/page";

export default function DashboardPage() {
  const { isVillaOwner } = useAuth();
  return isVillaOwner ? <VillaDashboardPage /> : <HotelDashboardPage />;
}
