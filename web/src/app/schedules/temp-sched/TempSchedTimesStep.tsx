import React from 'react'
import { Grid, DialogContentText, Typography } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import { FormField } from '../../forms'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { contentText, StepContainer, Value } from './sharedUtils'
import { ScheduleTZFilter } from '../ScheduleTZFilter'
import { isISOBefore } from '../../util/shifts'
import { DateTime } from 'luxon'
import { useURLParam } from '../../actions'

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
  const [zone] = useURLParam('tz', 'local')
  const now = DateTime.local()
    .setZone(zone)
    .startOf('day')
    .toFormat("yyyy-MM-dd'T'HH:mm:ss")

  function validate(): Error | null {
    if (isISOBefore(value.start, value.end)) return null
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
            During a temporary schedule, all on-call shifts will be set as
            configured on the next step. A temporary schedule ignores all rules,
            rotations, and overrides. On-call will be exactly as configured here
            for the entire duration.
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
            validate={() => validate()}
            min={now}
          />
        </Grid>
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            required
            name='end'
            validate={() => validate()}
            min={now}
          />
        </Grid>
      </Grid>
    </StepContainer>
  )
}
