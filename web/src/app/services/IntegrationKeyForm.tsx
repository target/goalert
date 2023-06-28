import React, { ReactElement } from 'react'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import MenuItem from '@mui/material/MenuItem'
import { FormContainer, FormField } from '../forms'
import { Config, ConfigData } from '../util/RequireConfig'
import { IntegrationKeyType } from '../../schema'
import { FieldError } from '../util/errutil'

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
          <Config>
            {(cfg: ConfigData): ReactElement => (
              <FormField
                fullWidth
                component={TextField}
                select
                required
                label='Type'
                name='type'
              >
                {cfg['Mailgun.Enable'] && (
                  <MenuItem value='email'>Email</MenuItem>
                )}
                <MenuItem value='generic'>Generic API</MenuItem>
                <MenuItem value='grafana'>Grafana</MenuItem>
                <MenuItem value='site24x7'>Site24x7</MenuItem>
                <MenuItem value='prometheusAlertmanager'>
                  Prometheus Alertmanager
                </MenuItem>
              </FormField>
            )}
          </Config>
        </Grid>
      </Grid>
    </FormContainer>
  )
}
