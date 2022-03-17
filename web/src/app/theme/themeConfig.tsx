import React, { ReactNode, useEffect, useState } from 'react'
import { PaletteOptions } from '@mui/material/styles/createPalette'
import { isCypress } from '../env'
import {
  createTheme,
  Theme,
  ThemeProvider as MUIThemeProvider,
} from '@mui/material/styles'
import {
  argbFromHex,
  hexFromArgb,
  themeFromSourceColor,
  Scheme,
} from '@material/material-color-utilities'
import {
  blueGrey,
  teal,
  green,
  deepPurple,
  pink,
  red,
  amber,
} from '@mui/material/colors'

interface ThemeProviderProps {
  children: ReactNode
}

interface ThemeContextParams {
  themeMode: string
  setThemeMode: (newMode: ThemeModeOption) => void
  sourceColor: string
  setSourceColor: (newColor: string) => void
}

type MUIThemeMode = 'dark' | 'light'
type ThemeModeOption = 'dark' | 'light' | 'system'

export const sourceColors = [
  blueGrey[700],
  teal[700],
  green[700],
  deepPurple[700],
  pink[700],
  red[700],
  amber[700],
]

export const ThemeContext = React.createContext<ThemeContextParams>({
  themeMode: '',
  setThemeMode: (): void => {},
  sourceColor: '',
  setSourceColor: (): void => {},
})
ThemeContext.displayName = 'ThemeContext'

function makePalette(
  scheme: Scheme,
  useSurfaceVariant?: boolean,
): PaletteOptions {
  return {
    primary: {
      main: hexFromArgb(scheme.primary),
    },
    secondary: {
      main: hexFromArgb(scheme.secondary),
    },
    background: {
      default: hexFromArgb(scheme.background),
      paper: useSurfaceVariant
        ? hexFromArgb(scheme.surfaceVariant)
        : hexFromArgb(scheme.surface),
    },
    error: {
      main: hexFromArgb(scheme.error),
    },
  }
}

// palette generated from https://material-foundation.github.io/material-theme-builder/#/custom
function getPalette(
  mode: MUIThemeMode,
  sourceColorHex: string,
): PaletteOptions {
  const sourceColor = argbFromHex(sourceColorHex)
  const theme = themeFromSourceColor(sourceColor)

  if (mode === 'dark') {
    return {
      mode: 'dark',
      ...makePalette(theme.schemes.dark),
    }
  }

  return {
    mode: 'light',
    ...makePalette(theme.schemes.light, true),
  }
}

function makeTheme(mode: MUIThemeMode, sourceColor: string): Theme {
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
    palette: getPalette(mode, sourceColor),
    components: {
      MuiIconButton: {
        defaultProps: {
          color: 'primary',
        },
      },
    },
    ...testOverrides,
  })
}

function saveThemeMode(theme: ThemeModeOption): void {
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

function saveThemeColor(hex: string): void {
  if (!window.localStorage) return
  window.localStorage.setItem('themeColor', hex)
}
function loadThemeColor(): string {
  return window?.localStorage?.getItem('themeColor') ?? sourceColors[0]
}

export function ThemeProvider(props: ThemeProviderProps): JSX.Element {
  const [savedThemeMode, setSavedThemeMode] = useState(loadTheme())
  const [sourceColor, setSourceColor] = useState(loadThemeColor())

  // used for watching if system theme mode changes
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
          saveThemeMode(newMode)
        },
        sourceColor,
        setSourceColor: (newColor: string) => {
          setSourceColor(newColor)
          saveThemeColor(newColor)
        },
      }}
    >
      <MUIThemeProvider
        theme={makeTheme(
          savedThemeMode === 'system' ? systemThemeMode : savedThemeMode,
          sourceColor,
        )}
      >
        {props.children}
      </MUIThemeProvider>
    </ThemeContext.Provider>
  )
}
