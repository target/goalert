import { useTheme, Breakpoint, Theme } from '@mui/material/styles'
import useMediaQuery from '@mui/material/useMediaQuery'

export function useIsWidthUp(breakpoint: Breakpoint): boolean {
  const theme = useTheme() as Theme
  return useMediaQuery(theme.breakpoints.up(breakpoint))
}

export function useIsWidthDown(breakpoint: Breakpoint): boolean {
  const theme = useTheme() as Theme
  return useMediaQuery(theme.breakpoints.down(breakpoint))
}
