import React from 'react'
import { NavLink } from 'react-router-dom'
import List from '@material-ui/core/List'
import { makeStyles } from '@material-ui/styles'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import ListItemIcon from '@material-ui/core/ListItemIcon'
import Typography from '@material-ui/core/Typography'
import { styles } from '../../styles/materialStyles'
import ExpandMoreIcon from '@material-ui/icons/ExpandMore'
import { Collapse } from '@material-ui/core'
import { urlPathSelector } from '../../selectors/url'
import { useSelector } from 'react-redux'

const useStyles = makeStyles(theme => {
  const { nav } = styles(theme)
  return {
    nav,
    subMenu: {
      backgroundColor: theme.palette.primary['500'],
      padding: '0',
    },
    link: {
      textDecoration: 'none',
      display: 'block',
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
    subMenuSelected: {
      color: '#616161',
      backgroundColor: '#ebebeb',
      borderRight: '3px solid #D3D3D3',
    },
  }
})

export default function NavSubMenu(props) {
  const { parentIcon, parentTitle, path, subMenuRoutes } = props
  const classes = useStyles()
  const pathname = useSelector(urlPathSelector)
  let isRoute = pathname.startsWith(path)

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
        <ExpandMoreIcon />
      </ListItem>
    )
  }

  function renderSubMenu(subMenuRoutes) {
    let subMenu = subMenuRoutes.map((route, key) => {
      return (
        <NavLink
          exact
          activeClassName={classes.subMenuSelected}
          key={key}
          className={classes.link}
          to={route.path}
        >
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
