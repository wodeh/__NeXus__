"use client";

import { useState, useEffect } from "react";
import { useParams } from "next/navigation";
import { Tv, Youtube, Calendar, Moon, Sun, User, Home } from "lucide-react";
import { getIPTVWelcome, IPTVWelcomeScreen } from "@/lib/api";

export default function IPTVWelcomePage() {
  const params = useParams();
  const villaId = params.villa_id as string;

  const [welcome, setWelcome] = useState<IPTVWelcomeScreen | null>(null);
  const [loading, setLoading] = useState(true);
  const [currentTime, setCurrentTime] = useState(new Date());
  const [showYoutube, setShowYoutube] = useState(false);

  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

  useEffect(() => {
    if (!villaId) return;
    loadWelcome();
  }, [villaId]);

  async function loadWelcome() {
    setLoading(true);
    try {
      const data = await getIPTVWelcome(villaId);
      setWelcome(data);
    } catch (e) {
      console.error("Failed to load welcome screen", e);
      // Fallback welcome
      setWelcome({
        villa_id: villaId,
        villa_name: "Villa",
        guest_name: "Guest",
        check_in_date: "",
        check_out_date: "",
        nights: 0,
        welcome_message: "Welcome to your villa",
        has_reservation: false,
      });
    } finally {
      setLoading(false);
    }
  }

  const hour = currentTime.getHours();
  const isNight = hour >= 18 || hour < 6;

  if (showYoutube) {
    return (
      <div className="fixed inset-0 bg-black flex flex-col">
        <div className="flex items-center justify-between px-6 py-3 bg-black/80 border-b border-white/10">
          <div className="flex items-center gap-2 text-white">
            <Youtube className="h-5 w-5 text-red-500" />
            <span className="font-medium">YouTube</span>
          </div>
          <button
            onClick={() => setShowYoutube(false)}
            className="text-sm text-white/70 hover:text-white px-3 py-1 rounded border border-white/20"
          >
            Back to Welcome
          </button>
        </div>
        <div className="flex-1">
          <iframe
            src="https://www.youtube.com/embed?listType=PL&list=PLrAXtmErZgOeiKm4sgPOknG-ih2TwDwXt&autoplay=0"
            className="w-full h-full"
            allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
            allowFullScreen
          />
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-950 flex items-center justify-center">
        <div className="text-center space-y-4">
          <Tv className="h-16 w-16 text-nexus-400 mx-auto animate-pulse" />
          <p className="text-xl text-white/70">Loading your welcome screen...</p>
        </div>
      </div>
    );
  }

  return (
    <div className={`min-h-screen flex flex-col transition-colors ${isNight ? "bg-slate-950" : "bg-slate-900"}`}>
      {/* Header Bar */}
      <div className="flex items-center justify-between px-8 py-4 border-b border-white/10">
        <div className="flex items-center gap-3">
          <Home className="h-6 w-6 text-nexus-400" />
          <span className="text-lg font-bold text-white">{welcome?.villa_name || "Villa"}</span>
        </div>
        <div className="flex items-center gap-4 text-white/70">
          {isNight ? <Moon className="h-5 w-5" /> : <Sun className="h-5 w-5 text-amber-400" />}
          <span className="text-2xl font-mono">
            {currentTime.toLocaleTimeString("en-US", { hour: "2-digit", minute: "2-digit", hour12: false })}
          </span>
          <span className="text-sm">
            {currentTime.toLocaleDateString("en-US", { weekday: "long", month: "long", day: "numeric" })}
          </span>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 flex flex-col items-center justify-center px-8 py-12">
        <div className="text-center space-y-8 max-w-2xl">
          {/* Greeting */}
          <div className="space-y-2">
            <h1 className="text-5xl md:text-7xl font-bold text-white leading-tight">
              {welcome?.has_reservation ? (
                <>
                  Welcome, <span className="text-nexus-400">{welcome.guest_name}</span>
                </>
              ) : (
                "Welcome"
              )}
            </h1>
            <p className="text-xl text-white/60">
              {isNight ? "Good Evening" : hour < 12 ? "Good Morning" : "Good Afternoon"}
            </p>
          </div>

          {/* Stay Details */}
          {welcome?.has_reservation && (
            <div className="flex items-center justify-center gap-6 text-white/70">
              <div className="flex items-center gap-2">
                <Calendar className="h-5 w-5 text-nexus-400" />
                <span className="text-lg">
                  {new Date(welcome.check_in_date).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
                  {" — "}
                  {new Date(welcome.check_out_date).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
                </span>
              </div>
              <div className="h-4 w-px bg-white/20" />
              <span className="text-lg">{welcome.nights} nights</span>
            </div>
          )}

          {/* Message */}
          <p className="text-lg text-white/50 max-w-lg mx-auto">
            {welcome?.welcome_message || "Enjoy your stay. Your smart TV is ready."}
          </p>

          {/* YouTube Button */}
          <button
            onClick={() => setShowYoutube(true)}
            className="group inline-flex items-center gap-3 bg-red-600 hover:bg-red-500 text-white px-8 py-4 rounded-xl text-xl font-medium transition-all hover:scale-105 active:scale-95"
          >
            <Youtube className="h-8 w-8" />
            <span>Watch YouTube</span>
          </button>
        </div>
      </div>

      {/* Footer */}
      <div className="px-8 py-4 border-t border-white/10 text-center text-sm text-white/30">
        Press any button on your remote to begin · Villa IPTV System
      </div>
    </div>
  );
}
