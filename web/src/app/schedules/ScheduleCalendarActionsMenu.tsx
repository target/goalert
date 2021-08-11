import React, { useContext, useState } from 'react'
import GroupAdd from '@material-ui/icons/GroupAdd'
import { Button, Menu, MenuItem } from '@material-ui/core'
import { ScheduleCalendarContext } from './ScheduleDetails'

function ScheduleCalendarActionsSelect(): JSX.Element {
  const { onNewTempSched, setOverrideDialog } = useContext(
    ScheduleCalendarContext,
  )
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null)

  const handleClick = (variant: string): void => {
    if (variant === 'temp') {
      onNewTempSched()
    } else {
      setOverrideDialog({ variant })
    }
    setAnchorEl(null)
  }

  return (
    <React.Fragment>
      <Button
        aria-controls='calendar-override-menu'
        aria-haspopup='true'
        onClick={(e) => setAnchorEl(e.currentTarget)}
        size='medium'
        variant='contained'
        color='primary'
        startIcon={<GroupAdd />}
      >
        Override
      </Button>
      <Menu
        id='calendar-override-menu'
        anchorEl={anchorEl}
        keepMounted
        open={Boolean(anchorEl)}
        onClose={() => setAnchorEl(null)}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'right',
        }}
        transformOrigin={{
          vertical: -8,
          horizontal: 'right',
        }}
        getContentAnchorEl={null} // unset in order to tranform origin
      >
        <MenuItem onClick={() => handleClick('add')}>Add User</MenuItem>
        <MenuItem onClick={() => handleClick('remove')}>Remove User</MenuItem>
        <MenuItem onClick={() => handleClick('replace')}>Replace User</MenuItem>
        <MenuItem onClick={() => handleClick('temp')}>
          Create Temporary Schedule
        </MenuItem>
      </Menu>
    </React.Fragment>
  )
}

export default ScheduleCalendarActionsSelect
