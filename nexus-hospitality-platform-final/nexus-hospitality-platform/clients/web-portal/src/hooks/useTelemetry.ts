import { createContext, useContext } from 'react'

interface TelemetryContextValue {
  trackEvent: (name: string, props?: Record<string, any>) => void
}

const TelemetryContext = createContext<TelemetryContextValue>({
  trackEvent: () => {},
})

export const TelemetryProvider = TelemetryContext.Provider

export const useTelemetry = () => useContext(TelemetryContext)
