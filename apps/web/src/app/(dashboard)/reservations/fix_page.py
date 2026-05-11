#!/usr/bin/env python3
import re

with open("page.tsx", "r") as f:
    content = f.read()

# 1. Remove duplicate showQuickBook state
content = content.replace(
    "  const [showQuickBook, setShowQuickBook] = useState(false);\n\n  const [showQuickBook, setShowQuickBook] = useState(false);",
    "  const [showQuickBook, setShowQuickBook] = useState(false);"
)

# 2. Add normalizeDate helper at top-level
if "function normalizeDate" not in content:
    content = content.replace(
        "function addDays(dateStr: string, days: number): string {",
        "function normalizeDate(d: string): string {\n  return d.split(\"T\")[0];\n}\n\nfunction addDays(dateStr: string, days: number): string {"
    )

# 3. Fix isDateInRange to use normalizeDate
content = content.replace(
    """function isDateInRange(date: string, start: string, end: string): boolean {
  const t = new Date(date + "T00:00:00").getTime();
  const s = new Date(start + "T00:00:00").getTime();
  const e = new Date(end + "T00:00:00").getTime();
  return t >= s && t < e;
}""",
    """function isDateInRange(date: string, start: string, end: string): boolean {
  const t = normalizeDate(date);
  const s = normalizeDate(start);
  const e = normalizeDate(end);
  return t >= s && t < e;
}"""
)

# 4. Fix getReservationsForRoomDate to be more robust
content = content.replace(
    "    return reservations.filter((r) => r.room_number === roomNumber && isDateInRange(date, r.check_in, r.check_out));",
    "    return reservations.filter((r) => {\n      if (r.room_number !== roomNumber) return false;\n      return isDateInRange(date, r.check_in, r.check_out);\n    });"
)

# 5. Fix getSpanForRes to use normalizeDate
content = content.replace(
    """  const getSpanForRes = useCallback((res: Reservation, roomNumber: string) => {
    if (res.room_number !== roomNumber) return 0;
    const overlapStart = new Date(Math.max(new Date(startDate + "T00:00:00").getTime(), new Date(res.check_in + "T00:00:00").getTime()));
    const overlapEnd = new Date(Math.min(new Date(addDays(startDate, dayCount) + "T00:00:00").getTime(), new Date(res.check_out + "T00:00:00").getTime()));
    const days = Math.round((overlapEnd.getTime() - overlapStart.getTime()) / (1000 * 60 * 60 * 24));
    return Math.max(0, days);
  }, [startDate, dayCount]);""",
    """  const getSpanForRes = useCallback((res: Reservation, roomNumber: string) => {
    if (res.room_number !== roomNumber) return 0;
    const resStart = normalizeDate(res.check_in);
    const resEnd = normalizeDate(res.check_out);
    const gridStart = startDate;
    const gridEnd = addDays(startDate, dayCount);
    const overlapStart = new Date(Math.max(new Date(gridStart + "T00:00:00").getTime(), new Date(resStart + "T00:00:00").getTime()));
    const overlapEnd = new Date(Math.min(new Date(gridEnd + "T00:00:00").getTime(), new Date(resEnd + "T00:00:00").getTime()));
    const days = Math.round((overlapEnd.getTime() - overlapStart.getTime()) / (1000 * 60 * 60 * 24));
    return Math.max(0, days);
  }, [startDate, dayCount]);"""
)

# 6. Fix spanRes check_in comparison in tapechart grid
content = content.replace(
    "const spanRes = resList.find((r) => r.check_in === date);",
    "const spanRes = resList.find((r) => normalizeDate(r.check_in) === date);"
)

# 7. Fix occupancyStats date comparisons
content = content.replace(
    """    const departingToday = reservations.filter((r) => r.check_out === today && r.status === "checked_in").length;
    const arrivingToday = reservations.filter((r) => r.check_in === today && r.status !== "checked_in").length;""",
    """    const departingToday = reservations.filter((r) => normalizeDate(r.check_out) === today && r.status === "checked_in").length;
    const arrivingToday = reservations.filter((r) => normalizeDate(r.check_in) === today && r.status !== "checked_in").length;"""
)

# 8. Remove inline normalizeDate definition inside component (it's now top-level)
content = content.replace(
    """  const today = useMemo(() => new Date().toISOString().split("T")[0], []);

  function normalizeDate(d: string): string {
    return d.split("T")[0];
  }

  const fetchData = useCallback""",
    """  const today = useMemo(() => new Date().toISOString().split("T")[0], []);

  const fetchData = useCallback"""
)

# 9. Add QuickBookModal component at the end, before Shared Components section
quick_book_modal = """
/* ─── Quick Book Modal ─── */
function QuickBookModal({ rooms, onClose, onCreate }: { rooms: Room[]; onClose: () => void; onCreate: (payload: any) => Promise<void> }) {
  const [form, setForm] = useState({ guest_name: "", email: "", phone: "", nights: 1, adults: 2, children: 0, special_requests: "", room_number: "", check_in: new Date().toISOString().split("T")[0] });
  const [loading, setLoading] = useState(false);

  const roomData = rooms.find((r) => r.number === form.room_number);
  const rate = roomData?.type === "Suite" ? 229 : roomData?.type === "Deluxe" || roomData?.type === "Deluxe King" ? 129 : 89;
  const total = form.nights * rate;
  const checkOut = addDays(form.check_in, form.nights);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
      <div className="w-full max-w-md rounded-xl border border-slate-700 bg-slate-900 p-6 shadow-2xl">
        <div className="flex items-center justify-between">
          <h3 className="text-lg font-bold text-white">Quick Book</h3>
          <button onClick={onClose} className="text-slate-400 hover:text-white"><X className="h-5 w-5" /></button>
        </div>
        <div className="mt-4 space-y-3">
          <div>
            <label className="mb-1 block text-xs text-slate-400">Guest Name *</label>
            <input className="input w-full" value={form.guest_name} onChange={(e) => setForm({ ...form, guest_name: e.target.value })} placeholder="Full name" />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Email</label>
              <input className="input w-full" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} placeholder="guest@email.com" />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Phone</label>
              <input className="input w-full" value={form.phone} onChange={(e) => setForm({ ...form, phone: e.target.value })} placeholder="+1-555-0000" />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Check In</label>
              <input type="date" className="input w-full" value={form.check_in} onChange={(e) => setForm({ ...form, check_in: e.target.value })} />
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Nights</label>
              <input type="number" min={1} max={30} className="input w-full" value={form.nights} onChange={(e) => setForm({ ...form, nights: parseInt(e.target.value) || 1 })} />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <div>
              <label className="mb-1 block text-xs text-slate-400">Room</label>
              <select className="input w-full" value={form.room_number} onChange={(e) => setForm({ ...form, room_number: e.target.value })}>
                <option value="">Select room...</option>
                {rooms.filter((r) => r.status === "vacant_clean").map((r) => (
                  <option key={r.number} value={r.number}>{r.number} · {r.type}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-xs text-slate-400">Guests</label>
              <div className="flex items-center gap-2">
                <input type="number" min={1} max={6} className="input w-16" value={form.adults} onChange={(e) => setForm({ ...form, adults: parseInt(e.target.value) || 1 })} />
                <span className="text-xs text-slate-500">+</span>
                <input type="number" min={0} max={4} className="input w-16" value={form.children} onChange={(e) => setForm({ ...form, children: parseInt(e.target.value) || 0 })} />
              </div>
            </div>
          </div>
          <div>
            <label className="mb-1 block text-xs text-slate-400">Special Requests</label>
            <textarea className="input w-full" rows={2} value={form.special_requests} onChange={(e) => setForm({ ...form, special_requests: e.target.value })} placeholder="Any special needs..." />
          </div>
        </div>
        <div className="mt-4 flex items-center justify-between border-t border-slate-700 pt-4">
          <div className="text-sm text-slate-400">{form.nights} nights · <span className="font-bold text-white">${total}</span></div>
          <div className="flex gap-2">
            <button onClick={onClose} className="btn-secondary text-xs">Cancel</button>
            <button
              onClick={async () => {
                if (!form.guest_name) return alert("Guest name is required");
                if (!form.room_number) return alert("Room is required");
                setLoading(true);
                await onCreate({
                  guest_name: form.guest_name,
                  email: form.email,
                  phone: form.phone,
                  room_type: roomData?.type || "Standard",
                  room_number: form.room_number,
                  check_in: form.check_in,
                  check_out: checkOut,
                  adults: form.adults,
                  children: form.children,
                  source: "direct",
                  special_requests: form.special_requests,
                });
                setLoading(false);
              }}
              disabled={loading}
              className="btn-primary text-xs disabled:opacity-50"
            >
              <Save className="h-4 w-4" />
              {loading ? "Creating..." : "Create Booking"}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

/* ─── Shared Components ─── */"""

content = content.replace(
    "/* ─── Shared Components ─── */",
    quick_book_modal
)

# 10. Add QuickBookModal usage in main return before closing </div>
content = content.replace(
    """      {/* Reservation Detail Modal */}
      {selectedRes && (
        <ReservationDetailModal""",
    """      {/* Quick Book Modal */}
      {showQuickBook && (
        <QuickBookModal
          rooms={rooms}
          onClose={() => setShowQuickBook(false)}
          onCreate={async (payload) => {
            try {
              await createReservation(payload);
              fetchData();
              setShowQuickBook(false);
            } catch (e: any) {
              alert(e.message);
            }
          }}
        />
      )}

      {/* Reservation Detail Modal */}
      {selectedRes && (
        <ReservationDetailModal"""
)

with open("page.tsx", "w") as f:
    f.write(content)

print("Fixed!")
