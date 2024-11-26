import { useContext, useEffect, useState } from 'react'
import { ThemeContext, MUIThemeMode } from './themeConfig'

export function useTheme(): MUIThemeMode {
  const ctx = useContext(ThemeContext)

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

  if (ctx.themeMode === 'system') {
    return systemThemeMode
  }

  return ctx.themeMode as MUIThemeMode
}
