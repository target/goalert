import React, { ChangeEvent, useState } from 'react'
import { useQuery, useMutation } from '@apollo/client'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import Grid from '@material-ui/core/Grid'
import RadioGroup from '@material-ui/core/RadioGroup'
import Radio from '@material-ui/core/Radio'
import { makeStyles } from '@material-ui/core/styles'
import _ from 'lodash'

import { query, setMutation } from './ScheduleOnCallNotificationsList'
import { Rule } from './ScheduleOnCallNotificationAction'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'
import { FormContainer, FormField } from '../../forms'
import { SlackChannelSelect } from '../../selection'
import { ISOTimePicker } from '../../util/ISOPickers'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'
import MaterialSelect, { SelectOption } from '../../selection/MaterialSelect'
import { OnCallNotificationRuleInput, WeekdayFilter } from '../../../schema'
import { isoToGQLClockTime, days } from '../util'

interface ScheduleOnCallNotificationFormProps {
  scheduleID: string
  onClose: () => void

  // if set, the form will default with these values
  rule?: Rule
}

type Value = {
  target: string
  time?: string
  weekdayFilter?: Array<SelectOption> | null
}

const useStyles = makeStyles({
  timeFields: {
    display: 'flex',
  },
  timeField: {
    paddingLeft: '2.5rem',
    paddingRight: 8,
    width: 'fit-content',
  },
  weekdayFilterField: {
    paddingRight: '2.5rem',
    paddingLeft: 8,
    width: '100%',
  },
})

// getWeekdayFilter takes the selected days and returns a full
// week represented as booleans.
// e.g. [{ label: 'Monday', value: '1' }, { label: 'Wednesday', value: '3' }]
// -> [true, false, true, false, false, false, true]
function getWeekdayFilter(days: Array<SelectOption>): WeekdayFilter {
  const res = new Array(7).fill(false) as WeekdayFilter

  days.forEach((day) => {
    res[parseInt(day.value, 10)] = true
  })

  return res
}

// getSelectedDays takes WeekdayFilter and returns the included truthy days
// as their given day-index in a week
// e.g. [false, true, true, false, false, true, false]
// -> [{ label: 'Monday', value: '1' }, { label: 'Wednesday', value: '3' }]
export function getSelectedDays(
  weekdayFilter?: WeekdayFilter,
): Array<SelectOption> {
  if (!weekdayFilter) {
    return []
  }
  return weekdayFilter
    .map((dayVal, idx) => ({
      label: days[idx],
      value: dayVal ? idx.toString() : '-1',
    }))
    .filter((dayVal) => dayVal.value !== '-1')
}

export function mapDataToInput(
  rules: Array<Rule> = [],
): Array<OnCallNotificationRuleInput> {
  return rules.map((nr: Rule) => {
    const n = _.pick(nr, 'id', 'target', 'time', 'weekdayFilter')
    n.target = _.pick(n.target, 'id', 'type')
    return n
  }) as Array<OnCallNotificationRuleInput>
}

export default function ScheduleOnCallNotificationFormDialog(
  p: ScheduleOnCallNotificationFormProps,
): JSX.Element {
  const classes = useStyles()
  const [value, setValue] = useState<Value>({
    target: p.rule?.target.id ?? '',
    time: p.rule?.time ?? '',
    weekdayFilter: getSelectedDays(p.rule?.weekdayFilter),
  })

  const [notifyOnUpdate, setNotifyOnUpdate] = useState(true)
  const [mutate, mutationStatus] = useMutation(setMutation)

  // load all rules to set
  const { loading, error, data } = useQuery(query, {
    variables: {
      id: p.scheduleID,
    },
    nextFetchPolicy: 'cache-first',
  })

  function handleOnSubmit(): void {
    const rules = mapDataToInput(data?.schedule?.onCallNotificationRules)
    const newRule: OnCallNotificationRuleInput = {
      target: {
        id: value.target,
        type: 'slackChannel',
      },
      time: isoToGQLClockTime(value.time),
      weekdayFilter: getWeekdayFilter(value?.weekdayFilter ?? []),
    }

    if (notifyOnUpdate) {
      delete newRule.time
      delete newRule.weekdayFilter
    }

    // handle editing vs creating
    const newRules = rules
    if (p.rule) {
      const idx = _.findIndex(rules, ['id', p.rule.id])
      newRules[idx] = {
        ...rules[idx],
        ...newRule,
      }
    } else {
      newRules.push(newRule)
    }

    mutate({
      variables: {
        input: {
          scheduleID: p.scheduleID,
          rules: newRules,
        },
      },
      optimisticResponse: () => p.onClose(),
    })
  }

  if (loading && !data?.schedule) return <Spinner />
  if (error) return <GenericError error={error.message} />

  function handleRadioOnChange(event: ChangeEvent<HTMLInputElement>): void {
    setNotifyOnUpdate(event?.target?.value === 'true')
  }

  function handleOnChange(value: Value): void {
    setValue(value)
  }

  return (
    <FormDialog
      title={(p.rule ? 'Edit ' : 'Create ') + 'Notification Rule'}
      errors={nonFieldErrors(mutationStatus.error)}
      onClose={() => p.onClose()}
      onSubmit={() => handleOnSubmit()}
      form={
        <FormContainer
          value={value}
          onChange={(value: Value) => handleOnChange(value)}
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
                  label='Notify when on-call hands off to a new user'
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
            <Grid className={classes.timeFields} item>
              <div className={classes.timeField}>
                <FormField
                  component={ISOTimePicker}
                  label='Time'
                  name='time'
                  disabled={notifyOnUpdate}
                />
              </div>
              <div className={classes.weekdayFilterField}>
                <FormField
                  component={MaterialSelect}
                  name='weekdayFilter'
                  label='Select Days'
                  multiple
                  fullWidth
                  disabled={notifyOnUpdate}
                  options={[
                    {
                      label: 'Sunday',
                      value: '0',
                    },
                    {
                      label: 'Monday',
                      value: '1',
                    },
                    {
                      label: 'Tuesday',
                      value: '2',
                    },
                    {
                      label: 'Wednesday',
                      value: '3',
                    },
                    {
                      label: 'Thursday',
                      value: '4',
                    },
                    {
                      label: 'Friday',
                      value: '5',
                    },
                    {
                      label: 'Saturday',
                      value: '6',
                    },
                  ]}
                />
              </div>
            </Grid>
          </Grid>
        </FormContainer>
      }
    />
  )
}
