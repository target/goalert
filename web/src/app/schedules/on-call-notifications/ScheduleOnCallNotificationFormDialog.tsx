import React, { ChangeEvent, useState } from 'react'
import { useQuery, useMutation } from '@apollo/client'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import Grid from '@material-ui/core/Grid'
import RadioGroup from '@material-ui/core/RadioGroup'
import Radio from '@material-ui/core/Radio'

import { query, setMutation } from './ScheduleOnCallNotificationsList'
import { Rule } from './ScheduleOnCallNotificationAction'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'
import { FormContainer, FormField } from '../../forms'
import { SlackChannelSelect } from '../../selection'
import { ISOTimePicker } from '../../util/ISOPickers'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'
import MaterialSelect from '../../selection/MaterialSelect'
import { OnCallNotificationRuleInput, WeekdayFilter } from '../../../schema'

interface ScheduleOnCallNotificationFormProps {
  scheduleID: string
  rule?: Rule
  onClose: () => void
}

type Value = {
  target: string
  time?: string
  weekdayFilter?: Array<number>
}

// Set weekdayFilter
// e.g. ['Monday', 'Wednesday', 'Saturday']
// -> [true, false, true, false, false, false, true]
function getWeekdayFilter(days: Array<number>): WeekdayFilter {
  const d = Array.apply(null, Array(7)).map(Boolean.prototype.valueOf, false) as WeekdayFilter
  days.forEach((day) => {
    d[day] = true
  })
  return d
}

// e.g. [false, true, true, false, false, true, false]
// -> [1, 2, 5]
function getSelectedDays(weekdayFilter: WeekdayFilter): Array<number> {
  return weekdayFilter.map((day, idx) => day ? idx : -1).filter((day) => day < 0)
}

export default function ScheduleOnCallNotificationFormDialog(
  p: ScheduleOnCallNotificationFormProps,
): JSX.Element {

  let initialVal: Value | null = null
  if (p.rule) {
    initialVal = {
      target: p.rule.target.id,
      time: p.rule.time,
      weekdayFilter: getSelectedDays(p?.rule?.weekdayFilter)
    }
  }

  const [value, setValue] = useState<Value | null>(initialVal)
  const [notifyOnUpdate, setNotifyOnUpdate] = useState(true)

  // load all rules if editing
  const { loading, error, data } = useQuery(query, {
    variables: {
      id: p.scheduleID,
    },
    nextFetchPolicy: 'cache-first',
    skip: !p.rule,
  })

  // add form value to rules
  let rules = data?.schedule?.notificationRules ?? []
  if (value) {
    let newRule: OnCallNotificationRuleInput = {
      target: {
        id: value.target,
        type: 'slackChannel',
      },
      time: value.time,
      weekdayFilter: getWeekdayFilter(value?.weekdayFilter ?? [])
    }

    if (notifyOnUpdate) {
      delete newRule.time
      delete newRule.weekdayFilter
    }

    rules = [...rules, value]
  }

  const [mutate, mutationStatus] = useMutation(setMutation, {
    variables: {
      scheduleID: p.scheduleID,
      rules,
    },
  })

  if (loading && !data?.schedule) return <Spinner />
  if (error) return <GenericError error={error.message} />

  function handleRadioOnChange(event: ChangeEvent<HTMLInputElement>): void {
    setNotifyOnUpdate(event?.target?.value === 'true')
  }

  return (
    <FormDialog
      title={(p.rule ? 'Edit ' : 'Create ') + 'Notification Rule'}
      errors={nonFieldErrors(mutationStatus.error)}
      onClose={() => p.onClose()}
      onSubmit={() => mutate()}
      form={
        <FormContainer
          value={value}
          onChange={(value: Value) => setValue(value)}
          errors={mutationStatus.error}
        >
          <Grid container spacing={2} direction='column'>
            <Grid item>
              <FormField
                component={SlackChannelSelect}
                fullWidth
                label='Select Channel'
                name='target'
              />
            </Grid>
            <Grid item>
              <RadioGroup onChange={handleRadioOnChange}>
                <FormControlLabel
                  label='Notify when a on-call hands off to a new user'
                  value='true'
                  control={<Radio />}
                />
                <FormControlLabel
                  label='Notify at a specific day and time every week'
                  value='false'
                  control={<Radio />}
                />
              </RadioGroup>
            </Grid>
            <Grid item container spacing={2}>
              <Grid item>
                <FormField
                  component={ISOTimePicker}
                  label='Time'
                  name='time'
                  disabled={notifyOnUpdate}
                />
              </Grid>
              <Grid item>
                <FormField
                  component={MaterialSelect}
                  name='weekdayFilter'
                  label='Select Days'
                  required
                  multiple
                  fullWidth
                  disabled={notifyOnUpdate}
                  options={[
                    {
                      label: 'Sunday',
                      value: 0,
                    },
                    {
                      label: 'Monday',
                      value: 1,
                    },
                    {
                      label: 'Tuesday',
                      value: 2,
                    },
                    {
                      label: 'Wednesday',
                      value: 3,
                    },
                    {
                      label: 'Thursday',
                      value: 4,
                    },
                    {
                      label: 'Friday',
                      value: 5,
                    },
                    {
                      label: 'Saturday',
                      value: 6,
                    },
                  ]}
                />
              </Grid>
            </Grid>
          </Grid>
        </FormContainer>
      }
    />
  )
}
