"use client";

import { useState } from "react";
import Link from "next/link";
import { ArrowLeft, Plus, Save, User, CalendarDays, BedDouble, Users, CreditCard, FileText, CheckCircle2, ChevronRight } from "lucide-react";

const roomTypes = [
  { name: "Standard", price: 89, maxGuests: 2 },
  { name: "Deluxe", price: 129, maxGuests: 2 },
  { name: "Deluxe King", price: 129, maxGuests: 2 },
  { name: "Suite", price: 229, maxGuests: 4 },
  { name: "Accessible", price: 99, maxGuests: 2 },
];

export default function NewReservationPage() {
  const [step, setStep] = useState(1);
  const [form, setForm] = useState({
    guest_name: "",
    email: "",
    phone: "",
    address: "",
    country: "",
    room_type: "Standard",
    room_number: "",
    check_in: "2026-05-10",
    check_out: "2026-05-12",
    adults: 1,
    children: 0,
    source: "direct" as string,
    rate_plan: "rack_rate" as string,
    special_requests: "",
    payment_method: "card" as string,
    deposit: 0,
    vip: false,
  });

  const selectedType = roomTypes.find((t) => t.name === form.room_type);
  const nights = Math.max(1, Math.ceil((new Date(form.check_out).getTime() - new Date(form.check_in).getTime()) / (1000 * 60 * 60 * 24)));
  const roomTotal = (selectedType?.price || 89) * nights;
  const tax = roomTotal * 0.12;
  const total = roomTotal + tax;

  const steps = [
    { num: 1, label: "Guest", icon: User },
    { num: 2, label: "Room & Dates", icon: CalendarDays },
    { num: 3, label: "Payment", icon: CreditCard },
    { num: 4, label: "Confirm", icon: CheckCircle2 },
  ];

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <Link href="/reservations" className="rounded-lg border border-slate-700 bg-slate-800 p-2 text-slate-400 hover:text-white">
          <ArrowLeft className="h-4 w-4" />
        </Link>
        <div>
          <h1 className="text-2xl font-bold text-white">New Reservation</h1>
          <p className="text-sm text-slate-400">Full booking form with all details</p>
        </div>
      </div>

      {/* Step Indicator */}
      <div className="flex items-center gap-2">
        {steps.map((s, i) => {
          const Icon = s.icon;
          const active = step === s.num;
          const done = step > s.num;
          return (
            <div key={s.num} className="flex items-center gap-2">
              <div
                className={`flex items-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition-colors ${
                  active
                    ? "bg-nexus-500 text-white"
                    : done
                    ? "bg-emerald-500/10 text-emerald-400"
                    : "bg-slate-800 text-slate-400"
                }`}
              >
                <Icon className="h-4 w-4" />
                {s.label}
              </div>
              {i < steps.length - 1 && <div className="h-px w-4 bg-slate-700" />}
            </div>
          );
        })}
      </div>

      {/* Step 1: Guest Info */}
      {step === 1 && (
        <div className="card space-y-4">
          <h3 className="text-lg font-semibold text-white">Guest Information</h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Full Name *</label>
              <input
                className="input w-full"
                value={form.guest_name}
                onChange={(e) => setForm({ ...form, guest_name: e.target.value })}
                placeholder="Guest full name"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Email</label>
              <input
                className="input w-full"
                value={form.email}
                onChange={(e) => setForm({ ...form, email: e.target.value })}
                placeholder="guest@email.com"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Phone *</label>
              <input
                className="input w-full"
                value={form.phone}
                onChange={(e) => setForm({ ...form, phone: e.target.value })}
                placeholder="+1-555-0000"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Country</label>
              <input
                className="input w-full"
                value={form.country}
                onChange={(e) => setForm({ ...form, country: e.target.value })}
                placeholder="United States"
              />
            </div>
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Address</label>
            <input
              className="input w-full"
              value={form.address}
              onChange={(e) => setForm({ ...form, address: e.target.value })}
              placeholder="Street address"
            />
          </div>
          <div className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={form.vip}
              onChange={(e) => setForm({ ...form, vip: e.target.checked })}
              className="h-4 w-4 rounded border-slate-600 bg-slate-700"
            />
            <label className="text-sm text-slate-300">VIP Guest</label>
          </div>
        </div>
      )}

      {/* Step 2: Room & Dates */}
      {step === 2 && (
        <div className="card space-y-4">
          <h3 className="text-lg font-semibold text-white">Room & Dates</h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Room Type</label>
              <div className="grid grid-cols-2 gap-2">
                {roomTypes.map((rt) => (
                  <button
                    key={rt.name}
                    onClick={() => setForm({ ...form, room_type: rt.name })}
                    className={`rounded-lg border p-3 text-left transition-colors ${
                      form.room_type === rt.name
                        ? "border-nexus-500 bg-nexus-500/10"
                        : "border-slate-700 bg-slate-800 hover:border-slate-600"
                    }`}
                  >
                    <p className="text-sm font-medium text-white">{rt.name}</p>
                    <p className="text-xs text-slate-400">${rt.price}/night · Max {rt.maxGuests}</p>
                  </button>
                ))}
              </div>
            </div>
            <div className="space-y-3">
              <div>
                <label className="mb-1 block text-xs text-slate-400">Check In</label>
                <input
                  type="date"
                  className="input w-full"
                  value={form.check_in}
                  onChange={(e) => setForm({ ...form, check_in: e.target.value })}
                />
              </div>
              <div>
                <label className="mb-1 block text-xs text-slate-400">Check Out</label>
                <input
                  type="date"
                  className="input w-full"
                  value={form.check_out}
                  onChange={(e) => setForm({ ...form, check_out: e.target.value })}
                />
              </div>
              <div>
                <label className="mb-1 block text-xs text-slate-400">Guests</label>
                <div className="flex items-center gap-3">
                  <div>
                    <span className="text-xs text-slate-500">Adults</span>
                    <input
                      type="number"
                      min={1}
                      max={4}
                      className="input mt-1 w-20"
                      value={form.adults}
                      onChange={(e) => setForm({ ...form, adults: parseInt(e.target.value) || 1 })}
                    />
                  </div>
                  <div>
                    <span className="text-xs text-slate-500">Children</span>
                    <input
                      type="number"
                      min={0}
                      max={3}
                      className="input mt-1 w-20"
                      value={form.children}
                      onChange={(e) => setForm({ ...form, children: parseInt(e.target.value) || 0 })}
                    />
                  </div>
                </div>
              </div>
            </div>
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Special Requests</label>
            <textarea
              className="input w-full"
              rows={3}
              value={form.special_requests}
              onChange={(e) => setForm({ ...form, special_requests: e.target.value })}
              placeholder="Any special requirements..."
            />
          </div>
        </div>
      )}

      {/* Step 3: Payment */}
      {step === 3 && (
        <div className="card space-y-4">
          <h3 className="text-lg font-semibold text-white">Payment & Pricing</h3>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Source</label>
              <select
                className="input w-full"
                value={form.source}
                onChange={(e) => setForm({ ...form, source: e.target.value })}
              >
                <option value="direct">Direct</option>
                <option value="ota">OTA (Booking.com, Expedia)</option>
                <option value="walk_in">Walk-in</option>
                <option value="agent">Travel Agent</option>
                <option value="corporate">Corporate</option>
              </select>
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Rate Plan</label>
              <select
                className="input w-full"
                value={form.rate_plan}
                onChange={(e) => setForm({ ...form, rate_plan: e.target.value })}
              >
                <option value="rack_rate">Rack Rate</option>
                <option value="corporate">Corporate Rate</option>
                <option value="advance">Advance Purchase</option>
                <option value="long_stay">Long Stay</option>
                <option value="promo">Promotion</option>
              </select>
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Payment Method</label>
              <select
                className="input w-full"
                value={form.payment_method}
                onChange={(e) => setForm({ ...form, payment_method: e.target.value })}
              >
                <option value="card">Credit Card</option>
                <option value="cash">Cash</option>
                <option value="bank">Bank Transfer</option>
                <option value="invoice">Invoice / Company</option>
              </select>
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Deposit ($)</label>
              <input
                type="number"
                className="input w-full"
                value={form.deposit}
                onChange={(e) => setForm({ ...form, deposit: parseFloat(e.target.value) || 0 })}
              />
            </div>
          </div>
        </div>
      )}

      {/* Step 4: Confirm */}
      {step === 4 && (
        <div className="card space-y-4">
          <h3 className="text-lg font-semibold text-white">Confirm Reservation</h3>
          <div className="rounded-lg border border-slate-700 bg-slate-800/50 p-4 space-y-2">
            <div className="flex justify-between text-sm">
              <span className="text-slate-400">Guest</span>
              <span className="text-white font-medium">{form.guest_name || "—"}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-slate-400">Room Type</span>
              <span className="text-white">{form.room_type}</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-slate-400">Dates</span>
              <span className="text-white">{form.check_in} → {form.check_out} ({nights} nights)</span>
            </div>
            <div className="flex justify-between text-sm">
              <span className="text-slate-400">Guests</span>
              <span className="text-white">{form.adults} adults, {form.children} children</span>
            </div>
            <div className="border-t border-slate-700 pt-2 mt-2">
              <div className="flex justify-between text-sm">
                <span className="text-slate-400">Room Charges</span>
                <span className="text-white">${roomTotal.toFixed(2)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-slate-400">Taxes (12%)</span>
                <span className="text-white">${tax.toFixed(2)}</span>
              </div>
              <div className="flex justify-between text-sm">
                <span className="text-slate-400">Deposit</span>
                <span className="text-white">-${form.deposit.toFixed(2)}</span>
              </div>
              <div className="flex justify-between text-lg font-bold text-white pt-2 border-t border-slate-700">
                <span>Balance Due</span>
                <span>${(total - form.deposit).toFixed(2)}</span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Navigation */}
      <div className="flex items-center justify-between">
        <button
          onClick={() => setStep(Math.max(1, step - 1))}
          disabled={step === 1}
          className="btn-secondary disabled:opacity-50"
        >
          <ArrowLeft className="h-4 w-4" />
          Back
        </button>
        {step < 4 ? (
          <button onClick={() => setStep(step + 1)} className="btn-primary">
            Next
            <ChevronRight className="h-4 w-4" />
          </button>
        ) : (
          <button className="btn-primary">
            <Save className="h-4 w-4" />
            Create Reservation
          </button>
        )}
      </div>
    </div>
  );
}
