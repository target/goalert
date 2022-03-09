import { useTheme } from '@mui/material'

type StatusColors = {
  ok: string
  warn: string
  err: string
}

function useStatusColors(): StatusColors {
  const theme = useTheme()
  return {
    ok: theme.palette.success.main,
    warn: theme.palette.warning.main,
    err: theme.palette.error.main,
  }
}

export default useStatusColors
