"use client";

import { useState, useEffect } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import { ShieldCheck, Search, FileText, Download, Lock, ArrowUpRight, Activity, User, Server, AlertCircle } from "lucide-react";
import { getAuditLogs, getAuditStats, AuditLog, AuditStats } from "@/lib/api";

export default function AuditPage() {
  const { config } = useTenant();
  const [search, setSearch] = useState("");
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [stats, setStats] = useState<AuditStats | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    setLoading(true);
    try {
      const [l, s] = await Promise.all([
        getAuditLogs(),
        getAuditStats(),
      ]);
      setLogs(l);
      setStats(s);
    } catch (e) {
      console.error("Failed to load audit data", e);
    } finally {
      setLoading(false);
    }
  }

  const hasAudit = hasCapability(config, CAPABILITIES.CORE.AUDIT_LOGS);

  if (!hasAudit) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-center">
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-slate-800">
          <Lock className="h-8 w-8 text-slate-500" />
        </div>
        <h2 className="mt-6 text-xl font-bold text-white">Audit Logs Locked</h2>
        <p className="mt-2 max-w-md text-sm text-slate-400">
          Audit logging requires a higher license tier.
        </p>
        <button className="btn-primary mt-6">
          <ArrowUpRight className="h-4 w-4" />
          Upgrade License
        </button>
      </div>
    );
  }

  const filteredLogs = logs.filter((l) =>
    l.action.toLowerCase().includes(search.toLowerCase()) ||
    (l.user_email || "").toLowerCase().includes(search.toLowerCase()) ||
    l.resource.toLowerCase().includes(search.toLowerCase())
  );

  const actionIcons: Record<string, React.ReactNode> = {
    reservation_created: <FileText className="h-4 w-4 text-emerald-400" />,
    guest_checkin: <User className="h-4 w-4 text-nexus-400" />,
    room_status_updated: <Activity className="h-4 w-4 text-sky-400" />,
    rate_plan_modified: <FileText className="h-4 w-4 text-amber-400" />,
    system_backup: <Server className="h-4 w-4 text-slate-400" />,
    channel_sync: <ShieldCheck className="h-4 w-4 text-emerald-400" />,
  };

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-white">Audit Logs</h2>
          <p className="text-sm text-slate-400">Activity tracking and compliance</p>
        </div>
        <div className="flex gap-2">
          <button className="btn-secondary">
            <Download className="h-4 w-4" /> Export
          </button>
          <button className="btn-secondary">
            <FileText className="h-4 w-4" /> GDPR Request
          </button>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><Activity className="h-4 w-4" /> Total Events</div>
          <p className="text-2xl font-bold text-white">{stats?.total_events || 0}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><ShieldCheck className="h-4 w-4" /> Today</div>
          <p className="text-2xl font-bold text-white">{stats?.today_events || 0}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><User className="h-4 w-4" /> User Actions</div>
          <p className="text-2xl font-bold text-white">{stats?.user_actions || 0}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><AlertCircle className="h-4 w-4" /> Failed</div>
          <p className="text-2xl font-bold text-white">{stats?.failed_actions || 0}</p>
        </div>
      </div>

      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
        <input className="input w-full max-w-md pl-9" placeholder="Search audit logs..." value={search} onChange={(e) => setSearch(e.target.value)} />
      </div>

      <div className="card">
        {loading ? (
          <p className="text-sm text-slate-500">Loading audit logs...</p>
        ) : filteredLogs.length === 0 ? (
          <p className="text-sm text-slate-500">No audit logs match your search.</p>
        ) : (
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-800 text-xs uppercase text-slate-400">
              <tr>
                <th className="px-4 py-3">Time</th>
                <th className="px-4 py-3">User</th>
                <th className="px-4 py-3">Action</th>
                <th className="px-4 py-3">Resource</th>
                <th className="px-4 py-3">Details</th>
                <th className="px-4 py-3">IP</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {filteredLogs.map((log) => (
                <tr key={log.id} className="hover:bg-slate-800/50">
                  <td className="px-4 py-3 text-slate-400 whitespace-nowrap">{log.created_at?.split("T")?.join(" ")?.substring(0, 16) || ""}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      {log.user_email ? (
                        <>
                          <User className="h-3 w-3 text-slate-400" />
                          <span className="text-white">{log.user_email}</span>
                        </>
                      ) : (
                        <>
                          <Server className="h-3 w-3 text-slate-400" />
                          <span className="text-slate-400">System</span>
                        </>
                      )}
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      {actionIcons[log.action] || <Activity className="h-4 w-4 text-slate-400" />}
                      <span className="text-white capitalize">{log.action.replace(/_/g, " ")}</span>
                    </div>
                  </td>
                  <td className="px-4 py-3">
                    <span className="rounded bg-slate-700 px-2 py-0.5 text-xs text-slate-300">{log.resource}</span>
                  </td>
                  <td className="px-4 py-3 text-slate-400 max-w-xs truncate">
                    {log.details ? JSON.stringify(log.details) : "—"}
                  </td>
                  <td className="px-4 py-3 text-slate-500">{log.ip_address || "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
}
