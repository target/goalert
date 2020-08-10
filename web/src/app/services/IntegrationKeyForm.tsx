import React from 'react'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import MenuItem from '@material-ui/core/MenuItem'
import { Help } from '@material-ui/icons'
import { makeStyles } from '@material-ui/core/styles'
import Tooltip from '@material-ui/core/Tooltip'
import InputAdornment from '@material-ui/core/InputAdornment'
import { FormContainer, FormField } from '../forms'
import { Config } from '../util/RequireConfig'
import { AppLink } from '../util/AppLink'
import { IntegrationKeyType } from '../../schema'

const useStyles = makeStyles((theme) => ({
  infoIcon: {
    color: theme.palette.primary.main,
  },
}))

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
  const classes = useStyles()
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
                InputProps={{
                  endAdornment: (
                    <InputAdornment position='end'>
                      <Tooltip title='API Documentation' placement='right'>
                        <AppLink to='/docs' newTab>
                          <Help className={classes.infoIcon} />
                        </AppLink>
                      </Tooltip>
                    </InputAdornment>
                  ),
                }}
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
