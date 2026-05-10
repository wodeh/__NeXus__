"use client";

import { useState } from "react";
import { ShieldCheck, Search, FileText, Trash2, Download } from "lucide-react";

const mockAudit = [
  { id: "a-001", action: "reservation_created", resource: "Reservation r-001", user: "frontdesk@demo.com", time: "2026-05-10 08:23:00" },
  { id: "a-002", action: "guest_updated", resource: "Guest g-002", user: "manager@demo.com", time: "2026-05-10 09:15:00" },
  { id: "a-003", action: "room_status_changed", resource: "Room 412", user: "housekeeping@demo.com", time: "2026-05-10 10:05:00" },
  { id: "a-004", action: "pricing_rule_applied", resource: "Recommendation rec-001", user: "revenue@demo.com", time: "2026-05-10 11:30:00" },
];

export default function AuditPage() {
  const [search, setSearch] = useState("");
  const filtered = mockAudit.filter((a) => a.action.includes(search.toLowerCase()) || a.resource.toLowerCase().includes(search.toLowerCase()));

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

      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
        <input className="input w-full max-w-md pl-9" placeholder="Search audit logs..." value={search} onChange={(e) => setSearch(e.target.value)} />
      </div>

      <div className="card overflow-hidden p-0">
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-800 text-xs uppercase text-slate-400">
            <tr>
              <th className="px-4 py-3">Action</th>
              <th className="px-4 py-3">Resource</th>
              <th className="px-4 py-3">User</th>
              <th className="px-4 py-3">Time</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            {filtered.map((a) => (
              <tr key={a.id} className="hover:bg-slate-800/50">
                <td className="px-4 py-3">
                  <div className="flex items-center gap-2">
                    <ShieldCheck className="h-4 w-4 text-nexus-400" />
                    <span className="font-medium text-white">{a.action.replace(/_/g, " ")}</span>
                  </div>
                </td>
                <td className="px-4 py-3 text-slate-300">{a.resource}</td>
                <td className="px-4 py-3 text-slate-300">{a.user}</td>
                <td className="px-4 py-3 text-slate-400">{a.time}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
