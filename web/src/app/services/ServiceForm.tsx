import React from 'react'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import { EscalationPolicySelect } from '../selection/EscalationPolicySelect'
import { FormContainer, FormField } from '../forms'

interface Value {
  name: string
  description: string
  escalationPolicyID: string
}

interface ServiceFormProps {
  value: Value

  errors: {
    field: 'name' | 'description' | 'escalationPolicyID'
    message: string
  }[]

  onChange: (val: Value) => void

  disabled: boolean

  epRequired: boolean
}

export default function ServiceForm(props: ServiceFormProps): JSX.Element {
  const { epRequired, ...containerProps } = props
  return (
    <FormContainer {...containerProps} optionalLabels={epRequired}>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            fullWidth
            label='Name'
            name='name'
            required
            component={TextField}
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            label='Description'
            name='description'
            multiline
            component={TextField}
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            label='Escalation Policy'
            name='escalation-policy'
            fieldName='escalationPolicyID'
            required={epRequired}
            component={EscalationPolicySelect}
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
