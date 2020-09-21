import React from 'react'
import {
  Grid,
  DialogContentText,
  Typography,
  makeStyles,
} from '@material-ui/core'
import { DateTime } from 'luxon'
import ScheduleCalendar from '../ScheduleCalendar'
import { Value, contentText, StepContainer } from './sharedUtils'

interface ReviewStepProps {
  scheduleID: string
  value: Value
}

const useStyles = makeStyles({
  contentText,
  calendarContainer: {
    height: 'fit-content',
    paddingTop: 16,
    // matches 75% width in rest of app
    marginRight: '12.5%',
    marginLeft: '12.5%',
    marginBottom: 5, // room for bottom drop-shadow from card
  },
})

export default function ReviewStep({ scheduleID, value }: ReviewStepProps) {
  const { start, end, shifts: _shifts } = value
  const classes = useStyles()

  const fmt = (t: string) =>
    DateTime.fromISO(t).toLocaleString(DateTime.DATETIME_MED)

  // map user label/values to name/ids
  const shifts = _shifts.map((s) => ({
    ...s,
    user: {
      id: s.user.value,
      name: s.user.label,
    },
  }))

  return (
    <React.Fragment>
      <StepContainer>
        <Grid container spacing={2}>
          <Grid item xs={12}>
            <Typography variant='h6' component='h2'>
              Review your fixed schedule.
            </Typography>
          </Grid>
          <Grid item xs={12}>
            <DialogContentText className={classes.contentText}>
              This schedule will be fixed to the following schedule from{' '}
              {fmt(start)} to {fmt(end)}.
            </DialogContentText>
          </Grid>
        </Grid>
      </StepContainer>
      <div className={classes.calendarContainer}>
        <ScheduleCalendar
          scheduleID={scheduleID}
          shifts={shifts}
          readOnly
          CardProps={{ elevation: 3 }}
        />
      </div>
    </React.Fragment>
  )
}
