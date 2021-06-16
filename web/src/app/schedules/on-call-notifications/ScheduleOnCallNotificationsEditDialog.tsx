import React, { useState } from 'react'
import { gql, useMutation, useQuery } from '@apollo/client'
import { DateTime } from 'luxon'
import { withoutTypeName } from './util'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors, fieldErrors, FieldError } from '../../util/errutil'
import ScheduleOnCallNotificationsForm, {
  Value,
} from './ScheduleOnCallNotificationsForm'

interface ScheduleOnCallNotificationsEditDialogProps {
  onClose: () => void

  scheduleID: string
  ruleID: string
}

const query = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
      onCallNotificationRules {
        id
        target {
          id
          type
        }
        time
        weekdayFilter
      }
    }
  }
`

const mutation = gql`
  mutation ($input: SetScheduleOnCallNotificationRulesInput!) {
    setScheduleOnCallNotificationRules(input: $input)
  }
`

export default function ScheduleOnCallNotificationsEditDialog(
  p: ScheduleOnCallNotificationsEditDialogProps,
): JSX.Element {
  const [_value, setValue] = useState<Value | null>(null)

  const { data, loading, error } = useQuery(query, {
    variables: { id: p.scheduleID },
  })
  const existingRules = (data?.schedule?.onCallNotificationRules || []).filter(
    (r) => r,
  )
  const rule = existingRules.find((r) => r.id === p.ruleID)

  const schedTZ = data?.schedule?.timeZone ?? ''
  const value = _value || {
    time: DateTime.fromFormat(rule?.time || '', 'HH:mm', {
      zone: schedTZ,
    }).toISO(),
    weekdayFilter: rule?.weekdayFilter ?? [
      false,
      false,
      false,
      false,
      false,
      false,
      false,
    ],
    slackChannelID: rule?.target?.id,
  }

  const [mutate, mutationStatus] = useMutation(mutation, {
    variables: {
      input: {
        scheduleID: p.scheduleID,
        rules: existingRules
          .filter((r) => r.id !== p.ruleID)
          .map(withoutTypeName)
          .concat({
            id: rule?.id,
            weekdayFilter: value.time ? value.weekdayFilter : null,
            time: DateTime.fromISO(value.time || '')
              .setZone(data?.schedule?.timeZone)
              .toFormat('HH:mm'),
            target: { type: 'slackChannel', id: value.slackChannelID },
          }),
      },
    },
    onCompleted: () => p.onClose(),
  })

  const formErrors = fieldErrors(mutationStatus.error).concat(
    nonFieldErrors(mutationStatus.error) as FieldError[], // NOTE:
  )

  return (
    <FormDialog
      title='Create Notification Rule'
      errors={formErrors}
      loading={loading}
      onClose={() => p.onClose()}
      onSubmit={() => mutate()}
      form={
        <ScheduleOnCallNotificationsForm
          scheduleID={p.scheduleID}
          errors={fieldErrors([])}
          value={value}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
