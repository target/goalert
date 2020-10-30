import React, { useEffect, useState } from 'react'
import {
  DialogContentText,
  Fab,
  Grid,
  Typography,
  makeStyles,
} from '@material-ui/core'
import { Add as AddIcon } from '@material-ui/icons'
import { contentText, Shift, StepContainer } from './sharedUtils'
import { FormContainer } from '../../forms'
import _ from 'lodash-es'
import TempSchedShiftsList from './TempSchedShiftsList'
import TempSchedAddShiftForm from './TempSchedAddShiftForm'
import { ScheduleTZFilter } from '../ScheduleTZFilter'
import { DateTime, Interval } from 'luxon'
import { FieldError } from '../../util/errutil'
import { isISOAfter, isISOBefore } from '../../util/shifts'

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
  listOuterContainer: {
    position: 'relative',
    overflowY: 'auto',
  },
  listInnerContainer: {
    position: 'absolute',
    width: '100%',
  },
  mainContainer: {
    height: '100%',
  },
  shiftFormContainer: {
    maxHeight: '100%',
  },
}))

type AddShiftsStepProps = {
  value: Shift[]
  onChange: (newValue: Shift[]) => void
  start: string
  end: string

  scheduleID: string
  stepText: string
}

type DTShift = {
  userID: string
  span: Interval
}

function shiftsToDT(shifts: Shift[]): DTShift[] {
  return shifts.map((s) => ({
    userID: s.userID,
    span: Interval.fromDateTimes(
      DateTime.fromISO(s.start),
      DateTime.fromISO(s.end),
    ),
  }))
}

function DTToShifts(shifts: DTShift[]): Shift[] {
  return shifts.map((s) => ({
    userID: s.userID,
    start: s.span.start.toISO(),
    end: s.span.end.toISO(),
  }))
}

function shiftEquals(a: Shift, b: Shift): boolean {
  return a.start === b.start && a.end === b.end && a.userID === b.userID
}

// mergeShifts will take the incoming shifts and merge them with
// the shifts stored in value. Using Luxon's Interval, overlaps
// and edge cases when merging are handled for us.
function mergeShifts(_shifts: Shift[]): Shift[] {
  const byUser = _.groupBy(shiftsToDT(_shifts), 'userID')

  return DTToShifts(
    _.flatten(
      _.values(
        _.mapValues(byUser, (shifts, userID) => {
          return Interval.merge(_.map(shifts, 'span')).map((span) => ({
            userID,
            span,
          }))
        }),
      ),
    ),
  )
}

export default function TempSchedAddShiftsStep({
  scheduleID,
  stepText,
  onChange,
  start,
  end,
  value,
}: AddShiftsStepProps): JSX.Element {
  const classes = useStyles()
  const [shift, setShift] = useState(null as Shift | null)
  const [submitted, setSubmitted] = useState(false)

  // set start equal to the temporary schedule's start
  // can't this do on mount since the step renderer puts everyone on the DOM at once
  useEffect(() => {
    setShift({
      start,
      end: DateTime.fromISO(start).plus({ hours: 8 }).toISO(),
      userID: '',
    })
  }, [start])

  // fieldErrors handles errors manually through the client
  // as this step form is nested inside the greater form
  // that makes the network request.
  function fieldErrors(s = submitted): FieldError[] {
    const result: FieldError[] = []
    const requiredMsg = 'this field is required'
    const add = (field: string, message: string): void => {
      result.push({ field, message } as FieldError)
    }

    if (!shift) return result
    if (s) {
      if (!shift.userID) add('userID', requiredMsg)
      if (!shift.start) add('start', requiredMsg)
      if (!shift.end) add('end', requiredMsg)
    }

    if (!isISOAfter(shift.end, shift.start)) {
      add('end', 'must be after shift start time')
      add('start', 'must be before shift end time')
    }
    if (isISOBefore(shift.start, start)) {
      add('start', 'must not be before temporary schedule start time')
    }
    if (isISOAfter(shift.end, end)) {
      add('end', 'must not extend beyond temporary schedule end time')
    }
    return result
  }

  function handleAddShift(): void {
    if (fieldErrors(true).length) {
      setSubmitted(true)
      return
    }
    if (!shift) return // ts sanity check

    onChange(mergeShifts(value.concat(shift)))
    const end = DateTime.fromISO(shift.end)
    const diff = end.diff(DateTime.fromISO(shift.start))
    setShift({
      userID: '',
      start: shift.end,
      end: end.plus(diff).toISO(),
    })
    setSubmitted(false)
  }

  return (
    <StepContainer>
      {/* main container for fields | button | shifts */}
      <Grid container spacing={0} className={classes.mainContainer}>
        {/* title + fields container */}
        <Grid
          item
          xs={5}
          container
          spacing={2}
          direction='column'
          className={classes.shiftFormContainer}
        >
          <Grid item>
            <Typography variant='body2'>{stepText}</Typography>
            <Typography variant='h6' component='h2'>
              Specify on-call shifts.
            </Typography>
          </Grid>
          <Grid item>
            <DialogContentText className={classes.contentText}>
              During a temporary schedule, all on-call shifts will be set as
              configured here. A temporary schedule ignores all rules,
              rotations, and overrides. On-call will be exactly as configured
              here for the entire duration.
            </DialogContentText>
          </Grid>
          <Grid item>
            <ScheduleTZFilter
              label={(tz) => `Configure in ${tz}`}
              scheduleID={scheduleID}
            />
          </Grid>
          <FormContainer
            errors={fieldErrors()}
            value={shift}
            onChange={(val: Shift) => setShift(val)}
          >
            <TempSchedAddShiftForm />
          </FormContainer>
        </Grid>

        {/* add button container */}
        <Grid item xs={2} className={classes.addButtonContainer}>
          <Fab
            className={classes.addButton}
            aria-label='Add Shift'
            title='Add Shift'
            onClick={handleAddShift}
            size='medium'
            color='primary'
            type='button'
            disabled={Boolean(fieldErrors().length)}
          >
            <AddIcon />
          </Fab>
        </Grid>

        {/* shifts list container */}
        <Grid item xs={5} className={classes.listOuterContainer}>
          <div className={classes.listInnerContainer}>
            <TempSchedShiftsList
              value={value}
              start={start}
              end={end}
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
