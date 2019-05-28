import React from 'react'
import p from 'prop-types'

import Menu from '@material-ui/core/Menu'
import MenuItem from '@material-ui/core/MenuItem'

export default class OtherActionsMenuDesktop extends React.PureComponent {
  static propTypes = {
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
  static defaultProps = {
    onClose: () => {},
    actions: [],
    placement: 'left',
  }

  render() {
    // anchorDir is the opposite of the wanted menu placement
    const anchorDir = this.props.placement === 'left' ? 'right' : 'left'
    return (
      <Menu
        anchorEl={() => this.props.anchorEl}
        getContentAnchorEl={null}
        open={this.props.isOpen}
        onClose={this.props.onClose}
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
        {this.props.actions.map((o, idx) => (
          <MenuItem
            key={idx}
            onClick={() => {
              this.props.onClose()
              o.onClick()
            }}
          >
            {o.label}
          </MenuItem>
        ))}
      </Menu>
    )
  }
}
