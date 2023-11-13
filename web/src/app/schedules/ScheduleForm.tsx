import React from 'react'
import { FormContainer, FormField } from '../forms'
import { TextField, Grid } from '@mui/material'
import { TimeZoneSelect } from '../selection'

export interface Value {
  name: string
  description: string
  timeZone: string
  favorite?: boolean
}

interface ScheduleFormProps {
  value: Value
  onChange: (value: Value) => void

  // These can be removed when we convert FormContainer.js to typescript.
  errors?: Error[]
  disabled?: boolean
}

interface Error {
  message: string
  field: string
  helpLink?: string
}

export default function ScheduleForm(
  props: ScheduleFormProps,
): React.ReactNode {
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
