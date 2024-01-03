// Import styles, initialize component theme here.
// import '../src/common.css';
import React from 'react'
import { beforeMount } from '@playwright/experimental-ct-react/hooks'
import { ConfigProvider } from '../web/src/app/util/RequireConfig'
import { Provider as URQLProvider } from 'urql'
import { client as urqlClient } from '../web/src/app/urql'
import { StyledEngineProvider } from '@mui/material'
import { ThemeProvider } from '../web/src/app/theme/themeConfig'

import { Settings } from 'luxon'
import { DestTypeProvider } from '../web/src/app/util/useDestinationTypes'
Settings.throwOnInvalid = true

beforeMount(async ({ App }) => {
  return (
    <StyledEngineProvider injectFirst>
      <ThemeProvider>
        <URQLProvider value={urqlClient}>
          <React.Suspense>
            <ConfigProvider>
              <DestTypeProvider>
                <App />
              </DestTypeProvider>
            </ConfigProvider>
          </React.Suspense>
        </URQLProvider>
      </ThemeProvider>
    </StyledEngineProvider>
  )
})
