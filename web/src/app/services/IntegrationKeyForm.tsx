import React from 'react'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import MenuItem from '@material-ui/core/MenuItem'
import { FormContainer, FormField } from '../forms'
import { Config } from '../util/RequireConfig'
import { IntegrationKeyType } from '../../schema'

interface Value {
  name: string
  type: IntegrationKeyType
}

interface IntegrationKeyFormProps {
  value: Value

  errors: {
    field: 'name' | 'type'
    message: string
  }[]

  onChange: (val: Value) => void
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
            {(cfg: { [x: string]: unknown }) => (
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
