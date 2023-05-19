import React from 'react'
import FormControlLabel from '@mui/material/FormControlLabel'
import Grid from '@mui/material/Grid'
import RadioGroup from '@mui/material/RadioGroup'
import Radio from '@mui/material/Radio'
import { DateTime } from 'luxon'
import { Checkbox, MenuItem, TextField, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'

import { FormContainer, FormField } from '../../forms'
import { ISOTimePicker } from '../../util/ISOPickers'
import {
  Value,
  NO_DAY,
  EVERY_DAY,
  RuleFieldError,
  ChannelFields,
  SlackFields,
} from './util'
import { Time } from '../../util/Time'
import { useScheduleTZ } from '../useScheduleTZ'
import SlackFieldsForm from './channel-type-fields/SlackFieldsForm'
import { useExpFlag } from '../../util/useExpFlag'
import { useConfigValue } from '../../util/RequireConfig'
import WebhookFieldsForm from './channel-type-fields/WebhookFieldsForm'

const days = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat']

interface ScheduleOnCallNotificationsFormProps {
  scheduleID: string
  value: Value
  errors: RuleFieldError[]
  onChange: (val: Value) => void
}

const useStyles = makeStyles({
  margin0: { margin: 0 },
  tzNote: { fontStyle: 'italic' },
})

export default function ScheduleOnCallNotificationsForm(
  props: ScheduleOnCallNotificationsFormProps,
): JSX.Element {
  const { scheduleID, ...formProps } = props
  const classes = useStyles()
  const [slackEnabled] = useConfigValue('Slack.Enable')
  const webhookEnabled = useExpFlag('chan-webhook')
  const { zone } = useScheduleTZ(scheduleID)

  const handleRuleChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
    if (e.target.value === 'on-change') {
      props.onChange({ ...formProps.value, time: null, weekdayFilter: NO_DAY })
      return
    }
    props.onChange({
      ...props.value,
      weekdayFilter: EVERY_DAY,
      time: DateTime.fromObject({ hour: 9 }, { zone }).toISO(),
    })
  }

  const handleChannelDataChange = (channelFields: ChannelFields): void => {
    props.onChange({
      ...props.value,
      channelFields,
    })
  }

  return (
    <FormContainer {...formProps}>
      <Grid container spacing={2} direction='column'>
        <Grid item xs={12}>
          <FormField
            fullWidth
            name='type'
            required
            select
            component={TextField}
          >
            {slackEnabled && <MenuItem value='SLACK'>SLACK</MenuItem>}
            {webhookEnabled && <MenuItem value='WEBHOOK'>WEBHOOK</MenuItem>}
          </FormField>
        </Grid>
        <Grid item>
          <RadioGroup
            name='ruleType'
            value={formProps.value.time ? 'time-of-day' : 'on-change'}
            onChange={handleRuleChange}
          >
            <FormControlLabel
              data-cy='notify-on-change'
              label='Notify when on-call changes'
              value='on-change'
              control={<Radio />}
            />
            <FormControlLabel
              data-cy='notify-at-time'
              label='Notify at a specific day and time every week'
              value='time-of-day'
              control={<Radio />}
            />
          </RadioGroup>
        </Grid>
        {props.value.time && (
          <Grid item xs={12}>
            <Typography color='textSecondary' className={classes.tzNote}>
              Times shown in schedule timezone ({zone})
            </Typography>
          </Grid>
        )}
        <Grid item container spacing={2} alignItems='center'>
          <Grid item xs={12} sm={5} md={4}>
            <FormField
              component={ISOTimePicker}
              timeZone={zone}
              fullWidth
              name='time'
              disabled={!props.value.time}
              required={!!props.value.time}
              hint={
                <Time
                  format='clock'
                  time={props.value.time}
                  suffix=' in local time'
                />
              }
            />
          </Grid>
          <Grid item xs={12} sm={7} md={8}>
            <Grid container justifyContent='space-between'>
              {days.map((day, i) => (
                <FormControlLabel
                  key={i}
                  label={day}
                  labelPlacement='top'
                  classes={{ labelPlacementTop: classes.margin0 }}
                  control={
                    <FormField
                      noError
                      component={Checkbox}
                      checkbox
                      name={`weekdayFilter[${i}]`}
                      disabled={!props.value.time}
                    />
                  }
                />
              ))}
            </Grid>
          </Grid>
        </Grid>
        {formProps.value.type === 'SLACK' && (
          <SlackFieldsForm
            slackFields={
              (formProps.value.channelFields as SlackFields) ?? {
                slackChannelID: null,
              }
            }
            onChange={handleChannelDataChange}
          />
        )}
        {formProps.value.type === 'WEBHOOK' && <WebhookFieldsForm />}
      </Grid>
    </FormContainer>
  )
}
