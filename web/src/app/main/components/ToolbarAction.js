import React from 'react'
import Hidden from '@mui/material/Hidden'
import IconButton from '@mui/material/IconButton'
import { Menu as MenuIcon, ChevronLeft } from '@mui/icons-material'
import { useIsWidthDown } from '../../util/useWidth'
import { PropTypes as p } from 'prop-types'
import { Route, Switch, useLocation } from 'wouter'
import { routes } from '../AppRoutes'

function removeLastPartOfPath(path) {
  const parts = path.split('/')
  parts.pop()
  return parts.join('/')
}

function ToolbarAction(props) {
  const fullScreen = useIsWidthDown('md')

  const [, navigate] = useLocation()

  function renderToolbarAction() {
    const route = removeLastPartOfPath(window.location.pathname)

    // only show back button on mobile
    if (!fullScreen) return null

    return (
      <IconButton
        aria-label='Back a Page'
        data-cy='nav-back-icon'
        onClick={() => navigate(route)}
        size='large'
        sx={(theme) => ({
          color: theme.palette.mode === 'light' ? 'white' : undefined,
        })}
      >
        <ChevronLeft />
      </IconButton>
    )
  }

  function renderToolbarMenu() {
    return (
      <Hidden mdUp>
        <IconButton
          aria-label='Open Navigation Menu'
          aria-expanded={props.showMobileSidebar}
          data-cy='nav-menu-icon'
          onClick={() => props.openMobileSidebar(true)}
          size='large'
          sx={(theme) => ({
            color: theme.palette.mode === 'light' ? 'inherit' : undefined,
          })}
        >
          <MenuIcon />
        </IconButton>
      </Hidden>
    )
  }

  const getRoute = (route) => (
    <Route path={route} children={renderToolbarAction()} />
  )

  return (
    <Switch>
      {Object.keys(routes)
        .filter((path) => path.split('/').length > 3)
        .map(getRoute)}
      <Route path='/:type' children={renderToolbarMenu()} />
      <Route path='/:type/:id' children={renderToolbarMenu()} />
    </Switch>
  )
}

ToolbarAction.propTypes = {
  showMobileSidebar: p.bool,
  openMobileSidebar: p.func,
}

export default ToolbarAction
