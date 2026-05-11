"use client";

import { useState } from "react";
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

const mockConversations = [
  { id: "w1", guestPhone: "+1-555-0101", guestName: "Alice Chen", currentState: "complete", lastMessageAt: "2026-05-10 14:30", messages: 8, bookingCreated: true },
  { id: "w2", guestPhone: "+1-555-0102", guestName: "Bob Jones", currentState: "ask_dates", lastMessageAt: "2026-05-10 13:15", messages: 4, bookingCreated: false },
  { id: "w3", guestPhone: "+1-555-0103", guestName: "Carol White", currentState: "complete", lastMessageAt: "2026-05-10 12:00", messages: 12, bookingCreated: true },
  { id: "w4", guestPhone: "+1-555-0104", guestName: "David Kim", currentState: "greeting", lastMessageAt: "2026-05-10 11:45", messages: 2, bookingCreated: false },
  { id: "w5", guestPhone: "+1-555-0105", guestName: null, currentState: "support", lastMessageAt: "2026-05-10 10:20", messages: 6, bookingCreated: false },
];

const mockMessages = [
  { id: "m1", direction: "inbound", body: "Hi, I'd like to book a room for May 15-17", sentAt: "14:20", status: "read" },
  { id: "m2", direction: "outbound", body: "Welcome! I'd be happy to help. Let me check availability for May 15-17. How many guests?", sentAt: "14:21", status: "delivered" },
  { id: "m3", direction: "inbound", body: "2 adults", sentAt: "14:22", status: "read" },
  { id: "m4", direction: "outbound", body: "Great! We have a Deluxe King available for $180/night. Total: $360 + taxes. Would you like to proceed?", sentAt: "14:22", status: "delivered" },
  { id: "m5", direction: "inbound", body: "Yes please", sentAt: "14:23", status: "read" },
  { id: "m6", direction: "outbound", body: "Perfect! Please provide your email for confirmation:", sentAt: "14:23", status: "delivered" },
  { id: "m7", direction: "inbound", body: "alice@email.com", sentAt: "14:24", status: "read" },
  { id: "m8", direction: "outbound", body: "Reservation confirmed! Ref: WA-001. Check-in: May 15, 3PM. Room: 201. See you soon!", sentAt: "14:25", status: "delivered" },
];

export default function WhatsAppPage() {
  const { config } = useTenant();
  const [activeTab, setActiveTab] = useState<"conversations" | "settings">("conversations");
  const [selectedConv, setSelectedConv] = useState<string | null>("w1");
  const [botEnabled, setBotEnabled] = useState(true);
  const [bookingEnabled, setBookingEnabled] = useState(true);
  const [autoReplyEnabled, setAutoReplyEnabled] = useState(true);

  if (!hasCapability(config, CAPABILITIES.REVENUE.WHATSAPP_BOT)) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <MessageCircle className="h-16 w-16 text-slate-600" />
        <h2 className="mt-4 text-xl font-bold text-white">WhatsApp Bot</h2>
        <p className="mt-2 text-slate-400">Upgrade to Revenue tier to enable WhatsApp automation.</p>
      </div>
    );
  }

  const activeConversations = mockConversations.filter((c) => c.currentState !== "complete").length;
  const bookingsToday = mockConversations.filter((c) => c.bookingCreated).length;

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
            <button onClick={() => setBotEnabled(!botEnabled)} className="text-nexus-400">
              {botEnabled ? <ToggleRight className="h-6 w-6" /> : <ToggleLeft className="h-6 w-6 text-slate-600" />}
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
              {mockConversations.map((conv) => (
                <button
                  key={conv.id}
                  onClick={() => setSelectedConv(conv.id)}
                  className={`w-full p-3 text-left transition-colors ${selectedConv === conv.id ? "bg-nexus-500/10" : "hover:bg-slate-800/50"}`}
                >
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium text-white">{conv.guestName || conv.guestPhone}</span>
                    <span className="text-[10px] text-slate-500">{conv.lastMessageAt.split(" ")[1]}</span>
                  </div>
                  <div className="mt-1 flex items-center gap-2">
                    <span className={`rounded px-1.5 py-0.5 text-[10px] ${
                      conv.currentState === "complete" ? "bg-emerald-500/10 text-emerald-400" :
                      conv.currentState === "greeting" ? "bg-slate-500/10 text-slate-400" :
                      "bg-nexus-500/10 text-nexus-400"
                    }`}>
                      {conv.currentState}
                    </span>
                    {conv.bookingCreated && <span className="rounded bg-emerald-500/10 px-1.5 py-0.5 text-[10px] text-emerald-400">Booked</span>}
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
                  <p className="text-sm font-medium text-white">Alice Chen</p>
                  <p className="text-xs text-slate-400">+1-555-0101 · Active now</p>
                </div>
              </div>
            </div>

            <div className="flex-1 space-y-3 overflow-y-auto py-3">
              {mockMessages.map((msg) => (
                <div key={msg.id} className={`flex ${msg.direction === "inbound" ? "justify-start" : "justify-end"}`}>
                  <div className={`max-w-[70%] rounded-lg px-3 py-2 ${
                    msg.direction === "inbound"
                      ? "bg-slate-800 text-slate-200"
                      : "bg-nexus-500/20 text-white"
                  }`}>
                    <p className="text-sm">{msg.body}</p>
                    <div className={`mt-1 flex items-center gap-1 ${msg.direction === "outbound" ? "justify-end" : ""}`}>
                      <span className="text-[10px] text-slate-500">{msg.sentAt}</span>
                      {msg.direction === "outbound" && msg.status === "delivered" && <CheckCircle2 className="h-3 w-3 text-slate-500" />}
                      {msg.direction === "outbound" && msg.status === "read" && <CheckCircle2 className="h-3 w-3 text-nexus-400" />}
                    </div>
                  </div>
                </div>
              ))}
            </div>

            <div className="border-t border-slate-700 pt-3">
              <div className="flex gap-2">
                <input className="input flex-1" placeholder="Type a message..." />
                <button className="btn-primary gap-2"><Send className="h-4 w-4" /></button>
              </div>
              <div className="mt-2 flex gap-2">
                <button className="rounded bg-slate-800 px-2 py-1 text-[10px] text-slate-400 hover:text-white">Book room</button>
                <button className="rounded bg-slate-800 px-2 py-1 text-[10px] text-slate-400 hover:text-white">Check-in info</button>
                <button className="rounded bg-slate-800 px-2 py-1 text-[10px] text-slate-400 hover:text-white">WiFi password</button>
                <button className="rounded bg-slate-800 px-2 py-1 text-[10px] text-slate-400 hover:text-white">Late checkout</button>
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
                <button onClick={() => setBotEnabled(!botEnabled)} className="text-nexus-400">
                  {botEnabled ? <ToggleRight className="h-6 w-6" /> : <ToggleLeft className="h-6 w-6 text-slate-600" />}
                </button>
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-white">Booking via Chat</p>
                  <p className="text-xs text-slate-400">Allow guests to book rooms via WhatsApp</p>
                </div>
                <button onClick={() => setBookingEnabled(!bookingEnabled)} className="text-nexus-400">
                  {bookingEnabled ? <ToggleRight className="h-6 w-6" /> : <ToggleLeft className="h-6 w-6 text-slate-600" />}
                </button>
              </div>
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-white">Auto-Reply</p>
                  <p className="text-xs text-slate-400">Instant responses to common questions</p>
                </div>
                <button onClick={() => setAutoReplyEnabled(!autoReplyEnabled)} className="text-nexus-400">
                  {autoReplyEnabled ? <ToggleRight className="h-6 w-6" /> : <ToggleLeft className="h-6 w-6 text-slate-600" />}
                </button>
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm text-slate-400">Welcome Message</label>
              <textarea className="input min-h-[80px] text-sm" defaultValue="Welcome to our hotel! How can I help you today? Reply BOOK to make a reservation, CHECKIN for check-in info, or WIFI for the password."></textarea>
            </div>

            <div className="space-y-2">
              <label className="text-sm text-slate-400">Phone Number ID</label>
              <input className="input text-sm" placeholder="Enter Meta Phone Number ID" />
            </div>

            <div className="space-y-2">
              <label className="text-sm text-slate-400">Access Token</label>
              <input className="input text-sm" type="password" placeholder="Enter Meta API token" />
            </div>

            <button className="btn-primary w-full">Save Configuration</button>
          </div>

          <div className="card space-y-6">
            <h3 className="text-lg font-semibold text-white">Webhook Setup</h3>
            <div className="rounded-lg bg-slate-800/50 p-4 space-y-3">
              <p className="text-sm text-slate-400">Configure this webhook URL in your Meta Developer Dashboard:</p>
              <div className="flex items-center gap-2 rounded bg-slate-900 p-3">
                <code className="flex-1 text-xs text-nexus-400">https://api.nexus.com/webhooks/whatsapp/demo</code>
                <button className="rounded bg-slate-800 px-2 py-1 text-xs text-white hover:bg-slate-700">Copy</button>
              </div>
            </div>

            <div className="space-y-3">
              <h4 className="text-sm font-medium text-white">Quick Response Templates</h4>
              {[
                { trigger: "BOOK", response: "Let's book your stay! What dates? (YYYY-MM-DD format)" },
                { trigger: "CHECKIN", response: "Check-in starts at 3PM. Your room number will be sent on arrival day." },
                { trigger: "WIFI", response: "Network: NexusGuest | Password: Welcome2026" },
                { trigger: "CHECKOUT", response: "Check-out is at 11AM. Late checkout available until 1PM for $30." },
                { trigger: "HELP", response: "Front desk: ext. 101 | Housekeeping: ext. 102 | Emergency: ext. 911" },
              ].map((t) => (
                <div key={t.trigger} className="flex items-start gap-3 rounded bg-slate-800/50 p-3">
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
