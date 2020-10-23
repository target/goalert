import React, { useState } from 'react'
import { Grid, TextField } from '@material-ui/core'
import { DateTime } from 'luxon'
import { useURLParam } from '../../actions'
import { FormField } from '../../forms'
import { UserSelect } from '../../selection'
import ClickableText from '../../util/ClickableText'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { Value } from './sharedUtils'

export default function TempSchedAddShiftForm(): JSX.Element {
  const [manualEntry, setManualEntry] = useState(false)
  const [zone] = useURLParam('tz', 'local')
  const [now] = useState(DateTime.local()
  .setZone(zone)
  .startOf('minute')
  .toISO())

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
          min={now}
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
            min={now}
            hint={
              <ClickableText
                text='Configure as duration'
                onClick={() => setManualEntry(false)}
              />
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
            mapValue={(nextVal: string, formValue: Value) => {
              const nextValDT = DateTime.fromISO(nextVal)
              if (!formValue || !nextValDT.isValid) return ''
              return nextValDT
                .diff(DateTime.fromISO(formValue.start), 'hours')
                .hours.toString()
            }}
            // value held in state
            mapOnChangeValue={(nextVal: string, formValue: Value) => {
              if (!nextVal) return ''
              return DateTime.fromISO(formValue.start)
                .plus({ hours: parseInt(nextVal, 10) })
                .toISO()
            }}
            min={0.25}
            hint={
              <ClickableText
                text='Configure as date/time'
                onClick={() => setManualEntry(true)}
              />
            }
          />
        )}
      </Grid>
    </React.Fragment>
  )
}
