"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useApi } from "@/hooks/useApi";
import { Sidebar } from "@/components/sidebar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { TrendingUp, ChevronLeft, ChevronRight, DollarSign } from "lucide-react";

interface DailyRate {
  date: string;
  base_rate: number;
  derived_rate: number;
  available_rooms: number;
  total_rooms: number;
}

export default function RatesPage() {
  const { get } = useApi();
  const [currentDate, setCurrentDate] = useState(new Date());
  const [planId, setPlanId] = useState("default");

  const startOfMonth = new Date(currentDate.getFullYear(), currentDate.getMonth(), 1);
  const endOfMonth = new Date(currentDate.getFullYear(), currentDate.getMonth() + 1, 0);

  const { data: rates = [], isLoading } = useQuery<DailyRate[]>({
    queryKey: ["daily-rates", planId, currentDate.getMonth()],
    queryFn: () =>
      get(
        `/rate-plans/${planId}/daily-rates?start=${startOfMonth.toISOString().split("T")[0]}&end=${endOfMonth.toISOString().split("T")[0]}`
      ),
  });

  const monthName = currentDate.toLocaleDateString("en-US", { month: "long", year: "numeric" });

  const daysInMonth = endOfMonth.getDate();
  const firstDayOfWeek = startOfMonth.getDay();

  const calendarDays = Array.from({ length: firstDayOfWeek }, () => null).concat(
    Array.from({ length: daysInMonth }, (_, i) => i + 1)
  );

  const rateMap = new Map(rates.map((r) => [r.date, r]));

  const prevMonth = () => setCurrentDate(new Date(currentDate.getFullYear(), currentDate.getMonth() - 1, 1));
  const nextMonth = () => setCurrentDate(new Date(currentDate.getFullYear(), currentDate.getMonth() + 1, 1));

  return (
    <div className="flex min-h-screen bg-background">
      <Sidebar />
      <main className="flex-1 p-6">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Rate Calendar</h1>
            <p className="text-sm text-muted-foreground">Manage pricing and availability</p>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" size="sm" onClick={prevMonth}>
              <ChevronLeft className="h-4 w-4" />
            </Button>
            <span className="min-w-[140px] text-center text-sm font-medium">{monthName}</span>
            <Button variant="outline" size="sm" onClick={nextMonth}>
              <ChevronRight className="h-4 w-4" />
            </Button>
          </div>
        </div>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between pb-2">
            <CardTitle className="text-sm">Daily Rates & Availability</CardTitle>
            <select
              value={planId}
              onChange={(e) => setPlanId(e.target.value)}
              className="h-8 rounded-md border border-input bg-background px-2 text-sm outline-none"
            >
              <option value="default">Default Plan</option>
              <option value="weekend">Weekend Special</option>
              <option value="corporate">Corporate Rate</option>
            </select>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="py-8 text-center text-muted-foreground">Loading rates...</div>
            ) : (
              <div className="grid grid-cols-7 gap-1 text-center text-sm">
                {["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"].map((d) => (
                  <div key={d} className="py-2 font-medium text-muted-foreground">
                    {d}
                  </div>
                ))}
                {calendarDays.map((day, i) => {
                  if (day === null) return <div key={`empty-${i}`} />;
                  const dateStr = `${currentDate.getFullYear()}-${String(currentDate.getMonth() + 1).padStart(2, "0")}-${String(day).padStart(2, "0")}`;
                  const rate = rateMap.get(dateStr);
                  const isWeekend = new Date(dateStr).getDay() === 0 || new Date(dateStr).getDay() === 6;

                  return (
                    <div
                      key={day}
                      className={`min-h-[80px] rounded-md border p-2 transition-colors hover:border-nexus-400 ${
                        isWeekend ? "bg-amber-50/50" : "bg-background"
                      }`}
                    >
                      <div className="font-medium">{day}</div>
                      {rate ? (
                        <div className="mt-1 space-y-0.5">
                          <div className="flex items-center gap-1 text-xs font-semibold text-nexus-700">
                            <DollarSign className="h-3 w-3" />
                            {rate.derived_rate}
                          </div>
                          <div className="text-xs text-muted-foreground">
                            {rate.available_rooms}/{rate.total_rooms} avail
                          </div>
                        </div>
                      ) : (
                        <div className="mt-1 text-xs text-muted-foreground">No rate</div>
                      )}
                    </div>
                  );
                })}
              </div>
            )}
          </CardContent>
        </Card>
      </main>
    </div>
  );
}
