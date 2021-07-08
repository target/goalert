import React, { ReactNode, useState } from 'react'
import { AppContext } from '../context'
import { PaletteOptions } from '@material-ui/core/styles/createPalette'
import { isCypress } from '../../env'
import {
  createMuiTheme,
  MuiThemeProvider,
  Theme,
} from '@material-ui/core/styles'

interface ThemeProviderProps {
  children: Array<ReactNode> | ReactNode
}

function getThemePreference(): string {
  const theme = localStorage.getItem('theme')
  if (!theme) {
    const mq = window.matchMedia('(prefers-color-scheme: dark)')
    return mq.matches ? 'dark' : 'light'
  }
  return theme
}

function handleSetTheme(theme: string): string {
  if (theme === 'dark') {
    localStorage.setItem('theme', 'light')
    return 'light'
  }
  localStorage.setItem('theme', 'dark')
  return 'dark'
}

function getPalette(theme: string): PaletteOptions {
  switch (theme) {
    case 'dark':
      return {
        type: 'dark',
        primary: {
          main: '#BB86FC',
        },
      }

    case 'light':
    default:
      return {
        primary: {
          main: '#6200ee',
        },
      }
  }
}

function makeTheme(theme: string): Theme {
  let testOverrides = {}
  if (isCypress) {
    testOverrides = {
      transitions: {
        // So we have `transition: none;` everywhere
        create: () => 'none',
      },
    }
  }

  return createMuiTheme({
    palette: getPalette(theme),

    // override default props
    props: {
      MuiCard: {
        variant: 'outlined',
      },
      MuiFilledInput: {
        disableUnderline: true,
        style: {
          borderRadius: 4,
        },
      },
    },

    ...testOverrides,
  })
}

export function ThemeProvider(props: ThemeProviderProps): JSX.Element {
  const [theme, setTheme] = useState(getThemePreference())

  return (
    <AppContext.Provider
      value={{
        theme,
        setTheme: (theme: string) => {
          const newTheme = handleSetTheme(theme)
          setTheme(newTheme)
        },
      }}
    >
      <MuiThemeProvider theme={makeTheme(theme)}>
        {props.children}
      </MuiThemeProvider>
    </AppContext.Provider>
  )
}
