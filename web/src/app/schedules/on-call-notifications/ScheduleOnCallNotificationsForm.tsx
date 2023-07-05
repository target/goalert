import {
  Checkbox,
  FormControlLabel,
  Grid,
  Radio,
  RadioGroup,
  TextField,
  Typography,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { DateTime } from 'luxon'
import React, { useMemo } from 'react'

import { TargetType } from '../../../schema'
import { FormContainer, FormField } from '../../forms'
import { SlackChannelSelect, SlackUserGroupSelect } from '../../selection'
import {
  renderMenuItem,
  sortDisableableMenuItems,
} from '../../selection/DisableableMenuItem'
import { ISOTimePicker } from '../../util/ISOPickers'
import { useConfigValue } from '../../util/RequireConfig'
import { Time } from '../../util/Time'
import { useExpFlag } from '../../util/useExpFlag'
import { useScheduleTZ } from '../useScheduleTZ'
import { EVERY_DAY, NO_DAY, RuleFieldError, Value } from './util'

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
  const [webhookEnabled] = useConfigValue('Webhook.Enable')
  const webhookChannelEnabled = useExpFlag('chan-webhook')
  const slackUGEnabled = useExpFlag('slack-ug')
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

  const handleTypeChange = (e: React.ChangeEvent<HTMLInputElement>): void => {
    const newType = e.target.value as TargetType
    if (props.value.type !== newType) {
      props.onChange({
        ...props.value,
        type: newType,
        targetID: null,
      })
    }
  }

  const notificationChannels = useMemo(
    () =>
      [
        {
          value: 'slackChannel',
          label: 'SLACK CHANNEL',
          disabledMessage: 'Slack must be configured by an administrator',
          disabled: !slackEnabled,
        },
        ...(slackUGEnabled
          ? [
              {
                value: 'slackUserGroup',
                label: 'SLACK USER GROUP',
                disabledMessage: 'Slack must be configured by an administrator',
                disabled: !slackEnabled,
              },
            ]
          : []),
        ...(webhookChannelEnabled
          ? [
              {
                value: 'chanWebhook',
                label: 'WEBHOOK',
                disabledMessage: 'Webhooks must be enabled by an administrator',
                disabled: !webhookEnabled,
              },
            ]
          : []),
      ].sort(sortDisableableMenuItems),
    [slackEnabled, slackUGEnabled, webhookEnabled, webhookChannelEnabled],
  )

  function renderTypeFields(type: TargetType): JSX.Element {
    switch (type) {
      case 'slackUserGroup':
        return (
          <Grid item>
            <FormField
              component={SlackUserGroupSelect}
              fullWidth
              name='targetID'
              label='Slack User Group'
            />
          </Grid>
        )
      case 'slackChannel':
        return (
          <Grid item>
            <FormField
              component={SlackChannelSelect}
              fullWidth
              required
              label='Slack Channel'
              name='targetID'
            />
          </Grid>
        )
      case 'chanWebhook':
        return (
          <Grid item>
            <FormField
              component={TextField}
              fullWidth
              required
              label='Webhook'
              name='targetID'
            />
          </Grid>
        )
      default:
        // unsupported type
        return <Grid item />
    }
  }

  return (
    <FormContainer {...formProps}>
      <Grid container spacing={2} direction='column'>
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
        <Grid item>
          <TextField
            fullWidth
            value={props.value.type}
            required
            name='notificationType'
            label='Type'
            select
            onChange={handleTypeChange}
            disabled={notificationChannels.length <= 1}
          >
            {notificationChannels.map(renderMenuItem)}
          </TextField>
        </Grid>
        {renderTypeFields(formProps.value.type)}
      </Grid>
    </FormContainer>
  )
}
