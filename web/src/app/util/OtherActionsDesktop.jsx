import React from 'react'
import p from 'prop-types'

import Menu from '@mui/material/Menu'
import MenuItem from '@mui/material/MenuItem'
import { Tooltip } from '@mui/material'

export default function OtherActionsMenuDesktop({
  placement = 'left',
  anchorEl,
  isOpen,
  actions = [],
  onClose = () => {},
}) {
  // anchorDir is the opposite of the wanted menu placement
  const anchorDir = placement === 'left' ? 'right' : 'left'

  return (
    <Menu
      anchorEl={anchorEl}
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
        // tooltip for alt text on menuitem hover
        // wrapped with div to allow tooltip to work on disabled menuitem
        <Tooltip key={idx} title={o.tooltip}>
          <div>
            <MenuItem
              onClick={() => {
                onClose()
                o.onClick()
              }}
              disabled={o.disabled}
            >
              {o.label}
            </MenuItem>
          </div>
        </Tooltip>
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
      disabled: p.bool,
      tooltip: p.string,
    }),
  ),

  placement: p.oneOf(['left', 'right']),
}
