import React, { useState } from 'react'
import { NavLink } from 'react-router-dom'
import Collapse from '@material-ui/core/Collapse'
import List from '@material-ui/core/List'
import { makeStyles } from '@material-ui/styles'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import Typography from '@material-ui/core/Typography'
import ArrowDropDownIcon from '@material-ui/icons/ArrowDropDown'

const useStyles = makeStyles(theme => {
  return {
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
    parentItem: {
      color: theme.palette.primary['500'],
    },
    subMenuLinkText: {
      paddingLeft: '3rem',
    },
  }
})

export default function NavSubMenu(props) {
  const {
    parentIcon,
    parentTitle,
    unActiveClass,
    activeClass,
    path,
    key,
    subMenuRoutes,
  } = props
  const [open, setOpen] = useState(false)
  const classes = useStyles()

  function handleClick(bool) {
    setOpen(bool)
  }

  function renderParentLink(IconComponent, label) {
    return (
      <ListItem>
        <ListItemIcon>
          <IconComponent />
        </ListItemIcon>
        <ListItemText
          className={classes.parentItem}
          disableTypography
          primary={<Typography variant='subtitle1'>{label}</Typography>}
        />
        <ArrowDropDownIcon className={classes.parentItem} />
      </ListItem>
    )
  }

  function renderSubMenu(subMenuRoutes) {
    let subMenu = subMenuRoutes.map((route, key) => {
      return (
        <NavLink key={key} className={classes.link} to={route.path}>
          <ListItem>
            <ListItemText className={classes.subMenuLinkText}>
              {route.title}
            </ListItemText>
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
        {renderParentLink(parentIcon, parentTitle)}
      </NavLink>
      <Collapse in={open} timeout='auto' unmountOnExit>
        <List className={classes.subMenu}>{renderSubMenu(subMenuRoutes)}</List>
      </Collapse>
    </React.Fragment>
  )
}
