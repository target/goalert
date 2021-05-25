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

interface ScheduleOnCallNotificationFormProps {
  scheduleID: string
  rule?: Rule
  onClose: () => void
}

type Value = {
  target: {
    type: 'slackChannel'
    id: string
  }
  time?: string
  weekdayFilter?: string
}

export default function ScheduleOnCallNotificationFormDialog(
  p: ScheduleOnCallNotificationFormProps,
): JSX.Element {
  const [value, setValue] = useState<Value>(p?.rule as Value)
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
    if (notifyOnUpdate) {
      delete value.time
      delete value.weekdayFilter
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
                      value: 'sunday',
                    },
                    {
                      label: 'Monday',
                      value: 'monday',
                    },
                    {
                      label: 'Tuesday',
                      value: 'tuesday',
                    },
                    {
                      label: 'Wednesday',
                      value: 'wednesday',
                    },
                    {
                      label: 'Thursday',
                      value: 'thursday',
                    },
                    {
                      label: 'Friday',
                      value: 'friday',
                    },
                    {
                      label: 'Saturday',
                      value: 'saturday',
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
