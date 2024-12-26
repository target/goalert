import React from 'react'
import { useLocation } from 'wouter'
import List from '@mui/material/List'
import makeStyles from '@mui/styles/makeStyles'
import ListItemText from '@mui/material/ListItemText'
import ListItemIcon from '@mui/material/ListItemIcon'
import Typography from '@mui/material/Typography'
import { styles } from '../styles/materialStyles'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import { Collapse, ListItemButton, Theme } from '@mui/material'
import AppLink from '../util/AppLink'
import { OpenInNew } from '@mui/icons-material'

const useStyles = makeStyles((theme: Theme) => {
  const { nav, navSelected } = styles(theme)
  return {
    nav,
    navSelected,
    subMenu: {
      padding: '0',
    },
    subMenuLinkText: {
      paddingLeft: '3.5rem',
      '& span': {
        fontSize: '.9rem',
      },
    },
    dropdown: {
      transition: theme.transitions.create(['transform'], {
        duration: theme.transitions.duration.short,
      }),
    },
    dropdownOpen: {
      transform: 'rotate(0)',
    },
    dropdownClosed: {
      transform: 'rotate(-90deg)',
    },
  }
})

export type NavBarSubLinkProps = {
  to: string
  title: string
  newTab?: boolean
}
export function NavBarSubLink({
  to,
  title,
  newTab,
}: NavBarSubLinkProps): JSX.Element {
  const { navSelected, nav, subMenuLinkText } = useStyles()
  const [path] = useLocation()
  return (
    <AppLink
      className={path.startsWith(to) ? navSelected : nav}
      to={to}
      newTab={newTab}
    >
      <ListItemButton tabIndex={-1}>
        <ListItemText className={subMenuLinkText}>
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
  icon: JSX.Element
  title: string
  to: string
  children?: React.ReactNode[] | React.ReactNode
}

export default function NavBarLink({
  icon,
  title,
  to,
  children,
}: NavBarLinkProps): JSX.Element {
  const classes = useStyles()
  const [path] = useLocation()
  const isRoute = path.startsWith(to)

  return (
    <React.Fragment>
      <AppLink
        to={to}
        className={!children && isRoute ? classes.navSelected : classes.nav}
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
            <ExpandMoreIcon
              color='action'
              className={
                classes.dropdown +
                ' ' +
                (isRoute ? classes.dropdownOpen : classes.dropdownClosed)
              }
            />
          )}
        </ListItemButton>
      </AppLink>
      {children && (
        <Collapse in={isRoute} mountOnEnter>
          <List className={classes.subMenu}>{children}</List>
        </Collapse>
      )}
    </React.Fragment>
  )
}
