"use client";

import { useState, useEffect } from "react";
import { useSearchParams } from "next/navigation";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Hotel, Tv, Wifi, Lock, Phone, MapPin, ChevronRight } from "lucide-react";

export default function IPTVHomePage() {
  const searchParams = useSearchParams();
  const [roomNumber, setRoomNumber] = useState("");
  const [guestName, setGuestName] = useState("Guest");
  const [currentTime, setCurrentTime] = useState(new Date());

  useEffect(() => {
    const room = searchParams.get("room");
    const token = searchParams.get("token");
    if (room) setRoomNumber(room);
    // In production, validate token with backend
    const timer = setInterval(() => setCurrentTime(new Date()), 60000);
    return () => clearInterval(timer);
  }, [searchParams]);

  const menuItems = [
    { label: "Movies & TV", icon: Tv, color: "bg-red-500", href: "#entertainment" },
    { label: "Room Service", icon: Hotel, color: "bg-emerald-500", href: "#dining" },
    { label: "Smart Lock", icon: Lock, color: "bg-blue-500", href: "#lock" },
    { label: "Hotel Info", icon: MapPin, color: "bg-amber-500", href: "#info" },
    { label: "WiFi & Help", icon: Wifi, color: "bg-purple-500", href: "#wifi" },
    { label: "Concierge", icon: Phone, color: "bg-pink-500", href: "#concierge" },
  ];

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 text-white">
      {/* Header Bar */}
      <header className="flex items-center justify-between px-8 py-6">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-white/10 backdrop-blur">
            <Hotel className="h-5 w-5" />
          </div>
          <div>
            <h1 className="text-lg font-bold">Nexus IPTV</h1>
            <p className="text-xs text-white/60">Guru Hospitality Platform</p>
          </div>
        </div>
        <div className="text-right">
          <p className="text-sm font-medium">Room {roomNumber || "—"}</p>
          <p className="text-xs text-white/60">Welcome, {guestName}</p>
          <p className="mt-1 text-xl font-light">
            {currentTime.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
          </p>
        </div>
      </header>

      {/* Hero Banner */}
      <div className="mx-8 mb-8 rounded-2xl bg-gradient-to-r from-nexus-600 to-nexus-800 p-8">
        <h2 className="text-3xl font-bold">Welcome to Your Stay</h2>
        <p className="mt-2 text-white/80">
          Experience luxury at your fingertips. Order room service, unlock your door,
          and explore entertainment — all from your TV.
        </p>
      </div>

      {/* Menu Grid */}
      <div className="mx-8 grid grid-cols-2 gap-4 md:grid-cols-3">
        {menuItems.map((item) => (
          <Card
            key={item.label}
            className="group cursor-pointer border-white/10 bg-white/5 backdrop-blur transition-all hover:bg-white/10"
          >
            <CardContent className="flex items-center gap-4 p-6">
              <div className={`flex h-12 w-12 items-center justify-center rounded-xl ${item.color}`}>
                <item.icon className="h-6 w-6 text-white" />
              </div>
              <div className="flex-1">
                <p className="font-semibold">{item.label}</p>
                <p className="text-sm text-white/50">Tap to explore</p>
              </div>
              <ChevronRight className="h-5 w-5 text-white/30 transition-transform group-hover:translate-x-1" />
            </CardContent>
          </Card>
        ))}
      </div>

      {/* Weather / Info Footer */}
      <footer className="mx-8 mt-8 rounded-xl bg-white/5 p-4 text-center text-sm text-white/40">
        Nexus PMS · Guru Hospitality Platform · Press * for Help
      </footer>
    </div>
  );
}
