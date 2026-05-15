// @ts-nocheck
"use client";

import { useState, useEffect, useCallback } from "react";
import {
  Home, Plus, Edit3, Trash2, CheckCircle2, X, RefreshCw, AlertTriangle,
  MapPin, Bed, Bath, Users, DollarSign, Shield, Star,
  ChevronRight, Phone, MessageSquare, Clock,
} from "lucide-react";
import {
  getVillas, createVilla, updateVilla, deleteVilla,
  VillaProperty,
} from "@/lib/api";

export default function VillaPropertiesPage() {
  const [villas, setVillas] = useState<VillaProperty[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreate, setShowCreate] = useState(false);
  const [editing, setEditing] = useState<VillaProperty | null>(null);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const v = await getVillas();
      setVillas(v);
    } catch (e: any) {
      setError(e.message || "Failed to load");
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  const handleCreate = async (villa: Partial<VillaProperty>) => {
    try {
      await createVilla(villa as any);
      setShowCreate(false);
      fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleUpdate = async (id: string, updates: Partial<VillaProperty>) => {
    try {
      await updateVilla(id, updates);
      setEditing(null);
      fetchData();
    } catch (e: any) {
      alert(e.message);
    }
  };

  const handleDelete = async (id: string) => {
    if (!confirm("Delete this villa?")) return;
    try {
      await deleteVilla(id);
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
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-white flex items-center gap-2">
            <Home className="h-6 w-6 text-nexus-400" /> My Villas
          </h1>
          <p className="text-sm text-slate-400">Manage your properties · Set pricing · Track availability</p>
        </div>
        <button onClick={() => setShowCreate(true)} className="btn-primary gap-2">
          <Plus className="h-4 w-4" /> Add Villa
        </button>
      </div>

      <div className="grid grid-cols-2 gap-4">
        {villas.map((villa) => (
          <div key={villa.id} className="card space-y-4">
            <div className="flex items-start justify-between">
              <div className="space-y-1">
                <p className="text-lg font-bold text-white">{villa.name}</p>
                <p className="text-xs text-slate-400">{villa.address}, {villa.city}</p>
              </div>
              <div className="flex items-center gap-1">
                <button onClick={() => setEditing(villa)} className="rounded p-1.5 text-slate-400 hover:text-white"><Edit3 className="h-4 w-4" /></button>
                <button onClick={() => handleDelete(villa.id)} className="rounded p-1.5 text-slate-400 hover:text-rose-400"><Trash2 className="h-4 w-4" /></button>
              </div>
            </div>

            <div className="grid grid-cols-4 gap-2 text-xs">
              <StatBadge icon={Bed} value={`${villa.bedrooms} BR`} />
              <StatBadge icon={Bath} value={`${villa.bathrooms} BA`} />
              <StatBadge icon={Users} value={`${villa.max_guests} guests`} />
              <StatBadge icon={DollarSign} value={`$${Math.round(villa.price_per_night || 0)}/night`} />
            </div>

            <div className="flex items-center justify-between text-xs">
              <div className="flex items-center gap-2">
                <span className={`rounded px-2 py-0.5 ${villa.status === "available" ? "bg-emerald-500/10 text-emerald-400" : "bg-amber-500/10 text-amber-400"}`}>
                  {villa.status}
                </span>
                <span className="text-slate-500">Cleaning: ${villa.cleaning_fee}</span>
                <span className="text-slate-500">Deposit: ${villa.security_deposit}</span>
              </div>
            </div>

            {(villa.amenities || []).length > 0 && (
              <div className="flex flex-wrap gap-1">
                {(villa.amenities || []).map((a) => (
                  <span key={a} className="rounded bg-slate-800 px-2 py-0.5 text-[10px] text-slate-400 capitalize">{a}</span>
                ))}
              </div>
            )}
          </div>
        ))}
      </div>

      {showCreate && <VillaModal onSave={handleCreate} onClose={() => setShowCreate(false)} />}
      {editing && <VillaModal villa={editing} onSave={(v) => handleUpdate(editing.id, v)} onClose={() => setEditing(null)} />}
    </div>
  );
}

function StatBadge({ icon: Icon, value }: { icon: React.ElementType; value: string }) {
  return (
    <div className="flex items-center gap-1.5 rounded bg-slate-800/50 px-2 py-1.5 text-slate-300">
      <Icon className="h-3 w-3 text-slate-500" /> {value}
    </div>
  );
}

function VillaModal({ villa, onSave, onClose }: { villa?: VillaProperty; onSave: (v: Partial<VillaProperty>) => void; onClose: () => void }) {
  const [form, setForm] = useState({
    name: villa?.name || "",
    description: villa?.description || "",
    address: villa?.address || "",
    city: villa?.city || "",
    country: villa?.country || "Palestine",
    bedrooms: villa?.bedrooms || 2,
    bathrooms: villa?.bathrooms || 1,
    max_guests: villa?.max_guests || 4,
    price_per_night: villa?.price_per_night || 200,
    currency: villa?.currency || "USD",
    cleaning_fee: villa?.cleaning_fee || 30,
    security_deposit: villa?.security_deposit || 200,
    elevation: villa?.elevation || 0,
    amenities: villa?.amenities?.join(", ") || "wifi, parking",
  });

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-lg rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl space-y-4 max-h-[90vh] overflow-auto">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">{villa ? "Edit Villa" : "Add New Villa"}</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>

        <div className="space-y-3">
          <input className="input w-full" placeholder="Villa Name" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} />
          <textarea className="input w-full" rows={2} placeholder="Description" value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} />
          <input className="input w-full" placeholder="Address" value={form.address} onChange={(e) => setForm({ ...form, address: e.target.value })} />
          <div className="grid grid-cols-3 gap-3">
            <input className="input w-full" placeholder="City" value={form.city} onChange={(e) => setForm({ ...form, city: e.target.value })} />
            <input className="input w-full" placeholder="Country" value={form.country} onChange={(e) => setForm({ ...form, country: e.target.value })} />
            <input type="number" className="input w-full" placeholder="Elevation (m)" value={form.elevation} onChange={(e) => setForm({ ...form, elevation: parseFloat(e.target.value) || 0 })} />
          </div>
          <div className="grid grid-cols-3 gap-3">
            <input type="number" className="input w-full" placeholder="Bedrooms" value={form.bedrooms} onChange={(e) => setForm({ ...form, bedrooms: parseInt(e.target.value) || 1 })} />
            <input type="number" className="input w-full" placeholder="Bathrooms" value={form.bathrooms} onChange={(e) => setForm({ ...form, bathrooms: parseInt(e.target.value) || 1 })} />
            <input type="number" className="input w-full" placeholder="Max Guests" value={form.max_guests} onChange={(e) => setForm({ ...form, max_guests: parseInt(e.target.value) || 1 })} />
          </div>
          <div className="grid grid-cols-3 gap-3">
            <input type="number" className="input w-full" placeholder="Price/Night" value={form.price_per_night} onChange={(e) => setForm({ ...form, price_per_night: parseFloat(e.target.value) || 0 })} />
            <input className="input w-full" placeholder="Currency" value={form.currency} onChange={(e) => setForm({ ...form, currency: e.target.value })} />
            <input type="number" className="input w-full" placeholder="Cleaning Fee" value={form.cleaning_fee} onChange={(e) => setForm({ ...form, cleaning_fee: parseFloat(e.target.value) || 0 })} />
          </div>
          <input type="number" className="input w-full" placeholder="Security Deposit" value={form.security_deposit} onChange={(e) => setForm({ ...form, security_deposit: parseFloat(e.target.value) || 0 })} />
          <input className="input w-full" placeholder="Amenities (comma separated)" value={form.amenities} onChange={(e) => setForm({ ...form, amenities: e.target.value })} />
        </div>

        <div className="flex justify-end gap-2 pt-2">
          <button onClick={onClose} className="btn-secondary">Cancel</button>
          <button onClick={() => onSave({ ...form, amenities: form.amenities.split(",").map((s) => s.trim()) })} className="btn-primary gap-2"><CheckCircle2 className="h-4 w-4" /> Save</button>
        </div>
      </div>
    </div>
  );
}
