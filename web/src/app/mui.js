import { createTheme } from '@mui/styles'
import grey from '@mui/material/colors/grey'
import red from '@mui/material/colors/red'
import { isCypress } from './env'

let testOverrides = {}
if (isCypress) {
  testOverrides = {
    transitions: {
      // So we have `transition: none;` everywhere
      create: () => 'none',
    },
  }
}

export const theme = createTheme({
  palette: {
    primary: {
      ...grey,
      main: '#616161',
      500: '#616161',
      400: '#757575',
    },
    secondary: grey,
    error: red,
  },

  props: {
    MuiTextField: {
      variant: 'outlined',
    },
  },

  ...testOverrides,
})
