import { Grid, TextField, Typography, makeStyles } from '@material-ui/core'
import { DateTime } from 'luxon'
import React, { useState } from 'react'
import { FormField } from '../../forms'
import { UserSelect } from '../../selection'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { Value } from './sharedUtils'

const useStyles = makeStyles({
  typography: {
    '&:hover': {
      cursor: 'pointer',
      textDecoration: 'underline',
    },
  },
})

export default function FixedSchedAddShiftForm(): JSX.Element {
  const classes = useStyles()
  const [manualEntry, setManualEntry] = useState(false)

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
