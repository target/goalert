import createMuiTheme from '@material-ui/core/styles/createMuiTheme'
import grey from '@material-ui/core/colors/grey'
import red from '@material-ui/core/colors/red'

let testOverrides = {}
if (global.Cypress) {
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
  typography: {
    useNextVariants: true,
  },
  ...testOverrides,
})
