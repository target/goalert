import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import { FormContainer, FormField } from '../forms'
import { MenuItem, Typography } from '@material-ui/core'
// import { getCountryCode, stripCountryCode } from './util'

export default class UserContactMethodForm extends React.PureComponent {
  static propTypes = {
    value: p.shape({
      name: p.string.isRequired,
      type: p.oneOf(['SMS', 'VOICE']).isRequired,
      value: p.string.isRequired,
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
          {/* <Grid item xs={12} sm={12} md={6}> */}
          {/*  <TextField */}
          {/*    fullWidth */}
          {/*    name='countryCode' */}
          {/*    label='Country Code' */}
          {/*    value={getCountryCode(this.props.value.value)} */}
          {/*    onChange={e => */}
          {/*      this.props.onChange({ */}
          {/*        ...this.props.value, */}
          {/*        value: */}
          {/*          e.target.value + stripCountryCode(this.props.value.value), */}
          {/*      }) */}
          {/*    } */}
          {/*    select */}
          {/*  > */}
          {/*    <MenuItem value='+1'>+1 (USA)</MenuItem> */}
          {/*    <MenuItem value='+91'>+91 (India)</MenuItem> */}
          {/*  </TextField> */}
          {/* </Grid> */}
          <Grid item xs={12} sm={12}>
            <FormField
              placeholder={'+1 123-456-7891'}
              aria-labelledby={'countryCodeIndicator'}
              fullWidth
              name='value'
              required
              label='Phone Number'
              type='tel'
              component={TextField}
            />
            <Typography
              variant={'caption'}
              component={'p'}
              id={'countryCodeIndicator'}
            >
              Please provide your country code eg. +1 (USA), +91 (India) +44
              (UK)
            </Typography>
            {/* <Link href={'https://countrycode.org/'} target={'_blank'}> */}
            {/*  Find Your Country Code */}
            {/* </Link> */}
          </Grid>
        </Grid>
      </FormContainer>
    )
  }
}
