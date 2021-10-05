import { useTheme, Breakpoint } from '@mui/material/styles'
import useMediaQuery from '@mui/material/useMediaQuery'

export function useIsWidthUp(breakpoint: Breakpoint): boolean {
  const theme = useTheme()
  return useMediaQuery(theme.breakpoints.up(breakpoint))
}

export function useIsWidthDown(breakpoint: Breakpoint): boolean {
  const theme = useTheme()
  return useMediaQuery(theme.breakpoints.down(breakpoint))
}
