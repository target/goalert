import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import { FormContainer, FormField } from '../forms'
import { MenuItem, Typography } from '@material-ui/core'
import InputAdornment from '@material-ui/core/InputAdornment'

export default class UserContactMethodForm extends React.PureComponent {
  static propTypes = {
    value: p.shape({
      name: p.string.isRequired,
      type: p.oneOf(['SMS', 'VOICE']).isRequired,
      value: p.string.isRequired,
      // disclaimer text to display at the bottom of the form
      disclaimer: p.string,
    }).isRequired,

    errors: p.arrayOf(
      p.shape({
        field: p.oneOf(['name', 'type', 'value']).isRequired,
        message: p.string.isRequired,
      }),
    ),

    onChange: p.func,

    disabled: p.bool,

    edit: p.bool,
  }

  static defaultProps = {
    onChange: () => {},
  }

  render() {
    const cleanValue = (val) => {
      val = val.replace(/[^0-9]/g, '')

      if (!val) {
        return ''
      }
      return '+' + val
    }
    return (
      <FormContainer {...this.props} optionalLabels>
        <Grid container spacing={2}>
          <Grid item xs={12} sm={12} md={6}>
            <FormField fullWidth name='name' required component={TextField} />
          </Grid>
          <Grid item xs={12} sm={12} md={6}>
            <FormField
              fullWidth
              name='type'
              required
              select
              disabled={this.props.edit}
              component={TextField}
            >
              <MenuItem value='SMS'>SMS</MenuItem>
              <MenuItem value='VOICE'>VOICE</MenuItem>
            </FormField>
          </Grid>
          <Grid item xs={12}>
            <FormField
              placeholder='11235550123'
              aria-labelledby='countryCodeIndicator'
              fullWidth
              name='value'
              required
              label='Phone Number'
              type='tel'
              InputProps={{
                startAdornment: (
                  <InputAdornment style={{ marginBottom: '0.1em' }}>
                    +
                  </InputAdornment>
                ),
              }}
              component={TextField}
              mapValue={(val) => val.replace(/^\+/, '')}
              mapOnChangeValue={cleanValue}
              disabled={this.props.edit}
            />
            {!this.props.edit && (
              <Typography
                variant='caption'
                component='p'
                id='countryCodeIndicator'
              >
                Please provide your country code e.g. +1 (USA), +91 (India) +44
                (UK)
              </Typography>
            )}
          </Grid>
          <Grid item xs={12}>
            <Typography variant='caption'>{this.props.disclaimer}</Typography>
          </Grid>
        </Grid>
      </FormContainer>
    )
  }
}
