import React, { useEffect, useState } from 'react'
import {
  Button,
  DialogContentText,
  Grid,
  Typography,
  makeStyles,
  FormControlLabel,
  Checkbox,
  FormHelperText,
} from '@material-ui/core'
import ArrowRightAltIcon from '@material-ui/icons/ArrowRightAlt'
import { contentText, Shift, StepContainer } from './sharedUtils'
import { FormContainer } from '../../forms'
import _ from 'lodash'
import TempSchedShiftsList from './TempSchedShiftsList'
import TempSchedAddShiftForm from './TempSchedAddShiftForm'
import { DateTime, Interval } from 'luxon'
import { FieldError } from '../../util/errutil'
import { isISOAfter } from '../../util/shifts'
import { Alert, AlertTitle } from '@material-ui/lab'

const useStyles = makeStyles((theme) => ({
  contentText,
  avatar: {
    backgroundColor: theme.palette.primary.main,
  },
  shiftsListContainer: {
    height: '100%',
    display: 'flex',
    flexDirection: 'column',
  },
  listOuterContainer: {
    height: '100%',
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
    paddingRight: '2rem',
  },
  noCoverageError: {
    marginTop: '.5rem',
    marginBottom: '.5rem',
  },
}))

type AddShiftsStepProps = {
  value: Shift[]
  onChange: (newValue: Shift[]) => void
  start: string
  end: string

  scheduleID: string
  edit?: boolean

  isAllowingNoCoverage: boolean
  setIsAllowingNoCoverage: (isAllowing: boolean) => void
  isShowingNoCoverageWarning: boolean
  hasNoCoverageGaps: boolean
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
  onChange,
  start,
  end,
  value,
  edit,
  isAllowingNoCoverage,
  setIsAllowingNoCoverage,
  isShowingNoCoverageWarning,
  hasNoCoverageGaps,
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
    <StepContainer data-cy='add-shifts-step'>
      {/* main container for fields | button | shifts */}
      <Grid container spacing={0} className={classes.mainContainer}>
        {/* title + fields container */}
        <Grid
          item
          xs={6}
          container
          spacing={2}
          direction='column'
          className={classes.shiftFormContainer}
        >
          <Grid item>
            {!edit && <Typography variant='body2'>STEP 2 OF 2</Typography>}
            <Typography variant='h6' component='h2'>
              Specify on-call shifts.
            </Typography>
          </Grid>
          <Grid item>
            <DialogContentText className={classes.contentText}>
              The schedule will be exactly as configured here for the entire
              duration (ignoring all assignments and overrides).
            </DialogContentText>
          </Grid>
          <FormContainer
            errors={fieldErrors()}
            value={shift}
            onChange={(val: Shift) => setShift(val)}
          >
            <TempSchedAddShiftForm
              value={shift}
              min={edit ? start : undefined}
              scheduleID={scheduleID}
            />
          </FormContainer>
          <Grid item>
            <Button
              data-cy='add-shift'
              color='secondary'
              variant='contained'
              fullWidth
              onClick={handleAddShift}
              endIcon={<ArrowRightAltIcon />}
            >
              Add Shift
            </Button>
          </Grid>
        </Grid>

        {/* shifts list container */}
        <Grid item xs={6} className={classes.shiftsListContainer}>
          <div className={classes.listOuterContainer}>
            <div className={classes.listInnerContainer}>
              <TempSchedShiftsList
                scheduleID={scheduleID}
                value={value}
                start={start}
                end={end}
                onRemove={(shift: Shift) => {
                  setShift(shift)
                  onChange(value.filter((s) => !shiftEquals(shift, s)))
                }}
                edit={edit}
              />
            </div>
          </div>
          {isShowingNoCoverageWarning && hasNoCoverageGaps && (
            <Alert severity='error' className={classes.noCoverageError}>
              <AlertTitle>Gaps in coverage</AlertTitle>
              <FormHelperText>
                There are gaps in coverage. During these gaps nobody will
                receive alerts. If you still want to proceed, check the box and
                then click Retry.
              </FormHelperText>
              <FormControlLabel
                label='Allow gaps in coverage'
                labelPlacement='end'
                control={
                  <Checkbox
                    data-cy='no-coverage-checkbox'
                    checked={isAllowingNoCoverage}
                    onChange={(e) => setIsAllowingNoCoverage(e.target.checked)}
                    name='isAwareOfNoCoverage'
                  />
                }
              />
            </Alert>
          )}
        </Grid>
        {/* <Grid item>
          
        </Grid> */}
      </Grid>
    </StepContainer>
  )
}
