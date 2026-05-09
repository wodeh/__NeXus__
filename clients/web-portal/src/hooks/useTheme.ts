import { createContext, useContext } from 'react'

interface ThemeContextValue {
  theme: string
  setTheme: (theme: string) => void
}

const ThemeContext = createContext<ThemeContextValue>({
  theme: 'light',
  setTheme: () => {},
})

export const ThemeProvider = ThemeContext.Provider

export const useTheme = () => useContext(ThemeContext)
