import React from 'react'
import { useLocation } from 'wouter'
import List from '@mui/material/List'
import ListItemButton from '@mui/material/ListItemButton'
import ListItemText from '@mui/material/ListItemText'
import ListItemIcon from '@mui/material/ListItemIcon'
import Typography from '@mui/material/Typography'
import { styles } from '../styles/materialStyles'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import { Collapse, Theme } from '@mui/material'
import type { SxProps } from '@mui/material'
import AppLink from '../util/AppLink'
import { OpenInNew } from '@mui/icons-material'
import { useTheme } from '@mui/material/styles'

function useNavClasses(theme: Theme): {
  nav: SxProps<Theme>
  navSelected: SxProps<Theme>
  subMenuLinkText: SxProps<Theme>
  dropdown: (open: boolean) => SxProps<Theme>
} {
  const { nav, navSelected } = styles(theme)
  return {
    nav: nav as SxProps<Theme>,
    navSelected: navSelected as SxProps<Theme>,
    subMenuLinkText: {
      paddingLeft: '3.5rem',
      '& span': {
        fontSize: '.9rem',
      },
    } as SxProps<Theme>,
    dropdown: (open: boolean): SxProps<Theme> => ({
      transition: theme.transitions.create(['transform'], {
        duration: theme.transitions.duration.short,
      }),
      transform: open ? 'rotate(0)' : 'rotate(-90deg)',
    }),
  }
}

export type NavBarSubLinkProps = {
  to: string
  title: string
  newTab?: boolean
}
export function NavBarSubLink({
  to,
  title,
  newTab,
}: NavBarSubLinkProps): React.JSX.Element {
  const theme = useTheme()
  const { navSelected, nav, subMenuLinkText } = useNavClasses(theme)
  const [path] = useLocation()
  return (
    <AppLink
      sx={path.startsWith(to) ? navSelected : nav}
      to={to}
      newTab={newTab}
    >
      <ListItemButton tabIndex={-1}>
        <ListItemText sx={subMenuLinkText}>
          {title}
          {newTab && (
            <OpenInNew fontSize='small' style={{ paddingLeft: '1em' }} />
          )}
        </ListItemText>
      </ListItemButton>
    </AppLink>
  )
}

export type NavBarLinkProps = {
  icon: React.JSX.Element
  title: string
  to: string
  children?: React.ReactNode[] | React.ReactNode
}

export default function NavBarLink({
  icon,
  title,
  to,
  children,
}: NavBarLinkProps): React.JSX.Element {
  const theme = useTheme()
  const classes = useNavClasses(theme)
  const [path] = useLocation()
  const isRoute = path.startsWith(to)

  return (
    <React.Fragment>
      <AppLink
        to={to}
        sx={!children && isRoute ? classes.navSelected : classes.nav}
      >
        <ListItemButton tabIndex={-1}>
          <ListItemIcon>{icon}</ListItemIcon>
          <ListItemText
            disableTypography
            primary={
              <Typography variant='subtitle1' component='p'>
                {title}
              </Typography>
            }
          />
          {children && (
            <ExpandMoreIcon color='action' sx={classes.dropdown(isRoute)} />
          )}
        </ListItemButton>
      </AppLink>
      {children && (
        <Collapse in={isRoute} mountOnEnter>
          <List sx={{ p: 0 }}>{children}</List>
        </Collapse>
      )}
    </React.Fragment>
  )
}
