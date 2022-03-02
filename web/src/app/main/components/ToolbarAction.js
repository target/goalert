import React from 'react'
import { Routes, Route, useNavigate } from 'react-router-dom'
import Hidden from '@mui/material/Hidden'
import IconButton from '@mui/material/IconButton'
import { Menu as MenuIcon, ChevronLeft } from '@mui/icons-material'
import { useIsWidthDown } from '../../util/useWidth'
import { PropTypes as p } from 'prop-types'

function removeLastPartOfPath(path) {
  const parts = path.split('/')
  parts.pop()
  return parts.join('/')
}

function ToolbarAction(props) {
  const fullScreen = useIsWidthDown('md')

  const navigate = useNavigate()

  function renderToolbarAction() {
    const route = removeLastPartOfPath(window.location.pathname)

    // only show back button on mobile
    if (!fullScreen) return null

    return (
      <IconButton
        aria-label='Back a Page'
        color='inherit'
        data-cy='nav-back-icon'
        onClick={() => navigate(route, { replace: true })}
        size='large'
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
          color='inherit'
          data-cy='nav-menu-icon'
          onClick={() => props.openMobileSidebar(true)}
          size='large'
        >
          <MenuIcon />
        </IconButton>
      </Hidden>
    )
  }

  const getRoute = (route) => (
    <Route path={route} element={renderToolbarAction()} />
  )

  return (
    <Routes>
      {getRoute('/schedules/:scheduleID/assignments')}
      {getRoute('/schedules/:scheduleID/escalation-policies')}
      {getRoute('/schedules/:scheduleID/overrides')}
      {getRoute('/schedules/:scheduleID/shifts')}
      {getRoute('/schedules/:scheduleID/on-call-notifications')}
      {getRoute('/escalation-policies/:escalationPolicyID/services')}
      {getRoute('/services/:serviceID/alerts')}
      {getRoute('/services/:serviceID/integration-keys')}
      {getRoute('/services/:serviceID/heartbeat-monitors')}
      {getRoute('/services/:serviceID/labels')}
      {getRoute('/users/:userID/on-call-assignments')}
      {getRoute('/users/:userID/sessions')}
      {getRoute('/users/:userID/schedule-calendar-subscriptions')}
      {getRoute('/profile/on-call-assignments')}
      {getRoute('/profile/schedule-calendar-subscriptions')}
      <Route path='/:type' element={renderToolbarMenu()} />
      <Route path='/:type/:id' element={renderToolbarMenu()} />
    </Routes>
  )
}

ToolbarAction.propTypes = {
  showMobileSidebar: p.bool,
  openMobileSidebar: p.func,
}

export default ToolbarAction
