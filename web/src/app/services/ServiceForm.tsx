import React from 'react'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import { EscalationPolicySelect } from '../selection/EscalationPolicySelect'
import { FormContainer, FormField } from '../forms'
import { FieldError } from '../util/errutil'

const MaxDetailsLength = 6 * 1024 // 6KiB

export interface Value {
  name: string
  description: string
  escalationPolicyID?: string
}

interface ServiceFormProps {
  value: Value

  errors: FieldError[]

  onChange: (val: Value) => void

  disabled?: boolean

  epRequired?: boolean
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
            rows={4}
            component={TextField}
            charCount={MaxDetailsLength}
            hint='Markdown Supported'
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
