import React, { useState } from 'react'
import { Grid, Typography, makeStyles } from '@material-ui/core'
import { DateTime } from 'luxon'
import { FormField } from '../../forms'
import { UserSelect } from '../../selection'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { Value } from './sharedUtils'
import NumberField from '../../util/NumberField'

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
            component={NumberField}
            label='Shift Duration (hours)'
            name='end'
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
                .plus({ hours: parseFloat(nextVal) })
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
