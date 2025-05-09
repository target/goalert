import React, { Suspense, useEffect, useLayoutEffect, useState } from 'react'
import AppBar from '@mui/material/AppBar'
import Hidden from '@mui/material/Hidden'
import Toolbar from '@mui/material/Toolbar'
import ToolbarPageTitle from './components/ToolbarPageTitle'
import ToolbarAction from './components/ToolbarAction'
import ErrorBoundary from './ErrorBoundary'
import Grid from '@mui/material/Grid'
import { PageActionContainer, PageActionProvider } from '../util/PageActions'
import SwipeableDrawer from '@mui/material/SwipeableDrawer'
import LazyWideSideBar, { drawerWidth } from './WideSideBar'
import LazyNewUserSetup from './components/NewUserSetup'
import { SkipToContentLink } from '../util/SkipToContentLink'
import { SearchContainer, SearchProvider } from '../util/AppBarSearchContainer'
import makeStyles from '@mui/styles/makeStyles'
import { useIsWidthDown } from '../util/useWidth'
import { isIOS } from '../util/browsers'
import UserSettingsPopover from './components/UserSettingsPopover'
import { Theme } from '@mui/material/styles'
import AppRoutes from './AppRoutes'
import { useURLKey } from '../actions'
import NavBar from './NavBar'
import AuthLink from './components/AuthLink'
import { useExpFlag } from '../util/useExpFlag'
import { NotificationProvider } from './SnackbarNotification'
import ReactGA from 'react-ga4'
import { useConfigValue } from '../util/RequireConfig'
import Spinner from '../loading/components/Spinner'

const useStyles = makeStyles((theme: Theme) => ({
  root: {
    flexGrow: 1,
    zIndex: 1,
    position: 'relative',
    display: 'flex',
    backgroundColor: theme.palette.background.default,
    height: '100%',
  },
  main: {
    width: '100%',
    overflowY: 'auto',
    marginTop: '64px',
  },
  mainContainer: { position: 'relative', height: '100%' },
  appBar: {
    [theme.breakpoints.up('md')]: { width: `calc(100% - ${drawerWidth})` },
    zIndex: theme.zIndex.drawer + 1,
  },
  containerClass: {
    padding: '1em',
    [theme.breakpoints.up('md')]: { width: '75%' },
    [theme.breakpoints.down('md')]: { width: '100%' },
  },
}))

export default function App(): React.JSX.Element {
  const [analyticsID] = useConfigValue('General.GoogleAnalyticsID') as [string]

  useEffect(() => {
    if (analyticsID) ReactGA.initialize(analyticsID)
  }, [analyticsID])

  const classes = useStyles()
  const [showMobile, setShowMobile] = useState(false)
  const fullScreen = useIsWidthDown('md')
  const marginLeft = fullScreen ? 0 : drawerWidth
  const urlKey = useURLKey()
  const hasExampleFlag = useExpFlag('example')

  useLayoutEffect(() => {
    setShowMobile(false)
  }, [urlKey])

  let cyFormat = 'wide'
  if (fullScreen) cyFormat = 'mobile'
  return (
    <div className={classes.root} id='app-root'>
      <PageActionProvider>
        <SearchProvider>
          <NotificationProvider>
            <AppBar
              position='fixed'
              className={classes.appBar}
              data-cy='app-bar'
              data-cy-format={cyFormat}
            >
              <SkipToContentLink />
              <Toolbar>
                <ToolbarAction
                  showMobileSidebar={showMobile}
                  openMobileSidebar={() => setShowMobile(true)}
                />
                <Suspense>
                  <ToolbarPageTitle />
                </Suspense>
                <div style={{ flex: 1 }} />
                <PageActionContainer />
                <SearchContainer />
                <UserSettingsPopover />
              </Toolbar>
            </AppBar>

            <Hidden mdDown>
              <LazyWideSideBar>
                <NavBar />
              </LazyWideSideBar>
            </Hidden>
            <Hidden mdUp>
              <SwipeableDrawer
                disableDiscovery={isIOS}
                open={showMobile}
                onOpen={() => setShowMobile(true)}
                onClose={() => setShowMobile(false)}
                SlideProps={{
                  unmountOnExit: true,
                }}
              >
                <NavBar />
              </SwipeableDrawer>
            </Hidden>

            <main
              id='content'
              className={classes.main}
              style={{ marginLeft }}
              data-exp-flag-example={String(hasExampleFlag)}
            >
              <ErrorBoundary>
                <Suspense fallback={<Spinner />}>
                  <LazyNewUserSetup />
                  <AuthLink />
                  <Grid
                    container
                    justifyContent='center'
                    className={classes.mainContainer}
                  >
                    <Grid className={classes.containerClass} item>
                      <AppRoutes />
                    </Grid>
                  </Grid>
                </Suspense>
              </ErrorBoundary>
            </main>
          </NotificationProvider>
        </SearchProvider>
      </PageActionProvider>
    </div>
  )
}
