import React, { ReactNode, useEffect, useState } from 'react'
import { PaletteOptions } from '@mui/material/styles/createPalette'
import { grey } from '@mui/material/colors'
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

function getPalette(mode: string, prefersDark: boolean): PaletteOptions {
  if (mode === 'dark' || (mode === 'system' && prefersDark)) {
    return {
      mode: 'dark',
      secondary: grey,
    }
  }

  if (mode === 'light' || (mode === 'system' && !prefersDark)) {
    return {
      mode: 'light',
      primary: {
        ...grey,
        main: '#616161',
      },
      secondary: grey,
    }
  }

  return {}
}

function makeTheme(mode: string, prefersDark: boolean): Theme {
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
    palette: getPalette(mode, prefersDark),
    ...testOverrides,
  })
}

export function ThemeProvider(props: ThemeProviderProps): JSX.Element {
  const storedTheme = localStorage.getItem('theme')
  const [themeMode, setThemeMode] = useState(storedTheme ?? 'system')
  const [prefersDark, setPrefersDark] = useState(
    window.matchMedia('(prefers-color-scheme: dark)').matches,
  )

  useEffect(() => {
    if (!storedTheme) {
      localStorage.setItem('theme', 'system')
    }

    if (themeMode === 'system') {
      const setTheme = (e: { matches: boolean }): void => {
        setPrefersDark(e.matches)
      }
      window
        .matchMedia('(prefers-color-scheme: dark)')
        .addEventListener('change', setTheme)

      return window
        .matchMedia('(prefers-color-scheme: dark)')
        .removeEventListener('change', setTheme)
    }
  }, [])

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
      <MUIThemeProvider theme={makeTheme(themeMode, prefersDark)}>
        {props.children}
      </MUIThemeProvider>
    </ThemeContext.Provider>
  )
}
