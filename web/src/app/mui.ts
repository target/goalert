import { createMuiTheme } from '@material-ui/core/styles'
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
      main: '#6200ee',
    },
  },

  // override default props
  props: {
    MuiList: {
      disablePadding: true,
    },
  },

  ...testOverrides,
})
