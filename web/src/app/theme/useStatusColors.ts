import { useTheme } from '@mui/material'
import { yellow } from '@mui/material/colors'

type StatusColors = {
  ok: string
  warn: string
  err: string
}

function useStatusColors(): StatusColors {
  const theme = useTheme()
  return {
    ok: theme.palette.success.main,
    warn: yellow[600], // TODO if practical, use theme.palette.warning.main,
    err: theme.palette.error.main,
  }
}

export default useStatusColors
