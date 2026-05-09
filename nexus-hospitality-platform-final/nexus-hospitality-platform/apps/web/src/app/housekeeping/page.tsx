"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { useApi } from "@/hooks/useApi";
import { Sidebar } from "@/components/sidebar";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { CheckCircle2, Clock, AlertCircle, Sparkles } from "lucide-react";

interface HKTask {
  id: string;
  room_number: string;
  task_type: string;
  priority: string;
  status: string;
  assigned_to: string;
  estimated_minutes: number;
  due_time: string;
  notes: string;
}

export default function HousekeepingPage() {
  const { get } = useApi();
  const [statusFilter, setStatusFilter] = useState("all");

  const { data: tasks = [], isLoading } = useQuery<HKTask[]>({
    queryKey: ["housekeeping-tasks"],
    queryFn: () => get("/housekeeping-tasks?limit=200"),
  });

  const filtered = statusFilter === "all" ? tasks : tasks.filter((t) => t.status === statusFilter);

  const statusConfig: Record<string, { color: string; icon: any }> = {
    pending: { color: "bg-amber-100 text-amber-700", icon: Clock },
    assigned: { color: "bg-blue-100 text-blue-700", icon: AlertCircle },
    in_progress: { color: "bg-purple-100 text-purple-700", icon: Sparkles },
    completed: { color: "bg-green-100 text-green-700", icon: CheckCircle2 },
    inspected: { color: "bg-emerald-100 text-emerald-700", icon: CheckCircle2 },
  };

  const counts = {
    pending: tasks.filter((t) => t.status === "pending").length,
    assigned: tasks.filter((t) => t.status === "assigned").length,
    in_progress: tasks.filter((t) => t.status === "in_progress").length,
    completed: tasks.filter((t) => t.status === "completed").length,
    inspected: tasks.filter((t) => t.status === "inspected").length,
  };

  return (
    <div className="flex min-h-screen bg-background">
      <Sidebar />
      <main className="flex-1 p-6">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Housekeeping Board</h1>
            <p className="text-sm text-muted-foreground">Manage room cleaning and inspection tasks</p>
          </div>
          <Button>Assign Tasks</Button>
        </div>

        <div className="mb-4 flex flex-wrap gap-2">
          {Object.entries(counts).map(([status, count]) => (
            <button
              key={status}
              onClick={() => setStatusFilter(statusFilter === status ? "all" : status)}
              className={`rounded-md px-3 py-1.5 text-sm font-medium capitalize transition-colors ${
                statusFilter === status
                  ? statusConfig[status]?.color + " ring-2 ring-offset-1"
                  : "bg-muted text-muted-foreground hover:bg-muted/80"
              }`}
            >
              {status.replace("_", " ")}: {count}
            </button>
          ))}
        </div>

        <Card>
          <CardHeader>
            <CardTitle className="text-sm">Tasks ({filtered.length})</CardTitle>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="py-8 text-center text-muted-foreground">Loading tasks...</div>
            ) : filtered.length === 0 ? (
              <div className="py-8 text-center text-muted-foreground">No tasks found</div>
            ) : (
              <div className="space-y-2">
                {filtered.map((task) => {
                  const config = statusConfig[task.status] || statusConfig.pending;
                  const Icon = config.icon;
                  return (
                    <div
                      key={task.id}
                      className="flex items-center justify-between rounded-lg border p-3"
                    >
                      <div className="flex items-center gap-3">
                        <div className={`rounded-full p-2 ${config.color}`}>
                          <Icon className="h-4 w-4" />
                        </div>
                        <div>
                          <p className="font-medium">
                            Room {task.room_number} · {task.task_type}
                          </p>
                          <p className="text-sm text-muted-foreground">
                            {task.notes || "No notes"} · {task.estimated_minutes} min
                          </p>
                        </div>
                      </div>
                      <div className="text-right">
                        <Badge variant="outline" className={config.color}>
                          {task.status.replace("_", " ")}
                        </Badge>
                        <p className="mt-1 text-xs text-muted-foreground">
                          Due {task.due_time}
                        </p>
                        {task.assigned_to && (
                          <p className="text-xs text-muted-foreground">Assigned to {task.assigned_to}</p>
                        )}
                      </div>
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
