import { useState, useEffect, useCallback } from 'react';

export interface AuthUser {
  id: string;
  email: string;
  firstName: string;
  lastName: string;
  tenantId: string;
  roles: string[];
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
}

const STORAGE_KEY = 'nexus_auth';

function loadTokens(): AuthTokens | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as AuthTokens;
    if (parsed.expiresAt && parsed.expiresAt < Date.now()) {
      localStorage.removeItem(STORAGE_KEY);
      return null;
    }
    return parsed;
  } catch {
    return null;
  }
}

function saveTokens(tokens: AuthTokens | null) {
  if (typeof window === 'undefined') return;
  if (tokens) {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(tokens));
  } else {
    localStorage.removeItem(STORAGE_KEY);
  }
}

function decodeToken(token: string): AuthUser | null {
  try {
    const payload = token.split('.')[1];
    const decoded = JSON.parse(atob(payload));
    return {
      id: decoded.sub,
      email: decoded.email,
      firstName: decoded.first_name || '',
      lastName: decoded.last_name || '',
      tenantId: decoded.tenant_id,
      roles: decoded.roles || [],
    };
  } catch {
    return null;
  }
}

export function useAuth() {
  const [tokens, setTokensState] = useState<AuthTokens | null>(null);
  const [user, setUser] = useState<AuthUser | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loaded = loadTokens();
    setTokensState(loaded);
    if (loaded) {
      setUser(decodeToken(loaded.accessToken));
    }
    setIsLoading(false);
  }, []);

  const setTokens = useCallback((tokens: AuthTokens | null) => {
    saveTokens(tokens);
    setTokensState(tokens);
    setUser(tokens ? decodeToken(tokens.accessToken) : null);
  }, []);

  const login = useCallback(async (email: string, password: string) => {
    const res = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });
    if (!res.ok) throw new Error('Login failed');
    const data = await res.json();
    const tokens: AuthTokens = {
      accessToken: data.access_token,
      refreshToken: data.refresh_token,
      expiresAt: Date.now() + data.expires_in * 1000,
    };
    setTokens(tokens);
    return tokens;
  }, [setTokens]);

  const logout = useCallback(() => {
    setTokens(null);
  }, [setTokens]);

  const getToken = useCallback(() => {
    return tokens?.accessToken || null;
  }, [tokens]);

  return {
    user,
    isAuthenticated: !!user,
    isLoading,
    login,
    logout,
    getToken,
    tenantId: user?.tenantId || '',
  };
}
