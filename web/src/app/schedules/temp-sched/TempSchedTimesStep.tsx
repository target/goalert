import React, { useState } from 'react'
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
import { isISOAfter } from '../../util/shifts'
import { DateTime } from 'luxon'

const useStyles = makeStyles({
  contentText,
})

type TempSchedTimesStepProps = {
  scheduleID: string
  stepText: string
  value: Value
  edit?: boolean
}

export default function TempSchedTimesStep({
  scheduleID,
  stepText,
  value,
  edit,
}: TempSchedTimesStepProps): JSX.Element {
  const classes = useStyles()
  const [now] = useState(DateTime.utc().startOf('minute').toISO())

  function validate(): Error | null {
    if (isISOAfter(value.start, value.end)) return new Error('Start date/time cannot be after end date/time.')
    return null
  }

  return (
    <StepContainer width='35%' data-cy='sched-times-step'>
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
            min={edit ? value.start : now}
            validate={() => validate()}
          />
        </Grid>
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            required
            name='end'
            min={edit ? value.start : now}
            validate={() => validate()}
          />
        </Grid>
      </Grid>
    </StepContainer>
  )
}
