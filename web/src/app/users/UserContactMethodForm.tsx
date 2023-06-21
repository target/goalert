import React from 'react'
import Grid from '@mui/material/Grid'
import TextField from '@mui/material/TextField'
import { FormContainer, FormField } from '../forms'
import TelTextField from '../util/TelTextField'
import { Checkbox, FormControlLabel, MenuItem, Typography } from '@mui/material'
import { ContactMethodType, StatusUpdateState } from '../../schema'
import { useConfigValue } from '../util/RequireConfig'
import { FieldError } from '../util/errutil'

type Value = {
  name: string
  type: ContactMethodType
  value: string
  statusUpdates?: StatusUpdateState
}

export type UserContactMethodFormProps = {
  value: Value

  errors?: Array<FieldError>

  disabled?: boolean
  edit?: boolean

  onChange?: (CMValue: Value) => void
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
        component={TelTextField}
        disabled={edit}
      />
    </React.Fragment>
  )
}

function renderURLField(edit: boolean): JSX.Element {
  return (
    <FormField
      placeholder='https://example.com'
      fullWidth
      name='value'
      required
      label='Webhook URL'
      type='url'
      component={TextField}
      disabled={edit}
    />
  )
}

function renderSlackField(edit: boolean): JSX.Element {
  return (
    <FormField
      fullWidth
      name='value'
      required
      label='Slack Member ID'
      placeholder='member ID'
      component={TextField}
      disabled={edit}
      // @ts-expect-error TS2322 -- FormField has not been converted to ts, and inferred type is incorrect.
      helperText='Go to your Slack profile, click the three dots, and select "Copy member ID".'
    />
  )
}

function renderTypeField(type: ContactMethodType, edit: boolean): JSX.Element {
  switch (type) {
    case 'SMS':
    case 'VOICE':
      return renderPhoneField(edit)
    case 'EMAIL':
      return renderEmailField(edit)
    case 'WEBHOOK':
      return renderURLField(edit)
    case 'SLACK_DM':
      return renderSlackField(edit)
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

const isPhoneType = (val: Value): boolean =>
  val.type === 'SMS' || val.type === 'VOICE'

export default function UserContactMethodForm(
  props: UserContactMethodFormProps,
): JSX.Element {
  const { value, edit = false, ...other } = props

  const [
    smsVoiceEnabled,
    emailEnabled,
    webhookEnabled,
    slackEnabled,
    disclaimer,
  ] = useConfigValue(
    'Twilio.Enable',
    'SMTP.Enable',
    'Webhook.Enable',
    'Slack.Enable',
    'General.NotificationDisclaimer',
  )

  const statusUpdateChecked =
    value.statusUpdates === 'ENABLED' ||
    value.statusUpdates === 'ENABLED_FORCED' ||
    false

  return (
    <FormContainer
      {...other}
      value={value}
      mapOnChangeValue={(newValue: Value): Value => {
        // if switching between phone types (or same type), keep the value
        if (
          (isPhoneType(value) && isPhoneType(newValue)) ||
          value.type === newValue.type
        ) {
          return newValue
        }

        return {
          ...newValue,
          value: '',
        }
      }}
      optionalLabels
    >
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
            {(edit || webhookEnabled) && (
              <MenuItem value='WEBHOOK'>WEBHOOK</MenuItem>
            )}
            {(edit || slackEnabled) && (
              <MenuItem value='SLACK_DM'>SLACK DM</MenuItem>
            )}
          </FormField>
        </Grid>
        <Grid item xs={12}>
          {renderTypeField(value.type, edit)}
        </Grid>
        <Grid item xs={12}>
          <Typography variant='caption'>{disclaimer}</Typography>
        </Grid>
        {edit && (
          <Grid item xs={12}>
            <FormControlLabel
              label='Enable status updates'
              control={
                <Checkbox
                  name='enableStatusUpdates'
                  disabled={
                    value.statusUpdates === 'DISABLED_FORCED' ||
                    value.statusUpdates === 'ENABLED_FORCED'
                  }
                  checked={statusUpdateChecked}
                  onChange={(v) =>
                    props.onChange &&
                    props.onChange({
                      ...value,
                      statusUpdates: v.target.checked ? 'ENABLED' : 'DISABLED',
                    })
                  }
                />
              }
            />
          </Grid>
        )}
      </Grid>
    </FormContainer>
  )
}
