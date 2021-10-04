// set webpack public path for loading additional assets
import { GOALERT_VERSION } from './env'

import React from 'react'
import ReactDOM from 'react-dom'
import { Provider as ReduxProvider } from 'react-redux'
import { ConnectedRouter } from 'connected-react-router'
import { ApolloProvider } from '@apollo/client'
import { ThemeProvider, Theme, StyledEngineProvider } from '@mui/material/styles';
import { theme } from './mui'
import { GraphQLClient } from './apollo'
import './styles'
import App from './main/App'
import MuiPickersUtilsProvider from './mui-pickers'
import history from './history'
import store from './reduxStore'
import GoogleAnalytics from './util/GoogleAnalytics'
import { Config, ConfigProvider, ConfigData } from './util/RequireConfig'
import { warn } from './util/debug'
import NewVersionCheck from './NewVersionCheck'

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

const LazyGARouteTracker = React.memo((props: { trackingID?: string }) => {
  if (!props.trackingID) {
    return null
  }

  const GAOptions = {
    titleCase: true,
    debug: false,
  }

  if (!GoogleAnalytics.init(props.trackingID, GAOptions)) {
    return null
  }

  return <GoogleAnalytics.RouteTracker />
})
LazyGARouteTracker.displayName = 'LazyGARouteTracker'

ReactDOM.render(
  <StyledEngineProvider injectFirst>
    <ThemeProvider theme={theme}>
      <ApolloProvider client={GraphQLClient}>
        <ReduxProvider store={store}>
          <ConnectedRouter history={history}>
            <MuiPickersUtilsProvider>
              <ConfigProvider>
                <NewVersionCheck />
                <Config>
                  {(config: ConfigData) => (
                    <LazyGARouteTracker
                      trackingID={config['General.GoogleAnalyticsID'] as string}
                    />
                  )}
                </Config>
                <App />
              </ConfigProvider>
            </MuiPickersUtilsProvider>
          </ConnectedRouter>
        </ReduxProvider>
      </ApolloProvider>
    </ThemeProvider>
  </StyledEngineProvider>,
  document.getElementById('app'),
)
