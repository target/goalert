import React from 'react'
import Hidden from '@mui/material/Hidden'
import { Theme } from '@mui/material/styles'
import IconButton from '@mui/material/IconButton'
import { Menu as MenuIcon, ChevronLeft } from '@mui/icons-material'
import { useIsWidthDown } from '../../util/useWidth'
import { Route, Switch, useLocation } from 'wouter'
import { routes } from '../AppRoutes'

interface ToolbarActionProps {
  showMobileSidebar: boolean
  openMobileSidebar: (arg: boolean) => void
}

function removeLastPartOfPath(path: string): string {
  const parts = path.split('/')
  parts.pop()
  return parts.join('/')
}

function ToolbarAction(props: ToolbarActionProps): React.ReactNode {
  const fullScreen = useIsWidthDown('md')

  const [, navigate] = useLocation()

  function renderToolbarAction(): React.ReactNode {
    const route = removeLastPartOfPath(window.location.pathname)

    // only show back button on mobile
    if (!fullScreen) return <React.Fragment />

    return (
      <IconButton
        aria-label='Back a Page'
        data-cy='nav-back-icon'
        onClick={() => navigate(route)}
        size='large'
        sx={(theme: Theme) => ({
          color: theme.palette.mode === 'light' ? 'white' : undefined,
        })}
      >
        <ChevronLeft />
      </IconButton>
    )
  }

  function renderToolbarMenu(): React.ReactNode {
    return (
      <Hidden mdUp>
        <IconButton
          aria-label='Open Navigation Menu'
          aria-expanded={props.showMobileSidebar}
          data-cy='nav-menu-icon'
          onClick={() => props.openMobileSidebar(true)}
          size='large'
          sx={(theme: Theme) => ({
            color: theme.palette.mode === 'light' ? 'inherit' : undefined,
          })}
        >
          <MenuIcon />
        </IconButton>
      </Hidden>
    )
  }

  const getRoute = (route: string, idx: number): React.ReactNode => (
    <Route key={idx} path={route}>
      {renderToolbarAction()}
    </Route>
  )

  return (
    <Switch>
      {Object.keys(routes)
        .filter((path) => path.split('/').length > 3)
        .map(getRoute)}
      <Route path='/:type'>{renderToolbarMenu()}</Route>
      <Route path='/:type/:id'>{renderToolbarMenu()}</Route>
    </Switch>
  )
}

export default ToolbarAction
