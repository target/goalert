import React from 'react'
import { IconButton, Typography } from '@material-ui/core'
import FlatList from '../../lists/FlatList'
import { fmt, Shift,UserInfoObject, useUserInfo } from './sharedUtils'
import { UserAvatar } from '../../util/avatars'
import { Delete } from '@material-ui/icons'


type FixedSchedShiftsListProps = {
    value: Shift[]
    onRemove: (shift: Shift) => void
}

export default function FixedSchedShiftsList({value, onRemove}: FixedSchedShiftsListProps) {
    const shifts = useUserInfo(value)
    
    function items() {
        return shifts.map((shift: Shift & UserInfoObject, idx: number) => ({
          title: shift.user.name,
          subText: `From ${fmt(shift.start)} to ${fmt(shift.end)}`,
          icon: <UserAvatar userID={shift.userID} />,
          secondaryAction: (
            <IconButton onClick={() => onRemove(shift)}>
              <Delete />
            </IconButton>
          ),
        }))
      }

    return (
    <React.Fragment>
        <Typography variant='subtitle1' component='h3'>
        Shifts
      </Typography>
      <FlatList
        items={items()}
        emptyMessage='Add a user to the left to get started.'
        dense
        ListItemProps={{
          disableGutters: true,
          divider: true,
        }}
      />
    </React.Fragment>
    )
}