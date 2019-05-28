import React, { Component } from 'react'
import Button from '@material-ui/core/Button'
import Menu from '@material-ui/core/Menu'
import MenuItem from '@material-ui/core/MenuItem'
import { ExpandMore } from '@material-ui/icons'

/*
 * Takes an options object array to render a dropdown menu.
 *
 * Each option should contain a label, item action, and optionally
 * a flag disabling the item or not.
 *
 * E.g.
 * [{
 *   label: 'Menu option label',
 *   disabled: this.state.disabled,
 *   action: () => someAction(foo, bar)
 * }]
 */
export class BaseActionsMenu extends Component {
  constructor(props) {
    super(props)

    this.state = {
      anchorEl: null,
      open: false,
    }
  }

  handleClick = event => {
    this.setState({ open: true, anchorEl: event.currentTarget })
  }

  render() {
    // Forces actions menu to the right of the table (should always be the last table cell in a row)
    const style = {
      float: 'right',
    }

    return (
      <div style={style}>
        <Button onClick={this.handleClick}>
          Actions
          <ExpandMore style={{ height: '1em', width: '1em' }} />
        </Button>
        <Menu
          anchorEl={this.state.anchorEl}
          open={(this.state.anchorEl && this.state.open) || false}
          onExited={() => {
            if (this._fn) {
              this._fn()
              this._fn = null
            }
          }}
          onClose={() => this.setState({ open: false })}
        >
          {this.props.options.map(item => {
            return (
              <MenuItem
                key={item.label}
                disabled={item.disabled}
                onClick={() => {
                  this.setState({ open: false })
                  this._fn = item.action
                }}
              >
                {item.label}
              </MenuItem>
            )
          })}
        </Menu>
      </div>
    )
  }
}
