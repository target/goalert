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
import _ from 'lodash-es'
import FixedSchedShiftsList from './FixedSchedShiftsList'
import FixedSchedAddShiftForm from './FixedSchedAddShiftForm'
import { ScheduleTZFilter } from '../ScheduleTZFilter'
import { DateTime, Interval } from 'luxon'
import { FieldError } from '../../util/errutil'

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

function isAfter(a: string, b: string): boolean {
  return DateTime.fromISO(a) > DateTime.fromISO(b)
}

function isBefore(a: string, b: string): boolean {
  return DateTime.fromISO(a) < DateTime.fromISO(b)
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

export default function AddShiftsStep({
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

  // set start equal to the fixed schedule's start
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

    if (!shift) {
      return result
    }

    if (s) {
      const message = 'this field is required'
      if (!shift.userID) {
        result.push({
          field: 'userID',
          message,
        } as FieldError)
      }
      if (!shift.start) {
        result.push({
          field: 'start',
          message,
        } as FieldError)
      }
      if (!shift.end) {
        result.push({
          field: 'end',
          message,
        } as FieldError)
      }

      return result
    }

    if (!isAfter(shift.end, shift?.start)) {
      result.push({
        field: 'end',
        message: 'must be after shift start time',
      } as FieldError)
    }
    if (isBefore(shift.start, start)) {
      result.push({
        field: 'start',
        message: 'must not be before fixed schedule start time',
      } as FieldError)
    }
    if (isAfter(shift.end, end)) {
      result.push({
        field: 'end',
        message: 'must not extend beyond fixed schedule end time',
      } as FieldError)
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
          <FormContainer
            errors={fieldErrors()}
            value={shift}
            onChange={(val: Shift) => setShift(val)}
          >
            <FixedSchedAddShiftForm />
          </FormContainer>
        </Grid>

        {/* add button container */}
        <Grid item xs={2} className={classes.addButtonContainer}>
          <Fab
            className={classes.addButton}
            onClick={handleAddShift}
            size='medium'
            color='primary'
          >
            <AddIcon />
          </Fab>
        </Grid>

        {/* shifts list container */}
        <Grid item xs={5} className={classes.listContainer}>
          <div style={{ position: 'absolute', width: '100%' }}>
            <FixedSchedShiftsList
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
