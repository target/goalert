import React, { useEffect, useState } from 'react'
import { Grid, TextField, Typography, makeStyles } from '@material-ui/core'
import { DateTime, Interval } from 'luxon'
import { round } from 'lodash-es'
import { FormField } from '../../forms'
import { UserSelect } from '../../selection'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { Shift, Value } from './sharedUtils'

const useStyles = makeStyles({
  typography: {
    '&:hover': {
      cursor: 'pointer',
      textDecoration: 'underline',
    },
  },
})

function durToEnd(start: string, dur: string): string {
  return DateTime.fromISO(start)
    .plus({ hours: parseInt(dur, 10) })
    .toISO()
}

function endToDur(s: string, e: string): number {
  const start = DateTime.fromISO(s)
  const end = DateTime.fromISO(e)
  return Interval.fromDateTimes(start, end).toDuration('hours').hours
}

type FixedSchedAddShiftFormProps = {
  shift: Shift
}

export default function FixedSchedAddShiftForm({
  shift,
}: FixedSchedAddShiftFormProps): JSX.Element {
  const classes = useStyles()
  const [manualEntry, setManualEntry] = useState(false)
  const [duration, setDuration] = useState<number>()

  // update duration when start/end fields change
  useEffect(() => {
    if (shift?.start && shift?.end) {
      setDuration(round(endToDur(shift.start, shift.end), 2))
    }
  }, [shift?.start, shift?.end])

  return (
    <React.Fragment>
      <Grid item>
        <FormField
          fullWidth
          component={UserSelect}
          label='Select a User'
          name='userID'
        />
      </Grid>
      <Grid item>
        <FormField
          fullWidth
          component={ISODateTimePicker}
          label='Shift Start'
          name='start'
          mapOnChangeValue={(value: string, formValue: Value) => {
            if (!manualEntry) {
              const diff = DateTime.fromISO(value).diff(
                DateTime.fromISO(formValue.start),
              )
              formValue.end = DateTime.fromISO(formValue.end).plus(diff).toISO()
            }
            return value
          }}
        />
      </Grid>
      <Grid item>
        {manualEntry ? (
          <FormField
            fullWidth
            component={ISODateTimePicker}
            label='Shift End'
            name='end'
            hint={
              <Typography
                className={classes.typography}
                variant='caption'
                color='textSecondary'
                onClick={() => setManualEntry(false)}
              >
                Configure as duration?
              </Typography>
            }
          />
        ) : (
          <FormField
            fullWidth
            component={TextField}
            label='Shift Duration (hours)'
            name='end'
            type='number'
            float
            // value held in form input
            mapValue={() => duration?.toString() ?? ''}
            // value held in state
            mapOnChangeValue={(nextVal: number, formValue: Value) => {
              setDuration(round(nextVal, 2))
              if (isNaN(nextVal)) return ''
              return DateTime.fromISO(formValue.start)
                .plus({ hours: nextVal })
                .toISO()
            }}
            min={0.25}
            hint={
              <Typography
                className={classes.typography}
                variant='caption'
                color='textSecondary'
                onClick={() => setManualEntry(true)}
              >
                Configure as date/time?
              </Typography>
            }
          />
        )}
      </Grid>
    </React.Fragment>
  )
}
