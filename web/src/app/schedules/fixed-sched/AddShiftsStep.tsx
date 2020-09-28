import React, { useEffect, useState } from 'react'
import {
  DialogContentText,
  Fab,
  Grid,
  IconButton,
  Typography,
  makeStyles,
} from '@material-ui/core'
import { Add as AddIcon, Delete as DeleteIcon } from '@material-ui/icons'
import {
  fmt,
  Value,
  Shift,
  User,
  contentText,
  StepContainer,
} from './sharedUtils'
import { FormContainer, FormField } from '../../forms'
import { UserSelect } from '../../selection'
import { ISODateTimePicker } from '../../util/ISOPickers'
import FlatList from '../../lists/FlatList'
import { UserAvatar } from '../../util/avatars'
import { DateTime } from 'luxon'

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
  edit?: boolean
}

export default function AddShiftsStep({
  value,
  onChange,
  edit,
}: AddShiftsStepProps) {
  const classes = useStyles()
  const [shift, setShift] = useState(null as Shift | null)
  const { start, end, shifts } = value

  // set start equal to the fixed schedule's start
  // can't this do on mount since the step renderer puts everyone on the DOM at once
  useEffect(() => {
    if (!value.shifts?.length && value.start && !shift?.start) {
      setShift({
        ...shift,
        start: value.start,
      } as Shift)
    }
  }, [value.start])

  // don't allow user to set start after end, or end before start
  // start with value's start/end as min/max
  const f = (d: string) => DateTime.fromISO(d).toFormat("yyyy-MM-dd'T'HH:mm:ss")
  let min = f(start)
  let max = f(end)
  if (shift?.start) min = f(shift.start)
  if (shift?.end) max = f(shift.end)

  function handleAddShift() {
    if (!shift) return

    // update shifts value
    onChange({
      ...value,
      shifts: [...shifts, shift],
    })

    // set next start date equal to the end date just added
    setShift({
      start: shift.end,
      end: '',
      user: null,
    })
  }

  const shiftFieldsEmpty = !shift?.start || !shift.end || !shift.user?.label
  function handleRemoveShift(idx: number) {
    const newShifts = shifts.slice()
    newShifts.splice(idx, 1)

    // populate shift to be deleted in add shift form if it's currently empty
    if (shiftFieldsEmpty) {
      setShift({
        start: shifts[idx].start,
        end: shifts[idx].end,
        user: shifts[idx].user,
      })
    }

    // update shifts value
    onChange({
      ...value,
      shifts: newShifts,
    })
  }

  function mapShiftstoItems() {
    if (!shifts.length) return []

    return shifts.map((shift: Shift, idx: number) => ({
      title: shift?.user?.label,
      subText: `From ${fmt(shift.start)} to ${fmt(shift.end)}`,
      icon: <UserAvatar userID={shift?.user?.value ?? ''} />,
      secondaryAction: (
        <IconButton onClick={() => handleRemoveShift(idx)}>
          <DeleteIcon />
        </IconButton>
      ),
    }))
  }

  return (
    <StepContainer>
      {/* main container for fields | button | shifts */}
      <Grid container spacing={2}>
        {/* title + fields container */}
        <Grid item xs={5} container spacing={2} direction='column'>
          <Grid item>
            <Typography variant='body2'>
              {edit ? 'STEP 1 OF 2' : 'STEP 2 OF 3'}
            </Typography>
            <Typography variant='h6' component='h2'>
              Determine each user's on-call shift.
            </Typography>
          </Grid>
          <Grid item>
            <DialogContentText className={classes.contentText}>
              Configuring a fixed schedule from {fmt(value.start)} to{' '}
              {fmt(value.end)}. Select a user and when they will be on call to
              add them to this fixed schedule.
            </DialogContentText>
          </Grid>
          <FormContainer value={shift} onChange={(val: Shift) => setShift(val)}>
            <Grid item>
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
            <Grid item>
              <FormField
                fullWidth
                component={ISODateTimePicker}
                label='Shift Start'
                name='start'
                inputProps={{ min, max }}
              />
            </Grid>
            <Grid item>
              <FormField
                fullWidth
                component={ISODateTimePicker}
                label='Shift End'
                name='end'
                inputProps={{ min, max }}
              />
            </Grid>
          </FormContainer>
        </Grid>

        {/* add button container */}
        <Grid item xs={2} className={classes.addButtonContainer}>
          <Fab
            className={classes.addButton}
            onClick={handleAddShift}
            disabled={shiftFieldsEmpty}
            size='medium'
            color='primary'
          >
            <AddIcon />
          </Fab>
        </Grid>

        {/* shifts list container */}
        <Grid item xs={5}>
          <Typography variant='subtitle1' component='h3'>
            Shifts
          </Typography>
          <FlatList
            items={mapShiftstoItems()}
            emptyMessage='Add a user to the left to get started.'
            dense
            ListItemProps={{
              disableGutters: true,
              divider: true,
            }}
          />
        </Grid>
      </Grid>
    </StepContainer>
  )
}
