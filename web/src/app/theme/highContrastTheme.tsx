import { createTheme, Theme } from '@mui/material/styles'

const black = '#000000'
const white = '#ffffff'
const grey = '#aaaaaa' // meets 9:1 contrast using black text for component highlighting

export function makeHighContrastTheme(mode: 'light' | 'dark'): Theme {
  const primary = mode === 'light' ? black : white
  const bg = mode === 'light' ? white : black

  return createTheme({
    palette: {
      mode,
      primary: {
        main: primary,
      },
      secondary: {
        main: primary,
      },
      background: {
        default: bg,
        paper: bg,
      },
      text: {
        primary: primary,
        secondary: primary,
      },
      divider: primary,
      action: {
        active: grey,
        selected: grey,
      },
    },
    components: {
      MuiCard: {
        defaultProps: {
          variant: 'outlined',
        },
      },
      MuiSvgIcon: {
        defaultProps: {
          sx: { color: primary },
        },
      },
      MuiIcon: {
        defaultProps: {
          sx: { color: primary },
        },
      },
      MuiIconButton: {
        defaultProps: {
          sx: { color: primary },
        },
      },
      MuiListItem: {
        defaultProps: {
          sx: {
            '&.Mui-selected': {
              backgroundColor: grey,
            },
          },
        },
      },
    },
  })
}
