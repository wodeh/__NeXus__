'use client';

import { createContext, useContext, useState, useEffect, useCallback, ReactNode } from 'react';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface AuthUser {
  id: string;
  email: string;
  name: string;
  role: string;
  tenant_id: string;
  is_active: boolean;
}

export interface AuthTenant {
  id: string;
  name: string;
  external_id: string;
}

interface AuthContextType {
  user: AuthUser | null;
  tenant: AuthTenant | null;
  capabilities: string[];
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  isVillaOwner: boolean;
  isSuperAdmin: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [tenant, setTenant] = useState<AuthTenant | null>(null);
  const [capabilities, setCapabilities] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Restore session on mount via httpOnly cookie check
  useEffect(() => {
    let cancelled = false;

    async function restoreSession() {
      try {
        const res = await fetch(`${API_BASE}/v1/auth/me`, {
          method: 'GET',
          credentials: 'include',
          headers: { 'Content-Type': 'application/json' },
        });

        if (res.ok) {
          const data = await res.json();
          if (!cancelled) {
            setUser(data.user ?? null);
            setTenant(data.tenant ?? null);
            setCapabilities(data.capabilities || []);
          }
        }
      } catch {
        // Session not valid or network error — user stays logged out
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    }

    restoreSession();
    return () => { cancelled = true; };
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await fetch(`${API_BASE}/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      credentials: 'include',
      body: JSON.stringify({ email, password }),
    });

    if (!res.ok) {
      const body = await res.text();
      throw new Error(body);
    }

    const data = await res.json();

    // Store only non-sensitive user/tenant data in React state
    // The JWT token is handled by the backend as an httpOnly cookie
    setUser(data.user ?? null);
    setTenant(data.tenant ?? null);
    setCapabilities(data.capabilities || []);
  }, []);

  const logout = useCallback(async () => {
    try {
      // Call server-side logout to clear the httpOnly cookie
      await fetch(`${API_BASE}/v1/auth/logout`, {
        method: 'POST',
        credentials: 'include',
        headers: { 'Content-Type': 'application/json' },
      });
    } catch {
      // Best-effort: even if the network request fails, clear local state
    } finally {
      setUser(null);
      setTenant(null);
      setCapabilities([]);
    }
  }, []);

  const isVillaOwner = tenant?.external_id === 'villa-owners' || capabilities.includes('villa:dashboard');
  const isSuperAdmin = user?.role === 'super_admin' || user?.role === 'admin';

  return (
    <AuthContext.Provider
      value={{ user, tenant, capabilities, isLoading, login, logout, isVillaOwner, isSuperAdmin }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used inside AuthProvider');
  return ctx;
}
