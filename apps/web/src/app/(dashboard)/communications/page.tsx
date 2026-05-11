"use client";

import { useState } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import {
  Mail,
  MessageSquare,
  Smartphone,
  Send,
  Plus,
  Trash2,
  Copy,
  Clock,
  CheckCircle2,
  XCircle,
  AlertTriangle,
  CalendarDays,
  Eye,
  Edit3,
  Play,
  Pause,
  Globe,
  User,
  FileText,
} from "lucide-react";

const mockTemplates = [
  { id: "t1", name: "Booking Confirmation", subject: "Your reservation is confirmed", body: "Dear {{guest_name}},\n\nYour reservation at {{property_name}} is confirmed.\n\nDates: {{check_in}} - {{check_out}}\nRoom: {{room_number}}\n\nWe look forward to welcoming you!", channel: "email", category: "booking", isActive: true },
  { id: "t2", name: "Pre-Arrival", subject: "Your stay is coming up", body: "Hi {{guest_name}},\n\nYour check-in at {{property_name}} is tomorrow.\n\nCheck-in time: 3:00 PM\nWiFi: NexusGuest / Password: Welcome2026\n\nNeed anything? Reply to this email or call us.", channel: "email", category: "pre_arrival", isActive: true },
  { id: "t3", name: "Check-In Instructions", subject: "", body: "Welcome {{guest_name}}! Your room {{room_number}} is ready. Digital key: {{digital_key}}. Enjoy your stay!", channel: "sms", category: "checkin", isActive: true },
  { id: "t4", name: "Review Request", subject: "How was your stay?", body: "Hi {{guest_name}},\n\nThank you for staying with us! We'd love to hear about your experience.\n\nLeave a review: {{review_link}}\n\nIt takes less than 2 minutes.", channel: "email", category: "review", isActive: true },
  { id: "t5", name: "WhatsApp Welcome", subject: "", body: "Welcome to {{property_name}}! 🎉\nYour room {{room_number}} is ready.\n\nNeed help? Reply with:\n• BOOK - for future stays\n• CHECKIN - check-in info\n• WIFI - password\n• HELP - speak to front desk", channel: "whatsapp", category: "stay", isActive: true },
];

const mockSequences = [
  { id: "s1", name: "Guest Journey", description: "Full guest communication flow", trigger: "booking_created", isActive: true, steps: 5 },
  { id: "s2", name: "Post-Stay Review", description: "Review request sequence", trigger: "checkout", isActive: true, steps: 3 },
];

const mockScheduled = [
  { id: "sc1", guestName: "Alice Chen", channel: "email", subject: "Your reservation is confirmed", status: "scheduled", scheduledAt: "2026-05-10 15:00" },
  { id: "sc2", guestName: "Bob Jones", channel: "sms", subject: "", body: "Welcome! Your room 102 is ready.", status: "sent", scheduledAt: "2026-05-10 14:30", sentAt: "2026-05-10 14:30" },
  { id: "sc3", guestName: "Carol White", channel: "email", subject: "How was your stay?", status: "delivered", scheduledAt: "2026-05-11 10:00", sentAt: "2026-05-11 10:00" },
  { id: "sc4", guestName: "David Kim", channel: "whatsapp", subject: "", body: "Check-in reminder: tomorrow 3PM", status: "failed", scheduledAt: "2026-05-10 12:00", error: "Invalid phone number" },
];

export default function CommunicationsPage() {
  const { tenantConfig } = useTenant();
  const [activeTab, setActiveTab] = useState<"templates" | "sequences" | "scheduled">("templates");
  const [selectedTemplate, setSelectedTemplate] = useState<string | null>(null);

  if (!hasCapability(tenantConfig, CAPABILITIES.REVENUE.COMMUNICATIONS)) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <Mail className="h-16 w-16 text-slate-600" />
        <h2 className="mt-4 text-xl font-bold text-white">Communications</h2>
        <p className="mt-2 text-slate-400">Upgrade to Revenue tier to manage guest communications.</p>
      </div>
    );
  }

  const totalScheduled = mockScheduled.length;
  const sentToday = mockScheduled.filter((s) => s.status === "sent" || s.status === "delivered").length;
  const failedToday = mockScheduled.filter((s) => s.status === "failed").length;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Communications</h1>
          <p className="text-sm text-slate-400">Templates · Sequences · Scheduled messages · Logs</p>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><Mail className="h-4 w-4" /> Scheduled</div>
          <p className="text-2xl font-bold text-white">{totalScheduled}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><CheckCircle2 className="h-4 w-4" /> Sent Today</div>
          <p className="text-2xl font-bold text-white">{sentToday}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><XCircle className="h-4 w-4" /> Failed</div>
          <p className="text-2xl font-bold text-white">{failedToday}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><Globe className="h-4 w-4" /> Open Rate</div>
          <p className="text-2xl font-bold text-white">68%</p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        <button onClick={() => setActiveTab("templates")} className={`rounded-lg px-4 py-2 text-sm font-medium ${activeTab === "templates" ? "bg-slate-800 text-white" : "text-slate-400"}`}>Templates</button>
        <button onClick={() => setActiveTab("sequences")} className={`rounded-lg px-4 py-2 text-sm font-medium ${activeTab === "sequences" ? "bg-slate-800 text-white" : "text-slate-400"}`}>Sequences</button>
        <button onClick={() => setActiveTab("scheduled")} className={`rounded-lg px-4 py-2 text-sm font-medium ${activeTab === "scheduled" ? "bg-slate-800 text-white" : "text-slate-400"}`}>Scheduled</button>
      </div>

      {/* Templates Tab */}
      {activeTab === "templates" && (
        <div className="grid grid-cols-3 gap-4">
          <div className="col-span-1 card max-h-[600px] overflow-y-auto p-0">
            <div className="sticky top-0 border-b border-slate-700 bg-slate-800/50 p-3 flex items-center justify-between">
              <h3 className="text-sm font-semibold text-white">Templates</h3>
              <button className="text-nexus-400"><Plus className="h-4 w-4" /></button>
            </div>
            <div className="divide-y divide-slate-800">
              {mockTemplates.map((t) => (
                <button
                  key={t.id}
                  onClick={() => setSelectedTemplate(t.id)}
                  className={`w-full p-3 text-left transition-colors ${selectedTemplate === t.id ? "bg-nexus-500/10" : "hover:bg-slate-800/50"}`}
                >
                  <div className="flex items-center gap-2">
                    {t.channel === "email" && <Mail className="h-4 w-4 text-slate-400" />}
                    {t.channel === "sms" && <Smartphone className="h-4 w-4 text-slate-400" />}
                    {t.channel === "whatsapp" && <MessageSquare className="h-4 w-4 text-slate-400" />}
                    <span className="text-sm font-medium text-white">{t.name}</span>
                  </div>
                  <p className="text-xs text-slate-500 mt-1">{t.category}</p>
                </button>
              ))}
            </div>
          </div>

          <div className="col-span-2 card space-y-4">
            {selectedTemplate ? (
              <>
                <div className="flex items-center justify-between">
                  <h3 className="text-lg font-semibold text-white">{mockTemplates.find((t) => t.id === selectedTemplate)?.name}</h3>
                  <div className="flex gap-2">
                    <button className="btn-secondary text-xs gap-1"><Edit3 className="h-3 w-3" /> Edit</button>
                    <button className="rounded bg-rose-500/10 p-2 text-rose-400 hover:bg-rose-500/20"><Trash2 className="h-4 w-4" /></button>
                  </div>
                </div>
                <div className="space-y-2">
                  <label className="text-sm text-slate-400">Subject</label>
                  <input className="input text-sm" value={mockTemplates.find((t) => t.id === selectedTemplate)?.subject || ""} readOnly />
                </div>
                <div className="space-y-2">
                  <label className="text-sm text-slate-400">Body</label>
                  <textarea className="input min-h-[200px] text-sm font-mono" value={mockTemplates.find((t) => t.id === selectedTemplate)?.body || ""} readOnly />
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-slate-400">Channel:</span>
                  <span className="rounded bg-slate-800 px-2 py-0.5 text-xs text-slate-300">
                    {mockTemplates.find((t) => t.id === selectedTemplate)?.channel}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-slate-400">Category:</span>
                  <span className="rounded bg-slate-800 px-2 py-0.5 text-xs text-slate-300">
                    {mockTemplates.find((t) => t.id === selectedTemplate)?.category}
                  </span>
                </div>
                <div className="flex gap-2">
                  <button className="btn-primary gap-1 text-sm"><Send className="h-4 w-4" /> Send Test</button>
                  <button className="btn-secondary gap-1 text-sm"><Copy className="h-4 w-4" /> Duplicate</button>
                </div>
              </>
            ) : (
              <div className="flex flex-col items-center justify-center py-20">
                <FileText className="h-12 w-12 text-slate-600" />
                <p className="mt-4 text-slate-400">Select a template to view details</p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Sequences Tab */}
      {activeTab === "sequences" && (
        <div className="space-y-4">
          <div className="flex justify-end">
            <button className="btn-primary gap-2 text-sm"><Plus className="h-4 w-4" /> New Sequence</button>
          </div>
          <div className="grid grid-cols-2 gap-4">
            {mockSequences.map((seq) => (
              <div key={seq.id} className="card space-y-4">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-lg font-semibold text-white">{seq.name}</h3>
                    <p className="text-sm text-slate-400">{seq.description}</p>
                  </div>
                  <span className={`rounded px-2 py-0.5 text-xs ${seq.isActive ? "bg-emerald-500/10 text-emerald-400" : "bg-slate-500/10 text-slate-400"}`}>
                    {seq.isActive ? "Active" : "Paused"}
                  </span>
                </div>
                <div className="flex items-center gap-4 text-sm text-slate-400">
                  <span className="flex items-center gap-1"><Clock className="h-4 w-4" /> Trigger: {seq.trigger}</span>
                  <span className="flex items-center gap-1"><Mail className="h-4 w-4" /> {seq.steps} steps</span>
                </div>
                <div className="flex gap-2">
                  <button className="btn-secondary flex-1 text-xs gap-1"><Edit3 className="h-3 w-3" /> Edit</button>
                  <button className={`flex-1 rounded px-3 py-2 text-xs font-medium gap-1 inline-flex items-center justify-center ${seq.isActive ? "bg-amber-500/10 text-amber-400" : "bg-emerald-500/10 text-emerald-400"}`}>
                    {seq.isActive ? <><Pause className="h-3 w-3" /> Pause</> : <><Play className="h-3 w-3" /> Activate</>}
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Scheduled Tab */}
      {activeTab === "scheduled" && (
        <div className="card space-y-4">
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-800 text-xs uppercase text-slate-400">
              <tr>
                <th className="px-4 py-3">Guest</th>
                <th className="px-4 py-3">Channel</th>
                <th className="px-4 py-3">Subject/Preview</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Scheduled</th>
                <th className="px-4 py-3">Sent</th>
                <th className="px-4 py-3 text-right">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {mockScheduled.map((s) => (
                <tr key={s.id} className="hover:bg-slate-800/50">
                  <td className="px-4 py-3 text-white">{s.guestName}</td>
                  <td className="px-4 py-3">
                    <span className="flex items-center gap-1 text-slate-400">
                      {s.channel === "email" && <Mail className="h-3 w-3" />}
                      {s.channel === "sms" && <Smartphone className="h-3 w-3" />}
                      {s.channel === "whatsapp" && <MessageSquare className="h-3 w-3" />}
                      {s.channel}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-300 max-w-xs truncate">{s.subject || s.body}</td>
                  <td className="px-4 py-3">
                    <span className={`rounded px-2 py-0.5 text-xs ${
                      s.status === "delivered" ? "bg-emerald-500/10 text-emerald-400" :
                      s.status === "sent" ? "bg-nexus-500/10 text-nexus-400" :
                      s.status === "scheduled" ? "bg-sky-500/10 text-sky-400" :
                      "bg-rose-500/10 text-rose-400"
                    }`}>
                      {s.status}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-slate-400">{s.scheduledAt}</td>
                  <td className="px-4 py-3 text-slate-400">{s.sentAt || "—"}</td>
                  <td className="px-4 py-3 text-right">
                    {s.status === "scheduled" && (
                      <button className="text-xs text-rose-400 hover:text-rose-300">Cancel</button>
                    )}
                    {s.status === "failed" && (
                      <span className="text-xs text-rose-400">{s.error}</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
