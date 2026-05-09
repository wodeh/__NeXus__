"use client";

import { useState } from "react";
import { useSearchParams } from "next/navigation";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Lock, Unlock, KeyRound, ArrowLeft, BatteryMedium, Wifi } from "lucide-react";
import Link from "next/link";

export default function IPTVLockPage() {
  const searchParams = useSearchParams();
  const roomNumber = searchParams.get("room") || "";

  const [isLocked, setIsLocked] = useState(true);
  const [isLoading, setIsLoading] = useState(false);
  const [lastAction, setLastAction] = useState("Door locked at 14:32");

  const toggleLock = async () => {
    setIsLoading(true);
    // In production, call backend API
    await new Promise((r) => setTimeout(r, 800));
    setIsLocked(!isLocked);
    setLastAction(isLocked ? `Door unlocked at ${new Date().toLocaleTimeString()}` : `Door locked at ${new Date().toLocaleTimeString()}`);
    setIsLoading(false);
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-900 via-slate-800 to-slate-900 text-white">
      <header className="flex items-center gap-4 px-8 py-6">
        <Link href={`/iptv?room=${roomNumber}`} className="rounded-lg bg-white/10 p-2 hover:bg-white/20">
          <ArrowLeft className="h-5 w-5" />
        </Link>
        <h1 className="text-xl font-bold">Smart Lock Control</h1>
      </header>

      <div className="mx-8 grid grid-cols-1 gap-6 md:grid-cols-2">
        <Card className="border-white/10 bg-white/5 backdrop-blur">
          <CardHeader>
            <CardTitle className="text-white">Room {roomNumber}</CardTitle>
          </CardHeader>
          <CardContent className="flex flex-col items-center gap-6 py-8">
            <div
              className={`flex h-32 w-32 items-center justify-center rounded-full transition-all ${
                isLocked ? "bg-red-500/20 text-red-400" : "bg-green-500/20 text-green-400"
              }`}
            >
              {isLocked ? <Lock className="h-16 w-16" /> : <Unlock className="h-16 w-16" />}
            </div>

            <p className="text-2xl font-bold">
              {isLocked ? "Locked" : "Unlocked"}
            </p>
            <p className="text-sm text-white/50">{lastAction}</p>

            <Button
              onClick={toggleLock}
              disabled={isLoading}
              className={`h-14 w-full text-lg ${
                isLocked
                  ? "bg-red-600 hover:bg-red-700"
                  : "bg-green-600 hover:bg-green-700"
              }`}
            >
              {isLoading
                ? "Processing..."
                : isLocked
                ? "Unlock Door"
                : "Lock Door"}
            </Button>
          </CardContent>
        </Card>

        <Card className="border-white/10 bg-white/5 backdrop-blur">
          <CardHeader>
            <CardTitle className="text-white">Lock Status</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            {[
              { label: "Battery", value: "87%", icon: BatteryMedium, color: "text-green-400" },
              { label: "Connection", value: "Online", icon: Wifi, color: "text-blue-400" },
              { label: "Lock Model", value: "Orbita S5", icon: KeyRound, color: "text-amber-400" },
            ].map((item) => (
              <div key={item.label} className="flex items-center justify-between rounded-lg bg-white/5 p-3">
                <div className="flex items-center gap-3">
                  <item.icon className={`h-5 w-5 ${item.color}`} />
                  <span className="text-sm">{item.label}</span>
                </div>
                <span className="font-medium">{item.value}</span>
              </div>
            ))}

            <div className="mt-4 rounded-lg bg-amber-500/10 p-3 text-sm text-amber-300">
              <p className="font-medium">Digital Key Active</p>
              <p className="text-amber-300/70">Your phone key is valid until check-out.</p>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
