import React, { ReactNode } from 'react'
import Drawer from '@mui/material/Drawer'

export const drawerWidth = '190px'

interface WideSideBarProps {
  children: ReactNode
}

function WideSideBar(props: WideSideBarProps): React.JSX.Element {
  return (
    <Drawer
      variant='permanent'
      slotProps={{
        paper: {
          sx: (theme) => ({
            width: drawerWidth,
            position: 'fixed',
            transition: theme.transitions.create('width', {
              easing: theme.transitions.easing.sharp,
              duration: theme.transitions.duration.enteringScreen,
            }),
          }),
        },
      }}
    >
      {props.children}
    </Drawer>
  )
}

export default WideSideBar
