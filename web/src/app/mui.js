import { createMuiTheme } from '@material-ui/core/styles'
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

export const theme = createMuiTheme({
  palette: {
    primary: {
      ...grey,
      '500': '#616161',
      '400': '#757575',
    },
    secondary: grey,
    error: red,
  },

  // override default props
  props: {
    MuiList: {
      disablePadding: true,
    },
  },

  ...testOverrides,
})
