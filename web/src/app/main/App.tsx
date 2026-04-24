import React, { Suspense, useEffect, useLayoutEffect, useState } from 'react'
import AppBar from '@mui/material/AppBar'
import Box from '@mui/material/Box'
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
import { useIsWidthDown } from '../util/useWidth'
import { isIOS } from '../util/browsers'
import UserSettingsPopover from './components/UserSettingsPopover'
import { Theme } from '@mui/material/styles'
import type { SxProps } from '@mui/material'
import AppRoutes from './AppRoutes'
import { useURLKey } from '../actions'
import NavBar from './NavBar'
import AuthLink from './components/AuthLink'
import { useExpFlag } from '../util/useExpFlag'
import { NotificationProvider } from './SnackbarNotification'
import ReactGA from 'react-ga4'
import { useConfigValue } from '../util/RequireConfig'
import Spinner from '../loading/components/Spinner'

const classes = {
  root: (theme: Theme) => ({
    flexGrow: 1,
    zIndex: 1,
    position: 'relative' as const,
    display: 'flex',
    backgroundColor: theme.palette.background.default,
    height: '100%',
  }),
  main: {
    width: '100%',
    overflowY: 'auto',
    marginTop: '64px',
  },
  mainContainer: { position: 'relative', height: '100%' },
  appBar: (theme: Theme) => ({
    [theme.breakpoints.up('md')]: { width: `calc(100% - ${drawerWidth})` },
    zIndex: theme.zIndex.drawer + 1,
  }),
  containerClass: (theme: Theme) => ({
    padding: '1em',
    [theme.breakpoints.up('md')]: { width: '75%' },
    [theme.breakpoints.down('md')]: { width: '100%' },
  }),
} satisfies Record<string, SxProps<Theme>>

export default function App(): React.JSX.Element {
  const [analyticsID] = useConfigValue('General.GoogleAnalyticsID') as [string]

  useEffect(() => {
    if (analyticsID) ReactGA.initialize(analyticsID)
  }, [analyticsID])

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
    <Box sx={classes.root} id='app-root'>
      <PageActionProvider>
        <SearchProvider>
          <NotificationProvider>
            <AppBar
              position='fixed'
              sx={classes.appBar}
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

            <Box sx={{ display: { xs: 'none', md: 'block' } }}>
              <LazyWideSideBar>
                <NavBar />
              </LazyWideSideBar>
            </Box>
            <Box sx={{ display: { xs: 'block', md: 'none' } }}>
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
            </Box>

            <Box
              component='main'
              id='content'
              sx={classes.main}
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
                    sx={classes.mainContainer}
                  >
                    <Grid sx={classes.containerClass} >
                      <AppRoutes />
                    </Grid>
                  </Grid>
                </Suspense>
              </ErrorBoundary>
            </Box>
          </NotificationProvider>
        </SearchProvider>
      </PageActionProvider>
    </Box>
  )
}
