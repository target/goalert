import React, { useState, useEffect } from 'react'
import { useMutation, gql } from '@apollo/client'
import Checkbox from '@mui/material/Checkbox'
import DialogContentText from '@mui/material/DialogContentText'
import FormControlLabel from '@mui/material/FormControlLabel'
import FormHelperText from '@mui/material/FormHelperText'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import makeStyles from '@mui/styles/makeStyles'
import Alert from '@mui/material/Alert'
import AlertTitle from '@mui/material/AlertTitle'
import _ from 'lodash'
import { DateTime } from 'luxon'

import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import { contentText, fmtLocal, Shift, Value } from './sharedUtils'
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
    padding: '0.5rem',
  },
  noCoverageError: {
    marginTop: '.5rem',
    marginBottom: '.5rem',
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

export default function TempSchedDialog({
  onClose,
  scheduleID,
  value: _value,
}: TempScheduleDialogProps): JSX.Element {
  const classes = useStyles()
  const edit = Boolean(_value)
  const { q, zone, isLocalZone } = useScheduleTZ(scheduleID)
  const [now] = useState(DateTime.utc().startOf('minute').toISO())
  const [value, setValue] = useState({
    start: _value?.start ?? '',
    end: _value?.end ?? '',
    shifts: (_value?.shifts ?? []).map((s) =>
      _.pick(s, 'start', 'end', 'userID'),
    ),
  })
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
      (s) => !schedInterval.engulfs(parseInterval(s, zone)),
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

  const hasCoverageGaps = (() => {
    if (q.loading) return false
    const schedInterval = parseInterval(value, zone)
    return getCoverageGapItems(schedInterval, value.shifts, zone).length > 0
  })()

  const [submit, { loading, error }] = useMutation(mutation, {
    onCompleted: () => onClose(),
    variables: {
      input: {
        ...value,
        scheduleID,
      },
    },
  })

  const handleSubmit = (): void => {
    setHasSubmitted(true)

    if (hasCoverageGaps && !allowNoCoverage) {
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
      maxWidth='lg'
      title='Define a Temporary Schedule'
      onClose={onClose}
      loading={loading}
      errors={errs}
      notices={
        !value.start ||
        DateTime.fromISO(value.start) > DateTime.utc().minus({ hour: 1 }) ||
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
          onChange={(newValue: Value) => setValue(newValue)}
        >
          <Grid container className={classes.formContainer}>
            <Grid
              item
              xs={12}
              md={6}
              container
              alignContent='flex-start'
              spacing={2}
              style={{ paddingRight: '1rem' }}
            >
              <Grid item xs={12}>
                <DialogContentText className={classes.contentText}>
                  The schedule will be exactly as configured here for the entire
                  duration (ignoring all assignments and overrides).
                </DialogContentText>
              </Grid>

              {!isLocalZone && (
                <Grid item xs={12}>
                  <Typography color='textSecondary' className={classes.tzNote}>
                    Configuring in {zone}
                  </Typography>
                </Grid>
              )}

              <Grid item xs={12} md={6}>
                <FormField
                  fullWidth
                  component={ISODateTimePicker}
                  required
                  name='start'
                  label='Schedule Start'
                  min={now}
                  validate={() => validate()}
                  timeZone={zone}
                  disabled={q.loading || edit}
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
                  validate={() => validate()}
                  timeZone={zone}
                  disabled={q.loading || edit}
                  hint={isLocalZone ? '' : fmtLocal(value.end)}
                />
              </Grid>

              <Grid item xs={12}>
                <TempSchedAddNewShift
                  value={value}
                  onChange={(shifts: Shift[]) => setValue({ ...value, shifts })}
                  scheduleID={scheduleID}
                  edit={edit}
                />
              </Grid>
            </Grid>

            {/* shifts list container */}
            <Grid item xs={12} md={6}>
              <Typography variant='subtitle1' component='h3'>
                Shifts
              </Typography>

              {/* coverage warning */}
              {hasSubmitted && hasCoverageGaps && (
                <Alert severity='error' className={classes.noCoverageError}>
                  <AlertTitle>Gaps in coverage</AlertTitle>
                  <FormHelperText>
                    There are gaps in coverage. During these gaps, nobody on the
                    schedule will receive alerts. If you still want to proceed,
                    check the box below and retry.
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
                    shifts: value.shifts.filter((s) => !shiftEquals(shift, s)),
                  })
                }}
                edit={edit}
              />
            </Grid>
          </Grid>
        </FormContainer>
      }
      onSubmit={handleSubmit}
    />
  )
}
