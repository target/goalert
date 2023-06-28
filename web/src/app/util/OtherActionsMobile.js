import React from 'react'
import p from 'prop-types'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import SwipeableDrawer from '@mui/material/SwipeableDrawer'

export default function OtherActionsMobile({ isOpen, onClose, actions }) {
  return (
    <SwipeableDrawer
      anchor='bottom'
      disableDiscovery
      disableSwipeToOpen
      open={isOpen}
      onOpen={() => null}
      onClose={onClose}
      SlideProps={{
        unmountOnExit: true,
      }}
    >
      <List data-cy='mobile-actions' role='menu'>
        {actions.map((o, idx) => (
          <ListItem
            key={idx}
            role='menuitem'
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

OtherActionsMobile.propTypes = {
  onClose: p.func,
  isOpen: p.bool,
  actions: p.arrayOf(
    p.shape({
      label: p.string.isRequired,
      onClick: p.func.isRequired,
    }),
  ),
}
