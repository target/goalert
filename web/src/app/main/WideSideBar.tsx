import React, { ReactNode } from 'react'
import Drawer from '@material-ui/core/Drawer'
import { makeStyles } from '@material-ui/core'

const drawerWidth = '12em'
const useStyles = makeStyles((theme) => ({
  sidebarPaper: {
    width: drawerWidth,
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
      {props.children}
    </Drawer>
  )
}

export default WideSideBar
