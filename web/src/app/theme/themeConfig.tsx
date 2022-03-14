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
  setThemeMode: (newMode: ThemeModeOption) => void
}

type MUIThemeMode = 'dark' | 'light'
type ThemeModeOption = 'dark' | 'light' | 'system'

export const ThemeContext = React.createContext<ThemeContextParams>({
  themeMode: '',
  setThemeMode: (): void => {},
})
ThemeContext.displayName = 'ThemeContext'

// palette generated from https://material-foundation.github.io/material-theme-builder/#/custom
function getPalette(mode: MUIThemeMode): PaletteOptions {
  if (mode === 'dark') {
    return {
      mode: 'dark',
      primary: { main: '#64d3ff', light: '#bbe9ff', dark: '#004d65' },
      secondary: { main: '#b5cad6', light: '#d0e6f3', dark: '#354a54' },
      background: {
        default: '#191c1e',
        paper: '#191c1e', // m3 surface
      },
      error: { main: '#ffb4a9', light: '#ffdad4', dark: '#930006' },
    }
  }

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

function makeTheme(mode: MUIThemeMode): Theme {
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

function saveTheme(theme: ThemeModeOption): void {
  if (!window.localStorage) return
  window.localStorage.setItem('theme', theme)
}

function loadTheme(): ThemeModeOption {
  if (!window.localStorage) return 'system'

  const theme = window.localStorage.getItem('theme')
  switch (theme) {
    case 'dark':
    case 'light':
      return theme
  }

  return 'system'
}

export function ThemeProvider(props: ThemeProviderProps): JSX.Element {
  const [savedThemeMode, setSavedThemeMode] = useState(loadTheme())
  const [systemThemeMode, setSystemThemeMode] = useState<MUIThemeMode>(
    window.matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark'
      : 'light',
  )

  useEffect(() => {
    const listener = (e: { matches: boolean }): void => {
      setSystemThemeMode(e.matches ? 'dark' : 'light')
    }
    window
      .matchMedia('(prefers-color-scheme: dark)')
      .addEventListener('change', listener)

    return () =>
      window
        .matchMedia('(prefers-color-scheme: dark)')
        .removeEventListener('change', listener)
  }, [])

  return (
    <ThemeContext.Provider
      value={{
        themeMode: savedThemeMode,
        setThemeMode: (newMode: ThemeModeOption) => {
          setSavedThemeMode(newMode)
          saveTheme(newMode)
        },
      }}
    >
      <MUIThemeProvider
        theme={makeTheme(
          savedThemeMode === 'system' ? systemThemeMode : savedThemeMode,
        )}
      >
        {props.children}
      </MUIThemeProvider>
    </ThemeContext.Provider>
  )
}
