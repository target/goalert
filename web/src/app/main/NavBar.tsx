import React from 'react'
import Divider from '@mui/material/Divider'
import List from '@mui/material/List'
import makeStyles from '@mui/styles/makeStyles'
import { styles as globalStyles } from '../styles/materialStyles'
import {
  Group,
  Layers,
  Notifications,
  RotateRight,
  Today,
  VpnKey,
  Build,
  DeveloperBoard,
} from '@mui/icons-material'
import { WizardHat as WizardIcon } from 'mdi-material-ui'
import { Theme, useTheme } from '@mui/material/styles'
import RequireConfig from '../util/RequireConfig'
import logo from '../public/logos/black/goalert-alt-logo.png'
import darkModeLogo from '../public/logos/white/goalert-alt-logo-white.png'
import NavBarLink, { NavBarSubLink } from './NavBarLink'
import { ExpFlag } from '../util/useExpFlag'

const useStyles = makeStyles((theme: Theme) => ({
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

export default function NavBar(): JSX.Element {
  const classes = useStyles()
  const theme = useTheme()

  let localDev = null
  if (process.env.NODE_ENV !== 'production') {
    localDev = <NavBarLink to='/dev' title='Dev' icon={<DeveloperBoard />} />
  }

  return (
    <React.Fragment>
      <a href='/' aria-hidden className={classes.logoDiv}>
        <img
          height={38}
          src={theme.palette.mode === 'dark' ? darkModeLogo : logo}
          alt='GoAlert Logo'
        />
      </a>
      <Divider />
      <nav>
        <List role='navigation' className={classes.list} data-cy='nav-list'>
          <NavBarLink to='/alerts' title='Alerts' icon={<Notifications />} />
          <NavBarLink
            to='/rotations'
            title='Rotations'
            icon={<RotateRight />}
          />
          <NavBarLink to='/schedules' title='Schedules' icon={<Today />} />
          <NavBarLink
            to='/escalation-policies'
            title='Escalation Policies'
            icon={<Layers />}
          />
          <NavBarLink to='/services' title='Services' icon={<VpnKey />} />
          <NavBarLink to='/users' title='Users' icon={<Group />} />

          <RequireConfig isAdmin>
            <NavBarLink to='/admin' title='Admin' icon={<Build />}>
              <NavBarSubLink to='/admin/config' title='Config' />
              <NavBarSubLink to='/admin/limits' title='System Limits' />
              <NavBarSubLink to='/admin/toolbox' title='Toolbox' />
              <NavBarSubLink to='/admin/message-logs' title='Message Logs' />
              <NavBarSubLink to='/admin/alert-counts' title='Alert Counts' />
              <NavBarSubLink
                to='/admin/service-metrics'
                title='Service Metrics'
              />
              <NavBarSubLink to='/admin/switchover' title='Switchover' />
              <ExpFlag flag='gql-api-keys'>
                <NavBarSubLink to='/admin/api-keys' title='API Keys' />
              </ExpFlag>
            </NavBarLink>
          </RequireConfig>

          <NavBarLink to='/wizard' title='Wizard' icon={<WizardIcon />} />

          {localDev}
        </List>
      </nav>
    </React.Fragment>
  )
}
