import React, { useState, useEffect, useRef } from 'react'
import { useMutation, gql } from '@apollo/client'
import Checkbox from '@material-ui/core/Checkbox'
import DialogContentText from '@material-ui/core/DialogContentText'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import FormHelperText from '@material-ui/core/FormHelperText'
import Grid from '@material-ui/core/Grid'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core'
import Alert from '@material-ui/lab/Alert'
import AlertTitle from '@material-ui/lab/AlertTitle'
import _ from 'lodash'
import { DateTime, Interval } from 'luxon'

import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import {
  contentText,
  dtToDuration,
  fmtLocal,
  Shift,
  Value,
} from './sharedUtils'
import { FormContainer, FormField } from '../../forms'
import TempSchedAddNewShift from './TempSchedAddNewShift'
import { isISOAfter, parseInterval } from '../../util/shifts'
import { getNextWeekday } from '../../util/luxon-helpers'
import { useScheduleTZ } from './hooks'
import TempSchedShiftsList from './TempSchedShiftsList'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { getCoverageGapItems } from './shiftsListUtil'

const mutation = gql`
  mutation ($input: SetTemporaryScheduleInput!) {
    setTemporarySchedule(input: $input)
  }
`

function shiftEquals(a: Shift, b: Shift): boolean {
  return a.start === b.start && a.end === b.end && a.userID === b.userID
}

const useStyles = makeStyles((theme) => ({
  contentText,
  avatar: {
    backgroundColor: theme.palette.primary.main,
  },
  formContainer: {
    height: '100%',
  },
  noCoverageError: {
    marginTop: '.5rem',
    marginBottom: '.5rem',
  },
  rightPane: {
    [theme.breakpoints.down('md')]: {
      marginTop: '1rem',
    },
    overflow: 'hidden',
  },
  sticky: {
    position: 'sticky',
    top: 0,
  },
  tzNote: {
    fontStyle: 'italic',
  },
}))

type TempScheduleDialogProps = {
  onClose: () => void
  scheduleID: string
  value?: Value
}

const clampForward = (nowISO: string, iso: string | undefined): string => {
  if (!iso) return ''

  const now = DateTime.fromISO(nowISO)
  const dt = DateTime.fromISO(iso)
  if (dt < now) {
    return now.toISO()
  }
  return iso
}

export default function TempSchedDialog({
  onClose,
  scheduleID,
  value: _value,
}: TempScheduleDialogProps): JSX.Element {
  const classes = useStyles()
  const edit = Boolean(_value)
  const { q, zone, isLocalZone } = useScheduleTZ(scheduleID)
  const [now] = useState(DateTime.utc().startOf('minute').toISO())
  const [showForm, setShowForm] = useState(false)
  const [value, setValue] = useState({
    start: clampForward(now, _value?.start),
    end: _value?.end ?? '',
    clearStart: _value?.start ?? null,
    clearEnd: _value?.end ?? null,
    shifts: (_value?.shifts ?? [])
      .map((s) => _.pick(s, 'start', 'end', 'userID'))
      .filter((s) => {
        if (DateTime.fromISO(s.end) > DateTime.fromISO(now)) {
          s.start = clampForward(now, s.start)
        }
        return true
      }),
  })
  const [shift, setShift] = useState<Shift | null>(null)
  const [allowNoCoverage, setAllowNoCoverage] = useState(false)
  const [hasSubmitted, setHasSubmitted] = useState(false)

  useEffect(() => {
    // set default start, end times when zone is ready
    if (!value.start && !value.end && !q.loading && zone) {
      const nextMonday = getNextWeekday(1, DateTime.now(), zone)
      const nextFriday = nextMonday.plus({ days: 5 }) // thru to the end of Friday
      setValue({
        ...value,
        start: nextMonday.toISO(),
        end: nextFriday.toISO(),
      })
    }
  }, [q.loading, zone])

  function validate(): Error | null {
    if (isISOAfter(value.start, value.end)) {
      return new Error('Start date/time cannot be after end date/time.')
    }
    return null
  }

  const hasInvalidShift = (() => {
    if (q.loading) return false
    const schedInterval = parseInterval(value, zone)
    return value.shifts.some(
      (s) =>
        DateTime.fromISO(s.end) > DateTime.fromISO(now) &&
        !schedInterval.engulfs(parseInterval(s, zone)),
    )
  })()

  const shiftErrors = hasInvalidShift
    ? [
        {
          message:
            'One or more shifts extend beyond the start and/or end of this temporary schedule',
        },
      ]
    : []

  function handleCoverageGapClick(coverageGap: Interval): void {
    if (!showForm) setShowForm(true)

    // make sure duration remains the same (evaluated off of the end timestamp)
    const startDT = DateTime.fromISO(shift?.start ?? '', { zone })
    const endDT = DateTime.fromISO(shift?.end ?? '', { zone })
    const duration = dtToDuration(startDT, endDT)
    const nextStart = coverageGap?.start
    const nextEnd = nextStart.plus({ hours: duration })

    setShift({
      userID: shift?.userID ?? '',
      start: nextStart.toISO(),
      end: nextEnd.toISO(),
    })
  }

  const hasCoverageGaps = (() => {
    if (q.loading) return false
    const schedInterval = parseInterval(value, zone)
    return (
      getCoverageGapItems(
        schedInterval,
        value.shifts,
        zone,
        handleCoverageGapClick,
      ).length > 0
    )
  })()

  const [submit, { loading, error }] = useMutation(mutation, {
    onCompleted: () => onClose(),
    variables: {
      input: {
        start: value.start,
        end: value.end,
        clearStart: value.clearStart,
        clearEnd: value.clearEnd,
        shifts: value.shifts.filter((s) => {
          // clamp/filter out shifts that are in the past
          if (DateTime.fromISO(s.end) <= DateTime.fromISO(now)) {
            return false
          }

          s.start = clampForward(now, s.start)
          return true
        }),
        scheduleID,
      },
    },
  })

  const shiftListRef = useRef<HTMLDivElement | null>(null)

  const handleSubmit = (): void => {
    setHasSubmitted(true)

    if (hasCoverageGaps && !allowNoCoverage) {
      // Scroll to show gap in coverage error on top of shift list
      if (shiftListRef?.current) {
        shiftListRef.current.scrollIntoView({ behavior: 'smooth' })
      }
      return
    }

    submit()
  }

  const nonFieldErrs = nonFieldErrors(error).map((e) => ({
    message: e.message,
  }))
  const fieldErrs = fieldErrors(error).map((e) => ({
    message: `${e.field}: ${e.message}`,
  }))
  const noCoverageErrs =
    hasSubmitted && hasCoverageGaps && !allowNoCoverage
      ? [new Error('This temporary schedule has gaps in coverage.')]
      : []
  const errs = nonFieldErrs
    .concat(fieldErrs)
    .concat(shiftErrors)
    .concat(noCoverageErrs)

  return (
    <FormDialog
      fullHeight
      maxWidth='lg'
      title='Define a Temporary Schedule'
      onClose={onClose}
      loading={loading}
      errors={errs}
      notices={
        !value.start ||
        DateTime.fromISO(value.start, { zone }) >
          DateTime.utc().minus({ hour: 1 }) ||
        edit
          ? []
          : [
              {
                type: 'WARNING',
                message: 'Start time occurs in the past',
                details:
                  'Any shifts or changes made to shifts in the past will be ignored when submitting.',
              },
            ]
      }
      form={
        <FormContainer
          optionalLabels
          disabled={loading}
          value={value}
          onChange={(newValue: Value) => setValue({ ...value, ...newValue })}
        >
          <Grid
            container
            className={classes.formContainer}
            justifyContent='space-between'
          >
            {/* left pane */}
            <Grid
              item
              xs={12}
              md={6}
              container
              alignContent='flex-start'
              spacing={2}
            >
              <Grid item xs={12}>
                <DialogContentText className={classes.contentText}>
                  The schedule will be exactly as configured here for the entire
                  duration (ignoring all assignments and overrides).
                </DialogContentText>
              </Grid>

              <Grid item xs={12}>
                <Typography color='textSecondary' className={classes.tzNote}>
                  Configuring in {zone}
                </Typography>
              </Grid>

              <Grid item xs={12} md={6}>
                <FormField
                  fullWidth
                  component={ISODateTimePicker}
                  required
                  name='start'
                  label='Schedule Start'
                  min={now}
                  max={DateTime.fromISO(now, { zone })
                    .plus({ year: 1 })
                    .toISO()}
                  validate={() => validate()}
                  timeZone={zone}
                  disabled={q.loading}
                  hint={isLocalZone ? '' : fmtLocal(value.start)}
                />
              </Grid>
              <Grid item xs={12} md={6}>
                <FormField
                  fullWidth
                  component={ISODateTimePicker}
                  required
                  name='end'
                  label='Schedule End'
                  min={value.start}
                  max={DateTime.fromISO(value.start, { zone })
                    .plus({ month: 3 })
                    .toISO()}
                  validate={() => validate()}
                  timeZone={zone}
                  disabled={q.loading}
                  hint={isLocalZone ? '' : fmtLocal(value.end)}
                />
              </Grid>

              <Grid item xs={12} className={classes.sticky}>
                <TempSchedAddNewShift
                  value={value}
                  onChange={(shifts: Shift[]) => setValue({ ...value, shifts })}
                  scheduleID={scheduleID}
                  edit={edit}
                  showForm={showForm}
                  setShowForm={setShowForm}
                  shift={shift}
                  setShift={setShift}
                />
              </Grid>
            </Grid>

            {/* right pane */}
            <Grid
              item
              xs={12}
              md={6}
              container
              spacing={2}
              className={classes.rightPane}
            >
              <Grid item xs={12} ref={shiftListRef}>
                <Typography variant='subtitle1' component='h3'>
                  Shifts
                </Typography>

                {hasSubmitted && hasCoverageGaps && (
                  <Alert severity='error' className={classes.noCoverageError}>
                    <AlertTitle>Gaps in coverage</AlertTitle>
                    <FormHelperText>
                      There are gaps in coverage. During these gaps, nobody on
                      the schedule will receive alerts. If you still want to
                      proceed, check the box below and retry.
                    </FormHelperText>
                    <FormControlLabel
                      label='Allow gaps in coverage'
                      labelPlacement='end'
                      control={
                        <Checkbox
                          data-cy='no-coverage-checkbox'
                          checked={allowNoCoverage}
                          onChange={(e) => setAllowNoCoverage(e.target.checked)}
                          name='allowCoverageGaps'
                        />
                      }
                    />
                  </Alert>
                )}

                <TempSchedShiftsList
                  scheduleID={scheduleID}
                  value={value.shifts}
                  start={value.start}
                  end={value.end}
                  onRemove={(shift: Shift) => {
                    setValue({
                      ...value,
                      shifts: value.shifts.filter(
                        (s) => !shiftEquals(shift, s),
                      ),
                    })
                  }}
                  edit={edit}
                  handleCoverageGapClick={handleCoverageGapClick}
                />
              </Grid>
            </Grid>
          </Grid>
        </FormContainer>
      }
      onSubmit={handleSubmit}
    />
  )
}
