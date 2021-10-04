import React from 'react'
import { NavLink } from 'react-router-dom'
import List from '@mui/material/List'
import makeStyles from '@mui/styles/makeStyles';
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import ListItemIcon from '@mui/material/ListItemIcon'
import Typography from '@mui/material/Typography'
import { styles } from '../../styles/materialStyles'
import ExpandMoreIcon from '@mui/icons-material/ExpandMore'
import { Collapse } from '@mui/material'
import { urlPathSelector } from '../../selectors/url'
import { useSelector } from 'react-redux'
import { PropTypes as p } from 'prop-types'

const useStyles = makeStyles((theme) => {
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

export default function NavSubMenu(props) {
  const { parentIcon, parentTitle, path, subMenuRoutes, closeMobileSidebar } =
    props
  const classes = useStyles()
  const pathname = useSelector(urlPathSelector)
  const isRoute = pathname.startsWith(path)

  function renderParentLink(IconComponent, label) {
    return (
      <ListItem button tabIndex={-1}>
        <ListItemIcon>
          <IconComponent />
        </ListItemIcon>
        <ListItemText
          disableTypography
          primary={
            <Typography variant='subtitle1' component='p'>
              {label}
            </Typography>
          }
        />
        <ExpandMoreIcon
          color='action'
          className={
            classes.dropdown +
            ' ' +
            (isRoute ? classes.dropdownOpen : classes.dropdownClosed)
          }
        />
      </ListItem>
    )
  }

  function renderSubMenu(subMenuRoutes) {
    const subMenu = subMenuRoutes.map((route, key) => {
      return (
        <NavLink
          exact
          activeClassName={classes.navSelected}
          key={key}
          className={classes.nav}
          to={route.path}
          onClick={closeMobileSidebar}
        >
          <ListItem button tabIndex={-1}>
            <ListItemText className={classes.subMenuLinkText}>
              {route.title}
            </ListItemText>
          </ListItem>
        </NavLink>
      )
    })

    return subMenu
  }

  return (
    <React.Fragment>
      <NavLink to={path} className={classes.nav}>
        {renderParentLink(parentIcon, parentTitle)}
      </NavLink>
      <Collapse in={isRoute} mountOnEnter>
        <List className={classes.subMenu}>{renderSubMenu(subMenuRoutes)}</List>
      </Collapse>
    </React.Fragment>
  )
}

NavSubMenu.propTypes = {
  parentIcon: p.object.isRequired,
  parentTitle: p.string.isRequired,
  path: p.string.isRequired,
  subMenuRoutes: p.array.isRequired,
  closeMobileSidebar: p.func.isRequired,
}
