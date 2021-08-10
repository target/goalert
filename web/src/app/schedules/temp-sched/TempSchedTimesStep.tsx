import React, { useState } from 'react'
import {
  Grid,
  DialogContentText,
  Typography,
  makeStyles,
} from '@material-ui/core'
import { FormField } from '../../forms'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { contentText, fmtLocal, StepContainer, Value } from './sharedUtils'
import { isISOAfter } from '../../util/shifts'
import { DateTime } from 'luxon'
import { useScheduleTZ } from './hooks'

const useStyles = makeStyles({
  contentText,
  tzNote: {
    fontStyle: 'italic',
  },
})

type TempSchedTimesStepProps = {
  scheduleID: string
  value: Value
  edit?: boolean
}

export default function TempSchedTimesStep({
  scheduleID,
  value,
  edit,
}: TempSchedTimesStepProps): JSX.Element {
  const classes = useStyles()
  const { q, zone, isLocalZone } = useScheduleTZ(scheduleID)
  const [now] = useState(DateTime.utc().startOf('minute').toISO())

  function validate(): Error | null {
    if (isISOAfter(value.start, value.end)) {
      return new Error('Start date/time cannot be after end date/time.')
    }
    return null
  }

  return (
    <StepContainer width='35%' data-cy='sched-times-step'>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Typography variant='body2'>STEP 1 OF 2</Typography>
          <Typography variant='h6' component='h2'>
            Determine start and end times.
          </Typography>
        </Grid>
        <Grid item xs={12}>
          <DialogContentText className={classes.contentText}>
            The schedule will be exactly as configured on the next step for the
            entire duration (ignoring all rules/overrides).
          </DialogContentText>
        </Grid>
        {!isLocalZone && (
          <Grid item xs={12}>
            <Typography color='textSecondary' className={classes.tzNote}>
              Configuring in {zone}
            </Typography>
          </Grid>
        )}
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            required
            label='Start'
            name='start'
            min={edit ? value.start : now}
            validate={() => validate()}
            timeZone={zone}
            disabled={q.loading}
            hint={isLocalZone ? '' : fmtLocal(value.start)}
          />
        </Grid>
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            required
            label='End'
            name='end'
            min={edit ? value.start : now}
            validate={() => validate()}
            timeZone={zone}
            disabled={q.loading}
            hint={isLocalZone ? '' : fmtLocal(value.end)}
          />
        </Grid>
      </Grid>
    </StepContainer>
  )
}
