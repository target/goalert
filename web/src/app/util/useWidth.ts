import { Theme, useMediaQuery } from '@material-ui/core'
import { Breakpoint } from '@material-ui/core/styles/createBreakpoints'
import { useTheme } from '@material-ui/styles'

export function useIsWidthUp(breakpoint: Breakpoint): boolean {
  const theme = useTheme() as Theme
  return useMediaQuery(theme.breakpoints.up(breakpoint))
}

export function useIsWidthDown(breakpoint: Breakpoint): boolean {
  const theme = useTheme() as Theme
  return useMediaQuery(theme.breakpoints.down(breakpoint))
}
