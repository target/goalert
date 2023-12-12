import React from 'react'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import { EscalationPolicySelect } from '../selection/EscalationPolicySelect'
import { FormContainer, FormField } from '../forms'
import { FieldError } from '../util/errutil'
import { useConfigValue } from '../util/RequireConfig'
import { Label } from '../../schema'

const MaxDetailsLength = 6 * 1024 // 6KiB

export interface Value {
  name: string
  description: string
  escalationPolicyID?: string
  labels: Label[]
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

  const [reqLabels] = useConfigValue('Services.RequiredLabels') as [string[]]

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
        {reqLabels &&
          reqLabels.map((labelName: string) => (
            <Grid item xs={12} key={labelName}>
              <FormField
                fullWidth
                name={labelName}
                required
                component={TextField}
                fieldName='labels'
                mapOnChangeValue={(newVal: string, value: Value) => {
                  return [
                    ...value.labels.filter((l) => l.key !== labelName),
                    {
                      key: labelName,
                      value: newVal,
                    },
                  ]
                }}
                mapValue={(labels: Label[]) =>
                  labels.find((l) => l.key === labelName)?.value || ''
                }
              />
            </Grid>
          ))}
      </Grid>
    </FormContainer>
  )
}
