// clients/web-portal/src/hooks/useOffline.ts
// ============================================================
// OFFLINE-FIRST ARCHITECTURE
// Supports: PWA, service workers, IndexedDB, background sync
// ============================================================

import { useEffect, useState, useCallback, useRef } from 'react';
import { atom, useAtom } from 'jotai';

interface SyncQueueItem {
  id: string;
  operation: 'create' | 'update' | 'delete';
  entity: string;
  data: any;
  timestamp: number;
  retryCount: number;
  status: 'pending' | 'syncing' | 'failed' | 'completed';
}

interface OfflineState {
  isOnline: boolean;
  syncQueue: SyncQueueItem[];
  lastSync: number;
  pendingChanges: number;
}

const offlineStateAtom = atom<OfflineState>({
  isOnline: navigator.onLine,
  syncQueue: [],
  lastSync: 0,
  pendingChanges: 0,
});

const DB_NAME = 'nexus-offline-db';
const DB_VERSION = 1;

class OfflineDB {
  private db: IDBDatabase | null = null;

  async init(): Promise<void> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open(DB_NAME, DB_VERSION);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        this.db = request.result;
        resolve();
      };

      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result;

        // Store for each entity type
        const stores = [
          'reservations', 'guests', 'rooms', 'housekeeping',
          'maintenance', 'inventory', 'folios', 'staff',
          'syncQueue', 'settings', 'cache'
        ];

        stores.forEach(storeName => {
          if (!db.objectStoreNames.contains(storeName)) {
            const store = db.createObjectStore(storeName, { keyPath: 'id' });
            store.createIndex('tenantId', 'tenantId', { unique: false });
            store.createIndex('updatedAt', 'updatedAt', { unique: false });
            store.createIndex('syncStatus', 'syncStatus', { unique: false });
          }
        });
      };
    });
  }

  async get<T>(storeName: string, id: string): Promise<T | undefined> {
    if (!this.db) await this.init();
    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([storeName], 'readonly');
      const store = transaction.objectStore(storeName);
      const request = store.get(id);

      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error);
    });
  }

  async getAll<T>(storeName: string, index?: string, query?: IDBKeyRange): Promise<T[]> {
    if (!this.db) await this.init();
    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([storeName], 'readonly');
      const store = transaction.objectStore(storeName);
      const target = index ? store.index(index) : store;
      const request = query ? target.getAll(query) : target.getAll();

      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error);
    });
  }

  async put(storeName: string, data: any): Promise<void> {
    if (!this.db) await this.init();
    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([storeName], 'readwrite');
      const store = transaction.objectStore(storeName);
      const request = store.put({ ...data, updatedAt: Date.now(), syncStatus: 'pending' });

      request.onsuccess = () => resolve();
      request.onerror = () => reject(request.error);
    });
  }

  async delete(storeName: string, id: string): Promise<void> {
    if (!this.db) await this.init();
    return new Promise((resolve, reject) => {
      const transaction = this.db!.transaction([storeName], 'readwrite');
      const store = transaction.objectStore(storeName);
      const request = store.delete(id);

      request.onsuccess = () => resolve();
      request.onerror = () => reject(request.error);
    });
  }

  async addToSyncQueue(item: Omit<SyncQueueItem, 'id' | 'timestamp' | 'retryCount' | 'status'>): Promise<void> {
    const queueItem: SyncQueueItem = {
      ...item,
      id: `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      timestamp: Date.now(),
      retryCount: 0,
      status: 'pending',
    };

    await this.put('syncQueue', queueItem);
  }

  async getPendingSyncItems(): Promise<SyncQueueItem[]> {
    return this.getAll<SyncQueueItem>('syncQueue', 'syncStatus', IDBKeyRange.only('pending'));
  }

  async markSynced(id: string): Promise<void> {
    const item = await this.get<SyncQueueItem>('syncQueue', id);
    if (item) {
      await this.put('syncQueue', { ...item, status: 'completed', syncStatus: 'synced' });
    }
  }
}

export const offlineDB = new OfflineDB();

export const useOffline = () => {
  const [state, setState] = useAtom(offlineStateAtom);
  const syncInProgress = useRef(false);

  // Initialize IndexedDB on mount
  useEffect(() => {
    offlineDB.init().catch(console.error);
  }, []);

  // Monitor online status
  useEffect(() => {
    const handleOnline = () => {
      setState(prev => ({ ...prev, isOnline: true }));
      triggerSync();
    };

    const handleOffline = () => {
      setState(prev => ({ ...prev, isOnline: false }));
    };

    window.addEventListener('online', handleOnline);
    window.addEventListener('offline', handleOffline);

    return () => {
      window.removeEventListener('online', handleOnline);
      window.removeEventListener('offline', handleOffline);
    };
  }, [setState]);

  // Background sync when app comes to foreground
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.visibilityState === 'visible' && navigator.onLine) {
        triggerSync();
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => document.removeEventListener('visibilitychange', handleVisibilityChange);
  }, []);

  const queueOperation = useCallback(async (
    operation: 'create' | 'update' | 'delete',
    entity: string,
    data: any
  ) => {
    await offlineDB.addToSyncQueue({ operation, entity, data });

    setState(prev => ({
      ...prev,
      syncQueue: [...prev.syncQueue, {
        id: `${Date.now()}`,
        operation,
        entity,
        data,
        timestamp: Date.now(),
        retryCount: 0,
        status: 'pending',
      }],
      pendingChanges: prev.pendingChanges + 1,
    }));

    // If online, try to sync immediately
    if (navigator.onLine) {
      triggerSync();
    }
  }, [setState]);

  const triggerSync = useCallback(async () => {
    if (syncInProgress.current || !navigator.onLine) return;
    syncInProgress.current = true;

    try {
      const pendingItems = await offlineDB.getPendingSyncItems();

      for (const item of pendingItems) {
        try {
          // Attempt to sync with backend
          const response = await fetch(`/api/${item.entity}`, {
            method: item.operation === 'delete' ? 'DELETE' : 
                    item.operation === 'create' ? 'POST' : 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(item.data),
          });

          if (response.ok) {
            await offlineDB.markSynced(item.id);
          } else {
            // Mark for retry
            await offlineDB.put('syncQueue', {
              ...item,
              retryCount: item.retryCount + 1,
              status: item.retryCount > 3 ? 'failed' : 'pending',
            });
          }
        } catch (error) {
          console.error(`Sync failed for ${item.id}:`, error);
        }
      }

      setState(prev => ({
        ...prev,
        lastSync: Date.now(),
        pendingChanges: pendingItems.length,
      }));
    } finally {
      syncInProgress.current = false;
    }
  }, [setState]);

  const cacheData = useCallback(async (key: string, data: any, ttl?: number) => {
    await offlineDB.put('cache', {
      id: key,
      data,
      expiresAt: ttl ? Date.now() + ttl : null,
    });
  }, []);

  const getCachedData = useCallback(async <T>(key: string): Promise<T | undefined> => {
    const cached = await offlineDB.get<{ data: T; expiresAt: number | null }>('cache', key);

    if (!cached) return undefined;
    if (cached.expiresAt && cached.expiresAt < Date.now()) {
      await offlineDB.delete('cache', key);
      return undefined;
    }

    return cached.data;
  }, []);

  return {
    isOnline: state.isOnline,
    pendingChanges: state.pendingChanges,
    lastSync: state.lastSync,
    queueOperation,
    triggerSync,
    cacheData,
    getCachedData,
  };
};

export const OfflineProvider: React.FC<{ 
  children: React.ReactNode; 
  isOnline: boolean;
}> = ({ children }) => {
  return <>{children}</>;
};
