import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import { FormContainer, FormField } from '../forms'
import { MenuItem, Typography } from '@material-ui/core'
import exampleNumbers from 'libphonenumber-js/examples.mobile.json'
import { getExampleNumber } from 'libphonenumber-js'

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
    const locale = this.props.value

    let exampleNumber

    try {
      exampleNumber = getExampleNumber(
        locale.countryCode,
        exampleNumbers,
      ).formatInternational()
    } catch (e) {
      exampleNumber = '+1 631 746 3748'
    }

    const dialCode = exampleNumber.split(' ')[0]

    const targetHQs = ['US', 'IN']
    let localizedHelpText = ''
    if (!targetHQs.includes(locale.countryCode)) {
      localizedHelpText = `${dialCode} (${locale.countryName}), `
    }

    const cleanValue = val => {
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
              placeholder={exampleNumber}
              aria-labelledby='countryCodeIndicator'
              fullWidth
              name='value'
              required
              label='Phone Number'
              type='tel'
              component={TextField}
              mapOnChangeValue={cleanValue}
              disabled={this.props.edit}
            />
            {!this.props.edit && (
              <Typography
                variant='caption'
                component='p'
                id='countryCodeIndicator'
              >
                {`Please provide your country dialing code e.g. ${localizedHelpText}+1 (USA), +91 (India)`}
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
