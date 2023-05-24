import {
  Checkbox,
  FormControlLabel,
  Grid,
  MenuItem,
  Radio,
  RadioGroup,
  Select,
  SelectChangeEvent,
  Typography,
} from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { DateTime } from 'luxon'
import React from 'react'

import { FormContainer, FormField } from '../../forms'
import { SlackChannelSelect, SlackUserGroupSelect } from '../../selection'
import { ISOTimePicker } from '../../util/ISOPickers'
import { useConfigValue } from '../../util/RequireConfig'
import { Time } from '../../util/Time'
import { useExpFlag } from '../../util/useExpFlag'
import { useScheduleTZ } from '../useScheduleTZ'
import {
  EVERY_DAY,
  NO_DAY,
  NotificationChannelType,
  RuleFieldError,
  Value,
  getEmptyChannelFields,
} from './util'

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

  const handleTypeChange = (
    e: SelectChangeEvent<NotificationChannelType>,
  ): void => {
    const newType = e.target.value as NotificationChannelType
    if (props.value.type !== newType) {
      const channelFields = getEmptyChannelFields(newType)
      if (
        'slackChannelID' in props.value.channelFields &&
        'slackChannelID' in channelFields
      ) {
        channelFields.slackChannelID = props.value.channelFields
          .slackChannelID as string | null
      }
      props.onChange({ ...props.value, type: newType, channelFields })
    }
  }

  function renderTypeFields(type: NotificationChannelType): JSX.Element {
    switch (type) {
      case 'SLACK_UG':
        return (
          <React.Fragment>
            <Grid item>
              <FormField
                component={SlackUserGroupSelect}
                fullWidth
                label='Slack User Group'
                name='channelFields.slackUserGroup'
              />
            </Grid>
            <Grid item>
              <FormField
                component={SlackChannelSelect}
                fullWidth
                required
                label='Slack Channel (fallback)'
                name='channelFields.slackChannelID'
              />
            </Grid>
          </React.Fragment>
        )
      case 'SLACK_CHANNEL':
      default:
        return (
          <Grid item>
            <FormField
              component={SlackChannelSelect}
              fullWidth
              required
              label='Slack Channel'
              name='channelFields.slackChannelID'
            />
          </Grid>
        )
    }
  }

  return (
    <FormContainer {...formProps}>
      <Grid container spacing={2} direction='column'>
        <Grid item xs={12}>
          <Select
            fullWidth
            value={props.value.type}
            required
            onChange={handleTypeChange}
          >
            <MenuItem value='SLACK_CHANNEL' disabled={!slackEnabled}>
              SLACK CHANNEL
            </MenuItem>
            {slackUGEnabled && (
              <MenuItem value='SLACK_UG' disabled={!slackEnabled}>
                SLACK USER GROUP
              </MenuItem>
            )}
          </Select>
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
        {renderTypeFields(formProps.value.type)}
      </Grid>
    </FormContainer>
  )
}
