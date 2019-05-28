import React from 'react'
import p from 'prop-types'

import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import SwipeableDrawer from '@material-ui/core/SwipeableDrawer'

export default class OtherActionsMenuDesktop extends React.PureComponent {
  static propTypes = {
    anchorEl: p.object,

    onClose: p.func,
    isOpen: p.bool,

    actions: p.arrayOf(
      p.shape({
        label: p.string.isRequired,
        onClick: p.func.isRequired,
      }),
    ),
  }

  render() {
    return (
      <SwipeableDrawer
        anchor='bottom'
        disableDiscovery
        disableSwipeToOpen
        open={this.props.isOpen}
        onOpen={() => null}
        onClose={this.props.onClose}
      >
        <List data-cy='mobile-actions'>
          {this.props.actions.map((o, idx) => (
            <ListItem
              key={idx}
              button
              onClick={() => {
                this.props.onClose()
                o.onClick()
              }}
            >
              <ListItemText primary={o.label} />
            </ListItem>
          ))}
        </List>
      </SwipeableDrawer>
    )
  }
}
