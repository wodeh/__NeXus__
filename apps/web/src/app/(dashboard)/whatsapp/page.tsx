"use client";

import { useState, useEffect } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import {
  MessageCircle,
  Settings,
  Send,
  Bot,
  User,
  CheckCircle2,
  XCircle,
  Clock,
  RefreshCw,
  ToggleLeft,
  ToggleRight,
  QrCode,
  ArrowRight,
  Sparkles,
  Phone,
} from "lucide-react";
import {
  getWhatsAppConversations,
  getWhatsAppMessages,
  sendWhatsAppMessage,
  getWhatsAppConfig,
  updateWhatsAppConfig,
  getWhatsAppTemplates,
  WhatsAppConversation,
  WhatsAppMessage,
  WhatsAppBotConfig,
  WhatsAppTemplate,
} from "@/lib/api";

export default function WhatsAppPage() {
  const { config } = useTenant();
  const [activeTab, setActiveTab] = useState<"conversations" | "settings">("conversations");
  const [selectedConv, setSelectedConv] = useState<string | null>(null);
  const [conversations, setConversations] = useState<WhatsAppConversation[]>([]);
  const [messages, setMessages] = useState<WhatsAppMessage[]>([]);
  const [botConfig, setBotConfig] = useState<WhatsAppBotConfig | null>(null);
  const [templates, setTemplates] = useState<WhatsAppTemplate[]>([]);
  const [loading, setLoading] = useState(false);
  const [sending, setSending] = useState(false);
  const [replyText, setReplyText] = useState("");

  useEffect(() => {
    loadConversations();
    loadConfig();
    loadTemplates();
  }, []);

  useEffect(() => {
    if (selectedConv) {
      loadMessages(selectedConv);
    }
  }, [selectedConv]);

  async function loadConversations() {
    try {
      const data = await getWhatsAppConversations();
      setConversations(data);
      if (data.length > 0 && !selectedConv) {
        setSelectedConv(data[0].id);
      }
    } catch (e) {
      console.error("Failed to load conversations", e);
    }
  }

  async function loadMessages(convId: string) {
    try {
      const data = await getWhatsAppMessages(convId);
      setMessages(data);
    } catch (e) {
      console.error("Failed to load messages", e);
    }
  }

  async function loadConfig() {
    try {
      const data = await getWhatsAppConfig();
      setBotConfig(data);
    } catch (e) {
      console.error("Failed to load config", e);
    }
  }

  async function loadTemplates() {
    try {
      const data = await getWhatsAppTemplates();
      setTemplates(data);
    } catch (e) {
      console.error("Failed to load templates", e);
    }
  }

  async function handleSendMessage() {
    if (!selectedConv || !replyText.trim()) return;
    setSending(true);
    try {
      await sendWhatsAppMessage(selectedConv, replyText.trim());
      setReplyText("");
      await loadMessages(selectedConv);
      await loadConversations();
    } catch (e) {
      console.error("Failed to send message", e);
    } finally {
      setSending(false);
    }
  }

  async function handleUpdateConfig(updates: Partial<WhatsAppBotConfig>) {
    try {
      const updated = await updateWhatsAppConfig(updates);
      setBotConfig(updated);
    } catch (e) {
      console.error("Failed to update config", e);
    }
  }

  if (!hasCapability(config, CAPABILITIES.REVENUE.WHATSAPP_BOT)) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <MessageCircle className="h-16 w-16 text-slate-600" />
        <h2 className="mt-4 text-xl font-bold text-white">WhatsApp Bot</h2>
        <p className="mt-2 text-slate-400">Upgrade to Revenue tier to enable WhatsApp automation.</p>
      </div>
    );
  }

  const activeConversations = conversations.filter((c) => c.current_state !== "complete").length;
  const bookingsToday = conversations.filter((c) => c.booking_created).length;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">WhatsApp Bot</h1>
          <p className="text-sm text-slate-400">Automated guest conversations · Booking via chat</p>
        </div>
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2">
            <span className="text-xs text-slate-400">Bot</span>
            <button 
              onClick={() => botConfig && handleUpdateConfig({ bot_enabled: !botConfig.bot_enabled })} 
              className="text-nexus-400"
            >
              {botConfig?.bot_enabled ? <ToggleRight className="h-6 w-6" /> : <ToggleLeft className="h-6 w-6 text-slate-600" />}
            </button>
          </div>
        </div>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4">
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><MessageCircle className="h-4 w-4" /> Active Conversations</div>
          <p className="text-2xl font-bold text-white">{activeConversations}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><CheckCircle2 className="h-4 w-4" /> Bookings Today</div>
          <p className="text-2xl font-bold text-white">{bookingsToday}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><Clock className="h-4 w-4" /> Avg Response</div>
          <p className="text-2xl font-bold text-white">12s</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><Sparkles className="h-4 w-4" /> Conversion Rate</div>
          <p className="text-2xl font-bold text-white">68%</p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        <button onClick={() => setActiveTab("conversations")} className={`rounded-lg px-4 py-2 text-sm font-medium ${activeTab === "conversations" ? "bg-slate-800 text-white" : "text-slate-400"}`}>Conversations</button>
        <button onClick={() => setActiveTab("settings")} className={`rounded-lg px-4 py-2 text-sm font-medium ${activeTab === "settings" ? "bg-slate-800 text-white" : "text-slate-400"}`}>Settings</button>
      </div>

      {activeTab === "conversations" && (
        <div className="grid grid-cols-3 gap-4">
          {/* Conversation List */}
          <div className="card max-h-[600px] overflow-y-auto p-0">
            <div className="sticky top-0 border-b border-slate-700 bg-slate-800/50 p-3">
              <h3 className="text-sm font-semibold text-white">Conversations</h3>
            </div>
            <div className="divide-y divide-slate-800">
              {conversations.map((conv) => (
                <button
                  key={conv.id}
                  onClick={() => setSelectedConv(conv.id)}
                  className={`w-full p-3 text-left transition-colors ${selectedConv === conv.id ? "bg-nexus-500/10" : "hover:bg-slate-800/50"}`}
                >
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium text-white">{conv.guest_name || conv.guest_phone}</span>
                    <span className="text-[10px] text-slate-500">{conv.last_message_at?.split("T")[1]?.substring(0,5) || ""}</span>
                  </div>
                  <div className="mt-1 flex items-center gap-2">
                    <span className={`rounded px-1.5 py-0.5 text-[10px] ${
                      conv.current_state === "complete" ? "bg-emerald-500/10 text-emerald-400" :
                      conv.current_state === "greeting" ? "bg-slate-500/10 text-slate-400" :
                      "bg-nexus-500/10 text-nexus-400"
                    }`}>
                      {conv.current_state}
                    </span>
                    {conv.booking_created && <span className="rounded bg-emerald-500/10 px-1.5 py-0.5 text-[10px] text-emerald-400">Booked</span>}
                  </div>
                  <p className="mt-1 text-xs text-slate-500">{conv.messages} messages</p>
                </button>
              ))}
            </div>
          </div>

          {/* Chat Area */}
          <div className="col-span-2 card flex flex-col max-h-[600px]">
            <div className="border-b border-slate-700 pb-3">
              <div className="flex items-center gap-2">
                <div className="h-8 w-8 rounded-full bg-nexus-500/20 flex items-center justify-center">
                  <User className="h-4 w-4 text-nexus-400" />
                </div>
                <div>
                  <p className="text-sm font-medium text-white">{conversations.find(c => c.id === selectedConv)?.guest_name || "Guest"}</p>
                  <p className="text-xs text-slate-400">{conversations.find(c => c.id === selectedConv)?.guest_phone || ""} · Active now</p>
                </div>
              </div>
            </div>

            <div className="flex-1 space-y-3 overflow-y-auto py-3">
              {messages.map((msg) => (
                <div key={msg.id} className={`flex ${msg.direction === "inbound" ? "justify-start" : "justify-end"}`}>
                  <div className={`max-w-[70%] rounded-lg px-3 py-2 ${
                    msg.direction === "inbound"
                      ? "bg-slate-800 text-slate-200"
                      : "bg-nexus-500/20 text-white"
                  }`}>
                    <p className="text-sm">{msg.body}</p>
                    <div className={`mt-1 flex items-center gap-1 ${msg.direction === "outbound" ? "justify-end" : ""}`}>
                      <span className="text-[10px] text-slate-500">{msg.sent_at?.split("T")[1]?.substring(0,5) || ""}</span>
                      {msg.direction === "outbound" && msg.status === "delivered" && <CheckCircle2 className="h-3 w-3 text-slate-500" />}
                      {msg.direction === "outbound" && msg.status === "read" && <CheckCircle2 className="h-3 w-3 text-nexus-400" />}
                    </div>
                  </div>
                </div>
              ))}
            </div>

            <div className="border-t border-slate-700 pt-3">
              <div className="flex gap-2">
                <input 
                  className="input flex-1" 
                  placeholder="Type a message..." 
                  value={replyText}
                  onChange={(e) => setReplyText(e.target.value)}
                  onKeyDown={(e) => e.key === "Enter" && handleSendMessage()}
                />
                <button 
                  className="btn-primary gap-2" 
                  onClick={handleSendMessage}
                  disabled={sending}
                >
                  <Send className="h-4 w-4" />
                </button>
              </div>
              <div className="mt-2 flex gap-2">
                {templates.slice(0,4).map((t) => (
                  <button 
                    key={t.id}
                    onClick={() => setReplyText(t.response)}
                    className="rounded bg-slate-800 px-2 py-1 text-[10px] text-slate-400 hover:text-white"
                  >
                    {t.trigger}
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>
      )}

      {activeTab === "settings" && (
        <div className="grid grid-cols-2 gap-6">
          <div className="card space-y-6">
            <h3 className="text-lg font-semibold text-white">Bot Configuration</h3>
            <div className="space-y-4">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-white">Enable WhatsApp Bot</p>
                  <p className="text-xs text-slate-400">Turn on automated responses</p>
                </div>
                <button 
                  onClick={() => botConfig && handleUpdateConfig({ bot_enabled: !botConfig.bot_enabled })} 
                  className="text-nexus-400"
                >
                  {botConfig?.bot_enabled ? <ToggleRight className="h-6 w-6" /> : <ToggleLeft className="h-6 w-6 text-slate-600" />}
                </button>
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-white">Booking via Chat</p>
                  <p className="text-xs text-slate-400">Allow guests to book rooms via WhatsApp</p>
                </div>
                <button 
                  onClick={() => botConfig && handleUpdateConfig({ booking_enabled: !botConfig.booking_enabled })} 
                  className="text-nexus-400"
                >
                  {botConfig?.booking_enabled ? <ToggleRight className="h-6 w-6" /> : <ToggleLeft className="h-6 w-6 text-slate-600" />}
                </button>
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-white">Auto-Reply</p>
                  <p className="text-xs text-slate-400">Instant responses to common questions</p>
                </div>
                <button 
                  onClick={() => botConfig && handleUpdateConfig({ auto_reply_enabled: !botConfig.auto_reply_enabled })} 
                  className="text-nexus-400"
                >
                  {botConfig?.auto_reply_enabled ? <ToggleRight className="h-6 w-6" /> : <ToggleLeft className="h-6 w-6 text-slate-600" />}
                </button>
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm text-slate-400">Welcome Message</label>
              <textarea 
                className="input min-h-[80px] text-sm" 
                value={botConfig?.welcome_message || ""}
                onChange={(e) => botConfig && handleUpdateConfig({ welcome_message: e.target.value })}
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm text-slate-400">Phone Number ID</label>
              <input 
                className="input text-sm" 
                placeholder="Enter Meta Phone Number ID" 
                value={botConfig?.phone_number_id || ""}
                onChange={(e) => botConfig && handleUpdateConfig({ phone_number_id: e.target.value })}
              />
            </div>

            <div className="space-y-2">
              <label className="text-sm text-slate-400">Access Token</label>
              <input className="input text-sm" type="password" placeholder="Enter Meta API token" />
            </div>
          </div>

          <div className="card space-y-6">
            <h3 className="text-lg font-semibold text-white">Webhook Setup</h3>
            <div className="rounded-lg bg-slate-800/50 p-4 space-y-3">
              <p className="text-sm text-slate-400">Configure this webhook URL in your Meta Developer Dashboard:</p>
              <div className="flex items-center gap-2 rounded bg-slate-900 p-3">
                <code className="flex-1 text-xs text-nexus-400">{botConfig?.webhook_url || "https://api.nexus.com/webhooks/whatsapp/demo"}</code>
                <button 
                  className="rounded bg-slate-800 px-2 py-1 text-xs text-white hover:bg-slate-700"
                  onClick={() => botConfig?.webhook_url && navigator.clipboard.writeText(botConfig.webhook_url)}
                >Copy</button>
              </div>
            </div>

            <div className="space-y-3">
              <h4 className="text-sm font-medium text-white">Quick Response Templates</h4>
              {templates.map((t) => (
                <div key={t.id} className="flex items-start gap-3 rounded bg-slate-800/50 p-3">
                  <span className="rounded bg-nexus-500/10 px-2 py-0.5 text-xs font-medium text-nexus-400">{t.trigger}</span>
                  <p className="text-sm text-slate-300">{t.response}</p>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
