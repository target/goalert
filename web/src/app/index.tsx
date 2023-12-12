import React, { StrictMode, Suspense } from 'react'
import { createRoot } from 'react-dom/client'
import { Provider as ReduxProvider } from 'react-redux'
import { ApolloProvider } from '@apollo/client'
import { StyledEngineProvider } from '@mui/material/styles'

import { GOALERT_VERSION, pathPrefix } from './env'
import { ThemeProvider } from './theme/themeConfig'
import { GraphQLClient } from './apollo'
import './styles'
import App from './main/App'
import store from './reduxStore'
import { ConfigProvider } from './util/RequireConfig'
import { warn } from './util/debug'
import NewVersionCheck from './NewVersionCheck'
import { Provider as URQLProvider } from 'urql'
import { client as urqlClient } from './urql'
import { Router } from 'wouter'

import { Settings } from 'luxon'
import RequireAuth from './main/RequireAuth'
import Login from './main/components/Login'
import { DestTypeProvider } from './util/useDestinationTypes'
import Spinner from './loading/components/Spinner'

Settings.throwOnInvalid = true

declare module 'luxon' {
  interface TSSettings {
    throwOnInvalid: true
  }
}

// version check
if (
  document
    .querySelector('meta[http-equiv=x-goalert-version]')
    ?.getAttribute('content') !== GOALERT_VERSION
) {
  warn(
    'app.js version does not match HTML version',
    'index.html=' +
      document
        .querySelector('meta[http-equiv=x-goalert-version]')
        ?.getAttribute('content'),
    'app.js=' + GOALERT_VERSION,
  )
}

const rootElement = document.getElementById('app')
const root = createRoot(rootElement as HTMLElement)

root.render(
  <StrictMode>
    <StyledEngineProvider injectFirst>
      <ThemeProvider>
        <ApolloProvider client={GraphQLClient}>
          <ReduxProvider store={store}>
            <Router base={pathPrefix}>
              <URQLProvider value={urqlClient}>
                <NewVersionCheck />
                <Suspense fallback={<Spinner />}>
                  <RequireAuth fallback={<Login />}>
                    <ConfigProvider>
                      <DestTypeProvider>
                        <App />
                      </DestTypeProvider>
                    </ConfigProvider>
                  </RequireAuth>
                </Suspense>
              </URQLProvider>
            </Router>
          </ReduxProvider>
        </ApolloProvider>
      </ThemeProvider>
    </StyledEngineProvider>
  </StrictMode>,
)
