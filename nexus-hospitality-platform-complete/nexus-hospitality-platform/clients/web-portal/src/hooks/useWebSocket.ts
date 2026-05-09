// clients/web-portal/src/hooks/useWebSocket.ts
// ============================================================
// REAL-TIME WEBSOCKET ARCHITECTURE
// Supports: guest notifications, staff alerts, IoT events, 
// IPTV sync, PMS updates, AI inference streaming
// ============================================================

import { useEffect, useRef, useCallback, useState } from 'react';
import { atom, useAtom } from 'jotai';
import { io, Socket } from 'socket.io-client';

interface WebSocketConfig {
  url: string;
  tenantId: string;
  propertyId?: string;
  roomId?: string;
  guestId?: string;
  authToken: string;
  reconnectInterval?: number;
  maxReconnectAttempts?: number;
}

interface WebSocketState {
  connected: boolean;
  authenticated: boolean;
  lastPing: number;
  latency: number;
  reconnectAttempts: number;
}

const socketAtom = atom<Socket | null>(null);
const wsStateAtom = atom<WebSocketState>({
  connected: false,
  authenticated: false,
  lastPing: 0,
  latency: 0,
  reconnectAttempts: 0,
});

export const useWebSocket = () => {
  const [socket, setSocket] = useAtom(socketAtom);
  const [state, setState] = useAtom(wsStateAtom);
  const pingIntervalRef = useRef<NodeJS.Timeout>();
  const configRef = useRef<WebSocketConfig>();

  const connect = useCallback((config: WebSocketConfig) => {
    configRef.current = config;

    const newSocket = io(config.url, {
      transports: ['websocket'],
      auth: {
        token: config.authToken,
      },
      query: {
        tenantId: config.tenantId,
        propertyId: config.propertyId,
        roomId: config.roomId,
        guestId: config.guestId,
      },
      reconnection: true,
      reconnectionAttempts: config.maxReconnectAttempts || 10,
      reconnectionDelay: config.reconnectInterval || 5000,
      reconnectionDelayMax: 30000,
    });

    newSocket.on('connect', () => {
      setState(prev => ({
        ...prev,
        connected: true,
        reconnectAttempts: 0,
      }));

      // Start ping monitoring
      pingIntervalRef.current = setInterval(() => {
        const start = Date.now();
        newSocket.emit('ping', () => {
          const latency = Date.now() - start;
          setState(prev => ({ ...prev, latency, lastPing: Date.now() }));
        });
      }, 30000);
    });

    newSocket.on('disconnect', (reason) => {
      setState(prev => ({ ...prev, connected: false }));
      if (pingIntervalRef.current) {
        clearInterval(pingIntervalRef.current);
      }

      // Log disconnect for telemetry
      console.warn(`WebSocket disconnected: ${reason}`);
    });

    newSocket.on('auth_success', () => {
      setState(prev => ({ ...prev, authenticated: true }));
    });

    newSocket.on('auth_error', (error) => {
      setState(prev => ({ ...prev, authenticated: false }));
      console.error('WebSocket auth error:', error);
    });

    newSocket.on('reconnect_attempt', (attempt) => {
      setState(prev => ({ ...prev, reconnectAttempts: attempt }));
    });

    newSocket.on('reconnect_failed', () => {
      setState(prev => ({ ...prev, connected: false, authenticated: false }));
      // Trigger offline mode
      window.dispatchEvent(new CustomEvent('nexus:offline'));
    });

    // Subscribe to real-time channels based on user role
    newSocket.on('reservation_update', (data) => {
      window.dispatchEvent(new CustomEvent('nexus:reservation_update', { detail: data }));
    });

    newSocket.on('room_status_change', (data) => {
      window.dispatchEvent(new CustomEvent('nexus:room_status_change', { detail: data }));
    });

    newSocket.on('housekeeping_alert', (data) => {
      window.dispatchEvent(new CustomEvent('nexus:housekeeping_alert', { detail: data }));
    });

    newSocket.on('maintenance_urgent', (data) => {
      window.dispatchEvent(new CustomEvent('nexus:maintenance_urgent', { detail: data }));
    });

    newSocket.on('guest_notification', (data) => {
      window.dispatchEvent(new CustomEvent('nexus:guest_notification', { detail: data }));
    });

    newSocket.on('iptv_event', (data) => {
      window.dispatchEvent(new CustomEvent('nexus:iptv_event', { detail: data }));
    });

    newSocket.on('iot_telemetry', (data) => {
      window.dispatchEvent(new CustomEvent('nexus:iot_telemetry', { detail: data }));
    });

    newSocket.on('ai_inference_stream', (data) => {
      window.dispatchEvent(new CustomEvent('nexus:ai_stream', { detail: data }));
    });

    newSocket.on('system_alert', (data) => {
      window.dispatchEvent(new CustomEvent('nexus:system_alert', { detail: data }));
    });

    setSocket(newSocket);
    return newSocket;
  }, [setSocket, setState]);

  const disconnect = useCallback(() => {
    if (socket) {
      socket.disconnect();
      setSocket(null);
    }
    if (pingIntervalRef.current) {
      clearInterval(pingIntervalRef.current);
    }
  }, [socket, setSocket]);

  const emit = useCallback((event: string, data: any) => {
    if (socket?.connected) {
      socket.emit(event, data);
    } else {
      // Queue for retry when connection restored
      queueOfflineEvent(event, data);
    }
  }, [socket]);

  const subscribe = useCallback((channel: string, handler: (data: any) => void) => {
    if (socket) {
      socket.on(channel, handler);
      return () => socket.off(channel, handler);
    }
    return () => {};
  }, [socket]);

  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  return {
    socket,
    state,
    connect,
    disconnect,
    emit,
    subscribe,
  };
};

// Offline event queue
const offlineQueue: Array<{ event: string; data: any; timestamp: number }> = [];

function queueOfflineEvent(event: string, data: any) {
  offlineQueue.push({ event, data, timestamp: Date.now() });
  // Persist to IndexedDB for recovery
  persistOfflineQueue();
}

function persistOfflineQueue() {
  if ('indexedDB' in window) {
    const request = indexedDB.open('nexus-offline', 1);
    request.onupgradeneeded = (event) => {
      const db = (event.target as IDBOpenDBRequest).result;
      db.createObjectStore('events', { keyPath: 'id', autoIncrement: true });
    };
    request.onsuccess = (event) => {
      const db = (event.target as IDBOpenDBRequest).result;
      const transaction = db.transaction(['events'], 'readwrite');
      const store = transaction.objectStore('events');
      offlineQueue.forEach(item => store.add(item));
    };
  }
}

export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  return <>{children}</>;
};
