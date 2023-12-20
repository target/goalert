import React from 'react'
import { FormContainer, FormField } from '../forms'
import Grid from '@mui/material/Grid'

import NumberField from '../util/NumberField'
import { DestinationInput } from '../../schema'
import DestinationInputChip from '../util/DestinationInputChip'
import { Typography } from '@mui/material'

export type FormValue = {
  delayMinutes: number
  actions: DestinationInput[]
}

export type PolicyStepFormProps = {
  value: FormValue
  errors?: Array<{ field: 'targets' | 'delayMinutes'; message: string }>
  disabled?: boolean
  onChange?: (value: FormValue) => void
}

export default function PolicyStepForm(
  props: PolicyStepFormProps,
): React.ReactNode {
  function handleDelete(a: DestinationInput): void {
    if (!props.onChange) return

    props.onChange({
      ...props.value,
      actions: props.value.actions.filter((b) => a !== b),
    })
  }

  return (
    <FormContainer
      value={props.value}
      onChange={(newValue: FormValue) => {
        if (!props.onChange) return

        props.onChange(newValue)
      }}
    >
      <Grid container spacing={2}>
        <Grid item xs={12}>
          {props.value.actions.map((a, idx) => (
            <DestinationInputChip
              key={idx}
              value={a}
              onDelete={props.disabled ? undefined : () => handleDelete(a)}
            />
          ))}
          {props.value.actions.length === 0 && (
            <Typography variant='body2' color='textSecondary'>
              No actions
            </Typography>
          )}
        </Grid>
        <Grid item xs={12}>
          <FormField
            component={NumberField}
            disabled={props.disabled}
            fullWidth
            label='Delay (minutes)'
            name='delayMinutes'
            required
            min={1}
            max={9000}
            hint={
              props.value.delayMinutes === 0
                ? 'This will cause the step to immediately escalate'
                : `This will cause the step to escalate after ${props.value.delayMinutes}m`
            }
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
