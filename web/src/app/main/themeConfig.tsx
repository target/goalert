import React, { ReactNode, useEffect, useState } from 'react'
import { PaletteOptions } from '@mui/material/styles/createPalette'
import { isCypress } from '../env'
import {
  createTheme,
  Theme,
  ThemeProvider as MUIThemeProvider,
} from '@mui/material/styles'

interface ThemeProviderProps {
  children: ReactNode
}

interface ThemeContextParams {
  themeMode: string
  setThemeMode: (newMode: string) => void
}

export const ThemeContext = React.createContext<ThemeContextParams>({
  themeMode: '',
  setThemeMode: (): void => {},
})
ThemeContext.displayName = 'ThemeContext'

function handleStoreThemeMode(theme: string): boolean {
  if (theme === 'dark') {
    localStorage.setItem('theme', 'dark')
    return true
  }
  if (theme === 'light') {
    localStorage.setItem('theme', 'light')
    return true
  }
  if (theme === 'system') {
    localStorage.setItem('theme', 'system')
    return true
  }

  console.warn('unknown theme, aborting')
  return false
}

function getPalette(mode: string): PaletteOptions {
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)')

  if (mode === 'dark' || (mode === 'system' && prefersDark)) {
    return {
      mode: 'dark',
      primary: { main: '#62D3FF', light: '#B6EAFF', dark: '#004D62' },
      secondary: { main: '#B4CAD6', light: '#CFE6F1', dark: '#354A53' },
      background: {
        default: '#191C1E',
        paper: '#40484C',
      },
      error: { main: '#FFB4A9' },
    }
  }

  if (mode === 'light' || (mode === 'system' && !prefersDark)) {
    return {
      mode: 'light',
      primary: {
        main: '#006684',
        light: '#B6EAFF',
        dark: '#001F2A',
      },
      secondary: { main: '#4C616B', light: '#CFE6F1', dark: '#071E26' },
      background: {
        default: '#FBFCFE',
        paper: '#DCE4E9',
      },
      error: { main: '#BA1B1B' },
    }
  }

  return {}
}

function makeTheme(mode: string): Theme {
  let testOverrides = {}
  if (isCypress) {
    testOverrides = {
      transitions: {
        // So we have `transition: none;` everywhere
        create: () => 'none',
      },
    }
  }

  return createTheme({
    palette: getPalette(mode),
    ...testOverrides,
  })
}

export function ThemeProvider(props: ThemeProviderProps): JSX.Element {
  const storedTheme = localStorage.getItem('theme')
  useEffect(() => {
    if (!storedTheme) {
      localStorage.setItem('theme', 'system')
    }
  }, [])
  const [themeMode, setThemeMode] = useState(storedTheme ?? 'system')

  return (
    <ThemeContext.Provider
      value={{
        themeMode,
        setThemeMode: (newMode: string) => {
          if (handleStoreThemeMode(newMode)) {
            setThemeMode(newMode)
          }
        },
      }}
    >
      <MUIThemeProvider theme={makeTheme(themeMode)}>
        {props.children}
      </MUIThemeProvider>
    </ThemeContext.Provider>
  )
}
