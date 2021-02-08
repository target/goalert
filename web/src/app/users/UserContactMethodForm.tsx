import React from 'react'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import { FormContainer, FormField } from '../forms'
import TelTextField from '../util/TelTextField'
import { MenuItem, Typography } from '@material-ui/core'
import { ContactMethodType } from '../../schema'
import { useConfigValue } from '../util/RequireConfig'

export type UserContactMethodFormProps = {
  value: { name: string; type: ContactMethodType; value: string }
  disclaimer?: string

  errors?: Array<{ field: 'name' | 'type' | 'value'; message: string }>

  disabled?: boolean
  edit?: boolean
}

function renderEmailField(edit: boolean): JSX.Element {
  return (
    <FormField
      placeholder='foobar@example.com'
      fullWidth
      name='value'
      required
      label='Email Address'
      type='email'
      component={TextField}
      disabled={edit}
    />
  )
}

function renderPhoneField(edit: boolean): JSX.Element {
  return (
    <React.Fragment>
      <FormField
        placeholder='11235550123'
        aria-labelledby='countryCodeIndicator'
        fullWidth
        name='value'
        required
        label='Phone Number'
        type='tel'
        component={TelTextField}
        disabled={edit}
      />
      {!edit && (
        <Typography variant='caption' component='p' id='countryCodeIndicator'>
          Please provide your country code e.g. +1 (USA), +91 (India), +44 (UK)
        </Typography>
      )}
    </React.Fragment>
  )
}

function renderTypeField(type: ContactMethodType, edit: boolean): JSX.Element {
  switch (type) {
    case 'SMS':
    case 'VOICE':
      return renderPhoneField(edit)
    case 'EMAIL':
      return renderEmailField(edit)
    default:
  }

  // fallback to generic
  return (
    <FormField
      fullWidth
      name='value'
      required
      label='Value'
      component={TextField}
      disabled={edit}
    />
  )
}

export default function UserContactMethodForm(
  props: UserContactMethodFormProps,
): JSX.Element {
  const { value, edit = false, disclaimer, ...other } = props

  const [smsVoiceEnabled, emailEnabled] = useConfigValue(
    'Twilio.Enable',
    'SMTP.Enable',
  )

  return (
    <FormContainer {...other} value={value} optionalLabels>
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
            disabled={edit}
            component={TextField}
          >
            {(edit || smsVoiceEnabled) && <MenuItem value='SMS'>SMS</MenuItem>}
            {(edit || smsVoiceEnabled) && (
              <MenuItem value='VOICE'>VOICE</MenuItem>
            )}
            {(edit || emailEnabled) && <MenuItem value='EMAIL'>EMAIL</MenuItem>}
          </FormField>
        </Grid>
        <Grid item xs={12}>
          {renderTypeField(value.type, edit)}
        </Grid>
        <Grid item xs={12}>
          <Typography variant='caption'>{disclaimer}</Typography>
        </Grid>
      </Grid>
    </FormContainer>
  )
}
