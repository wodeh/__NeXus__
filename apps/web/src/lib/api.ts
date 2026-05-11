const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function api<T>(path: string, opts?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: {
      "Content-Type": "application/json",
      ...(opts?.headers || {}),
    },
    ...opts,
  });
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`API ${res.status}: ${body}`);
  }
  return res.json();
}

/* ─── Legacy axios-compatible client for existing hooks ─── */
export const apiClient = {
  async get<T>(path: string): Promise<T> {
    return api<T>(path);
  },
  async post<T>(path: string, payload: unknown): Promise<T> {
    return api<T>(path, { method: "POST", body: JSON.stringify(payload) });
  },
  async patch<T>(path: string, payload: unknown): Promise<T> {
    return api<T>(path, { method: "PATCH", body: JSON.stringify(payload) });
  },
  async delete<T>(path: string): Promise<T> {
    return api<T>(path, { method: "DELETE" });
  },
};

export interface Reservation {
  id: string;
  guest_name: string;
  email?: string;
  phone?: string;
  room_number?: string;
  room_type: string;
  check_in: string;
  check_out: string;
  adults: number;
  children: number;
  status: "confirmed" | "checked_in" | "checked_out" | "cancelled" | "no_show";
  source: "walk_in" | "ota" | "direct" | "agent";
  total: number;
  balance: number;
  special_requests?: string;
  vip: boolean;
  color?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateReservationPayload {
  guest_name: string;
  email?: string;
  phone?: string;
  room_type: string;
  room_number?: string;
  check_in: string;
  check_out: string;
  adults?: number;
  children?: number;
  source?: string;
  special_requests?: string;
  vip?: boolean;
}

export async function getReservations(): Promise<Reservation[]> {
  const data = await api<{ reservations: Reservation[] }>("/v1/reservations");
  return data.reservations;
}

export async function createReservation(payload: CreateReservationPayload): Promise<Reservation> {
  return api<Reservation>("/v1/reservations", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function checkInReservation(id: string): Promise<void> {
  await api(`/v1/reservations/${id}/checkin`, { method: "PATCH" });
}

export async function checkOutReservation(id: string): Promise<void> {
  await api(`/v1/reservations/${id}/checkout`, { method: "PATCH" });
}

export async function cancelReservation(id: string): Promise<void> {
  await api(`/v1/reservations/${id}/cancel`, { method: "PATCH" });
}

export async function assignRoom(id: string, roomNumber: string): Promise<void> {
  await api(`/v1/reservations/${id}/assign-room`, {
    method: "PATCH",
    body: JSON.stringify({ room_number: roomNumber }),
  });
}

/* ─── Rooms ─── */
export interface Room {
  id: string;
  number: string;
  type: string;
  floor: string;
  bed_type?: string;
  status: "occupied" | "vacant_clean" | "vacant_dirty" | "blocked" | "maintenance";
  rate_night: number;
}

export async function getRooms(): Promise<Room[]> {
  const data = await api<{ rooms: Room[] }>("/v1/rooms");
  return data.rooms;
}

export async function updateRoomStatus(number: string, status: string): Promise<void> {
  await api(`/v1/rooms/${number}/status`, {
    method: "PATCH",
    body: JSON.stringify({ status }),
  });
}

/* ─── Housekeeping ─── */
export interface HousekeepingTask {
  id: string;
  room_number: string;
  task_type: string;
  status: "pending" | "in_progress" | "completed" | "skipped" | "blocked";
  assigned_to?: string;
  priority: string;
  notes?: string;
  started_at?: string;
  completed_at?: string;
  created_at: string;
}

export interface HousekeepingStaff {
  id: string;
  name: string;
  role: string;
  active_shift: string;
  max_rooms_per_day: number;
  phone?: string;
  email?: string;
  active: boolean;
}

export async function getHousekeepingTasks(): Promise<HousekeepingTask[]> {
  const data = await api<{ tasks: HousekeepingTask[] }>("/v1/housekeeping/tasks");
  return data.tasks;
}

export async function createHousekeepingTask(payload: {
  room_number: string;
  task_type: string;
  priority?: string;
  notes?: string;
  assigned_to?: string;
}): Promise<HousekeepingTask> {
  return api<HousekeepingTask>("/v1/housekeeping/tasks", {
    method: "POST",
    body: JSON.stringify(payload),
  });
}

export async function updateHousekeepingTask(
  id: string,
  updates: { status?: string; assigned_to?: string; notes?: string }
): Promise<void> {
  await api(`/v1/housekeeping/tasks/${id}`, {
    method: "PATCH",
    body: JSON.stringify(updates),
  });
}

export async function getHousekeepingStaff(): Promise<HousekeepingStaff[]> {
  const data = await api<{ staff: HousekeepingStaff[] }>("/v1/housekeeping/staff");
  return data.staff;
}
