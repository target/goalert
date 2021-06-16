import React from 'react'
import { gql, useQuery } from '@apollo/client'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import Grid from '@material-ui/core/Grid'
import RadioGroup from '@material-ui/core/RadioGroup'
import Radio from '@material-ui/core/Radio'
import { DateTime } from 'luxon'
import { Checkbox, makeStyles, Typography } from '@material-ui/core'
import { FormContainer, FormField } from '../../forms'
import { SlackChannelSelect } from '../../selection'
import { ISOTimePicker } from '../../util/ISOPickers'
import { WeekdayFilter } from '../../../schema'

export type Value = {
  slackChannelID: string | null
  time: string | null
  weekdayFilter: WeekdayFilter
}

interface ScheduleOnCallNotificationsFormProps {
  scheduleID: string

  value: Value

  errors: {
    field: 'time' | 'weekdayFilter' | 'slackChannelID'
    message: string
  }[]

  onChange: (val: Value) => void
}

const query = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
    }
  }
`

function timeHint(schedTZ: string, timeISO: string | null): string {
  if (timeISO === null) return ''

  const dt = DateTime.fromISO(timeISO)
  const schedTime = dt.setZone(schedTZ).toLocaleString(DateTime.TIME_SIMPLE)
  const localTime = dt.setZone('local').toLocaleString(DateTime.TIME_SIMPLE)

  if (schedTime === localTime) return ''

  return `${localTime} ${dt.setZone('local').toFormat('ZZZZ')}`
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
  const { data, loading, error } = useQuery(query, {
    variables: { id: scheduleID },
  })
  const value = props.value
  const schedTZ = data?.schedule?.timeZone

  const handleRuleChange = (e) => {
    if (e.target.value === 'on-change') {
      formProps.onChange({
        ...formProps.value,
        time: null,
        weekdayFilter: [false, false, false, false, false, false, false],
      })
      return
    }

    formProps.onChange({
      ...formProps.value,
      weekdayFilter: [true, true, true, true, true, true, true],
      time: DateTime.fromObject({
        hour: 9,
        minute: 0,
        zone: schedTZ,
      }).toISO(),
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
            style={{ visibility: value.time ? 'visible' : 'hidden' }}
          >
            Configuring in {schedTZ}
          </Typography>
        </Grid>
        <Grid item>
          <Grid container spacing={2} alignItems='center'>
            <Grid item xs={12} sm={5} md={4}>
              <FormField
                component={ISOTimePicker}
                fullWidth
                name='time'
                disabled={!value.time}
                required={!!value.time}
                hint={timeHint(schedTZ, value.time)}
              />
            </Grid>
            <Grid item xs={12} sm={7} md={8}>
              <Grid container justify='space-between'>
                {days.map((day, i) => (
                  <FormControlLabel
                    key={i}
                    label={day.slice(0, 3)}
                    labelPlacement='top'
                    classes={{ labelPlacementTop: classes.margin0 }}
                    control={
                      <FormField
                        noError
                        component={Checkbox}
                        checkbox
                        name={`weekdayFilter[${i}]`}
                        disabled={!value.time}
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
