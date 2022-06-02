import { GOALERT_VERSION, pathPrefix } from './env'

import React, { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { Provider as ReduxProvider } from 'react-redux'
import { ApolloProvider } from '@apollo/client'
import { StyledEngineProvider } from '@mui/material/styles'
import { ThemeProvider } from './theme/themeConfig'
import { GraphQLClient } from './apollo'
import './styles'
import App from './main/App'
import MuiPickersUtilsProvider from './mui-pickers'
import store from './reduxStore'
import { ConfigProvider } from './util/RequireConfig'
import { warn } from './util/debug'
import NewVersionCheck from './NewVersionCheck'
import { Provider as URQLProvider } from 'urql'
import { client as urqlClient } from './urql'
import { Router } from 'wouter'

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
              <MuiPickersUtilsProvider>
                <URQLProvider value={urqlClient}>
                  <ConfigProvider>
                    <NewVersionCheck />
                    <App />
                  </ConfigProvider>
                </URQLProvider>
              </MuiPickersUtilsProvider>
            </Router>
          </ReduxProvider>
        </ApolloProvider>
      </ThemeProvider>
    </StyledEngineProvider>
  </StrictMode>,
)
