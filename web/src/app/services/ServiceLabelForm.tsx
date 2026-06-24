import React from 'react'
import { FormContainer, FormField } from '../forms'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import { LabelKeySelect } from '../selection/LabelKeySelect'
import { Config } from '../util/RequireConfig'
import { Label } from '../../schema'

function validateKey(value: string): Error | undefined {
  const parts = value.split('/')
  if (parts.length !== 2)
    return new Error('Must be in the format: "example/KeyName".')
}

interface LabelFormProps {
  value: Label
  errors: Error[]

  onChange: (value: Label) => void
  editValueOnly?: boolean
  disabled?: boolean
  create?: boolean
}

export default function LabelForm(props: LabelFormProps): React.JSX.Element {
  const { editValueOnly = false, create, ...otherProps } = props

  return (
    <FormContainer {...otherProps} optionalLabels>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Config>
            {(cfg) => (
              <FormField
                fullWidth
                disabled={editValueOnly}
                component={LabelKeySelect}
                label='Key'
                name='key'
                required
                onCreate={
                  !cfg['General.DisableLabelCreation']
                    ? (key: string) =>
                        otherProps.onChange({ ...otherProps.value, key })
                    : undefined
                }
                validate={validateKey}
              />
            )}
          </Config>
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            label='Value'
            name='value'
            required
          />
        </Grid>
      </Grid>
    </FormContainer>
  )
}
