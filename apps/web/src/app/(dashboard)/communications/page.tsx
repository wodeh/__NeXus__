"use client";

import { useState, useEffect } from "react";
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
import {
  getCommTemplates,
  getCommSequences,
  getCommScheduled,
  CommTemplate,
  CommSequence,
  CommScheduled,
} from "@/lib/api";

export default function CommunicationsPage() {
  const { config } = useTenant();
  const [activeTab, setActiveTab] = useState<"templates" | "sequences" | "scheduled">("templates");
  const [selectedTemplate, setSelectedTemplate] = useState<string | null>(null);
  const [templates, setTemplates] = useState<CommTemplate[]>([]);
  const [sequences, setSequences] = useState<CommSequence[]>([]);
  const [scheduled, setScheduled] = useState<CommScheduled[]>([]);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    setLoading(true);
    try {
      const [t, s, sc] = await Promise.all([
        getCommTemplates(),
        getCommSequences(),
        getCommScheduled(),
      ]);
      setTemplates(t);
      setSequences(s);
      setScheduled(sc);
    } catch (e) {
      console.error("Failed to load communications data", e);
    } finally {
      setLoading(false);
    }
  }

  if (!hasCapability(config, CAPABILITIES.REVENUE.COMMUNICATIONS)) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <Mail className="h-16 w-16 text-slate-600" />
        <h2 className="mt-4 text-xl font-bold text-white">Communications</h2>
        <p className="mt-2 text-slate-400">Upgrade to Revenue tier to manage guest communications.</p>
      </div>
    );
  }

  const totalScheduled = scheduled.length;
  const sentToday = scheduled.filter((s) => s.status === "sent" || s.status === "delivered").length;
  const failedToday = scheduled.filter((s) => s.status === "failed").length;

  const selectedTemplateData = templates.find((t) => t.id === selectedTemplate);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Communications</h1>
          <p className="text-sm text-slate-400">Templates · Sequences · Scheduled messages · Logs</p>
        </div>
      </div>

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

      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        <button onClick={() => setActiveTab("templates")} className={`rounded-lg px-4 py-2 text-sm font-medium ${activeTab === "templates" ? "bg-slate-800 text-white" : "text-slate-400"}`}>Templates</button>
        <button onClick={() => setActiveTab("sequences")} className={`rounded-lg px-4 py-2 text-sm font-medium ${activeTab === "sequences" ? "bg-slate-800 text-white" : "text-slate-400"}`}>Sequences</button>
        <button onClick={() => setActiveTab("scheduled")} className={`rounded-lg px-4 py-2 text-sm font-medium ${activeTab === "scheduled" ? "bg-slate-800 text-white" : "text-slate-400"}`}>Scheduled</button>
      </div>

      {loading ? (
        <p className="text-sm text-slate-500">Loading communications data...</p>
      ) : (
        <>
          {activeTab === "templates" && (
            <div className="grid grid-cols-3 gap-4">
              <div className="col-span-1 card max-h-[600px] overflow-y-auto p-0">
                <div className="sticky top-0 border-b border-slate-700 bg-slate-800/50 p-3 flex items-center justify-between">
                  <h3 className="text-sm font-semibold text-white">Templates</h3>
                  <button className="text-nexus-400"><Plus className="h-4 w-4" /></button>
                </div>
                <div className="divide-y divide-slate-800">
                  {templates.map((t) => (
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
                {selectedTemplateData ? (
                  <>
                    <div className="flex items-center justify-between">
                      <h3 className="text-lg font-semibold text-white">{selectedTemplateData.name}</h3>
                      <div className="flex gap-2">
                        <button className="btn-secondary text-xs gap-1"><Edit3 className="h-3 w-3" /> Edit</button>
                        <button className="rounded bg-rose-500/10 p-2 text-rose-400 hover:bg-rose-500/20"><Trash2 className="h-4 w-4" /></button>
                      </div>
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm text-slate-400">Subject</label>
                      <input className="input text-sm" value={selectedTemplateData.subject || ""} readOnly />
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm text-slate-400">Body</label>
                      <textarea className="input min-h-[200px] text-sm font-mono" value={selectedTemplateData.body} readOnly />
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-slate-400">Channel:</span>
                      <span className="rounded bg-slate-800 px-2 py-0.5 text-xs text-slate-300">{selectedTemplateData.channel}</span>
                    </div>
                    <div className="flex items-center gap-2">
                      <span className="text-xs text-slate-400">Category:</span>
                      <span className="rounded bg-slate-800 px-2 py-0.5 text-xs text-slate-300">{selectedTemplateData.category}</span>
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

          {activeTab === "sequences" && (
            <div className="space-y-4">
              <div className="flex justify-end">
                <button className="btn-primary gap-2 text-sm"><Plus className="h-4 w-4" /> New Sequence</button>
              </div>
              <div className="grid grid-cols-2 gap-4">
                {sequences.map((seq) => (
                  <div key={seq.id} className="card space-y-4">
                    <div className="flex items-center justify-between">
                      <div>
                        <h3 className="text-lg font-semibold text-white">{seq.name}</h3>
                        <p className="text-sm text-slate-400">{seq.description}</p>
                      </div>
                      <span className={`rounded px-2 py-0.5 text-xs ${seq.is_active ? "bg-emerald-500/10 text-emerald-400" : "bg-slate-500/10 text-slate-400"}`}>
                        {seq.is_active ? "Active" : "Paused"}
                      </span>
                    </div>
                    <div className="flex items-center gap-4 text-sm text-slate-400">
                      <span className="flex items-center gap-1"><Clock className="h-4 w-4" /> Trigger: {seq.trigger}</span>
                      <span className="flex items-center gap-1"><Mail className="h-4 w-4" /> {seq.steps} steps</span>
                    </div>
                    <div className="flex gap-2">
                      <button className="btn-secondary flex-1 text-xs gap-1"><Edit3 className="h-3 w-3" /> Edit</button>
                      <button className={`flex-1 rounded px-3 py-2 text-xs font-medium gap-1 inline-flex items-center justify-center ${seq.is_active ? "bg-amber-500/10 text-amber-400" : "bg-emerald-500/10 text-emerald-400"}`}>
                        {seq.is_active ? <><Pause className="h-3 w-3" /> Pause</> : <><Play className="h-3 w-3" /> Activate</>}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

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
                  {scheduled.map((s) => (
                    <tr key={s.id} className="hover:bg-slate-800/50">
                      <td className="px-4 py-3 text-white">{s.guest_name}</td>
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
                      <td className="px-4 py-3 text-slate-400">{s.scheduled_at?.split("T")?.join(" ")?.substring(0, 16) || "—"}</td>
                      <td className="px-4 py-3 text-slate-400">{s.sent_at?.split("T")?.join(" ")?.substring(0, 16) || "—"}</td>
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
        </>
      )}
    </div>
  );
}
