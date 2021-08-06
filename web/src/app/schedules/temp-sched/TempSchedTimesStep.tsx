import React, { useState } from 'react'
import {
  Grid,
  DialogContentText,
  Typography,
  makeStyles,
} from '@material-ui/core'
import { FormField } from '../../forms'
import { ISOKeyboardDateTimePicker } from '../../util/ISOPickers'
import { contentText, StepContainer, Value } from './sharedUtils'
import { isISOAfter } from '../../util/shifts'
import { DateTime } from 'luxon'
import { gql, useQuery } from '@apollo/client'

const useStyles = makeStyles({
  contentText,
})

const schedTZQuery = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
    }
  }
`

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
  const { data, error, loading } = useQuery(schedTZQuery, {
    variables: { id: scheduleID },
  })
  const zone = data?.schedule?.timeZone
  console.log(zone === undefined)

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
          <Typography variant='body2'>{stepText}</Typography>
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
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={ISOKeyboardDateTimePicker}
            required
            name='start'
            min={edit ? value.start : now}
            validate={() => validate()}
            timeZone={zone}
          />
        </Grid>
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={ISOKeyboardDateTimePicker}
            required
            name='end'
            min={edit ? value.start : now}
            validate={() => validate()}
            timeZone={zone}
          />
        </Grid>
      </Grid>
    </StepContainer>
  )
}
