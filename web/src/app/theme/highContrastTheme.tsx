import { createTheme } from '@mui/material/styles'

const black = '#000000'
const white = '#ffffff'

export const highContrastLightTheme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: black,
    },
    secondary: {
      main: black,
    },
    background: {
      default: white,
      paper: white,
    },
    text: {
      primary: black,
      secondary: black,
    },
    divider: 'black',
  },
  components: {
    MuiCard: {
      defaultProps: {
        variant: 'outlined',
      },
    },
    MuiIconButton: {
      defaultProps: {
        sx: { color: black },
      },
    },
  },
})

export const highContrastDarkTheme = createTheme({
  palette: {
    mode: 'dark',
    primary: {
      main: white,
    },
    secondary: {
      main: white,
    },
    background: {
      default: black,
      paper: black,
    },
    text: {
      primary: white,
      secondary: white,
    },
    divider: white,
  },
  components: {
    MuiCard: {
      defaultProps: {
        variant: 'outlined',
      },
    },
    MuiIconButton: {
      defaultProps: {
        sx: { color: white },
      },
    },
  },
})
