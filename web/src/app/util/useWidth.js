import { useTheme } from '@material-ui/core/styles'
import useMediaQuery from '@material-ui/core/useMediaQuery'

/**
 * https://material-ui.com/components/use-media-query/#migrating-from-withwidth
 *
 * Only works because the number of with default mui breakpoints.
 * It will break once you change the number of breakpoints.
 * See https://reactjs.org/docs/hooks-rules.html#only-call-hooks-at-the-top-level
 */
export default function useWidth() {
  const theme = useTheme()
  const keys = [...theme.breakpoints.keys].reverse()

  return (
    keys.reduce((output, key) => {
      const matches = useMediaQuery(theme.breakpoints.up(key))
      return !output && matches ? key : output
    }, null) || 'xs'
  )
}
