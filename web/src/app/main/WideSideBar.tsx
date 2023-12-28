import React, { ReactNode } from 'react'
import Drawer from '@mui/material/Drawer'
import makeStyles from '@mui/styles/makeStyles'
import { Theme } from '@mui/material/styles'

export const drawerWidth = '190px'

const useStyles = makeStyles((theme: Theme) => ({
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
