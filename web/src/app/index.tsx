import React, { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { Provider as ReduxProvider } from 'react-redux'
import { ApolloProvider } from '@apollo/client'
import { StyledEngineProvider } from '@mui/material/styles'

import { GOALERT_VERSION as version, pathPrefix } from './env'
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
    ?.getAttribute('content') !== version
) {
  warn(
    'app.js version does not match HTML version',
    'index.html=' +
      document
        .querySelector('meta[http-equiv=x-goalert-version]')
        ?.getAttribute('content'),
    'app.js=' + version,
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
                <RequireAuth fallback={<Login />}>
                  <ConfigProvider>
                    <App />
                  </ConfigProvider>
                </RequireAuth>
              </URQLProvider>
            </Router>
          </ReduxProvider>
        </ApolloProvider>
      </ThemeProvider>
    </StyledEngineProvider>
  </StrictMode>,
)
