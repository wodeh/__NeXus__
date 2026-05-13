
/* ─── WhatsApp API ─── */

export interface WhatsAppConfig {
  id: string;
  tenant_id: string;
  phone_number: string;
  api_key?: string;
  webhook_url?: string;
  is_active: boolean;
  welcome_message?: string;
  auto_reply_enabled: boolean;
  created_at: string;
  updated_at: string;
}

export async function getWhatsAppConfig(): Promise<WhatsAppConfig> {
  return api("/v1/whatsapp/config");
}

export async function updateWhatsAppConfig(payload: Partial<WhatsAppConfig>): Promise<WhatsAppConfig> {
  return api("/v1/whatsapp/config", { method: "PATCH", body: JSON.stringify(payload) });
}

/* ─── Tenant API ─── */

export async function getTenantConfig(): Promise<TenantConfig> {
  return api("/v1/tenant/config");
}

/* ─── Agent API ─── */

export interface Agent {
  id: string;
  tenant_id: string;
  name: string;
  email: string;
  role: string;
  phone?: string;
  commission_rate?: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export async function getAgents(): Promise<Agent[]> {
  const data = await api<{ agents: Agent[] }>("/v1/agents");
  return data.agents || [];
}

export async function updateAgent(id: string, payload: Partial<Agent>): Promise<Agent> {
  return api(`/v1/agents/${id}`, { method: "PATCH", body: JSON.stringify(payload) });
}

export async function deleteAgent(id: string): Promise<void> {
  return api(`/v1/agents/${id}`, { method: "DELETE" });
}
