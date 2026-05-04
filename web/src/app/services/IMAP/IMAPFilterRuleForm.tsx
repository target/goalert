import React, { ChangeEvent } from 'react'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import MenuItem from '@mui/material/MenuItem'
import { FormContainer, FormField } from '../../forms'
import { FieldError } from '../../util/errutil'
import { FormControlLabel, Switch } from '@mui/material'

export interface Value {
  name: string
  fromPattern: string
  subjectPattern: string
  toPattern: string
  matchMode: 'exact' | 'contains' | 'regex'
  excludeReplies: boolean
  enabled?: boolean
}

interface IMAPFilterRuleFormProps {
  value: Value

  errors: FieldError[]

  onChange: (val: Value) => void

  disabled: boolean

  edit?: boolean
}

export default function IMAPFilterRuleForm(
  props: IMAPFilterRuleFormProps,
): JSX.Element {
  const { edit = false, value, onChange, disabled, ...formProps } = props

  return (
    <FormContainer
      value={value}
      onChange={onChange}
      disabled={disabled}
      {...formProps}
      optionalLabels
    >
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            label='Name'
            name='name'
            required
            hint='A descriptive name for this filter rule'
          />
        </Grid>

        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            label='Subject Pattern'
            name='subjectPattern'
            hint='Pattern to match in the email subject line'
          />
        </Grid>

        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            label='From Pattern'
            name='fromPattern'
            hint='Pattern to match in the sender email address'
          />
        </Grid>

        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            label='To Pattern'
            name='toPattern'
            hint='Pattern to match in the recipient email address'
          />
        </Grid>

        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            select
            required
            label='Match Mode'
            name='matchMode'
            hint='How patterns should be matched against email content'
          >
            <MenuItem value='contains'>Contains</MenuItem>
            <MenuItem value='exact'>Exact Match</MenuItem>
            <MenuItem value='regex'>Regular Expression</MenuItem>
          </FormField>
        </Grid>

        <Grid item xs={12}>
          <FormControlLabel
            control={
              <Switch
                checked={value.excludeReplies}
                onChange={(e: ChangeEvent<HTMLInputElement>) =>
                  onChange({ ...value, excludeReplies: e.target.checked })
                }
                disabled={disabled}
              />
            }
            label='Exclude reply emails (Re:, Fwd:)'
          />
        </Grid>

        {edit && (
          <Grid item xs={12}>
            <FormControlLabel
              control={
                <Switch
                  checked={value.enabled ?? true}
                  onChange={(e: ChangeEvent<HTMLInputElement>) =>
                    onChange({ ...value, enabled: e.target.checked })
                  }
                  disabled={disabled}
                />
              }
              label='Enabled'
            />
          </Grid>
        )}
      </Grid>
    </FormContainer>
  )
}
