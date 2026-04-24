import React from 'react'
import Divider from '@mui/material/Divider'
import List from '@mui/material/List'
import Typography from '@mui/material/Typography'
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
import { useTheme } from '@mui/material/styles'
import RequireConfig from '../util/RequireConfig'

import NavBarLink, { NavBarSubLink } from './NavBarLink'
import logoImgSrc from '../public/logos/lightmode_logo.svg'
import darkModeLogoImgSrc from '../public/logos/darkmode_logo.svg'

export default function NavBar(): React.JSX.Element {
  const theme = useTheme()

  let localDevFooter = null
  let localDevRiver = null
  if (process.env.NODE_ENV !== 'production') {
    localDevFooter = (
      <NavBarLink to='/dev' title='Dev' icon={<DeveloperBoard />} />
    )
    localDevRiver = (
      <NavBarSubLink newTab to='/admin/riverui' title='Job Queues' />
    )
  }

  const logo =
    theme.palette.mode === 'dark' ? (
      <img src={darkModeLogoImgSrc} height={61} alt='GoAlert' />
    ) : (
      <img src={logoImgSrc} height={61} alt='GoAlert' />
    )

  return (
    <React.Fragment>
      <div
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'left',
        }}
      >
        <a href='/' aria-hidden style={{ ...(theme.mixins.toolbar as React.CSSProperties), paddingLeft: 8 }}>
          {logo}
        </a>
        <Typography variant='h5' sx={{ pl: 1 }}>
          <b>GoAlert</b>
        </Typography>
      </div>
      <Divider />
      <nav>
        <List role='navigation' sx={{ p: 0 }} data-cy='nav-list'>
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
              <NavBarSubLink to='/admin/maintenance' title='Maintenance' />
              <NavBarSubLink to='/admin/limits' title='System Limits' />
              <NavBarSubLink to='/admin/toolbox' title='Toolbox' />
              <NavBarSubLink to='/admin/message-logs' title='Message Logs' />
              <NavBarSubLink to='/admin/alert-counts' title='Alert Counts' />
              <NavBarSubLink
                to='/admin/service-metrics'
                title='Service Metrics'
              />
              <NavBarSubLink to='/admin/switchover' title='Switchover' />
              <NavBarSubLink to='/admin/api-keys' title='API Keys' />
              {localDevRiver}
            </NavBarLink>
          </RequireConfig>

          <NavBarLink to='/wizard' title='Wizard' icon={<WizardIcon />} />

          {localDevFooter}
        </List>
      </nav>
    </React.Fragment>
  )
}
