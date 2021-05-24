import React, { useState } from 'react'
import { useMutation } from '@apollo/client'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import Grid from '@material-ui/core/Grid'
import RadioGroup from '@material-ui/core/RadioGroup'
import Radio from '@material-ui/core/Radio'

import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'
import { setMutation } from './ScheduleOnCallNotifications'
import { FormContainer, FormField } from '../../forms'
import { SlackChannelSelect } from '../../selection'
import { ISODateTimePicker } from '../../util/ISOPickers'

interface ScheduleOnCallNotificationFormProps {
  scheduleID: string
  rule?: {
    channel: string
    onUpdate?: boolean
    atTime?: boolean
    day?:
      | 'Monday'
      | 'Tuesday'
      | 'Wednesday'
      | 'Thursday'
      | 'Friday'
      | 'Saturday'
      | 'Sunday'
    time?: string
  }
  onClose: () => void
}

export default function ScheduleOnCallNotificationFormDialog(
  p: ScheduleOnCallNotificationFormProps,
): JSX.Element {
  const [rule, setRule] = useState(p?.rule)
  const [mutate, mutationStatus] = useMutation(setMutation, {
    variables: {
      scheduleID: p.scheduleID,
      rules: [rule], // todo: append to rules
    },
  })

  console.log('rule: ', rule)

  return (
    <FormDialog
      title={(p.rule ? 'Edit ' : 'Create ') + 'Notification Rule'}
      errors={nonFieldErrors(mutationStatus.error)}
      onClose={() => p.onClose()}
      onSubmit={() => mutate()}
      form={
        <FormContainer value={rule} onChange={(value: any) => setRule(value)}>
          <Grid container spacing={2} direction='column'>
            <Grid item>
              <FormField
                component={SlackChannelSelect}
                fullWidth
                label='Select Channel'
                name='channel'
              />
            </Grid>
            <Grid item>
              <RadioGroup>
                <FormField
                  name='onUpdate'
                  render={() => (
                    <FormControlLabel
                      label='Notify when a on-call hands off to a new user'
                      value='onUpdate'
                      control={<Radio />}
                    />
                  )}
                />
                <FormField
                  name='atTime'
                  render={() => (
                    <FormControlLabel
                      label='Notify at a specific day and time every week'
                      value='atTime'
                      control={<Radio />}
                    />
                  )}
                />
              </RadioGroup>
            </Grid>
            <Grid item>
              <FormField
                component={ISODateTimePicker}
                fullWidth
                label='Time'
                disabled={!rule?.atTime}
                name='time'
              />
            </Grid>
          </Grid>
        </FormContainer>
      }
    />
  )
}
