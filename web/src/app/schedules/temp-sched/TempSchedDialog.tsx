import React, { useState, useEffect } from 'react'
import { useMutation, gql } from '@apollo/client'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import { contentText, fmtLocal, Shift, Value } from './sharedUtils'
import _ from 'lodash'
import { FormContainer, FormField } from '../../forms'
import TempSchedAddNewShift from './TempSchedAddNewShift'
import { isISOAfter, parseInterval } from '../../util/shifts'
import { DateTime } from 'luxon'
import { getNextWeekday } from '../../util/luxon-helpers'
import { useScheduleTZ } from './hooks'
import {
  DialogContentText,
  Grid,
  makeStyles,
  Typography,
} from '@material-ui/core'
import TempSchedShiftsList from './TempSchedShiftsList'
import { ISODateTimePicker } from '../../util/ISOPickers'

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
  mainContainer: {
    height: '100%',
    padding: '0.5rem',
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

  useEffect(() => {
    // set default start, end times when zone is ready
    if (!value.start && !value.end && !q.loading && zone) {
      const nextMonday = getNextWeekday(1, DateTime.now(), zone)
      const nextFriday = nextMonday.plus({ days: 5 })
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

  const schedInterval = parseInterval(value)
  const hasInvalidShift = value.shifts.some(
    (s) => !schedInterval.engulfs(parseInterval(s)),
  )

  const shiftErrors = hasInvalidShift
    ? [
        {
          message:
            'One or more shifts extend beyond the start and/or end of this temporary schedule',
        },
      ]
    : []

  const [submit, { loading, error }] = useMutation(mutation, {
    onCompleted: () => onClose(),
    variables: {
      input: {
        ...value,
        scheduleID,
      },
    },
  })

  const nonFieldErrs = nonFieldErrors(error).map((e) => ({
    message: e.message,
  }))
  const fieldErrs = fieldErrors(error).map((e) => ({
    message: `${e.field}: ${e.message}`,
  }))
  const errs = nonFieldErrs.concat(fieldErrs).concat(shiftErrors)

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
          <Grid container className={classes.mainContainer}>
            <Grid
              item
              xs={6}
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

              <Grid item xs={6}>
                <FormField
                  fullWidth
                  component={ISODateTimePicker}
                  required
                  name='start'
                  min={edit ? value.start : now}
                  validate={() => validate()}
                  timeZone={zone}
                  disabled={q.loading}
                  hint={isLocalZone ? '' : fmtLocal(value.start)}
                />
              </Grid>
              <Grid item xs={6}>
                <FormField
                  fullWidth
                  component={ISODateTimePicker}
                  required
                  name='end'
                  min={edit ? value.start : now}
                  validate={() => validate()}
                  timeZone={zone}
                  disabled={q.loading}
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
            <Grid item xs={6}>
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
      onSubmit={() => submit()}
    />
  )
}
