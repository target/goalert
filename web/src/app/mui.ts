import { createTheme, adaptV4Theme } from '@mui/material/styles'
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

export const theme = createTheme(
  adaptV4Theme({
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
  }),
)
