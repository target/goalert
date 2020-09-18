import React from 'react'
import {
  Grid,
  DialogContentText,
  IconButton,
  Typography,
  makeStyles,
} from '@material-ui/core'
import AddIcon from '@material-ui/icons/ArrowDownward'
import { DateTime } from 'luxon'
import { FormField } from '../../forms'
import { UserSelect } from '../../selection'
import { ISODateTimePicker } from '../../util/ISOPickers'
import FlatList from '../../lists/FlatList'
import { UserAvatar } from '../../util/avatars'

const useStyles = makeStyles({
  addButtonContainer: {
    display: 'flex',
    justifyContent: 'center',
  },
  contentText: {
    marginBottom: 0,
  },
})

interface AddShiftsStepProps {
  value: any
  onChange: (val: any) => any
}

interface Shift {
  start: string
  end: string
  user: User
}

interface User {
  label: string
  value: string
}

export default function AddShiftsStep({ value, onChange }: AddShiftsStepProps) {
  const classes = useStyles()
  const { shifts, _shift } = value

  const fmt = (t: string) =>
    DateTime.fromISO(t).toLocaleString(DateTime.DATETIME_MED)

  function handleAddShift() {
    return onChange({
      ...value,
      shifts: [...shifts, _shift],
    })
  }

  function mapShiftstoItems() {
    return shifts.map((shift: Shift) => ({
      title: shift.user.label,
      subText: `From ${fmt(shift.start)} to ${fmt(shift.end)}`,
      icon: <UserAvatar userID={'test'} />,
    }))
  }

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <Typography variant='h6' component='h2'>
          Determine each user's on-call shift.
        </Typography>
      </Grid>
      <Grid item xs={12}>
        <DialogContentText className={classes.contentText}>
          Configuring a fixed schedule from {fmt(value.start)} to{' '}
          {fmt(value.end)}. Select a user and when they will be on call to add
          them to this fixed schedule.
        </DialogContentText>
      </Grid>
      <Grid item xs={12}>
        <FormField
          fullWidth
          component={UserSelect}
          required
          saveLabelOnChange
          label='Select a User'
          name={`_shift.user`}
          mapValue={(u: User) => u?.value ?? ''}
        />
      </Grid>
      <Grid item xs={6}>
        <FormField
          fullWidth
          component={ISODateTimePicker}
          required
          label='Shift Start'
          name={`_shift.start`}
        />
      </Grid>
      <Grid item xs={6}>
        <FormField
          fullWidth
          component={ISODateTimePicker}
          required
          label='Shift End'
          name={`_shift.end`}
        />
      </Grid>
      <Grid className={classes.addButtonContainer} item xs={12}>
        <IconButton
          onClick={handleAddShift}
          disabled={!_shift.start || !_shift.end || !_shift.user?.value}
        >
          <AddIcon />
        </IconButton>
      </Grid>
      <Grid item xs={12}>
        <FlatList items={mapShiftstoItems()} emptyMessage='' />
      </Grid>
    </Grid>
  )
}
