import React from 'react'
import { PropTypes as p } from 'prop-types'
import Divider from '@mui/material/Divider'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import Typography from '@mui/material/Typography'
import makeStyles from '@mui/styles/makeStyles'
import { styles as globalStyles } from '../../styles/materialStyles'
import {
  Group as UsersIcon,
  Layers as EscalationPoliciesIcon,
  Notifications as AlertsIcon,
  RotateRight as RotationsIcon,
  Today as SchedulesIcon,
  VpnKey as ServicesIcon,
  Build as AdminIcon,
} from '@mui/icons-material'
import { WizardHat as WizardIcon } from 'mdi-material-ui'
import { NavLink } from 'react-router-dom'
import ListItemIcon from '@mui/material/ListItemIcon'
import { useTheme } from '@mui/material/styles'
import routeConfig, { getPath } from '../routes'
import RequireConfig from '../../util/RequireConfig'
import NavSubMenu from './NavSubMenu'
import logo from '../../public/logos/black/goalert-alt-logo.png'
import darkModeLogo from '../../public/logos/white/goalert-alt-logo-white.png'

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
    ...theme.mixins.toolbar,
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
  },
  navIcon: {
    width: '1em',
    height: '1em',
    fontSize: '24px',
  },
  list: {
    padding: 0,
  },
}))

export default function SideBarDrawerList(props) {
  const { closeMobileSidebar } = props
  const classes = useStyles()
  const theme = useTheme()

  function renderSidebarItem(IconComponent, label) {
    return (
      <ListItem button tabIndex={-1}>
        <ListItemIcon>
          <IconComponent className={classes.navIcon} />
        </ListItemIcon>
        <ListItemText
          disableTypography
          primary={
            <Typography variant='subtitle1' component='p'>
              {label}
            </Typography>
          }
        />
      </ListItem>
    )
  }

  function renderSidebarNavLink(icon, path, label, key) {
    return (
      <NavLink
        key={key}
        to={path}
        className={({ isActive }) =>
          isActive ? classes.navSelected : classes.nav
        }
        onClick={closeMobileSidebar}
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
        path={getPath(cfg).replace('/*', '')}
        subMenuRoutes={cfg.subRoutes}
        closeMobileSidebar={closeMobileSidebar}
      >
        {renderSidebarItem(navIcons[cfg.title], cfg.title)}
      </NavSubMenu>
    )
  }

  return (
    <React.Fragment>
      <div aria-hidden className={classes.logoDiv}>
        <img
          height={38}
          src={theme.palette.mode === 'dark' ? darkModeLogo : logo}
          alt='GoAlert Logo'
        />
      </div>
      <Divider />
      <nav>
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
                    path={getPath(cfg).replace('/*', '')}
                    subMenuRoutes={cfg.subRoutes}
                  >
                    {renderSidebarItem(navIcons[cfg.title], cfg.title)}
                  </NavSubMenu>
                )
              }
              return renderSidebarNavLink(
                navIcons[cfg.title],
                getPath(cfg).replace('/*', ''),
                cfg.title,
                idx,
              )
            })}
          <RequireConfig isAdmin>{renderAdmin()}</RequireConfig>

          {renderSidebarNavLink(WizardIcon, '/wizard', 'Wizard')}
        </List>
      </nav>
    </React.Fragment>
  )
}

SideBarDrawerList.propTypes = {
  closeMobileSidebar: p.func.isRequired,
}
