import React, { ReactNode, useState } from 'react'
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

export const ThemeContext = React.createContext({
  themeMode: '',
  toggleThemeMode: (): void => {},
})
ThemeContext.displayName = 'ThemeContext'

function getThemePreference(): string {
  const theme = localStorage.getItem('theme')
  if (!theme) {
    // get if system preference is set to dark mode
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    return mediaQuery.matches ? 'dark' : 'light'
  }
  return theme
}

function handleSetThemeMode(theme: string): string {
  if (theme === 'dark') {
    localStorage.setItem('theme', 'light')
    return 'light'
  }
  localStorage.setItem('theme', 'dark')
  return 'dark'
}

function getPalette(mode: string): PaletteOptions {
  switch (mode) {
    case 'dark':
      return {
        mode: 'dark',
      }

    case 'light':
    default:
      return {
        mode: 'light',
        primary: {
          ...grey,
          main: '#616161',
        },
      }
  }
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
  const [themeMode, setThemeMode] = useState(getThemePreference())

  return (
    <ThemeContext.Provider
      value={{
        themeMode,
        toggleThemeMode: () => {
          const newTheme = handleSetThemeMode(themeMode)
          setThemeMode(newTheme)
        },
      }}
    >
      <MUIThemeProvider theme={makeTheme(themeMode)}>
        {props.children}
      </MUIThemeProvider>
    </ThemeContext.Provider>
  )
}
