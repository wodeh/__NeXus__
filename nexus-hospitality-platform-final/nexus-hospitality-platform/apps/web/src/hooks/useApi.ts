import { useCallback } from 'react';
import axios, { AxiosRequestConfig, AxiosError } from 'axios';
import { useAuth } from './useAuth';

const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  if (typeof window !== 'undefined') {
    const raw = localStorage.getItem('nexus_auth');
    if (raw) {
      try {
        const tokens = JSON.parse(raw);
        if (tokens.accessToken) {
          config.headers.Authorization = `Bearer ${tokens.accessToken}`;
        }
      } catch {}
    }
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401 && typeof window !== 'undefined') {
      localStorage.removeItem('nexus_auth');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

export function useApi() {
  const { tenantId } = useAuth();

  const get = useCallback(
    async <T = unknown>(path: string, config?: AxiosRequestConfig): Promise<T> => {
      const url = path.startsWith('/tenants/') ? path : `/tenants/${tenantId}${path}`;
      const res = await api.get<T>(url, config);
      return res.data;
    },
    [tenantId]
  );

  const post = useCallback(
    async <T = unknown>(path: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> => {
      const url = path.startsWith('/tenants/') ? path : `/tenants/${tenantId}${path}`;
      const res = await api.post<T>(url, data, config);
      return res.data;
    },
    [tenantId]
  );

  const patch = useCallback(
    async <T = unknown>(path: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> => {
      const url = path.startsWith('/tenants/') ? path : `/tenants/${tenantId}${path}`;
      const res = await api.patch<T>(url, data, config);
      return res.data;
    },
    [tenantId]
  );

  const del = useCallback(
    async <T = unknown>(path: string, config?: AxiosRequestConfig): Promise<T> => {
      const url = path.startsWith('/tenants/') ? path : `/tenants/${tenantId}${path}`;
      const res = await api.delete<T>(url, config);
      return res.data;
    },
    [tenantId]
  );

  return { get, post, patch, del, raw: api };
}
