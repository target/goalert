import { Button, Grid, TextField } from '@material-ui/core'
import { DateTime } from 'luxon'
import React, { useState } from 'react'
import { FormField } from '../../forms'
import { UserSelect } from '../../selection'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { Value } from './sharedUtils'

type FixedSchedAddShiftFormProps = {
  setEndTime: (end: string) => void
}

export default function FixedSchedAddShiftForm({
  setEndTime,
}: FixedSchedAddShiftFormProps) {
  const [manualEntry, setManualEntry] = useState(false)

  return (
    <React.Fragment>
      <Grid item>
        <FormField
          fullWidth
          component={UserSelect}
          label='Select a User'
          name='userID'
          required
        />
      </Grid>
      <Grid item>
        <FormField
          fullWidth
          component={ISODateTimePicker}
          label='Shift Start'
          name='start'
          required
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
            required
            hint={
              <React.Fragment>
                <Button onClick={() => setManualEntry(false)}>
                  Click here
                </Button>{' '}
                to switch to shift duration.
              </React.Fragment>
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
            mapValue={(fieldValue: string, formValue: Value) => {
              if (!formValue) return '0'
              return DateTime.fromISO(fieldValue).diff(
                DateTime.fromISO(formValue.start),
                'hours',
              ).hours
            }}
            mapOnChangeValue={(newFieldValue: number, formValue: Value) => {
              return DateTime.fromISO(formValue.start)
                .plus({ hours: newFieldValue })
                .toISO()
            }}
            min={0.25}
            hint={
              <React.Fragment>
                <Button onClick={() => setManualEntry(true)}>Click here</Button>{' '}
                to enter an exact date-time.
              </React.Fragment>
            }
          />
        )}
      </Grid>
    </React.Fragment>
  )
}
