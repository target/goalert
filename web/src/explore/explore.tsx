import React, { useState } from 'react'
import { createRoot } from 'react-dom/client'
import { GraphiQL } from 'graphiql'
import { Provider as ReduxProvider } from 'react-redux'
import { StyledEngineProvider } from '@mui/material/styles'
import { ThemeProvider } from '../app/theme/themeConfig'
import makeStyles from '@mui/styles/makeStyles'
import Login from '../app/main/components/Login'
import store from '../app/reduxStore'

import 'graphiql/graphiql.css'
const useStyles = makeStyles((theme) => ({
  root: {
    flexGrow: 1,
    zIndex: 1,
    position: 'relative',
    display: 'flex',
    backgroundColor: theme.palette.background.default,
    height: '100%',
  },
}))

const App = (): React.ReactNode => {
  const path = location.host + location.pathname.replace(/\/explore.*$/, '')
  const [needLogin, setNeedLogin] = useState(false)
  const classes = useStyles()

  if (needLogin) {
    return (
      <div className={classes.root}>
        <Login />
      </div>
    )
  }

  return (
    <GraphiQL
      headerEditorEnabled
      fetcher={async (graphQLParams) => {
        const resp = await fetch(location.protocol + '//' + path, {
          method: 'POST',
          headers: {
            Accept: 'application/json',
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(graphQLParams),
          credentials: 'same-origin',
        })
        if (resp.status === 401) {
          setNeedLogin(true)
        }
        return resp.json().catch(() => resp.text())
      }}
    />
  )
}

const container = createRoot(document.getElementById('root') as HTMLElement)

container.render(
  <div style={{ height: '100vh' }}>
    <StyledEngineProvider injectFirst>
      <ThemeProvider>
        <ReduxProvider store={store}>
          <App />
        </ReduxProvider>
      </ThemeProvider>
    </StyledEngineProvider>
  </div>,
)
