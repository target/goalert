import React, { useState } from 'react'
import { NavLink } from 'react-router-dom'
import List from '@mui/material/List'
import makeStyles from '@mui/styles/makeStyles'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import ListItemIcon from '@mui/material/ListItemIcon'
import Typography from '@mui/material/Typography'
import { styles } from '../../styles/materialStyles'
import ExpandLess from '@mui/icons-material/ExpandLess'
import ExpandMore from '@mui/icons-material/ExpandMore'
import { Collapse } from '@mui/material'
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
  }
})

export default function NavSubMenu(props) {
  const { parentIcon, parentTitle, path, subMenuRoutes, closeMobileSidebar } =
    props
  const classes = useStyles()
  const [open, setOpen] = useState(false)

  function renderParentLink(IconComponent, label) {
    return (
      <ListItem button tabIndex={-1} onClick={() => setOpen(!open)}>
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
        {open ? <ExpandLess /> : <ExpandMore />}
      </ListItem>
    )
  }

  function renderSubMenu(subMenuRoutes) {
    const subMenu = subMenuRoutes.map((route, key) => {
      return (
        <NavLink
          key={key}
          className={({ isActive }) =>
            isActive ? classes.navSelected : classes.nav
          }
          to={path + route.path}
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
      <span className={classes.nav}>
        {renderParentLink(parentIcon, parentTitle)}
      </span>

      <Collapse in={open} timeout='auto' mountOnEnter unmountOnExit>
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
