import React from 'react'
import { PropTypes as p } from 'prop-types'
import Divider from '@material-ui/core/Divider'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core'
import { styles as globalStyles } from '../../styles/materialStyles'
import {
  Build as WizardIcon,
  Feedback as FeedbackIcon,
  Group as UsersIcon,
  Layers as EscalationPoliciesIcon,
  Notifications as AlertsIcon,
  PowerSettingsNew as LogoutIcon,
  RotateRight as RotationsIcon,
  Today as SchedulesIcon,
  VpnKey as ServicesIcon,
  Settings as AdminIcon,
} from '@material-ui/icons'

import routeConfig, { getPath } from '../routes'

import { NavLink } from 'react-router-dom'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import { CurrentUserAvatar } from '../../util/avatars'
import { authLogout } from '../../actions'
import { useDispatch } from 'react-redux'
import RequireConfig, { Config } from '../../util/RequireConfig'
import NavSubMenu from './NavSubMenu'

import logo from '../../public/goalert-alt-logo-scaled.png'
import AppLink from '../../util/AppLink'

const navIcons = {
  Alerts: AlertsIcon,
  Rotations: RotationsIcon,
  Schedules: SchedulesIcon,
  'Escalation Policies': EscalationPoliciesIcon,
  Services: ServicesIcon,
  Users: UsersIcon,
  Admin: AdminIcon,
}

const useStyles = makeStyles((theme) => ({
  ...globalStyles(theme),
  logoDiv: {
    width: '100%',
    display: 'flex',
    justifyContent: 'center',
  },
  logo: {
    padding: '0.5em',
  },
  navIcon: {
    width: '1em',
    height: '1em',
    fontSize: '24px',
  },
  list: {
    color: theme.palette.primary['500'],
    padding: 0,
  },
  listItemText: {
    color: theme.palette.primary['500'],
  },
}))

export default function SideBarDrawerList() {
  const classes = useStyles()
  const dispatch = useDispatch()
  const logout = () => dispatch(authLogout(true))

  function renderSidebarItem(IconComponent, label) {
    return (
      <ListItem button tabIndex={-1}>
        <ListItemIcon>
          <IconComponent className={classes.navIcon} />
        </ListItemIcon>
        <ListItemText
          disableTypography
          primary={
            <Typography
              variant='subtitle1'
              component='p'
              className={classes.listItemText}
            >
              {label}
            </Typography>
          }
        />
      </ListItem>
    )
  }

  function renderSidebarLink(icon, path, label, props = {}) {
    return (
      <AppLink to={path} className={classes.nav} {...props}>
        {renderSidebarItem(icon, label)}
      </AppLink>
    )
  }

  function renderSidebarNavLink(icon, path, label, key) {
    return (
      <NavLink
        key={key}
        to={path}
        className={classes.nav}
        activeClassName={classes.navSelected}
      >
        {renderSidebarItem(icon, label)}
      </NavLink>
    )
  }

  function renderAdmin() {
    const cfg = routeConfig.find((c) => c.title === 'Admin')

    return (
      <NavSubMenu
        parentIcon={navIcons[cfg.title]}
        parentTitle={cfg.title}
        path={getPath(cfg)}
        subMenuRoutes={cfg.subRoutes}
      >
        {renderSidebarItem(navIcons[cfg.title], cfg.title)}
      </NavSubMenu>
    )
  }

  function renderFeedback(url) {
    return (
      <AppLink to={url} className={classes.nav} newTab data-cy='feedback-link'>
        {renderSidebarItem(FeedbackIcon, 'Feedback')}
      </AppLink>
    )
  }

  return (
    <React.Fragment>
      <div aria-hidden className={classes.logoDiv}>
        <img
          className={classes.logo}
          height={32}
          src={logo}
          alt='GoAlert Logo'
        />
      </div>
      <Divider />
      <List role='navigation' className={classes.list} data-cy='nav-list'>
        {routeConfig
          .filter((cfg) => cfg.nav !== false)
          .map((cfg, idx) => {
            if (cfg.subRoutes) {
              return (
                <NavSubMenu
                  key={idx}
                  parentIcon={navIcons[cfg.title]}
                  parentTitle={cfg.title}
                  path={getPath(cfg)}
                  subMenuRoutes={cfg.subRoutes}
                >
                  {renderSidebarItem(navIcons[cfg.title], cfg.title)}
                </NavSubMenu>
              )
            }
            return renderSidebarNavLink(
              navIcons[cfg.title],
              getPath(cfg),
              cfg.title,
              idx,
            )
          })}
        <RequireConfig isAdmin>
          <Divider aria-hidden />
          {renderAdmin()}
        </RequireConfig>

        <Divider aria-hidden />
        {renderSidebarNavLink(WizardIcon, '/wizard', 'Wizard')}
        <Config>
          {(cfg) =>
            cfg['Feedback.Enable'] &&
            renderFeedback(
              cfg['Feedback.OverrideURL'] ||
                'https://www.surveygizmo.com/s3/4106900/GoAlert-Feedback',
            )
          }
        </Config>
        {renderSidebarLink(LogoutIcon, '/api/v2/identity/logout', 'Logout', {
          onClick: (e) => {
            e.preventDefault()
            logout()
          },
        })}
        {renderSidebarNavLink(CurrentUserAvatar, '/profile', 'Profile')}
      </List>
    </React.Fragment>
  )
}

SideBarDrawerList.propTypes = {
  onWizard: p.func.isRequired,
}
