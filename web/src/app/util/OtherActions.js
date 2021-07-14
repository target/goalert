import React, { useRef, useState } from 'react'
import p from 'prop-types'
import IconButton from '@material-ui/core/IconButton'
import { MoreHoriz as OptionsIcon } from '@material-ui/icons'
import Hidden from '@material-ui/core/Hidden'
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

export default function OtherActions({ color, icon, actions, placement }) {
  const [anchorEl, setAnchorEl] = useState(null)
  const onClose = cancelable(() => setAnchorEl(null))
  const ref = useRef(null)

  return (
    <React.Fragment>
      <span ref={ref}>
        {React.cloneElement(icon, {
          'aria-label': 'Other Actions',
          'data-cy': 'other-actions',
          color: color || 'inherit',
          'aria-expanded': Boolean(anchorEl),
          onClick: (e) => {
            onClose.cancel()
            setAnchorEl(e.currentTarget)
          },
        })}
      </span>
      <Hidden smDown>
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
    }),
  ).isRequired,
  color: p.string,
  icon: p.element,
  placement: p.oneOf(['left', 'right']),
}

OtherActions.defaultProps = {
  icon: (
    <IconButton>
      <OptionsIcon />
    </IconButton>
  ),
  placement: 'left',
}
