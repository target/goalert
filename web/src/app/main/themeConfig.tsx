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

// palette generated from https://material-foundation.github.io/material-theme-builder/#/custom
function getPalette(mode: string): PaletteOptions {
  const prefersDark = window.matchMedia('(prefers-color-scheme: dark)')

  if (mode === 'light' || (mode === 'system' && !prefersDark)) {
    return {
      mode: 'light',
      primary: {
        main: '#006684',
        light: '#bbe9ff',
        dark: '#001f2a',
      },
      secondary: { main: '#4d616b', light: '#d0e6f3', dark: '#081e27' },
      background: {
        default: '#fbfcfe',
        paper: '#dce3e8', // m3 surface variant
      },
      error: { main: '#ba1b1b', light: '#ffdad4', dark: '#410001' },
    }
  }

  if (mode === 'dark' || (mode === 'system' && prefersDark)) {
    return {
      mode: 'dark',
      primary: { main: '#64d3ff', light: '#bbe9ff', dark: '#004d65' },
      secondary: { main: '#b5cad6', light: '#d0e6f3', dark: '#354a54' },
      background: {
        default: '#191c1e',
        paper: '#40484c', // m3 surface variant
      },
      error: { main: '#ffb4a9', light: '#ffdad4', dark: '#930006' },
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
