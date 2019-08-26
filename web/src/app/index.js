/// #if HMR
import './rhl'
/// #endif

import React from 'react'
import ReactDOM from 'react-dom'
import { Provider as ReduxProvider } from 'react-redux'
import { ConnectedRouter } from 'connected-react-router'
import { ApolloProvider } from '@apollo/react-hooks'
import { MuiThemeProvider } from '@material-ui/core/styles'
import { theme } from './mui'
import { GraphQLClient } from './apollo'
import './styles'
import App from './main/NewApp'
import MuiPickersUtilsProvider from './mui-pickers'
import history from './history'
import store from './reduxStore'
import { GracefulUnmounterProvider } from './util/gracefulUnmount'
import GA from './util/GoogleAnalytics'
import { Config, ConfigProvider } from './util/RequireConfig'

const LazyGARouteTracker = React.memo(props => {
  if (!props.trackingID) {
    return null
  }

  const GAOptions = {
    titleCase: true,
    debug: false,
  }

  if (!GA.init(props.trackingID, GAOptions)) {
    return null
  }

  return <GA.RouteTracker />
})

ReactDOM.render(
  <MuiThemeProvider theme={theme}>
    <ApolloProvider client={GraphQLClient}>
      <ReduxProvider store={store}>
        <ConnectedRouter history={history}>
          <MuiPickersUtilsProvider>
            <ConfigProvider>
              <Config>
                {config => (
                  <LazyGARouteTracker
                    trackingID={config['General.GoogleAnalyticsID']}
                  />
                )}
              </Config>
              <GracefulUnmounterProvider>
                <App />
              </GracefulUnmounterProvider>
            </ConfigProvider>
          </MuiPickersUtilsProvider>
        </ConnectedRouter>
      </ReduxProvider>
    </ApolloProvider>
  </MuiThemeProvider>,
  document.getElementById('app'),
)
