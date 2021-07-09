import React from 'react'
import p from 'prop-types'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import SwipeableDrawer from '@material-ui/core/SwipeableDrawer'

export default function OtherActionsMenuDesktop({ isOpen, onClose, actions }) {
  return (
    <SwipeableDrawer
      anchor='bottom'
      disableDiscovery
      disableSwipeToOpen
      open={isOpen}
      onOpen={() => null}
      onClose={onClose}
    >
      <List data-cy='mobile-actions'>
        {actions.map((o, idx) => (
          <ListItem
            key={idx}
            button
            onClick={() => {
              onClose()
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

OtherActionsMenuDesktop.propTypes = {
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
