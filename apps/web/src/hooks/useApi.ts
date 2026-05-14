"use client";

import { useState, useEffect, useCallback } from "react";

const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

interface ApiError {
  message: string;
  status: number;
}

interface ApiResponse<T> {
  data: T | null;
  loading: boolean;
  error: ApiError | null;
  refetch: () => void;
}

function getTenantId(): string {
  return localStorage.getItem("tenantId") || "demo";
}

<<<<<<< HEAD
function getToken(): string | null {
  return localStorage.getItem("token");
}

async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const tenantId = getTenantId();
  const token = getToken();
=======
async function apiFetch<T>(path: string, options?: RequestInit): Promise<T> {
  const tenantId = getTenantId();
>>>>>>> phase1/security-stability
  const url = `${API_BASE}/tenants/${tenantId}${path}`;

  const headers: Record<string, string> = {
    "Content-Type": "application/json",
<<<<<<< HEAD
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
  };

  const res = await fetch(url, {
=======
  };

  const res = await fetch(url, {
    credentials: "include",
>>>>>>> phase1/security-stability
    ...options,
    headers: { ...headers, ...((options?.headers as Record<string, string>) || {}) },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({}));
    throw { message: body.error || `HTTP ${res.status}`, status: res.status };
  }

  return res.json();
}

// ==================== IPTV ====================

export interface IPTVChannel {
  id: string;
  name: string;
  number: number;
  category: string;
  streamUrl: string;
  logoUrl?: string;
  isPremium: boolean;
  isActive: boolean;
  language: string;
  description?: string;
}

export interface IPTVContent {
  id: string;
  title: string;
  type: "movie" | "series" | "documentary" | "music" | "kids";
  category: string;
  duration?: number;
  posterUrl?: string;
  description?: string;
  isPremium: boolean;
  isActive: boolean;
}

export interface IPTVRoomBinding {
  id: string;
  roomId: string;
  roomNumber: string;
  channelIds: string[];
  welcomeScreenId?: string;
  defaultChannelId?: string;
  guestName?: string;
  checkIn?: string;
  checkOut?: string;
}

export interface IPTVWelcomeScreen {
  id: string;
  roomId: string;
  title: string;
  message: string;
  backgroundUrl?: string;
  showWeather: boolean;
  showTime: boolean;
  language: string;
}

export interface IPTVAnalytics {
  totalChannels: number;
  activeChannels: number;
  totalContent: number;
  activeContent: number;
  totalBindings: number;
  viewership: { hour: string; viewers: number }[];
}

export function useIPTVChannels(): ApiResponse<IPTVChannel[]> {
  return useApiFetch<IPTVChannel[]>("/iptv/channels");
}

export function useIPTVContent(): ApiResponse<IPTVContent[]> {
  return useApiFetch<IPTVContent[]>("/iptv/content");
}

export function useIPTVRoomBindings(): ApiResponse<IPTVRoomBinding[]> {
  return useApiFetch<IPTVRoomBinding[]>("/iptv/room-bindings");
}

export function useIPTVWelcomeScreen(roomId: string): ApiResponse<IPTVWelcomeScreen> {
  return useApiFetch<IPTVWelcomeScreen>(`/iptv/rooms/${roomId}/welcome`);
}

export function useIPTVAnalytics(): ApiResponse<IPTVAnalytics> {
  return useApiFetch<IPTVAnalytics>("/iptv/analytics");
}

export async function createIPTVChannel(channel: Omit<IPTVChannel, "id">): Promise<IPTVChannel> {
  return apiFetch<IPTVChannel>("/iptv/channels", { method: "POST", body: JSON.stringify(channel) });
}

export async function updateIPTVChannel(id: string, channel: Partial<IPTVChannel>): Promise<IPTVChannel> {
  return apiFetch<IPTVChannel>(`/iptv/channels/${id}`, { method: "PATCH", body: JSON.stringify(channel) });
}

export async function deleteIPTVChannel(id: string): Promise<void> {
  return apiFetch<void>(`/iptv/channels/${id}`, { method: "DELETE" });
}

// ==================== SMART LOCKS ====================

export interface SmartLock {
  id: string;
  roomId: string;
  roomNumber: string;
  serialNumber: string;
  model: string;
  manufacturer: string;
  status: "online" | "offline" | "low_battery" | "warning";
  batteryLevel: number;
  lastCommunicationAt?: string;
  lastUnlockAt?: string;
  lastLockAt?: string;
  firmwareVersion: string;
  remoteUnlockEnabled: boolean;
  autoLockEnabled: boolean;
  accessCodeCount: number;
}

export interface LockEvent {
  id: string;
  lockId: string;
  eventType: string;
  eventSource: string;
  details: string;
  occurredAt: string;
}

export interface AccessCode {
  id: string;
  lockId: string;
  code: string;
  label: string;
  isActive: boolean;
  validFrom: string;
  validUntil: string;
  maxUses?: number;
  useCount: number;
}

export interface LockOverview {
  total: number;
  online: number;
  offline: number;
  lowBattery: number;
  activeCodes: number;
}

export function useSmartLocks(): ApiResponse<SmartLock[]> {
  return useApiFetch<SmartLock[]>("/locks");
}

export function useLockOverview(): ApiResponse<LockOverview> {
  return useApiFetch<LockOverview>("/locks/overview");
}

export function useLockEvents(lockId: string): ApiResponse<LockEvent[]> {
  return useApiFetch<LockEvent[]>(`/locks/${lockId}/events`);
}

export function useAccessCodes(lockId: string): ApiResponse<AccessCode[]> {
  return useApiFetch<AccessCode[]>(`/locks/${lockId}/access-codes`);
}

export function useLockByRoom(roomId: string): ApiResponse<SmartLock> {
  return useApiFetch<SmartLock>(`/locks/by-room/${roomId}`);
}

export async function remoteUnlock(lockId: string): Promise<void> {
  return apiFetch<void>(`/locks/${lockId}/unlock`, { method: "POST" });
}

export async function remoteLock(lockId: string): Promise<void> {
  return apiFetch<void>(`/locks/${lockId}/lock`, { method: "POST" });
}

export async function createAccessCode(lockId: string, code: Omit<AccessCode, "id" | "lockId">): Promise<AccessCode> {
  return apiFetch<AccessCode>(`/locks/${lockId}/access-codes`, { method: "POST", body: JSON.stringify(code) });
}

export async function revokeAccessCode(lockId: string, codeId: string): Promise<void> {
  return apiFetch<void>(`/locks/${lockId}/access-codes/${codeId}/revoke`, { method: "POST" });
}

// ==================== COMMAND CENTER ====================

export interface RoomOverview {
  id: string;
  number: string;
  floor: string;
  type: string;
  status: string;
  housekeeping: string;
  guestName?: string;
  checkIn?: string;
  checkOut?: string;
  vip: boolean;
  balance: number;
  lockStatus?: string;
  batteryLevel?: number;
}

export interface DashboardStats {
  totalRooms: number;
  occupied: number;
  available: number;
  blocked: number;
  arrivalsToday: number;
  departuresToday: number;
  inHouse: number;
  revenueToday: number;
  pendingBalance: number;
  alerts: { type: string; message: string; severity: string }[];
}

export function useRooms(propertyId?: string): ApiResponse<RoomOverview[]> {
  const path = propertyId ? `/properties/${propertyId}/rooms` : "/properties/demo/rooms";
  return useApiFetch<RoomOverview[]>(path);
}

export function useDashboardStats(): ApiResponse<DashboardStats> {
  // This would need a dedicated endpoint; using rooms as proxy for now
  return useApiFetch<DashboardStats>("/properties/demo/rooms");
}

// ==================== RESERVATIONS ====================

export interface Reservation {
  id: string;
  guest_name: string;
  email?: string;
  phone?: string;
  room_number: string;
  room_type: string;
  check_in: string;
  check_out: string;
  nights: number;
  adults: number;
  children: number;
  status: string;
  source: string;
  total: number;
  balance: number;
  vip?: boolean;
}

export function useReservations(): ApiResponse<Reservation[]> {
  return useApiFetch<Reservation[]>("/reservations");
}

export function useReservation(id: string): ApiResponse<Reservation> {
  return useApiFetch<Reservation>(`/reservations/${id}`);
}

export async function checkInReservation(id: string): Promise<void> {
  return apiFetch<void>(`/reservations/${id}/checkin`, { method: "POST" });
}

export async function checkOutReservation(id: string): Promise<void> {
  return apiFetch<void>(`/reservations/${id}/checkout`, { method: "POST" });
}

export async function cancelReservation(id: string): Promise<void> {
  return apiFetch<void>(`/reservations/${id}/cancel`, { method: "PATCH" });
}

// ==================== BASE HOOK ====================

function useApiFetch<T>(path: string): ApiResponse<T> {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<ApiError | null>(null);
  const [retryCount, setRetryCount] = useState(0);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await apiFetch<T>(path);
      setData(result);
    } catch (err) {
      setError(err as ApiError);
    } finally {
      setLoading(false);
    }
  }, [path]);

  useEffect(() => {
    fetchData();
  }, [fetchData, retryCount]);

  return {
    data,
    loading,
    error,
    refetch: () => setRetryCount((c) => c + 1),
  };
}

<<<<<<< HEAD
export { apiFetch, getTenantId, getToken };
=======
export { apiFetch, getTenantId };
>>>>>>> phase1/security-stability
