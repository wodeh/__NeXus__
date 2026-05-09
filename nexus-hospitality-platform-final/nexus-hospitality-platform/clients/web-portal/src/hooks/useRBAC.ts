import { createContext, useContext } from 'react'

interface RBACContextValue {
  hasPermission: (permission: string) => boolean
  hasRole: (role: string) => boolean
}

const RBACContext = createContext<RBACContextValue>({
  hasPermission: () => false,
  hasRole: () => false,
})

export const RBACProvider = RBACContext.Provider

export const useRBAC = () => useContext(RBACContext)
