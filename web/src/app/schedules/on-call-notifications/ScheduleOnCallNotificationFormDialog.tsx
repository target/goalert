import React, { ChangeEvent, useState } from 'react'
import { useQuery, useMutation } from '@apollo/client'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import Grid from '@material-ui/core/Grid'
import RadioGroup from '@material-ui/core/RadioGroup'
import Radio from '@material-ui/core/Radio'

import { query, setMutation } from './ScheduleOnCallNotificationsList'
import { Rule, mapDataToInput } from './util'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors, fieldErrors, FieldError } from '../../util/errutil'
import { FormContainer, FormField } from '../../forms'
import { SlackChannelSelect } from '../../selection'
import { ISOTimePicker } from '../../util/ISOPickers'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'
import { SelectOption } from '../../selection/MaterialSelect'
import { OnCallNotificationRuleInput, WeekdayFilter } from '../../../schema'
import { isoToGQLClockTime, days } from '../util'
import {
  Checkbox,
  Hidden,
  Table,
  TableBody,
  TableCell,
  TableRow,
} from '@material-ui/core'
import { DateTime } from 'luxon'
import useWidth from '../../util/useWidth'

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
  if (!rule) {
    return {
      slackChannelID: '',
      time: DateTime.local().set({ minute: 0, hour: 9 }).toISO(),
      weekdayFilter: new Array(7).fill(true) as WeekdayFilter,
      ruleType: RuleType.OnChange,
    }
  }

  const result: Value = {
    slackChannelID: rule.target.id,
    time: '',
    weekdayFilter: new Array(7).fill(false) as WeekdayFilter,
    ruleType: RuleType.OnChange,
  }

  // on schedule change
  if (rule.weekdayFilter) {
    result.weekdayFilter = rule.weekdayFilter
    result.time = rule.time as string
    result.ruleType = RuleType.OnSchedule
  }

  return result
}

interface ScheduleOnCallNotificationFormProps {
  scheduleID: string
  onClose: () => void

  // if set, populates form
  rule?: Rule
}

export default function ScheduleOnCallNotificationFormDialog(
  p: ScheduleOnCallNotificationFormProps,
): JSX.Element {
  const [value, setValue] = useState<Value>(getInitialValue(p.rule))
  const width = useWidth()
  console.log(value)

  const { loading, error, data } = useQuery(query, {
    variables: {
      id: p.scheduleID,
    },
    nextFetchPolicy: 'cache-first',
  })

  function makeRules(): OnCallNotificationRuleInput[] {
    let existingRules = mapDataToInput(data?.schedule?.onCallNotificationRules)

    // remove old rule when editing
    if (p.rule) {
      existingRules = existingRules.filter((r) => r.id !== p.rule?.id)
    }

    console.log('EXISTING', existingRules)

    let newRule: OnCallNotificationRuleInput
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

    console.log('NEW', newRule)
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
    setValue({ ...value, ruleType: event.target.value as RuleType })
  }

  function handleOnChange(value: Value): void {
    setValue(value)
  }

  const formErrors = fieldErrors(mutationStatus.error).concat(
    nonFieldErrors(mutationStatus.error) as FieldError[], // NOTE:
  )

  const scheduleTimeField = (
    <FormField
      component={ISOTimePicker}
      fullWidth
      label='Time'
      name='time'
      disabled={value.ruleType !== RuleType.OnSchedule}
      required={value.ruleType === RuleType.OnSchedule}
    />
  )

  return (
    <FormDialog
      title={(p.rule ? 'Edit ' : 'Create ') + 'Notification Rule'}
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
              <Table padding={width === 'xs' ? 'default' : 'none'}>
                <TableBody>
                  <Hidden smUp>
                    <TableRow>
                      <TableCell colSpan={7} padding='none'>
                        {scheduleTimeField}
                      </TableCell>
                    </TableRow>
                  </Hidden>
                  <TableRow>
                    <Hidden xsDown>
                      <TableCell rowSpan={2} padding='none'>
                        {scheduleTimeField}
                      </TableCell>
                    </Hidden>
                    {days.map((day, dayIdx) => (
                      <TableCell key={dayIdx} variant='head' align='center'>
                        {day.slice(0, 3)}
                      </TableCell>
                    ))}
                  </TableRow>
                  <TableRow>
                    {days.map((day, i) => (
                      <TableCell key={i} padding='checkbox'>
                        <FormField
                          noError
                          component={Checkbox}
                          checkbox
                          value={value.weekdayFilter[i]}
                          name={`weekday-filter[${i}]`}
                          fieldName={`weekdayFilter[${i}]`}
                          disabled={value.ruleType !== RuleType.OnSchedule}
                        />
                      </TableCell>
                    ))}
                  </TableRow>
                </TableBody>
              </Table>
            </Grid>
          </Grid>
        </FormContainer>
      }
    />
  )
}
