import React from 'react'
import { Grid, TextField } from '@material-ui/core'
import { FormField } from '../../../forms'

export function CreateAlertInfo() {
  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <FormField
          fullWidth
          label='Alert Summary'
          name='summary'
          required
          component={TextField}
        />
      </Grid>
      <Grid item xs={12}>
        <FormField
          fullWidth
          multiline
          rows={7}
          variant='outlined'
          placeholder='Alert Details'
          name='details'
          component={TextField}
        />
      </Grid>
    </Grid>
  )
}
