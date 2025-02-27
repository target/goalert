import React from 'react'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import { FormContainer, FormField } from '../forms'
import MaterialSelect from '../selection/MaterialSelect'
import { FieldError } from '../util/errutil'

export interface PolicyFormValue {
  name: string
  description: string
  repeat: {
    label: string
    value: string
  }
}

interface PolicyFormProps {
  value: PolicyFormValue
  errors?: Array<FieldError>
  disabled?: boolean
  onChange?: (value: PolicyFormValue) => void
}

function PolicyForm(props: PolicyFormProps): React.JSX.Element {
  return (
    <FormContainer optionalLabels {...props}>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            component={TextField}
            disabled={props.disabled}
            fieldName='name'
            fullWidth
            label='Name'
            name='name'
            required
            value={props.value.name}
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            component={TextField}
            disabled={props.disabled}
            fieldName='description'
            fullWidth
            label='Description'
            multiline
            name='description'
            value={props.value.description}
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            component={MaterialSelect}
            disabled={props.disabled}
            fieldName='repeat'
            fullWidth
            hint='The amount of times it will escalate through all steps'
            label='Repeat Count'
            name='repeat'
            options={[
              { label: '0', value: '0' },
              { label: '1', value: '1' },
              { label: '2', value: '2' },
              { label: '3', value: '3' },
              { label: '4', value: '4' },
              { label: '5', value: '5' },
            ]}
            required
            value={props.value.repeat.value}
            min={0}
            max={5}
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}

export default PolicyForm
