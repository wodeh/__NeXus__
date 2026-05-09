import { createContext, useContext } from 'react'

interface User {
  id: string
  role: string
}

interface AuthContextValue {
  user: User | null
  login: (user: User) => void
  logout: () => void
}

const AuthContext = createContext<AuthContextValue>({
  user: null,
  login: () => {},
  logout: () => {},
})

export const AuthProvider = AuthContext.Provider

export const useAuth = () => useContext(AuthContext)
