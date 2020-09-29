import { Button, Grid, TextField } from '@material-ui/core'
import { DateTime } from 'luxon'
import React, { useState } from 'react'
import { FormField } from '../../forms'
import { UserSelect } from '../../selection'
import { ISODateTimePicker } from '../../util/ISOPickers'

export default function FixedSchedAddShiftForm() {
  const [manualEntry, setManualEntry] = useState(false)

  return (
    <React.Fragment>
      <Grid item>
        <FormField
          fullWidth
          component={ISODateTimePicker}
          label='Shift Start'
          name='start'
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
            mapValue={(fieldValue, formValue) => {
              if (!formValue) return 0
              return DateTime.fromISO(fieldValue).diff(
                DateTime.fromISO(formValue.start),
                'hours',
              ).hours
            }}
            mapOnChangeValue={(newFieldValue, formValue) => {
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
      <Grid item>
        <FormField
          fullWidth
          component={UserSelect}
          label='Select a User'
          name='userID'
        />
      </Grid>
    </React.Fragment>
  )
}
