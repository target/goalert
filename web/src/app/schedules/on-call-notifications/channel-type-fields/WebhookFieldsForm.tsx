import Grid from '@mui/material/Grid'
import React from 'react'
import { FormField } from '../../../forms'
import { WebhookFields } from '../util'
import { TextField } from '@mui/material'

export type WebhookFieldsFormProps = {
  onChange: (val: WebhookFields) => void
}

export default function WebhookFieldsForm(): JSX.Element {
  return (
    <React.Fragment>
      <Grid item>
        <FormField
          component={TextField}
          fullWidth
          required
          label='Webhook URL'
          name='channelFields.webhookURL'
        />
      </Grid>
    </React.Fragment>
  )
}
