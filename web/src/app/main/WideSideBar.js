import React from 'react'
import withStyles from '@material-ui/core/styles/withStyles'
import Drawer from '@material-ui/core/Drawer'

const drawerWidth = '10.5em'
const styles = theme => ({
  sidebarPaper: {
    width: drawerWidth,
    position: 'fixed',
    transition: theme.transitions.create('width', {
      easing: theme.transitions.easing.sharp,
      duration: theme.transitions.duration.enteringScreen,
    }),
  },
})

@withStyles(styles)
export default class WideSideBar extends React.PureComponent {
  state = {
    show: false,
  }

  render() {
    const { classes } = this.props
    return (
      <Drawer
        variant='permanent'
        classes={{
          paper: classes.sidebarPaper,
        }}
      >
        {this.props.children}
      </Drawer>
    )
  }
}
