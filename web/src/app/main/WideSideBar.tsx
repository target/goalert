import React, { ReactNode } from 'react'
import Drawer from '@material-ui/core/Drawer'
import { makeStyles } from '@material-ui/core'
import logo from '../public/goalert-alt-logo-scaled.png'

export const drawerWidth = '220px'

const useStyles = makeStyles((theme) => ({
  logoDiv: {
    ...theme.mixins.toolbar,
    width: '100%',
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
  },
  sidebarPaper: {
    width: 220,
    position: 'fixed',
    transition: theme.transitions.create('width', {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  },
}))

interface WideSideBarProps {
  children: ReactNode
}

function WideSideBar(props: WideSideBarProps): JSX.Element {
  const classes = useStyles()

  return (
    <Drawer
      variant='permanent'
      classes={{
        paper: classes.sidebarPaper,
      }}
    >
      <div aria-hidden className={classes.logoDiv}>
        <img height={32} src={logo} alt='GoAlert Logo' />
      </div>
      {props.children}
    </Drawer>
  )
}

export default WideSideBar
