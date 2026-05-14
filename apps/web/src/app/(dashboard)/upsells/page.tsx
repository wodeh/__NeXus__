"use client";

import { useState, useEffect, useCallback } from "react";
import {
  TrendingUp, Plus, Edit3, Trash2, CheckCircle2, X, RefreshCw, AlertTriangle,
  ShoppingBag, DollarSign, Users, Package, Star, ArrowUpRight,
} from "lucide-react";
import {
  getUpsellOffers, getUpsellPurchases, createUpsellOffer, updateUpsellOffer, deleteUpsellOffer,
  UpsellOffer, UpsellPurchase,
} from "@/lib/api";

export default function UpsellsPage() {
  const [offers, setOffers] = useState<UpsellOffer[]>([]);
  const [purchases, setPurchases] = useState<UpsellPurchase[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [editing, setEditing] = useState<UpsellOffer | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [o, p] = await Promise.all([getUpsellOffers(), getUpsellPurchases()]);
      setOffers(o);
      setPurchases(p);
    } catch (e: any) {
      setError(e.message || "Failed to load");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const totalRevenue = offers.reduce((s, o) => s + (o.revenue_generated || 0), 0);
  const totalSold = offers.reduce((s, o) => s + (o.total_sold || 0), 0);
  const avgConversion = totalSold > 0 ? Math.round((totalSold / Math.max(offers.length, 1)) * 10) : 0;

  const handleCreate = async (offer: Partial<UpsellOffer>) => {
    try {
      await createUpsellOffer(offer as any);
      setShowCreate(false);
      fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleUpdate = async (id: string, updates: Partial<UpsellOffer>) => {
    try {
      await updateUpsellOffer(id, updates);
      setEditing(null);
      fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this offer?")) return;
    try {
      await deleteUpsellOffer(id);
      fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  };

  if (loading) {
    return (
      <div className="flex h-96 items-center justify-center">
        <RefreshCw className="h-8 w-8 animate-spin text-nexus-400" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex h-96 flex-col items-center justify-center gap-4">
        <AlertTriangle className="h-10 w-10 text-rose-400" />
        <p className="text-rose-400">{error}</p>
        <button onClick={fetchData} className="btn-primary">Retry</button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <ShoppingBag className="h-6 w-6 text-nexus-400" /> Upsell Engine
          </h1>
          <p className="text-sm text-slate-400">Room upgrades · Late checkout · Breakfast · Spa · Auto-offered via AI</p>
        </div>
        <button onClick={() => setShowCreate(true)} className="btn-primary gap-2">
          <Plus className="h-4 w-4" /> New Offer
        </button>
      </div>

      {/* KPIs */}
      <div className="grid grid-cols-4 gap-4">
        <KpiCard label="Total Revenue" value={`$${Math.round(totalRevenue).toLocaleString()}`} change="+18.2%" icon={DollarSign} color="nexus" />
        <KpiCard label="Items Sold" value={totalSold.toString()} change="+24" icon={Package} color="emerald" />
        <KpiCard label="Active Offers" value={offers.filter(o => o.is_active).length.toString()} change="+2" icon={Star} color="amber" />
        <KpiCard label="Avg Conversion" value={`${avgConversion}%`} change="+1.2%" icon={TrendingUp} color="violet" />
      </div>

      {/* Offers Grid */}
      <div className="grid grid-cols-3 gap-4">
        {offers.map((offer) => (
          <div key={offer.id} className="card space-y-4">
            <div className="flex items-center justify-between">
              <div className={`rounded px-2 py-1 text-[10px] uppercase font-bold ${
                offer.is_active ? "bg-emerald-500/10 text-emerald-400" : "bg-slate-700 text-slate-400"
              }`}>
                {offer.is_active ? "Active" : "Inactive"}
              </div>
              <div className="flex items-center gap-1">
                <button onClick={() => setEditing(offer)} className="rounded p-1 text-slate-400 hover:text-white"><Edit3 className="h-4 w-4" /></button>
                <button onClick={() => handleDelete(offer.id)} className="rounded p-1 text-slate-400 hover:text-rose-400"><Trash2 className="h-4 w-4" /></button>
              </div>
            </div>

            <div className="space-y-1">
              <p className="text-lg font-bold text-white">{offer.name}</p>
              <p className="text-xs text-slate-400">{offer.description}</p>
            </div>

            <div className="flex items-baseline gap-1">
              <span className="text-2xl font-bold text-white">${Math.round(offer.price)}</span>
              <span className="text-xs text-slate-500">{offer.currency}</span>
            </div>

            <div className="grid grid-cols-2 gap-2 text-xs">
              <div className="rounded bg-slate-800/50 p-2">
                <p className="text-slate-500">Sold</p>
                <p className="font-bold text-white">{offer.total_sold}</p>
              </div>
              <div className="rounded bg-slate-800/50 p-2">
                <p className="text-slate-500">Revenue</p>
                <p className="font-bold text-emerald-400">${Math.round(offer.revenue_generated || 0).toLocaleString()}</p>
              </div>
            </div>

            {offer.auto_offer && (
              <div className="flex items-center gap-1 text-[10px] text-nexus-400">
                <ArrowUpRight className="h-3 w-3" /> Auto-offered in guest journeys
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Recent Purchases */}
      <div className="card space-y-4">
        <h3 className="text-lg font-semibold text-white">Recent Purchases</h3>
        <table className="w-full text-left text-sm">
          <thead className="bg-slate-800 text-xs uppercase text-slate-400">
            <tr>
              <th className="px-4 py-2">Offer</th>
              <th className="px-4 py-2">Guest</th>
              <th className="px-4 py-2">Price</th>
              <th className="px-4 py-2">Status</th>
              <th className="px-4 py-2">Payment</th>
              <th className="px-4 py-2">Folio</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-800">
            {purchases.map((p) => (
              <tr key={p.id} className="hover:bg-slate-800/50">
                <td className="px-4 py-2 text-white">{p.offer_name}</td>
                <td className="px-4 py-2 text-slate-400">{p.guest_phone}</td>
                <td className="px-4 py-2 text-white">${p.price}</td>
                <td className="px-4 py-2">
                  <span className={`rounded px-2 py-0.5 text-xs ${
                    p.status === "confirmed" ? "bg-emerald-500/10 text-emerald-400" :
                    p.status === "pending" ? "bg-amber-500/10 text-amber-400" :
                    "bg-rose-500/10 text-rose-400"
                  }`}>
                    {p.status}
                  </span>
                </td>
                <td className="px-4 py-2 text-xs text-slate-400 capitalize">{p.payment_method.replace("_", " ")}</td>
                <td className="px-4 py-2">
                  {p.folio_posted ? <CheckCircle2 className="h-4 w-4 text-emerald-400" /> : <span className="text-xs text-slate-500">Pending</span>}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Modals */}
      {showCreate && <OfferModal onSave={handleCreate} onClose={() => setShowCreate(false)} />}
      {editing && <OfferModal offer={editing} onSave={(o) => handleUpdate(editing.id, o)} onClose={() => setEditing(null)} />}
    </div>
  );
}

function KpiCard({ label, value, change, icon: Icon, color }: { label: string; value: string; change: string; icon: React.ElementType; color: string }) {
  const colorMap: Record<string, string> = {
    nexus: "text-nexus-400 bg-nexus-500/10",
    emerald: "text-emerald-400 bg-emerald-500/10",
    amber: "text-amber-400 bg-amber-500/10",
    violet: "text-violet-400 bg-violet-500/10",
  };
  return (
    <div className="card space-y-3">
      <div className="flex items-center justify-between">
        <span className="text-xs text-slate-400 font-medium">{label}</span>
        <div className={`rounded-lg p-2 ${colorMap[color]}`}><Icon className="h-4 w-4" /></div>
      </div>
      <p className="text-2xl font-bold text-white">{value}</p>
      <div className="flex items-center gap-1 text-xs text-emerald-400">
        <ArrowUpRight className="h-3 w-3" />{change} vs last month
      </div>
    </div>
  );
}

function OfferModal({ offer, onSave, onClose }: { offer?: UpsellOffer; onSave: (o: Partial<UpsellOffer>) => void; onClose: () => void }) {
  const [form, setForm] = useState({
    name: offer?.name || "",
    description: offer?.description || "",
    category: offer?.category || "room_upgrade",
    price: offer?.price || 0,
    currency: offer?.currency || "USD",
    is_active: offer?.is_active ?? true,
    auto_offer: offer?.auto_offer ?? false,
    display_order: offer?.display_order || 0,
  });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">{offer ? "Edit Offer" : "New Upsell Offer"}</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>

        <div className="space-y-3">
          <input className="input w-full" placeholder="Offer Name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          <textarea className="input w-full" rows={2} placeholder="Description" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
          <select className="input w-full" value={form.category} onChange={(e) => setForm({ ...form, category: e.target.value })}>
            <option value="room_upgrade">Room Upgrade</option>
            <option value="late_checkout">Late Checkout</option>
            <option value="early_checkin">Early Check-in</option>
            <option value="breakfast">Breakfast</option>
            <option value="spa">Spa</option>
            <option value="parking">Parking</option>
          </select>
          <div className="grid grid-cols-2 gap-3">
            <input type="number" className="input w-full" placeholder="Price" value={form.price} onChange={(e) => setForm({ ...form, price: parseFloat(e.target.value) || 0 })} />
            <input className="input w-full" placeholder="Currency" value={form.currency} onChange={(e) => setForm({ ...form, currency: e.target.value })} />
          </div>
          <div className="flex items-center gap-4">
            <label className="flex items-center gap-2 text-sm text-white">
              <input type="checkbox" checked={form.is_active} onChange={(e) => setForm({ ...form, is_active: e.target.checked })} /> Active
            </label>
            <label className="flex items-center gap-2 text-sm text-white">
              <input type="checkbox" checked={form.auto_offer} onChange={(e) => setForm({ ...form, auto_offer: e.target.checked })} /> Auto-offer in journeys
            </label>
          </div>
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <button onClick={onClose} className="btn-secondary">Cancel</button>
          <button onClick={() => onSave(form)} className="btn-primary gap-2"><CheckCircle2 className="h-4 w-4" /> Save</button>
        </div>
      </div>
    </div>
  );
}
