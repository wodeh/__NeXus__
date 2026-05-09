import { create } from 'zustand';

interface AuthState {
  tenantId: string | null;
  propertyId: string | null;
  setTenant: (id: string) => void;
  setProperty: (id: string) => void;
}

export const useAuthStore = create<AuthState>((set) => ({
  tenantId: null,
  propertyId: null,
  setTenant: (id) => set({ tenantId: id }),
  setProperty: (id) => set({ propertyId: id }),
}));
