import React, { useRef, useState } from 'react'
import p from 'prop-types'
import IconButton from '@mui/material/IconButton'
import { MoreHoriz as OptionsIcon } from '@mui/icons-material'
import Hidden from '@mui/material/Hidden'
import OtherActionsDesktop from './OtherActionsDesktop'
import OtherActionsMobile from './OtherActionsMobile'

const cancelable = (_fn) => {
  let fn = _fn
  const cFn = (...args) => fn(...args)
  cFn.cancel = () => {
    fn = () => {}
  }
  return cFn
}

export default function OtherActions({
  color,
  IconComponent = OptionsIcon,
  actions,
  placement = 'left',
  disabled,
}) {
  const [anchorEl, setAnchorEl] = useState(null)
  const onClose = cancelable(() => setAnchorEl(null))
  const ref = useRef(null)

  return (
    <React.Fragment>
      <IconButton
        aria-label='Other Actions'
        data-cy='other-actions'
        aria-expanded={Boolean(anchorEl)}
        color='secondary'
        disabled={disabled}
        onClick={(e) => {
          onClose.cancel()
          setAnchorEl(e.currentTarget)
        }}
        ref={ref}
      >
        <IconComponent style={{ color }} />
      </IconButton>
      <Hidden mdDown>
        <OtherActionsDesktop
          isOpen={Boolean(anchorEl)}
          onClose={onClose}
          actions={actions}
          anchorEl={anchorEl}
          placement={placement}
        />
      </Hidden>
      <Hidden mdUp>
        <OtherActionsMobile
          isOpen={Boolean(anchorEl)}
          onClose={onClose}
          actions={actions}
        />
      </Hidden>
    </React.Fragment>
  )
}

OtherActions.propTypes = {
  actions: p.arrayOf(
    p.shape({
      label: p.string.isRequired,
      onClick: p.func.isRequired,
      disabled: p.bool,
      tooltip: p.string,
    }),
  ).isRequired,
  disabled: p.bool,
  color: p.string,
  IconComponent: p.elementType,
  placement: p.oneOf(['left', 'right']),
}
