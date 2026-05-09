// sdk/typescript-sdk/src/client.ts
import { EventEmitter } from 'events';

export interface NexusConfig {
  apiKey: string;
  tenantId: string;
  baseURL?: string;
  version?: string;
  timeout?: number;
  websocketURL?: string;
}

export class NexusClient extends EventEmitter {
  private config: Required<NexusConfig>;
  private ws: WebSocket | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;

  public pms: PMSService;
  public iptv: IPTVService;
  public iot: IoTService;
  public ai: AIService;
  public billing: BillingService;
  public guest: GuestService;

  constructor(config: NexusConfig) {
    super();
    this.config = {
      baseURL: 'https://api.nexus-platform.com',
      version: 'v1',
      timeout: 30000,
      websocketURL: 'wss://ws.nexus-platform.com',
      ...config,
    };

    this.pms = new PMSService(this);
    this.iptv = new IPTVService(this);
    this.iot = new IoTService(this);
    this.ai = new AIService(this);
    this.billing = new BillingService(this);
    this.guest = new GuestService(this);
  }

  async request<T>(
    method: string,
    path: string,
    body?: unknown,
    options?: RequestInit
  ): Promise<NexusResponse<T>> {
    const url = `${this.config.baseURL}/api/${this.config.version}${path}`;

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      'X-API-Key': this.config.apiKey,
      'X-Tenant-ID': this.config.tenantId,
      ...((options?.headers as Record<string, string>) || {}),
    };

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.config.timeout);

    try {
      const response = await fetch(url, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
        signal: controller.signal,
        ...options,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        throw new NexusError(
          error.message || `HTTP ${response.status}`,
          response.status,
          error.code
        );
      }

      const data = await response.json();
      return {
        data,
        status: response.status,
        headers: response.headers,
      };
    } catch (error) {
      clearTimeout(timeoutId);
      throw error;
    }
  }

  connectWebSocket(): void {
    if (this.ws?.readyState === WebSocket.OPEN) return;

    const wsUrl = `${this.config.websocketURL}?tenant=${this.config.tenantId}&token=${this.config.apiKey}`;
    this.ws = new WebSocket(wsUrl);

    this.ws.onopen = () => {
      this.reconnectAttempts = 0;
      this.emit('connected');
    };

    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data);
      this.emit(message.type, message.payload);
    };

    this.ws.onclose = () => {
      this.emit('disconnected');
      this.attemptReconnect();
    };

    this.ws.onerror = (error) => {
      this.emit('error', error);
    };
  }

  private attemptReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      this.emit('maxReconnectReached');
      return;
    }

    this.reconnectAttempts++;
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);

    setTimeout(() => {
      this.emit('reconnecting', this.reconnectAttempts);
      this.connectWebSocket();
    }, delay);
  }

  disconnect(): void {
    this.ws?.close();
    this.ws = null;
  }

  subscribe(channel: string, handler: (data: unknown) => void): () => void {
    this.on(channel, handler);
    return () => this.off(channel, handler);
  }
}

export interface NexusResponse<T> {
  data: T;
  status: number;
  headers: Headers;
}

export class NexusError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string
  ) {
    super(message);
    this.name = 'NexusError';
  }
}

// Service classes
class PMSService {
  constructor(private client: NexusClient) {}

  async getGuest(guestId: string): Promise<NexusResponse<Guest>> {
    return this.client.request('GET', `/guests/${guestId}`);
  }

  async createGuest(guest: GuestCreateRequest): Promise<NexusResponse<Guest>> {
    return this.client.request('POST', '/guests', guest);
  }

  async updateGuest(guestId: string, guest: GuestUpdateRequest): Promise<NexusResponse<Guest>> {
    return this.client.request('PATCH', `/guests/${guestId}`, guest);
  }

  async getReservation(reservationId: string): Promise<NexusResponse<Reservation>> {
    return this.client.request('GET', `/reservations/${reservationId}`);
  }

  async createReservation(reservation: ReservationCreateRequest): Promise<NexusResponse<Reservation>> {
    return this.client.request('POST', '/reservations', reservation);
  }

  async checkIn(reservationId: string, request: CheckInRequest): Promise<NexusResponse<CheckInResult>> {
    return this.client.request('POST', `/reservations/${reservationId}/checkin`, request);
  }

  async checkOut(reservationId: string, request: CheckOutRequest): Promise<NexusResponse<CheckOutResult>> {
    return this.client.request('POST', `/reservations/${reservationId}/checkout`, request);
  }

  async listRooms(propertyId: string, options?: ListOptions): Promise<NexusResponse<PaginatedResponse<Room>>> {
    const params = new URLSearchParams();
    if (options?.page) params.set('page', options.page.toString());
    if (options?.perPage) params.set('per_page', options.perPage.toString());

    return this.client.request('GET', `/properties/${propertyId}/rooms?${params}`);
  }

  async updateRoomStatus(roomId: string, status: string): Promise<NexusResponse<Room>> {
    return this.client.request('PATCH', `/rooms/${roomId}`, { status });
  }

  async createHousekeepingTask(task: HousekeepingTaskRequest): Promise<NexusResponse<HousekeepingTask>> {
    return this.client.request('POST', '/housekeeping/tasks', task);
  }

  async completeHousekeepingTask(
    taskId: string,
    request: HousekeepingCompletionRequest
  ): Promise<NexusResponse<HousekeepingTask>> {
    return this.client.request('POST', `/housekeeping/tasks/${taskId}/complete`, request);
  }

  async createMaintenanceRequest(request: MaintenanceRequest): Promise<NexusResponse<MaintenanceTicket>> {
    return this.client.request('POST', '/maintenance/requests', request);
  }
}

class IPTVService {
  constructor(private client: NexusClient) {}

  async getChannels(propertyId: string): Promise<NexusResponse<Channel[]>> {
    return this.client.request('GET', `/iptv/channels?property_id=${propertyId}`);
  }

  async startStream(request: StreamRequest): Promise<NexusResponse<StreamSession>> {
    return this.client.request('POST', '/iptv/streams', request);
  }

  async stopStream(streamId: string): Promise<NexusResponse<void>> {
    return this.client.request('DELETE', `/iptv/streams/${streamId}`);
  }

  async getEPG(channelId: number, date: string): Promise<NexusResponse<EPG>> {
    return this.client.request('GET', `/iptv/epg/${channelId}?date=${date}`);
  }
}

class IoTService {
  constructor(private client: NexusClient) {}

  async registerDevice(request: DeviceRegistrationRequest): Promise<NexusResponse<Device>> {
    return this.client.request('POST', '/iot/devices', request);
  }

  async getDevice(deviceId: string): Promise<NexusResponse<Device>> {
    return this.client.request('GET', `/iot/devices/${deviceId}`);
  }

  async sendCommand(deviceId: string, command: DeviceCommand): Promise<NexusResponse<CommandResult>> {
    return this.client.request('POST', `/iot/devices/${deviceId}/commands`, command);
  }

  async getTelemetry(deviceId: string, query?: TelemetryQuery): Promise<NexusResponse<TelemetryData>> {
    const params = new URLSearchParams();
    if (query?.startTime) params.set('start_time', query.startTime);
    if (query?.endTime) params.set('end_time', query.endTime);
    if (query?.interval) params.set('interval', query.interval);

    return this.client.request('GET', `/iot/devices/${deviceId}/telemetry?${params}`);
  }
}

class AIService {
  constructor(private client: NexusClient) {}

  async conciergeQuery(request: ConciergeRequest): Promise<NexusResponse<ConciergeResponse>> {
    return this.client.request('POST', '/ai/concierge', request);
  }

  async analyzeSentiment(request: SentimentRequest): Promise<NexusResponse<SentimentResponse>> {
    return this.client.request('POST', '/ai/sentiment', request);
  }

  async forecastOccupancy(request: ForecastRequest): Promise<NexusResponse<ForecastResponse>> {
    return this.client.request('POST', '/ai/forecast/occupancy', request);
  }

  async getRecommendations(request: RecommendationRequest): Promise<NexusResponse<RecommendationResponse>> {
    return this.client.request('POST', '/ai/recommendations', request);
  }
}

class BillingService {
  constructor(private client: NexusClient) {}
  // Billing operations
}

class GuestService {
  constructor(private client: NexusClient) {}
  // Guest portal operations
}

// Types
export interface ListOptions {
  page?: number;
  perPage?: number;
  sort?: string;
  order?: 'asc' | 'desc';
}

export interface PaginatedResponse<T> {
  data: T[];
  meta: {
    page: number;
    per_page: number;
    total: number;
    total_pages: number;
  };
}

export interface Guest {
  id: string;
  first_name: string;
  last_name: string;
  email?: string;
  phone?: string;
  nationality?: string;
  language?: string;
  preferences?: Record<string, unknown>;
  created_at: string;
  updated_at: string;
}

export interface GuestCreateRequest {
  first_name: string;
  last_name: string;
  email?: string;
  phone?: string;
  nationality?: string;
  language?: string;
  preferences?: Record<string, unknown>;
}

export interface GuestUpdateRequest {
  first_name?: string;
  last_name?: string;
  email?: string;
  phone?: string;
  preferences?: Record<string, unknown>;
}

export interface Reservation {
  id: string;
  confirmation_number: string;
  guest_id: string;
  room_id?: string;
  property_id: string;
  status: string;
  check_in_date: string;
  check_out_date: string;
  num_adults: number;
  num_children: number;
  total_amount: number;
  currency_code: string;
  booking_source: string;
  special_requests?: string;
  created_at: string;
  updated_at: string;
}

export interface ReservationCreateRequest {
  guest_id: string;
  property_id: string;
  check_in_date: string;
  check_out_date: string;
  num_adults: number;
  num_children?: number;
  room_type_id?: string;
  rate_plan_id?: string;
  special_requests?: string;
}

export interface CheckInRequest {
  room_id?: string;
  key_card_id?: string;
  preferences?: string[];
}

export interface CheckInResult {
  reservation_id: string;
  room_id: string;
  key_card_id: string;
  checked_in_at: string;
}

export interface CheckOutRequest {
  folio_charges?: Record<string, unknown>[];
}

export interface CheckOutResult {
  reservation_id: string;
  checked_out_at: string;
  balance: number;
}

export interface Room {
  id: string;
  room_number: string;
  property_id: string;
  room_type_id: string;
  status: string;
  housekeeping_status?: string;
  floor?: number;
  is_accessible?: boolean;
  iptv_device_id?: string;
  iot_gateway_id?: string;
}

export interface HousekeepingTask {
  id: string;
  room_id: string;
  task_type: string;
  priority: string;
  status: string;
  assigned_to?: string;
  scheduled_at: string;
  started_at?: string;
  completed_at?: string;
  duration_minutes?: number;
  score?: number;
}

export interface HousekeepingTaskRequest {
  room_id: string;
  task_type: string;
  priority?: string;
  assigned_to?: string;
  scheduled_at: string;
  checklist?: string[];
}

export interface HousekeepingCompletionRequest {
  completed_items: string[];
  issues?: string[];
  notes?: string;
}

export interface MaintenanceRequest {
  room_id: string;
  category: string;
  priority: string;
  title: string;
  description: string;
  reported_by?: string;
}

export interface MaintenanceTicket {
  id: string;
  room_id: string;
  category: string;
  priority: string;
  status: string;
  title: string;
  description: string;
  reported_at: string;
  assigned_to?: string;
  scheduled_date?: string;
}

export interface Channel {
  id: number;
  name: string;
  number: number;
  category: string;
  language: string;
  hd: boolean;
  logo_url?: string;
}

export interface StreamRequest {
  room_id: string;
  channel_id: number;
  device_type?: string;
  profile?: string;
}

export interface StreamSession {
  id: string;
  room_id: string;
  channel_id: number;
  manifest_url: string;
  token: string;
  expires_at: string;
}

export interface EPG {
  channel_id: number;
  date: string;
  programs: Program[];
}

export interface Program {
  id: string;
  title: string;
  description?: string;
  start_time: string;
  end_time: string;
  genre?: string;
  rating?: string;
}

export interface Device {
  id: string;
  protocol: string;
  hardware_id: string;
  device_type: string;
  manufacturer: string;
  model: string;
  room_id: string;
  property_id: string;
  capabilities: string[];
  state: Record<string, unknown>;
  status: string;
  last_seen: string;
}

export interface DeviceRegistrationRequest {
  protocol: string;
  hardware_id: string;
  device_type: string;
  manufacturer: string;
  model: string;
  room_id: string;
  property_id: string;
  capabilities: string[];
}

export interface DeviceCommand {
  action: string;
  params?: Record<string, unknown>;
}

export interface CommandResult {
  success: boolean;
  message?: string;
  new_state?: Record<string, unknown>;
}

export interface TelemetryQuery {
  start_time?: string;
  end_time?: string;
  interval?: string;
}

export interface TelemetryData {
  device_id: string;
  data_points: Record<string, unknown>[];
}

export interface ConciergeRequest {
  prompt: string;
  guest_id?: string;
  room_id?: string;
  language?: string;
  use_rag?: boolean;
  max_tokens?: number;
  temperature?: number;
}

export interface ConciergeResponse {
  id: string;
  content: string;
  finish_reason: string;
  sources?: Array<{
    document_id: string;
    content: string;
    score: number;
  }>;
  latency_ms: number;
}

export interface SentimentRequest {
  text: string;
  tenant_id?: string;
}

export interface SentimentResponse {
  score: number;
  confidence: number;
  label: string;
  aspects?: Array<{
    aspect: string;
    sentiment: number;
  }>;
}

export interface ForecastRequest {
  property_id: string;
  horizon_days: number;
  tenant_id?: string;
}

export interface ForecastResponse {
  property_id: string;
  forecast: Array<{
    date: string;
    predicted_occupancy: number;
    predicted_adr: number;
    predicted_revpar: number;
    confidence: number;
  }>;
  accuracy: number;
}

export interface RecommendationRequest {
  guest_id: string;
  context?: Record<string, unknown>;
}

export interface RecommendationResponse {
  guest_id: string;
  recommendations: Array<{
    type: string;
    title: string;
    description: string;
    score: number;
    image_url?: string;
  }>;
}

export { NexusClient };
export default NexusClient;
