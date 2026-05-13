
/* ─── WhatsApp API (messages) ─── */

export interface WhatsAppConversation {
  id: string;
  tenant_id: string;
  guest_phone: string;
  guest_name?: string;
  last_message: string;
  last_message_at: string;
  unread_count: number;
  is_active: boolean;
}

export interface WhatsAppMessage {
  id: string;
  conversation_id: string;
  direction: "inbound" | "outbound";
  body: string;
  status: "sent" | "delivered" | "read" | "failed";
  created_at: string;
}

export async function getWhatsAppConversations(): Promise<WhatsAppConversation[]> {
  const data = await api<{ conversations: WhatsAppConversation[] }>('/v1/whatsapp/conversations');
  return data.conversations || [];
}

export async function getWhatsAppMessages(conversationId: string): Promise<WhatsAppMessage[]> {
  const data = await api<{ messages: WhatsAppMessage[] }>(`/v1/whatsapp/conversations/${conversationId}/messages`);
  return data.messages || [];
}
