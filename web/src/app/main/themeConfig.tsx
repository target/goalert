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
  setThemeMode: (newMode: ThemeName) => void
}

type SystemTheme = 'dark' | 'light'
type ThemeName = 'dark' | 'light' | 'system'

export const ThemeContext = React.createContext<ThemeContextParams>({
  themeMode: '',
  setThemeMode: (): void => {},
})
ThemeContext.displayName = 'ThemeContext'

function getPalette(mode: ThemeName): PaletteOptions {
  if (mode === 'dark') {
    return {
      mode: 'dark',
      secondary: grey,
    }
  }

  return {
    mode: 'light',
    primary: {
      ...grey,
      main: '#616161',
    },
    secondary: grey,
  }
}

function makeTheme(mode: ThemeName): Theme {
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

function saveTheme(theme: ThemeName): void {
  if (!window.localStorage) return
  window.localStorage.setItem('theme', theme)
}
function loadTheme(): ThemeName {
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
  const [savedTheme, setSavedTheme] = useState(loadTheme())
  const [systemTheme, setSystemTheme] = useState<SystemTheme>(
    window.matchMedia('(prefers-color-scheme: dark)').matches
      ? 'dark'
      : 'light',
  )

  useEffect(() => {
    const listener = (e: { matches: boolean }) =>
      setSystemTheme(e.matches ? 'dark' : 'light')
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
        themeMode: savedTheme,
        setThemeMode: (newMode: ThemeName) => {
          setSavedTheme(newMode)
          saveTheme(newMode)
        },
      }}
    >
      <MUIThemeProvider
        theme={makeTheme(savedTheme === 'system' ? systemTheme : savedTheme)}
      >
        {props.children}
      </MUIThemeProvider>
    </ThemeContext.Provider>
  )
}
