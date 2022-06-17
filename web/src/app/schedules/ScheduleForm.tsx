import React from 'react'
import { FormContainer, FormField } from '../forms'
import { TextField, Grid } from '@mui/material'
import { TimeZoneSelect } from '../selection'

export interface Value {
  name: string
  description: string
  timeZone: string
}

interface ScheduleFormProps {
  value: Value
  onChange: (value: Value) => void
}

export default function ScheduleForm(props: ScheduleFormProps): JSX.Element {
  return (
    <FormContainer optionalLabels {...props}>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            name='name'
            label='Name'
            required
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            multiline
            name='description'
            label='Description'
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TimeZoneSelect}
            name='time-zone'
            fieldName='timeZone'
            label='Time Zone'
            required
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
