import React, { useMemo } from 'react'
import {
  Grid,
  DialogContentText,
  Typography,
  TextField,
  MenuItem,
  Divider,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { DateTime, Duration } from 'luxon'
import { FormField, FormContainer } from '../../forms'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { fmtLocal } from '../../util/timeFormat'
import { contentText, Shift, TempSchedValue } from './sharedUtils'
import TempSchedAddNewShift from './TempSchedAddNewShift'
import { useScheduleTZ } from '../useScheduleTZ'
import { DurationValues } from './TempSchedDialog'
import { isISOAfter } from '../../util/shifts'

const useStyles = makeStyles(() => ({
  contentText,
  sticky: {
    position: 'sticky',
    top: 0,
  },
  tzNote: {
    fontStyle: 'italic',
  },
}))

interface TempSchedFormProps {
  scheduleID: string
  duration: DurationValues
  setDuration: React.Dispatch<React.SetStateAction<DurationValues>>
  value: TempSchedValue
  setValue: React.Dispatch<React.SetStateAction<TempSchedValue>>
  showForm: boolean
  setShowForm: React.Dispatch<React.SetStateAction<boolean>>
  shift: Shift
  setShift: React.Dispatch<React.SetStateAction<Shift>>
}

export default function TempSchedForm(props: TempSchedFormProps): JSX.Element {
  const {
    scheduleID,
    duration,
    setDuration,
    value,
    setValue,
    showForm,
    setShowForm,
    shift,
    setShift,
  } = props

  const classes = useStyles()
  const now = useMemo(() => DateTime.utc().startOf('minute').toISO(), [])
  const { q, zone, isLocalZone } = useScheduleTZ(scheduleID)

  function validate(): Error | null {
    if (isISOAfter(value.start, value.end)) {
      return new Error('Start date/time cannot be after end date/time.')
    }
    return null
  }

  return (
    <Grid item xs={12} md={6} container alignContent='flex-start' spacing={2}>
      <Grid item xs={12}>
        <DialogContentText className={classes.contentText}>
          The schedule will be exactly as configured here for the entire
          duration (ignoring all assignments and overrides).
        </DialogContentText>
      </Grid>

      <Grid item xs={12}>
        <Typography color='textSecondary' className={classes.tzNote}>
          Times shown in schedule timezone ({zone})
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
          max={DateTime.fromISO(now, { zone }).plus({ year: 1 }).toISO()}
          softMax={value.end}
          softMaxLabel='Must be before end time.'
          softMin={DateTime.fromISO(value.end).plus({ month: -3 }).toISO()}
          softMinLabel='Must be within 3 months of end time.'
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
          min={now}
          softMin={value.start}
          softMinLabel='Must be after start time.'
          softMax={DateTime.fromISO(value.start).plus({ month: 3 }).toISO()}
          softMaxLabel='Must be within 3 months of start time.'
          validate={() => validate()}
          timeZone={zone}
          disabled={q.loading}
          hint={isLocalZone ? '' : fmtLocal(value.end)}
        />
      </Grid>

      <FormContainer
        value={duration}
        onChange={(newValue: DurationValues) => {
          setDuration({ ...duration, ...newValue })
          setValue({
            ...value,
            shiftDur: Duration.fromObject({
              [newValue.ivl]: newValue.dur,
            }),
          })
        }}
      >
        <Grid item xs={12} md={6}>
          <FormField
            fullWidth
            component={TextField}
            type='number'
            name='dur'
            min={1}
            label='Shift Duration'
            validate={() => validate()}
            disabled={q.loading}
          />
        </Grid>
        <Grid item xs={12} md={6}>
          <FormField
            fullWidth
            component={TextField}
            name='ivl'
            select
            label='Shift Interval'
            validate={() => validate()}
            disabled={q.loading}
          >
            <MenuItem value='hours'>Hour</MenuItem>
            <MenuItem value='days'>Day</MenuItem>
            <MenuItem value='weeks'>Week</MenuItem>
          </FormField>
        </Grid>
      </FormContainer>

      <Grid item xs={12}>
        <Divider />
      </Grid>

      <Grid item xs={12} className={classes.sticky}>
        <TempSchedAddNewShift
          value={value}
          onChange={(shifts: Shift[]) => setValue({ ...value, shifts })}
          scheduleID={scheduleID}
          showForm={showForm}
          setShowForm={setShowForm}
          shift={shift}
          setShift={setShift}
        />
      </Grid>
    </Grid>
  )
}
