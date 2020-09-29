import React from 'react'
import {
  Grid,
  DialogContentText,
  Typography,
  makeStyles,
} from '@material-ui/core'
import ScheduleCalendar from '../ScheduleCalendar'
import { fmt, Value, contentText, StepContainer } from './sharedUtils'
import { useUserInfo } from '../../util/useUserInfo'

type ReviewStepProps = {
  value: Value
  stepText: string
}

const useStyles = makeStyles({
  contentText,
  calendarContainer: {
    height: 'fit-content',
    marginBottom: 5, // room for bottom drop-shadow from card
  },
})

export default function ReviewStep({ stepText, value }: ReviewStepProps) {
  const { start, end, shifts: _shifts } = value
  const classes = useStyles()

  const shifts = useUserInfo(_shifts)

  return (
    <StepContainer width='75%'>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Typography variant='body2'>{stepText}</Typography>
          <Typography variant='h6' component='h2'>
            Review your fixed schedule.
          </Typography>
        </Grid>
        <Grid item xs={12}>
          <DialogContentText className={classes.contentText}>
            This schedule will be fixed to the following shifts from{' '}
            {fmt(start)} to {fmt(end)}.
          </DialogContentText>
        </Grid>
        <Grid className={classes.calendarContainer} item xs={12}>
          <ScheduleCalendar
            shifts={[]}
            fixedShifts={shifts}
            readOnly
            CardProps={{ elevation: 3 }}
          />
        </Grid>
      </Grid>
    </StepContainer>
  )
}
