import React from 'react'
import p from 'prop-types'

import Menu from '@material-ui/core/Menu'
import MenuItem from '@material-ui/core/MenuItem'

export default function OtherActionsMenuDesktop({placement,anchorEl,isOpen,actions,onClose}) {

    // anchorDir is the opposite of the wanted menu placement
    const anchorDir = placement === 'left' ? 'right' : 'left'
    
    return (
      <Menu
        anchorEl={() => anchorEl}
        getContentAnchorEl={null}
        open={isOpen}
        onClose={onClose}
        PaperProps={{
          style: {
            minWidth: '15em',
          },
        }}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: anchorDir,
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: anchorDir,
        }}
      >
        {actions.map((o, idx) => (
          <MenuItem
            key={idx}
            onClick={() => {
              onClose()
              o.onClick()
            }}
          >
            {o.label}
          </MenuItem>
        ))}
      </Menu>
    )
}

OtherActionsMenuDesktop.propTypes = {
    anchorEl: p.object,
    isOpen: p.bool,
    onClose: p.func,

    actions: p.arrayOf(
      p.shape({
        label: p.string.isRequired,
        onClick: p.func.isRequired,
      }),
    ),

    placement: p.oneOf(['left', 'right']),
  }

  OtherActionsMenuDesktop.defaultProps = {
    onClose: () => {},
    actions: [],
    placement: 'left',
  }
