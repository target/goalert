import React, { useState } from 'react'
import AppBar from '@material-ui/core/AppBar'
import Hidden from '@material-ui/core/Hidden'
import Toolbar from '@material-ui/core/Toolbar'
import ToolbarTitle from './components/ToolbarTitle'
import ToolbarAction from './components/ToolbarAction'
import ErrorBoundary from './ErrorBoundary'
import routeConfig, { renderRoutes } from './routes'
import { Switch, Route } from 'react-router-dom'
import Grid from '@material-ui/core/Grid'
import { useSelector } from 'react-redux'
import { authSelector } from '../selectors'
import { PageActionContainer, PageActionProvider } from '../util/PageActions'
import { PageNotFound as LazyPageNotFound } from '../error-pages/Errors'
import LazySideBarDrawerList from './components/SideBarDrawerList'
import LazyMobileSideBar from './MobileSideBar'
import LazyWideSideBar from './WideSideBar'
import LazyNewUserSetup from './components/NewUserSetup'
import Login from './components/Login'
import URLErrorDialog from './URLErrorDialog'
import { SkipToContentLink } from '../util/SkipToContentLink'
import { SearchContainer, SearchProvider } from '../util/AppBarSearchContainer'
import { isWidthDown, makeStyles } from '@material-ui/core'
import useWidth from '../util/useWidth'

const drawerWidth = '12em'

const useStyles = makeStyles((theme) => ({
  root: {
    flexGrow: 1,
    zIndex: 1,
    position: 'relative',
    display: 'flex',
    backgroundColor: 'lightgrey',
    height: '100%',
  },
  main: {
    width: '100%',
    overflowY: 'auto',
    marginTop: '64px',
  },
  appBar: {
    zIndex: theme.zIndex.drawer + 1,
  },
  icon: {
    marginRight: '0.25em',
    color: theme.palette.primary['500'],
  },
  toolbar: theme.mixins.toolbar,
  containerClass: {
    padding: '1em',
    [theme.breakpoints.up('md')]: { width: '75%' },
    [theme.breakpoints.down('sm')]: { width: '100%' },
  },
}))

export default function App() {
  const classes = useStyles()
  const [showMobile, setShowMobile] = useState(false)
  const width = useWidth()
  const fullScreen = isWidthDown('md', width)
  const authValid = useSelector(authSelector)

  if (!authValid) {
    return <Login />
  }
  const marginLeft = fullScreen ? 0 : drawerWidth

  let cyFormat = 'wide'
  if (fullScreen) cyFormat = 'mobile'
  return (
    <div className={classes.root}>
      <PageActionProvider>
        <SearchProvider>
          <AppBar
            position='fixed'
            className={classes.appBar}
            data-cy='app-bar'
            data-cy-format={cyFormat}
          >
            <SkipToContentLink />
            <Toolbar className={classes.toolbar}>
              <ToolbarAction
                handleShowMobileSidebar={() => setShowMobile(true)}
              />
              <ToolbarTitle />

              <PageActionContainer />
              <SearchContainer />
            </Toolbar>
          </AppBar>

          <Hidden smDown>
            <LazyWideSideBar>
              <div className={classes.toolbar} />
              <LazySideBarDrawerList onWizard={() => setShowMobile(true)} />
            </LazyWideSideBar>
          </Hidden>
          <Hidden mdUp>
            <LazyMobileSideBar
              show={showMobile}
              onChange={(val) => setShowMobile(val)}
            >
              <LazySideBarDrawerList onWizard={() => setShowMobile(true)} />
            </LazyMobileSideBar>
          </Hidden>

          <URLErrorDialog />

          <main id='content' className={classes.main} style={{ marginLeft }}>
            <ErrorBoundary>
              <LazyNewUserSetup />
              <Grid container justify='center'>
                <Grid className={classes.containerClass} item>
                  <Switch>
                    {renderRoutes(routeConfig)}
                    <Route component={() => <LazyPageNotFound />} />
                  </Switch>
                </Grid>
              </Grid>
            </ErrorBoundary>
          </main>
        </SearchProvider>
      </PageActionProvider>
    </div>
  )
}
