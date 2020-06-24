import React from 'react'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import { FormContainer, FormField } from '../forms'

function clampTimeout(val: string): number | string {
  if (!val) return ''
  const num = parseInt(val, 10)
  if (Number.isNaN(num)) return val

  // need to have the min be 1 here so you can type `10`
  return Math.min(Math.max(1, num), 9000)
}
interface HeartbeatMonitorFormProps {
  value: {
    name: string
    timeoutMinutes: [number, string]
  }

  errors: {
    field: ['name', 'timeoutMinutes']
    message: string
  }

  onChange: Function
}
export default function HeartbeatMonitorForm(
  props: HeartbeatMonitorFormProps,
): JSX.Element {
  const { ...formProps } = props
  return (
    <FormContainer {...formProps} optionalLabels>
      <Grid container spacing={2}>
        <Grid item style={{ flexGrow: 1 }} xs={12}>
          <FormField
            fullWidth
            component={TextField}
            label='Name'
            name='name'
            required
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            required
            type='number'
            label='Timeout (minutes)'
            name='timeoutMinutes'
            min={5}
            max={9000}
            mapOnChangeValue={clampTimeout}
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
