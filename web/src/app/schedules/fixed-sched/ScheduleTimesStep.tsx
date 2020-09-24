import React from 'react'
import {
  Grid,
  DialogContentText,
  Typography,
  makeStyles,
} from '@material-ui/core'
import { DateTime } from 'luxon'
import { FormField } from '../../forms'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { contentText, StepContainer, Value } from './sharedUtils'

const useStyles = makeStyles({
  contentText,
})

interface ScheduleTimesStepProps {
  value: Value
}

export default function ScheduleTimesStep({ value }: ScheduleTimesStepProps) {
  const classes = useStyles()

  // don't allow user to set start after end, or end before start
  const f = (d: string) => DateTime.fromISO(d).toFormat("yyyy-MM-dd'T'HH:mm:ss")
  let min = null
  let max = null
  if (value.start) min = f(value.start)
  if (value.end) max = f(value.end)

  return (
    <StepContainer>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Typography variant='body2'>STEP 1 OF 3</Typography>
          <Typography variant='h6' component='h2'>
            Determine start and end times.
          </Typography>
        </Grid>
        <Grid item xs={12}>
          <DialogContentText className={classes.contentText}>
            Selecting a start and end dates will define a span of time on this
            schedule with a fixed set of shifts. These shifts ignores all rules,
            rotations, and overrides and will behave exactly as configured here.
          </DialogContentText>
        </Grid>
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            required
            name='start'
            inputProps={{ max }}
          />
        </Grid>
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            required
            name='end'
            inputProps={{ min }}
          />
        </Grid>
      </Grid>
    </StepContainer>
  )
}
