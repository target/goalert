import React from 'react'
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

export default class OtherActions extends React.PureComponent {
  static propTypes = {
    actions: p.arrayOf(
      p.shape({
        label: p.string.isRequired,
        onClick: p.func.isRequired,
      }),
    ).isRequired,
    icon: p.element,
    placement: p.oneOf(['left', 'right']),
  }

  static defaultProps = {
    icon: (
      <IconButton>
        <OptionsIcon />
      </IconButton>
    ),
    placement: 'left',
  }

  state = {
    anchorEl: null,
  }

  render() {
    const onClose = cancelable(() =>
      this.setState({
        anchorEl: null,
      }),
    )
    return (
      <React.Fragment>
        <span ref={this.ref}>
          {React.cloneElement(this.props.icon, {
            'aria-label': 'Other Actions',
            'data-cy': 'other-actions',
            color: 'inherit',
            'aria-expanded': Boolean(this.state.anchorEl),
            onClick: (e) => {
              onClose.cancel()
              this.setState({
                anchorEl: e.currentTarget,
              })
            },
          })}
        </span>
        <Hidden smDown>
          <OtherActionsDesktop
            isOpen={Boolean(this.state.anchorEl)}
            onClose={onClose}
            actions={this.props.actions}
            anchorEl={this.state.anchorEl}
            placement={this.props.placement}
          />
        </Hidden>
        <Hidden mdUp>
          <OtherActionsMobile
            isOpen={Boolean(this.state.anchorEl)}
            onClose={onClose}
            actions={this.props.actions}
          />
        </Hidden>
      </React.Fragment>
    )
  }
}
