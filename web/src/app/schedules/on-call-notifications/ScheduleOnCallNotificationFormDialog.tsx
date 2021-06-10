import React, { ChangeEvent, useState } from 'react'
import { useQuery, useMutation } from '@apollo/client'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import Grid from '@material-ui/core/Grid'
import RadioGroup from '@material-ui/core/RadioGroup'
import Radio from '@material-ui/core/Radio'

import { query, setMutation } from './ScheduleOnCallNotificationsList'
import { Rule, RuleInput, mapDataToInput } from './util'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors, fieldErrors, FieldError } from '../../util/errutil'
import { FormContainer, FormField } from '../../forms'
import { SlackChannelSelect } from '../../selection'
import { ISOTimePicker } from '../../util/ISOPickers'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'
import { WeekdayFilter } from '../../../schema'
import { isoToGQLClockTime, days } from '../util'
import { Checkbox, makeStyles } from '@material-ui/core'
import { DateTime } from 'luxon'

enum RuleType {
  OnChange = 'ON_CHANGE',
  OnSchedule = 'ON_SCHEDULE',
}

type Value = {
  slackChannelID: string
  time: string
  weekdayFilter: WeekdayFilter
  ruleType: RuleType
}

function getInitialValue(rule?: Rule): Value {
  // defaults
  const result: Value = {
    slackChannelID: '',
    time: DateTime.local().set({ hour: 9, minute: 0 }).toISO(),
    weekdayFilter: new Array(7).fill(true) as WeekdayFilter,
    ruleType: RuleType.OnChange,
  }

  result.slackChannelID = rule?.target?.id ?? ''

  if (rule?.weekdayFilter) {
    result.weekdayFilter = rule.weekdayFilter
    result.time = rule.time as string
    result.ruleType = RuleType.OnSchedule
  }

  return result
}

const useStyles = makeStyles({ margin0: { margin: 0 } })

interface ScheduleOnCallNotificationFormProps {
  scheduleID: string
  onClose: () => void

  // if set, populates form
  rule?: Rule
}

export default function ScheduleOnCallNotificationFormDialog(
  p: ScheduleOnCallNotificationFormProps,
): JSX.Element {
  const [value, setValue] = useState(getInitialValue(p.rule))
  const classes = useStyles()

  const { loading, error, data } = useQuery(query, {
    variables: {
      id: p.scheduleID,
    },
    nextFetchPolicy: 'cache-first',
  })

  function makeRules(): RuleInput[] {
    let existingRules = mapDataToInput(data?.schedule?.onCallNotificationRules)

    // remove old rule when editing
    if (p.rule) {
      existingRules = existingRules.filter((r) => r.id !== p.rule?.id)
    }

    let newRule: RuleInput
    switch (value.ruleType) {
      case RuleType.OnChange:
        newRule = {
          target: {
            id: value.slackChannelID,
            type: 'slackChannel',
          },
        }
        break

      case RuleType.OnSchedule:
        newRule = {
          target: {
            id: value.slackChannelID,
            type: 'slackChannel',
          },
          // TODO add value.timeZone, pass as arg here
          time: isoToGQLClockTime(value.time),
          weekdayFilter: value.weekdayFilter,
        }
        break
      default:
        throw new Error('Unknown rule type')
    }

    return existingRules.concat(newRule)
  }

  const [mutate, mutationStatus] = useMutation(setMutation, {
    variables: {
      input: {
        scheduleID: p.scheduleID,
        rules: makeRules(),
      },
    },
    onCompleted: () => p.onClose(),
  })

  if (loading && !data?.schedule) return <Spinner />
  if (error) return <GenericError error={error.message} />

  function handleRadioOnChange(event: ChangeEvent<HTMLInputElement>): void {
    const ruleType = event.target.value as RuleType
    setValue({ ...value, ruleType })
  }

  function handleOnChange(value: Value): void {
    setValue(value)
  }

  const formErrors = fieldErrors(mutationStatus.error).concat(
    nonFieldErrors(mutationStatus.error) as FieldError[], // NOTE:
  )

  return (
    <FormDialog
      title={`${p.rule ? 'Edit' : 'Create'} Notification Rule`}
      errors={formErrors}
      onClose={() => p.onClose()}
      onSubmit={() => mutate()}
      form={
        <FormContainer
          value={value}
          onChange={(value: Value) => handleOnChange(value)}
          errors={formErrors}
        >
          <Grid container spacing={2} direction='column'>
            <Grid item>
              <FormField
                component={SlackChannelSelect}
                fullWidth
                label='Select Channel'
                name='slack-channel-id'
                fieldName='slackChannelID'
                required
              />
            </Grid>
            <Grid item>
              <RadioGroup onChange={handleRadioOnChange} value={value.ruleType}>
                <FormControlLabel
                  data-cy='notify-on-change'
                  label='Notify when on-call hands off to a new user'
                  value={RuleType.OnChange}
                  control={<Radio />}
                />
                <FormControlLabel
                  data-cy='notify-at-time'
                  label='Notify at a specific day and time every week'
                  value={RuleType.OnSchedule}
                  control={<Radio />}
                />
              </RadioGroup>
            </Grid>
            <Grid item>
              <Grid container spacing={2} alignItems='center'>
                <Grid item xs={12} sm={5} md={4}>
                  <FormField
                    component={ISOTimePicker}
                    fullWidth
                    label='Time'
                    name='time'
                    disabled={value.ruleType !== RuleType.OnSchedule}
                    required={value.ruleType === RuleType.OnSchedule}
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
                            value={value.weekdayFilter[i]}
                            name={`weekday-filter[${i}]`}
                            fieldName={`weekdayFilter[${i}]`}
                            disabled={value.ruleType !== RuleType.OnSchedule}
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
      }
    />
  )
}
