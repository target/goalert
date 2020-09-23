import React, { useState } from 'react'
import {
  Fab,
  Grid,
  DialogContentText,
  IconButton,
  Typography,
  Fade,
  makeStyles,
} from '@material-ui/core'
import AddIcon from '@material-ui/icons/Add'
import { DateTime } from 'luxon'
import { Value, Shift, User, contentText, StepContainer } from './sharedUtils'
import { FormContainer, FormField } from '../../forms'
import { UserSelect } from '../../selection'
import { ISODateTimePicker } from '../../util/ISOPickers'
import FlatList from '../../lists/FlatList'
import { UserAvatar } from '../../util/avatars'

const useStyles = makeStyles((theme) => ({
  contentText,
  addButton: {
    boxShadow: 'none',
  },
  addButtonContainer: {
    display: 'flex',
    justifyContent: 'center',
    alignItems: 'center',
  },
  avatar: {
    backgroundColor: theme.palette.primary.main,
  },
}))

interface AddShiftsStepProps {
  value: Value
  onChange: (val: Value) => any
}

export default function AddShiftsStep({ value, onChange }: AddShiftsStepProps) {
  const classes = useStyles()
  const [shift, setShift] = useState(null as Shift | null)
  const { shifts } = value

  const fmt = (t: string) =>
    DateTime.fromISO(t).toLocaleString(DateTime.DATETIME_MED)

  function handleAddShift() {
    if (!shift) return

    onChange({
      ...value,
      shifts: [...shifts, shift],
    })

    setShift({
      start: shift.end,
      end: '',
      user: null,
    })
  }

  function mapShiftstoItems() {
    if (!shifts.length) return []

    return shifts.map((shift: Shift) => ({
      title: shift?.user?.label,
      subText: `From ${fmt(shift.start)} to ${fmt(shift.end)}`,
      icon: <UserAvatar userID={shift?.user?.value ?? ''} />,
    }))
  }

  return (
    <StepContainer>
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

        <Grid item xs={12} container spacing={2}>
          <Grid item xs={10} container spacing={2}>
            <FormContainer
              value={shift}
              onChange={(val: Shift) => setShift(val)}
            >
              <Grid item xs={12}>
                <FormField
                  fullWidth
                  saveLabel
                  component={UserSelect}
                  saveLabelOnChange
                  label='Select a User'
                  name='user'
                  mapValue={(u: User) => u?.value}
                />
              </Grid>
              <Grid item xs={6}>
                <FormField
                  fullWidth
                  component={ISODateTimePicker}
                  label='Shift Start'
                  name='start'
                />
              </Grid>
              <Grid item xs={6}>
                <FormField
                  fullWidth
                  component={ISODateTimePicker}
                  label='Shift End'
                  name='end'
                />
              </Grid>
            </FormContainer>
          </Grid>
          <Grid className={classes.addButtonContainer} item xs={2}>
            <Fab
              className={classes.addButton}
              onClick={handleAddShift}
              disabled={!shift?.start || !shift.end || !shift.user?.label}
              size='medium'
              color='primary'
            >
              <AddIcon />
            </Fab>
          </Grid>
        </Grid>

        <Fade in={shifts.length > 0}>
          <Grid item xs={12}>
            <Typography variant='subtitle1' component='h3'>
              Shifts
            </Typography>
            <FlatList
              items={mapShiftstoItems()}
              emptyMessage='Add a user above to get started.' // fallback empty message
              dense
              ListItemProps={{
                disableGutters: true,
                divider: true,
              }}
            />
          </Grid>
        </Fade>
      </Grid>
    </StepContainer>
  )
}
