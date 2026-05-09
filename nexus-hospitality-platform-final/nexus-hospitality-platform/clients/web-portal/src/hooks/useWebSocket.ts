import { createContext, useContext } from 'react'

interface WebSocketContextValue {
  socket: any
}

const WebSocketContext = createContext<WebSocketContextValue>({
  socket: null,
})

export const WebSocketProvider = WebSocketContext.Provider

export const useWebSocket = () => useContext(WebSocketContext)
