import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import MenuItem from '@material-ui/core/MenuItem'
import { Help } from '@material-ui/icons'
import withStyles from '@material-ui/core/styles/withStyles'
import Tooltip from '@material-ui/core/Tooltip'
import InputAdornment from '@material-ui/core/InputAdornment'
import { FormContainer, FormField } from '../forms'
import { Config } from '../util/RequireConfig'
import { AppLink } from '../util/AppLink'

const styles = theme => ({
  infoIcon: {
    color: theme.palette.primary['500'],
  },
})

@withStyles(styles)
export default class IntegrationKeyForm extends React.PureComponent {
  static propTypes = {
    value: p.shape({
      name: p.string,
      type: p.oneOf(['generic', 'grafana', 'site24x7', 'email']),
    }).isRequired,

    errors: p.arrayOf(
      p.shape({
        field: p.oneOf(['name', 'type']).isRequired,
        message: p.string.isRequired,
      }),
    ),

    onChange: p.func,
  }

  render() {
    const { classes, ...formProps } = this.props
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
              {cfg => (
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
                          <AppLink to='/docs' target='_blank'>
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
                </FormField>
              )}
            </Config>
          </Grid>
        </Grid>
      </FormContainer>
    )
  }
}
