import React from 'react'
import Divider from '@mui/material/Divider'
import List from '@mui/material/List'
import Typography from '@mui/material/Typography'
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
import NavBarLink, { NavBarSubLink } from './NavBarLink'
import { ExpFlag } from '../util/useExpFlag'

import logoSrcSet1 from '../public/logos/lightmode/lightmode_logo@1x.webp'
import logoSrcSet2 from '../public/logos/lightmode/lightmode_logo@1.5x.webp'
import logoSrcSet3 from '../public/logos/lightmode/lightmode_logo@2x.webp'
import logoImgSrc from '../public/logos/lightmode/lightmode_logo@2x.png'

import darkModeLogoSrcSet1 from '../public/logos/darkmode/darkmode_logo@1x.webp'
import darkModeLogoSrcSet2 from '../public/logos/darkmode/darkmode_logo@1.5x.webp'
import darkModeLogoSrcSet3 from '../public/logos/darkmode/darkmode_logo@2x.webp'
import darkModeLogoImgSrc from '../public/logos/darkmode/darkmode_logo@2x.png'

const useStyles = makeStyles((theme: Theme) => ({
  ...globalStyles(theme),
  logoDiv: {
    ...theme.mixins.toolbar,
    paddingLeft: 8,
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

  const logo =
    theme.palette.mode === 'dark' ? (
      <picture>
        <source
          srcSet={`
          ${darkModeLogoSrcSet1},
          ${darkModeLogoSrcSet2} 1.5x,
          ${darkModeLogoSrcSet3} 2x
        `}
          type='image/webp'
        />
        <img src={darkModeLogoImgSrc} height={61} alt='GoAlert' />
      </picture>
    ) : (
      <picture>
        <source
          srcSet={`
          ${logoSrcSet1},
          ${logoSrcSet2} 1.5x,
          ${logoSrcSet3} 2x
        `}
          type='image/webp'
        />
        <img src={logoImgSrc} height={61} alt='GoAlert' />
      </picture>
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
        <a href='/' aria-hidden className={classes.logoDiv}>
          {logo}
        </a>
        <Typography variant='h5' sx={{ pl: 1 }}>
          <b>GoAlert</b>
        </Typography>
      </div>
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
