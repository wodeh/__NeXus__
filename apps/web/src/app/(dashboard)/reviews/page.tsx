"use client";

import { useState } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import {
  Star,
  MessageSquare,
  ThumbsUp,
  ThumbsDown,
  TrendingUp,
  TrendingDown,
  Send,
  Globe,
  CalendarDays,
  Filter,
  Search,
  Reply,
  CheckCircle2,
  Clock,
  Mail,
  Smartphone,
  AlertTriangle,
} from "lucide-react";

const mockReviews = [
  { id: "r1", guestName: "Alice Chen", roomNumber: "201", channel: "internal", overallRating: 5, cleanliness: 5, service: 5, location: 4, value: 5, amenities: 5, comment: "Absolutely wonderful stay! Staff went above and beyond. Will definitely return.", isPublished: true, staffResponse: "Thank you Alice! We can't wait to welcome you back.", respondedAt: "2026-05-11", createdAt: "2026-05-10" },
  { id: "r2", guestName: "Bob Jones", roomNumber: "102", channel: "booking_com", overallRating: 4, cleanliness: 4, service: 5, location: 5, value: 4, amenities: 3, comment: "Great location and friendly staff. Room was a bit small but comfortable.", isPublished: true, staffResponse: null, respondedAt: null, createdAt: "2026-05-09" },
  { id: "r3", guestName: "Carol White", roomNumber: "301", channel: "google", overallRating: 5, cleanliness: 5, service: 5, location: 5, value: 5, amenities: 5, comment: "Best hotel experience in years. The spa package was incredible.", isPublished: true, staffResponse: "So glad you enjoyed the spa! Hope to see you again soon.", respondedAt: "2026-05-10", createdAt: "2026-05-09" },
  { id: "r4", guestName: "David Kim", roomNumber: "105", channel: "internal", overallRating: 2, cleanliness: 2, service: 3, location: 4, value: 2, amenities: 3, comment: "AC was noisy all night. Couldn't sleep well. Front desk was helpful though.", isPublished: false, staffResponse: null, respondedAt: null, createdAt: "2026-05-08" },
  { id: "r5", guestName: "Emma Wilson", roomNumber: "202", channel: "tripadvisor", overallRating: 4, cleanliness: 4, service: 4, location: 5, value: 4, amenities: 4, comment: "Lovely boutique hotel. Breakfast buffet exceeded expectations.", isPublished: true, staffResponse: "Thank you Emma! Our chef will be thrilled to hear that.", respondedAt: "2026-05-09", createdAt: "2026-05-08" },
  { id: "r6", guestName: "Frank Brown", roomNumber: "401", channel: "expedia", overallRating: 3, cleanliness: 3, service: 3, location: 5, value: 3, amenities: 4, comment: "Average experience. Nothing special but nothing terrible either.", isPublished: true, staffResponse: null, respondedAt: null, createdAt: "2026-05-07" },
];

const mockReviewRequests = [
  { id: "rr1", guestName: "Alice Chen", channel: "internal", status: "completed", scheduledAt: "2026-05-11 10:00", sentAt: "2026-05-11 10:00", openedAt: "2026-05-11 10:05", completedAt: "2026-05-11 10:15" },
  { id: "rr2", guestName: "Bob Jones", channel: "booking_com", status: "sent", scheduledAt: "2026-05-10 14:00", sentAt: "2026-05-10 14:00", openedAt: null, completedAt: null },
  { id: "rr3", guestName: "Carol White", channel: "google", status: "opened", scheduledAt: "2026-05-10 12:00", sentAt: "2026-05-10 12:00", openedAt: "2026-05-10 12:30", completedAt: null },
  { id: "rr4", guestName: "David Kim", channel: "internal", status: "scheduled", scheduledAt: "2026-05-12 10:00", sentAt: null, openedAt: null, completedAt: null },
];

export default function ReviewsPage() {
  const { config } = useTenant();
  const [activeTab, setActiveTab] = useState<"reviews" | "requests" | "analytics">("reviews");
  const [search, setSearch] = useState("");
  const [selectedReview, setSelectedReview] = useState<string | null>(null);
  const [responseText, setResponseText] = useState("");

  if (!hasCapability(config, CAPABILITIES.REVENUE.GUEST_REVIEWS)) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <Star className="h-16 w-16 text-slate-600" />
        <h2 className="mt-4 text-xl font-bold text-white">Guest Reviews</h2>
        <p className="mt-2 text-slate-400">Upgrade to Revenue tier to manage reviews.</p>
      </div>
    );
  }

  const filteredReviews = mockReviews.filter((r) =>
    r.guestName.toLowerCase().includes(search.toLowerCase()) ||
    r.comment.toLowerCase().includes(search.toLowerCase())
  );

  const avgRating = (mockReviews.reduce((s, r) => s + r.overallRating, 0) / mockReviews.length).toFixed(1);
  const totalReviews = mockReviews.length;
  const responseRate = Math.round((mockReviews.filter((r) => r.staffResponse).length / totalReviews) * 100);
  const pendingResponse = mockReviews.filter((r) => !r.staffResponse && r.overallRating <= 3).length;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Guest Reviews</h1>
          <p className="text-sm text-slate-400">Review management · Response tracking · NPS analytics</p>
        </div>
      </div>

      {/* KPI Cards */}
      <div className="grid grid-cols-4 gap-4">
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><Star className="h-4 w-4" /> Average Rating</div>
          <p className="text-2xl font-bold text-white">{avgRating}</p>
          <div className="flex items-center gap-1 text-xs text-emerald-400"><TrendingUp className="h-3 w-3" /> +0.3 vs last month</div>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><MessageSquare className="h-4 w-4" /> Total Reviews</div>
          <p className="text-2xl font-bold text-white">{totalReviews}</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><Reply className="h-4 w-4" /> Response Rate</div>
          <p className="text-2xl font-bold text-white">{responseRate}%</p>
        </div>
        <div className="card space-y-2">
          <div className="flex items-center gap-2 text-xs text-slate-400"><AlertTriangle className="h-4 w-4" /> Pending Response</div>
          <p className="text-2xl font-bold text-white">{pendingResponse}</p>
        </div>
      </div>

      {/* Tabs */}
      <div className="flex items-center gap-1 border-b border-slate-700 pb-1">
        {(["reviews", "requests", "analytics"] as const).map((tab) => (
          <button key={tab} onClick={() => setActiveTab(tab)} className={`rounded-lg px-4 py-2 text-sm font-medium ${activeTab === tab ? "bg-slate-800 text-white" : "text-slate-400"}`}>
            {tab === "reviews" ? "All Reviews" : tab === "requests" ? "Review Requests" : "Analytics"}
          </button>
        ))}
      </div>

      {/* Reviews Tab */}
      {activeTab === "reviews" && (
        <div className="grid grid-cols-3 gap-4">
          <div className="col-span-2 card space-y-4">
            <div className="relative">
              <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-500" />
              <input className="input w-full pl-9" placeholder="Search reviews..." value={search} onChange={(e) => setSearch(e.target.value)} />
            </div>
            <div className="space-y-3">
              {filteredReviews.map((review) => (
                <div key={review.id} className="rounded-lg border border-slate-700 bg-slate-800/50 p-4 space-y-3">
                  <div className="flex items-start justify-between">
                    <div className="flex items-center gap-3">
                      <div className="h-10 w-10 rounded-full bg-nexus-500/20 flex items-center justify-center">
                        <span className="text-sm font-medium text-nexus-400">{review.guestName[0]}</span>
                      </div>
                      <div>
                        <p className="text-sm font-medium text-white">{review.guestName}</p>
                        <p className="text-xs text-slate-400">Room {review.roomNumber} · {review.channel}</p>
                      </div>
                    </div>
                    <div className="flex items-center gap-1">
                      {Array.from({ length: 5 }).map((_, i) => (
                        <Star key={i} className={`h-4 w-4 ${i < review.overallRating ? "text-amber-400 fill-amber-400" : "text-slate-600"}`} />
                      ))}
                    </div>
                  </div>
                  <p className="text-sm text-slate-300">{review.comment}</p>
                  <div className="flex items-center gap-4 text-xs text-slate-500">
                    <span>Cleanliness: {review.cleanliness}</span>
                    <span>Service: {review.service}</span>
                    <span>Location: {review.location}</span>
                    <span>Value: {review.value}</span>
                    <span>Amenities: {review.amenities}</span>
                  </div>
                  {review.staffResponse && (
                    <div className="rounded bg-slate-700/50 p-3">
                      <p className="text-xs text-nexus-400 font-medium">Staff Response ({review.respondedAt}):</p>
                      <p className="text-sm text-slate-300 mt-1">{review.staffResponse}</p>
                    </div>
                  )}
                  {!review.staffResponse && (
                    <div className="flex gap-2">
                      <input
                        className="input flex-1 text-sm"
                        placeholder="Write a response..."
                        value={selectedReview === review.id ? responseText : ""}
                        onChange={(e) => { setSelectedReview(review.id); setResponseText(e.target.value); }}
                      />
                      <button className="btn-primary text-xs gap-1"><Reply className="h-3 w-3" /> Reply</button>
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>

          <div className="space-y-4">
            <div className="card space-y-3">
              <h3 className="text-sm font-semibold text-white">Rating Breakdown</h3>
              {[5, 4, 3, 2, 1].map((rating) => {
                const count = mockReviews.filter((r) => r.overallRating === rating).length;
                const pct = (count / totalReviews) * 100;
                return (
                  <div key={rating} className="flex items-center gap-2">
                    <span className="text-xs text-slate-400 w-3">{rating}</span>
                    <Star className="h-3 w-3 text-amber-400" />
                    <div className="flex-1 h-2 rounded-full bg-slate-800">
                      <div className="h-full rounded-full bg-amber-400" style={{ width: `${pct}%` }} />
                    </div>
                    <span className="text-xs text-slate-400 w-6">{count}</span>
                  </div>
                );
              })}
            </div>

            <div className="card space-y-3">
              <h3 className="text-sm font-semibold text-white">By Channel</h3>
              {["internal", "booking_com", "google", "tripadvisor", "expedia"].map((ch) => {
                const count = mockReviews.filter((r) => r.channel === ch).length;
                return (
                  <div key={ch} className="flex items-center justify-between">
                    <span className="text-sm text-slate-400 capitalize">{ch.replace("_", ".")}</span>
                    <span className="text-sm text-white">{count}</span>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      )}

      {/* Requests Tab */}
      {activeTab === "requests" && (
        <div className="card space-y-4">
          <table className="w-full text-left text-sm">
            <thead className="bg-slate-800 text-xs uppercase text-slate-400">
              <tr>
                <th className="px-4 py-3">Guest</th>
                <th className="px-4 py-3">Channel</th>
                <th className="px-4 py-3">Status</th>
                <th className="px-4 py-3">Scheduled</th>
                <th className="px-4 py-3">Sent</th>
                <th className="px-4 py-3">Opened</th>
                <th className="px-4 py-3">Completed</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-800">
              {mockReviewRequests.map((req) => (
                <tr key={req.id} className="hover:bg-slate-800/50">
                  <td className="px-4 py-3 text-white">{req.guestName}</td>
                  <td className="px-4 py-3"><span className="rounded bg-slate-800 px-2 py-0.5 text-xs text-slate-300">{req.channel}</span></td>
                  <td className="px-4 py-3">
                    <span className={`rounded px-2 py-0.5 text-xs ${
                      req.status === "completed" ? "bg-emerald-500/10 text-emerald-400" :
                      req.status === "opened" ? "bg-sky-500/10 text-sky-400" :
                      req.status === "sent" ? "bg-nexus-500/10 text-nexus-400" :
                      "bg-slate-500/10 text-slate-400"
                    }`}>{req.status}</span>
                  </td>
                  <td className="px-4 py-3 text-slate-400">{req.scheduledAt}</td>
                  <td className="px-4 py-3 text-slate-400">{req.sentAt || "—"}</td>
                  <td className="px-4 py-3 text-slate-400">{req.openedAt || "—"}</td>
                  <td className="px-4 py-3 text-slate-400">{req.completedAt || "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Analytics Tab */}
      {activeTab === "analytics" && (
        <div className="grid grid-cols-2 gap-4">
          <div className="card space-y-4">
            <h3 className="text-lg font-semibold text-white">NPS Score</h3>
            <div className="flex items-center justify-center py-8">
              <div className="text-center">
                <p className="text-5xl font-bold text-emerald-400">72</p>
                <p className="text-sm text-slate-400 mt-2">Net Promoter Score</p>
                <p className="text-xs text-slate-500">Based on 128 responses</p>
              </div>
            </div>
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-emerald-400">Promoters (4-5★)</span>
                <span className="text-white">85%</span>
              </div>
              <div className="h-2 rounded-full bg-slate-800">
                <div className="h-full rounded-full bg-emerald-400" style={{ width: "85%" }} />
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-slate-400">Passive (3★)</span>
                <span className="text-white">10%</span>
              </div>
              <div className="h-2 rounded-full bg-slate-800">
                <div className="h-full rounded-full bg-slate-400" style={{ width: "10%" }} />
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-rose-400">Detractors (1-2★)</span>
                <span className="text-white">5%</span>
              </div>
              <div className="h-2 rounded-full bg-slate-800">
                <div className="h-full rounded-full bg-rose-400" style={{ width: "5%" }} />
              </div>
            </div>
          </div>

          <div className="card space-y-4">
            <h3 className="text-lg font-semibold text-white">Category Scores</h3>
            {[
              { label: "Cleanliness", score: 4.5, color: "bg-emerald-400" },
              { label: "Service", score: 4.7, color: "bg-emerald-400" },
              { label: "Location", score: 4.6, color: "bg-emerald-400" },
              { label: "Value", score: 4.0, color: "bg-nexus-400" },
              { label: "Amenities", score: 4.2, color: "bg-nexus-400" },
            ].map((cat) => (
              <div key={cat.label} className="space-y-1">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-slate-300">{cat.label}</span>
                  <span className="text-white">{cat.score}/5</span>
                </div>
                <div className="h-2 rounded-full bg-slate-800">
                  <div className={`h-full rounded-full ${cat.color}`} style={{ width: `${(cat.score / 5) * 100}%` }} />
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
