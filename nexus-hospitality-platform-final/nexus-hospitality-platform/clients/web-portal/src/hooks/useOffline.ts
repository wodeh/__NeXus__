import { createContext, useContext } from 'react'

interface OfflineContextValue {
  isOnline: boolean
}

const OfflineContext = createContext<OfflineContextValue>({
  isOnline: true,
})

export const OfflineProvider = OfflineContext.Provider

export const useOffline = () => useContext(OfflineContext)
