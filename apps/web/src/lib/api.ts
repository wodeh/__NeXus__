const API_BASE = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

async function api(path: string, init?: RequestInit) {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...init,
  });
  if (!res.ok) {
    const err = await res.text();
    throw new Error(err || `HTTP ${res.status}`);
  }
  return res.json();
}

export const apiClient = {
  get: (path: string) => api(path),
  post: (path: string, body: unknown) => api(path, { method: "POST", body: JSON.stringify(body) }),
  patch: (path: string, body: unknown) => api(path, { method: "PATCH", body: JSON.stringify(body) }),
  put: (path: string, body: unknown) => api(path, { method: "PUT", body: JSON.stringify(body) }),
  del: (path: string) => api(path, { method: "DELETE" }),
};

export type TenantID = string;
