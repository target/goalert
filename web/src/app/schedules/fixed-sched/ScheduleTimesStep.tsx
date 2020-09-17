import React from 'react'
import {
  Grid,
  DialogContentText,
  Typography,
  makeStyles,
} from '@material-ui/core'
import { FormField } from '../../forms'
import { ISODateTimePicker } from '../../util/ISOPickers'

const useStyles = makeStyles({
  contentText: {
    marginBottom: 0,
  },
})

export default function ScheduleTimesStep() {
  const classes = useStyles()

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
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
        />
      </Grid>
      <Grid item xs={6}>
        <FormField
          fullWidth
          component={ISODateTimePicker}
          required
          name='end'
        />
      </Grid>
    </Grid>
  )
}
