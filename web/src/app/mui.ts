import { createTheme } from '@mui/material/styles'
import { grey, red } from '@mui/material/colors'
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
    },
    secondary: grey,
    error: red,
  },

  ...testOverrides,
})
