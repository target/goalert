import React, { Component } from 'react'
import { Switch, Route, withRouter } from 'react-router-dom'
import Hidden from '@material-ui/core/Hidden'
import IconButton from '@material-ui/core/IconButton'
import { Menu as MenuIcon, ChevronLeft } from '@material-ui/icons'
import withWidth, { isWidthUp } from '@material-ui/core/withWidth'

@withWidth()
@withRouter
export default class ToolbarAction extends Component {
  removeLastPartOfPath = path => {
    const parts = path.split('/')
    parts.pop()
    return parts.join('/')
  }

  renderToolbarAction = () => {
    const route = this.removeLastPartOfPath(window.location.pathname)

    // only show back button on mobile
    if (isWidthUp('md', this.props.width)) return null

    return (
      <IconButton
        aria-label='Back a Page'
        color='inherit'
        data-cy='nav-back-icon'
        onClick={() => this.props.history.replace(route)}
      >
        <ChevronLeft />
      </IconButton>
    )
  }

  render() {
    const getRoute = route => (
      <Route path={route} render={() => this.renderToolbarAction()} />
    )

    return (
      <Switch>
        {getRoute('/schedules/:scheduleID/assignments')}
        {getRoute('/schedules/:scheduleID/escalation-policies')}
        {getRoute('/schedules/:scheduleID/overrides')}
        {getRoute('/schedules/:scheduleID/shifts')}
        {getRoute('/escalation-policies/:escalationPolicyID/services')}
        {getRoute('/services/:serviceID/alerts')}
        {getRoute('/services/:serviceID/integration-keys')}
        {getRoute('/services/:serviceID/heartbeat-monitors')}
        {getRoute('/services/:serviceID/labels')}
        {getRoute('/users/:userID/on-call-assignments')}
        {getRoute('/users/:userID/schedule-calendar-subscriptions')}
        {getRoute('/profile/on-call-assignments')}
        {getRoute('/profile/schedule-calendar-subscriptions')}
        <Route
          render={() => (
            <Hidden mdUp>
              <IconButton
                aria-label='Navigation Menu'
                color='inherit'
                data-cy='nav-menu-icon'
                onClick={() => this.props.handleShowMobileSidebar(true)}
              >
                <MenuIcon />
              </IconButton>
            </Hidden>
          )}
        />
      </Switch>
    )
  }
}
