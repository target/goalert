import React from 'react'
import {
  Grid,
  DialogContentText,
  Typography,
  makeStyles,
} from '@material-ui/core'
import { FormField } from '../../forms'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { contentText, StepContainer, Value } from './sharedUtils'
import { ScheduleTZFilter } from '../ScheduleTZFilter'

import { DateTime } from 'luxon'

const useStyles = makeStyles({
  contentText,
})

type TempSchedTimesStepProps = {
  scheduleID: string
  stepText: string
  value: Value
}

export default function TempSchedTimesStep({
  scheduleID,
  stepText,
  value,
}: TempSchedTimesStepProps): JSX.Element {
  const classes = useStyles()

  function isValid(): Error | null {
    if (DateTime.fromISO(value.start) < DateTime.fromISO(value.end)) return null
    return new Error('Start date/time cannot be after end date/time.')
  }

  return (
    <StepContainer width='35%'>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Typography variant='body2'>{stepText}</Typography>
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
        <Grid item xs={12}>
          <ScheduleTZFilter
            label={(tz) => `Configure in ${tz}`}
            scheduleID={scheduleID}
          />
        </Grid>
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            required
            name='start'
            validate={() => isValid()}
          />
        </Grid>
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            required
            name='end'
            validate={() => isValid()}
          />
        </Grid>
      </Grid>
    </StepContainer>
  )
}
