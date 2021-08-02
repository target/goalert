import { createTheme } from '@material-ui/core/styles'
import grey from '@material-ui/core/colors/grey'
import red from '@material-ui/core/colors/red'
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

  ...testOverrides,
})
