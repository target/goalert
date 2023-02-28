import React from 'react'
import FormControlLabel from '@mui/material/FormControlLabel'
import Grid from '@mui/material/Grid'
import RadioGroup from '@mui/material/RadioGroup'
import Radio from '@mui/material/Radio'
import { DateTime } from 'luxon'
import { Checkbox, Typography } from '@mui/material'

import makeStyles from '@mui/styles/makeStyles'

import { FormContainer, FormField } from '../../forms'
import { SlackChannelSelect } from '../../selection'
import { ISOTimePicker } from '../../util/ISOPickers'
import { Value, NO_DAY, EVERY_DAY, RuleFieldError } from './util'
import { Time } from '../../util/Time'
import { useScheduleTZ } from '../useScheduleTZ'

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

  return (
    <FormContainer {...formProps} optionalLabels>
      <Grid container spacing={2} direction='column'>
        <Grid item>
          <FormField
            component={SlackChannelSelect}
            fullWidth
            label='Slack Channel'
            name='slackChannelID'
            required
          />
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
        <Grid item xs={12}>
          <Typography
            color='textSecondary'
            className={classes.tzNote}
            style={{ visibility: props.value.time ? 'visible' : 'hidden' }}
          >
            Times shown in schedule timezone ({zone})
          </Typography>
        </Grid>
        <Grid item>
          <Grid container spacing={2} alignItems='center'>
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
        </Grid>
      </Grid>
    </FormContainer>
  )
}
