import React, { useEffect, useState } from 'react'
import {
  DialogContentText,
  Fab,
  Grid,
  Typography,
  makeStyles,
} from '@material-ui/core'
import { Add as AddIcon } from '@material-ui/icons'
import { fmt, Shift, contentText, StepContainer } from './sharedUtils'
import { FormContainer } from '../../forms'

import FixedSchedShiftsList from './FixedSchedShiftsList'
import FixedSchedAddShiftForm from './FixedSchedAddShiftForm'
import { ScheduleTZFilter } from '../ScheduleTZFilter'

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
  listContainer: {
    position: 'relative',
    overflowY: 'scroll',
  },
  mainContainer: {
    height: '100%',
  },
}))

const shiftEquals = (a: Shift, b: Shift) =>
  a.start === b.start && a.end === b.end && a.userID === b.userID

interface AddShiftsStepProps {
  value: Shift[]
  onChange: (newValue: Shift[]) => void
  start: string
  end: string

  scheduleID: string
  stepText: string
}

export default function AddShiftsStep({
  scheduleID,
  stepText,
  onChange,
  start,
  end,
  value,
}: AddShiftsStepProps) {
  const classes = useStyles()
  const [shift, setShift] = useState(null as Shift | null)

  // set start equal to the fixed schedule's start
  // can't this do on mount since the step renderer puts everyone on the DOM at once
  useEffect(() => {
    if (shift) return
    setShift({ start, end: '', userID: '' })
  }, [start])

  return (
    <StepContainer>
      {/* main container for fields | button | shifts */}
      <Grid container spacing={2} className={classes.mainContainer}>
        {/* title + fields container */}
        <Grid item xs={5} container spacing={2} direction='column'>
          <Grid item>
            <Typography variant='body2'>{stepText}</Typography>
            <Typography variant='h6' component='h2'>
              Determine each user's on-call shift.
            </Typography>
          </Grid>
          <Grid item>
            <DialogContentText className={classes.contentText}>
              Configuring a fixed schedule from {fmt(start)} to {fmt(end)}.
              Select a user and when they will be on call to add them to this
              fixed schedule.
            </DialogContentText>
          </Grid>
          <Grid item>
            <ScheduleTZFilter
              label={(tz) => `Configure in ${tz}`}
              scheduleID={scheduleID}
            />
          </Grid>
          <FormContainer value={shift} onChange={(val: Shift) => setShift(val)}>
            <FixedSchedAddShiftForm />
          </FormContainer>
        </Grid>

        {/* add button container */}
        <Grid item xs={2} className={classes.addButtonContainer}>
          <Fab
            className={classes.addButton}
            onClick={() => shift && onChange(value.concat(shift))}
            size='medium'
            color='primary'
          >
            <AddIcon />
          </Fab>
        </Grid>

        {/* shifts list container */}
        <Grid item xs={5} className={classes.listContainer}>
          <div style={{ position: 'absolute' }}>
            <FixedSchedShiftsList
              value={value}
              onRemove={(shift: Shift) => {
                setShift(shift)
                onChange(value.filter((s) => !shiftEquals(shift, s)))
              }}
            />
          </div>
        </Grid>
      </Grid>
    </StepContainer>
  )
}
