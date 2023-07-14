import React from 'react'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import MenuItem from '@mui/material/MenuItem'
import { FormContainer, FormField } from '../forms'
import { IntegrationKeyType } from '../../schema'
import { FieldError } from '../util/errutil'
import { useIntKeyTypes } from '../util/useIntKeyTypes'

export interface Value {
  name: string
  type: IntegrationKeyType
}

interface IntegrationKeyFormProps {
  value: Value

  errors: FieldError[]

  onChange: (val: Value) => void

  // can be deleted when FormContainer.js is converted to ts
  disabled: boolean
}

export default function IntegrationKeyForm(
  props: IntegrationKeyFormProps,
): JSX.Element {
  const types = useIntKeyTypes()

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
            select
            required
            label='Type'
            name='type'
          >
            {types.map((t) => (
              <MenuItem disabled={!t.enabled} key={t.id} value={t.id}>
                {t.name}
              </MenuItem>
            ))}
          </FormField>
        </Grid>
      </Grid>
    </FormContainer>
  )
}
