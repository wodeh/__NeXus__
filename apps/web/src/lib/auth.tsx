'use client';

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';

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
  token: string | null;
  capabilities: string[];
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  isVillaOwner: boolean;
  isSuperAdmin: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [tenant, setTenant] = useState<AuthTenant | null>(null);
  const [token, setToken] = useState<string | null>(null);
  const [capabilities, setCapabilities] = useState<string[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // Restore from localStorage on mount
  useEffect(() => {
    const saved = localStorage.getItem('nexus-auth');
    if (saved) {
      try {
        const parsed = JSON.parse(saved);
        if (parsed.token && parsed.user) {
          setToken(parsed.token);
          setUser(parsed.user);
          setTenant(parsed.tenant);
          setCapabilities(parsed.capabilities || []);
        }
      } catch {
        localStorage.removeItem('nexus-auth');
      }
    }
    setIsLoading(false);
  }, []);

  const login = async (email: string, password: string) => {
    const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });

    if (!res.ok) {
      const body = await res.text();
      throw new Error(body);
    }

    const data = await res.json();
    const authData = {
      token: data.token,
      user: data.user,
      tenant: data.tenant,
      capabilities: data.capabilities || [],
    };

    localStorage.setItem('nexus-auth', JSON.stringify(authData));
    localStorage.setItem('nexus-tenant', data.tenant?.external_id || 'demo');

    setToken(data.token);
    setUser(data.user);
    setTenant(data.tenant);
    setCapabilities(data.capabilities || []);
  };

  const logout = () => {
    localStorage.removeItem('nexus-auth');
    localStorage.removeItem('nexus-tenant');
    setToken(null);
    setUser(null);
    setTenant(null);
    setCapabilities([]);
  };

  const isVillaOwner = tenant?.external_id === 'villa-owners' || capabilities.includes('villa:dashboard');
  const isSuperAdmin = user?.role === 'super_admin' || user?.role === 'admin';

  return (
    <AuthContext.Provider
      value={{ user, tenant, token, capabilities, isLoading, login, logout, isVillaOwner, isSuperAdmin }}
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
