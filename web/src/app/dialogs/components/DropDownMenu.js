import React, { Component } from 'react'
import IconButton from '@material-ui/core/IconButton'
import Menu from '@material-ui/core/Menu'
import MenuItem from '@material-ui/core/MenuItem'
import withStyles from '@material-ui/core/styles/withStyles'
import { MoreVert as MoreVertIcon } from '@material-ui/icons'
import { styles } from '../../styles/materialStyles'

/*
  Takes a list of options each with a label and an onClick function
*/

@withStyles(styles)
export default class DropdownMenu extends Component {
  constructor(props) {
    super(props)

    this.state = {
      anchorEl: null,
    }
  }

  render() {
    const options = this.props.options

    return (
      <div>
        <IconButton
          aria-label={this.props['aria-label']}
          data-cy={this.props['data-cy']}
          onClick={(event) => this.setState({ anchorEl: event.currentTarget })}
          aria-haspopup='true'
          style={{ color: this.props.color || 'inherit' }}
        >
          <MoreVertIcon />
        </IconButton>
        <Menu
          anchorEl={this.state.anchorEl}
          open={!!this.state.anchorEl}
          onExited={() => {
            if (this._fn) {
              this._fn()
              this._fn = null
            }
          }}
          onClose={() => {
            this.setState({ anchorEl: null })
          }}
        >
          {options.map((option) => (
            <MenuItem
              key={option.label}
              onClick={() => {
                this.setState({ anchorEl: null })
                this._fn = option.onClick
              }}
            >
              {option.label}
            </MenuItem>
          ))}
        </Menu>
      </div>
    )
  }
}
