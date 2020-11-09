import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import IconButton from '@material-ui/core/IconButton'
import Menu from '@material-ui/core/Menu'
import MenuItem from '@material-ui/core/MenuItem'
import { MoreVert as MoreVertIcon } from '@material-ui/icons'

/*
  Takes a list of options each with a label and an onClick function
*/

function DropDownMenu(props) {
  const [anchorEl, setAnchorEl] = useState(null)
  const { options, onExited } = props

  return (
    <div>
      <IconButton
        aria-label={props['aria-label']}
        data-cy={props['data-cy']}
        onClick={(event) => setAnchorEl(event.currentTarget)}
        aria-haspopup='true'
        style={{ color: props.color || 'inherit' }}
      >
        <MoreVertIcon />
      </IconButton>
      <Menu
        anchorEl={anchorEl}
        open={!!anchorEl}
        onExited={onExited}
        onClose={() => {
          setAnchorEl(null)
        }}
      >
        {options.map((option) => (
          <MenuItem
            key={option.label}
            onClick={() => {
              setAnchorEl(null)
              option.onClick()
            }}
          >
            {option.label}
          </MenuItem>
        ))}
      </Menu>
    </div>
  )
}

DropDownMenu.propTypes = {
  options: p.arrayOf(
    p.shape({
      label: p.string,
      onClick: p.func,
    }),
  ),
  onExited: p.func,
}

export default DropDownMenu
