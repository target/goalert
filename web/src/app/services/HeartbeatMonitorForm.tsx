import React from 'react'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import { FormContainer, FormField } from '../forms'
import { FieldError } from '../util/errutil'
import { DurationField } from '../util/DurationField'
import { Duration } from 'luxon'

function clampTimeout(val: string): number | string {
  if (!val) return ''
  const dur = Duration.fromISO(val)
  return dur.as('minutes')
}
export interface Value {
  name: string
  timeoutMinutes: number
}
interface HeartbeatMonitorFormProps {
  value: Value

  errors: FieldError[]

  onChange: (val: Value) => void

  // can be deleted when FormContainer.js is converted to ts
  disabled: boolean
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
            component={DurationField}
            required
            label='Timeout'
            name='timeoutMinutes'
            min={5}
            max={540000}
            mapValue={(minutes) =>
              Duration.fromObject({
                minutes,
              }).toISO()
            }
            mapOnChangeValue={clampTimeout}
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
