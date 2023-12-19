import React, { useState } from 'react'

import { DateTime } from 'luxon'
import FormDialog from '../../dialogs/FormDialog'
import ScheduleOnCallNotificationsForm, {
  FormValue,
} from './ScheduleOnCallNotificationsForm'
import { EVERY_DAY, NO_DAY, destToDestInput, ruleToFormValue } from './util'
import { gql, useMutation, useQuery } from 'urql'
import {
  Destination,
  DestinationInput,
  OnCallNotificationRule,
  Schedule,
  SetScheduleOnCallNotificationRulesInput,
} from '../../../schema'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'

const rulesQuery = gql`
  query EditScheduleNotifyRules($id: ID!) {
    schedule(id: $id) {
      timeZone
      onCallNotificationRules {
        id
        time
        weekdayFilter
        dest {
          type
          values {
            fieldID
            value
          }
        }
      }
    }
  }
`

const rulesMutation = gql`
  mutation UpdateScheduleNotifyRules(
    $input: SetScheduleOnCallNotificationRulesInput!
  ) {
    setScheduleOnCallNotificationRules(input: $input)
  }
`

interface ScheduleOnCallNotificationsEditDialogProps {
  onClose: () => void

  scheduleID: string
  ruleID: string
}

export default function ScheduleOnCallNotificationsEditDialog(
  props: ScheduleOnCallNotificationsEditDialogProps,
): JSX.Element {
  const [{ data, error }] = useQuery<{ schedule: Schedule }>({
    query: rulesQuery,
    variables: { id: props.scheduleID },
  })
  if (error) throw error
  const [status, commit] = useMutation<
    unknown,
    {
      input: SetScheduleOnCallNotificationRulesInput
    }
  >(rulesMutation)

  const rules = data?.schedule.onCallNotificationRules || []
  const zone = data?.schedule.timeZone || ''
  const rule = rules.find((r) => r.id === props.ruleID)
  if (!rule) throw new Error('Rule not found') // should never happen (i.e. suspense)

  const [value, setValue] = useState<FormValue>({
    time: rule.time
      ? DateTime.fromFormat(rule.time, 'HH:mm', { zone }).toISO()
      : null,
    weekdayFilter: rule.time ? rule.weekdayFilter || EVERY_DAY : NO_DAY,
    dest: destToDestInput(rule.dest),
  })

  const fieldErrs = fieldErrors(status.error)
  const nonFieldErrs = nonFieldErrors(status.error)

  return (
    <FormDialog
      title='Edit Notification Rule'
      errors={nonFieldErrs}
      loading={status.fetching}
      onClose={() => props.onClose()}
      onSubmit={() =>
        commit({
          input: {
            scheduleID: props.scheduleID,
            rules: rules
              .filter((r) => r.id !== props.ruleID)
              .map(ruleToFormValue)
              .concat({
                id: props.ruleID,
                time: value.time
                  ? DateTime.fromISO(value.time, { zone }).toFormat('HH:mm')
                  : null,
                weekdayFilter: value.time ? value.weekdayFilter : null,
                dest: value.dest,
              }),
          },
        }).then((res) => {
          if (res.error) return

          props.onClose()
        })
      }
      form={
        <ScheduleOnCallNotificationsForm
          scheduleID={props.scheduleID}
          errors={fieldErrs}
          value={value}
          onChange={setValue}
        />
      }
    />
  )
}
