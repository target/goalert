import React, { useState } from 'react'
import { NavLink } from 'react-router-dom'
import Collapse from '@material-ui/core/Collapse'
import List from '@material-ui/core/List'
import { makeStyles } from '@material-ui/styles'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'

const useStyles = makeStyles({
  subMenu: {
    backgroundColor: '#616161',
  },
  link: {
    textDecoration: 'none',
    color: '#fff',
    '&:hover': {
      textDecoration: 'none',
    },
  },
})

export default function NavSubMenu(props) {
  const {
    unActiveClass,
    activeClass,
    path,
    key,
    subMenuRoutes,
    children,
  } = props
  const [open, setOpen] = useState(false)
  const classes = useStyles()

  function handleClick(bool) {
    setOpen(bool)
  }

  function renderSubMenu(subMenuRoutes) {
    let subMenu = subMenuRoutes.map((route, key) => {
      return (
        <NavLink key={key} className={classes.link} to={route.path}>
          <ListItem>
            <ListItemText>{route.title}</ListItemText>
          </ListItem>
        </NavLink>
      )
    })

    return subMenu
  }

  function activeLink(match) {
    if (!match) {
      return handleClick(false)
    }

    return handleClick(true)
  }

  return (
    <React.Fragment>
      <NavLink
        key={key}
        to={path}
        className={(unActiveClass, classes.link)}
        activeClassName={activeClass}
        isActive={activeLink}
      >
        {children}
      </NavLink>
      <Collapse in={open} timeout='auto' unmountOnExit>
        <List className={classes.subMenu}>{renderSubMenu(subMenuRoutes)}</List>
      </Collapse>
    </React.Fragment>
  )
}
