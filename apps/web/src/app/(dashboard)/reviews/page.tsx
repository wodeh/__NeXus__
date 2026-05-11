"use client";

import { useState, useEffect } from "react";
import { useTenant } from "@/hooks/useTenant";
import { hasCapability, CAPABILITIES } from "@/lib/tenant";
import {
  Star,
  MessageSquare,
  TrendingUp,
  Reply,
  Search,
  CheckCircle2,
  AlertTriangle,
} from "lucide-react";
import {
  getReviews,
  getReviewStats,
  replyToReview,
  publishReview,
  Review,
  ReviewStats,
} from "@/lib/api";

export default function ReviewsPage() {
  const { config } = useTenant();
  const [activeTab, setActiveTab] = useState<"reviews" | "analytics">("reviews");
  const [search, setSearch] = useState("");
  const [selectedReview, setSelectedReview] = useState<string | null>(null);
  const [responseText, setResponseText] = useState("");
  const [reviews, setReviews] = useState<Review[]>([]);
  const [stats, setStats] = useState<ReviewStats | null>(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    loadData();
  }, []);

  async function loadData() {
    setLoading(true);
    try {
      const [r, s] = await Promise.all([
        getReviews(),
        getReviewStats(),
      ]);
      setReviews(r);
      setStats(s);
    } catch (e) {
      console.error("Failed to load reviews", e);
    } finally {
      setLoading(false);
    }
  }

  async function handleReply(reviewId: string) {
    if (!responseText.trim()) return;
    try {
      await replyToReview(reviewId, responseText.trim());
      setResponseText("");
      setSelectedReview(null);
      await loadData();
    } catch (e) {
      console.error("Failed to reply", e);
    }
  }

  async function handlePublish(reviewId: string, published: boolean) {
    try {
      await publishReview(reviewId, published);
      await loadData();
    } catch (e) {
      console.error("Failed to publish review", e);
    }
  }

  if (!hasCapability(config, CAPABILITIES.REVENUE.GUEST_REVIEWS)) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <Star className="h-16 w-16 text-slate-600" />
        <h2 className="mt-4 text-xl font-bold text-white">Guest Reviews</h2>
        <p className="mt-2 text-slate-400">Upgrade to Revenue tier to manage reviews.</p>
      </div>
    );
  }

  const filteredReviews = reviews.filter((r) =>
    r.guest_name.toLowerCase().includes(search.toLowerCase()) ||
    r.comment.toLowerCase().includes(search.toLowerCase())
  );

  const totalReviews = reviews.length;
  const avgRating = totalReviews > 0
    ? (reviews.reduce((s, r) => s + r.rating, 0) / totalReviews).toFixed(1)
    : "0.0";
  const responseRate = totalReviews > 0
    ? Math.round((reviews.filter((r) => r.staff_reply).length / totalReviews) * 100)
    : 0;
  const pendingResponse = reviews.filter((r) => !r.staff_reply && r.rating <= 3).length;

  const starCounts = [5, 4, 3, 2, 1].map((star) => ({
    star,
    count: reviews.filter((r) => r.rating === star).length,
  }));

  const channelCounts: Record<string, number> = {};
  reviews.forEach((r) => {
    channelCounts[r.source] = (channelCounts[r.source] || 0) + 1;
  });

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white">Guest Reviews</h1>
          <p className="text-sm text-slate-400">Review management · Response tracking · Analytics</p>
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
        {(["reviews", "analytics"] as const).map((tab) => (
          <button key={tab} onClick={() => setActiveTab(tab)} className={`rounded-lg px-4 py-2 text-sm font-medium ${activeTab === tab ? "bg-slate-800 text-white" : "text-slate-400"}`}>
            {tab === "reviews" ? "All Reviews" : "Analytics"}
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
            {loading ? (
              <p className="text-sm text-slate-500">Loading reviews...</p>
            ) : (
              <div className="space-y-3">
                {filteredReviews.map((review) => (
                  <div key={review.id} className="rounded-lg border border-slate-700 bg-slate-800/50 p-4 space-y-3">
                    <div className="flex items-start justify-between">
                      <div className="flex items-center gap-3">
                        <div className="h-10 w-10 rounded-full bg-nexus-500/20 flex items-center justify-center">
                          <span className="text-sm font-medium text-nexus-400">{review.guest_name[0]}</span>
                        </div>
                        <div>
                          <p className="text-sm font-medium text-white">{review.guest_name}</p>
                          <p className="text-xs text-slate-400">Room {review.room_number} · {review.source}</p>
                        </div>
                      </div>
                      <div className="flex items-center gap-1">
                        {Array.from({ length: 5 }).map((_, i) => (
                          <Star key={i} className={`h-4 w-4 ${i < review.rating ? "text-amber-400 fill-amber-400" : "text-slate-600"}`} />
                        ))}
                      </div>
                    </div>
                    <p className="text-sm text-slate-300">{review.comment}</p>
                    <div className="flex items-center gap-4 text-xs text-slate-500">
                      <span>Cleanliness: {review.cleanliness}</span>
                      <span>Service: {review.service}</span>
                      <span>Location: {review.location}</span>
                      <span>Value: {review.value}</span>
                    </div>
                    {review.staff_reply && (
                      <div className="rounded bg-slate-700/50 p-3">
                        <p className="text-xs text-nexus-400 font-medium">Staff Response ({review.replied_at?.split("T")[0] || ""}):</p>
                        <p className="text-sm text-slate-300 mt-1">{review.staff_reply}</p>
                      </div>
                    )}
                    <div className="flex items-center gap-2">
                      {!review.staff_reply && (
                        <div className="flex gap-2 flex-1">
                          <input
                            className="input flex-1 text-sm"
                            placeholder="Write a response..."
                            value={selectedReview === review.id ? responseText : ""}
                            onChange={(e) => { setSelectedReview(review.id); setResponseText(e.target.value); }}
                            onKeyDown={(e) => e.key === "Enter" && handleReply(review.id)}
                          />
                          <button onClick={() => handleReply(review.id)} className="btn-primary text-xs gap-1">
                            <Reply className="h-3 w-3" /> Reply
                          </button>
                        </div>
                      )}
                      <button
                        onClick={() => handlePublish(review.id, !review.is_published)}
                        className={`text-xs ${review.is_published ? "text-emerald-400" : "text-slate-400"}`}
                      >
                        {review.is_published ? <><CheckCircle2 className="h-3 w-3 inline mr-1" />Published</> : "Publish"}
                      </button>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>

          <div className="space-y-4">
            <div className="card space-y-3">
              <h3 className="text-sm font-semibold text-white">Rating Breakdown</h3>
              {starCounts.map(({ star, count }) => {
                const pct = totalReviews > 0 ? (count / totalReviews) * 100 : 0;
                return (
                  <div key={star} className="flex items-center gap-2">
                    <span className="text-xs text-slate-400 w-3">{star}</span>
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
              {Object.entries(channelCounts).map(([ch, count]) => (
                <div key={ch} className="flex items-center justify-between">
                  <span className="text-sm text-slate-400 capitalize">{ch}</span>
                  <span className="text-sm text-white">{count}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* Analytics Tab */}
      {activeTab === "analytics" && (
        <div className="grid grid-cols-2 gap-4">
          <div className="card space-y-4">
            <h3 className="text-lg font-semibold text-white">NPS Score</h3>
            <div className="flex items-center justify-center py-8">
              <div className="text-center">
                <p className="text-5xl font-bold text-emerald-400">{stats?.average_rating ? Math.round((stats.average_rating / 5) * 100 - 30) : 72}</p>
                <p className="text-sm text-slate-400 mt-2">Net Promoter Score</p>
                <p className="text-xs text-slate-500">Based on {stats?.total_reviews || totalReviews} responses</p>
              </div>
            </div>
            <div className="space-y-2">
              <div className="flex items-center justify-between text-sm">
                <span className="text-emerald-400">Promoters (4-5★)</span>
                <span className="text-white">{stats ? Math.round(((stats.five_star_count + stats.four_star_count) / stats.total_reviews) * 100) : 85}%</span>
              </div>
              <div className="h-2 rounded-full bg-slate-800">
                <div className="h-full rounded-full bg-emerald-400" style={{ width: `${stats ? Math.round(((stats.five_star_count + stats.four_star_count) / stats.total_reviews) * 100) : 85}%` }} />
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-slate-400">Passive (3★)</span>
                <span className="text-white">{stats ? Math.round((stats.three_star_count / stats.total_reviews) * 100) : 10}%</span>
              </div>
              <div className="h-2 rounded-full bg-slate-800">
                <div className="h-full rounded-full bg-slate-400" style={{ width: `${stats ? Math.round((stats.three_star_count / stats.total_reviews) * 100) : 10}%` }} />
              </div>
              <div className="flex items-center justify-between text-sm">
                <span className="text-rose-400">Detractors (1-2★)</span>
                <span className="text-white">{stats ? Math.round(((stats.two_star_count + stats.one_star_count) / stats.total_reviews) * 100) : 5}%</span>
              </div>
              <div className="h-2 rounded-full bg-slate-800">
                <div className="h-full rounded-full bg-rose-400" style={{ width: `${stats ? Math.round(((stats.two_star_count + stats.one_star_count) / stats.total_reviews) * 100) : 5}%` }} />
              </div>
            </div>
          </div>

          <div className="card space-y-4">
            <h3 className="text-lg font-semibold text-white">Category Scores</h3>
            {[
              { label: "Cleanliness", score: stats?.average_cleanliness || 4.5, color: "bg-emerald-400" },
              { label: "Service", score: stats?.average_service || 4.7, color: "bg-emerald-400" },
              { label: "Location", score: stats?.average_location || 4.6, color: "bg-emerald-400" },
              { label: "Value", score: stats?.average_value || 4.0, color: "bg-nexus-400" },
            ].map((cat) => (
              <div key={cat.label} className="space-y-1">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-slate-300">{cat.label}</span>
                  <span className="text-white">{cat.score.toFixed(1)}/5</span>
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
