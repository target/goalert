import React, { useState } from 'react'
import FormDialog from '../../dialogs/FormDialog'
import ScheduleOnCallNotificationsForm, {
  FormValue,
} from './ScheduleOnCallNotificationsForm'
import { NO_DAY, ruleToFormValue } from './util'
import { useSchedOnCallNotifyTypes } from '../../util/useDestinationTypes'
import { useQuery, useMutation, gql } from 'urql'
import {
  Schedule,
  SetScheduleOnCallNotificationRulesInput,
} from '../../../schema'
import { fieldErrors, nonFieldErrors } from '../../util/errutil'

const rulesQuery = gql`
  query FetchScheduleNotifyRules($id: ID!) {
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

interface ScheduleOnCallNotificationsCreateDialogProps {
  onClose: () => void
  scheduleID: string
}

export default function ScheduleOnCallNotificationsCreateDialog(
  props: ScheduleOnCallNotificationsCreateDialogProps,
): JSX.Element {
  const [{ data, error }] = useQuery<{ schedule: Schedule }>({
    query: rulesQuery,
    variables: { id: props.scheduleID },
  })
  if (error) throw error // should not happen

  const oldRules: FormValue[] = (
    data?.schedule?.onCallNotificationRules ?? []
  ).map(ruleToFormValue)

  const destTypes = useSchedOnCallNotifyTypes()
  const [value, setValue] = useState<FormValue>({
    time: null,
    weekdayFilter: NO_DAY,
    dest: {
      type: destTypes[0].type,
      values: [],
    },
  })

  const [status, commit] = useMutation<
    unknown,
    { input: SetScheduleOnCallNotificationRulesInput }
  >(rulesMutation)

  return (
    <FormDialog
      title='Create Notification Rule'
      errors={nonFieldErrors(status.error)}
      loading={status.fetching}
      onClose={props.onClose}
      onSubmit={() =>
        commit({
          input: {
            scheduleID: props.scheduleID,
            rules: oldRules.concat([ruleToFormValue(value)]),
          },
        }).then((res) => {
          if (res.error) return
          props.onClose()
        })
      }
      form={
        <ScheduleOnCallNotificationsForm
          scheduleID={props.scheduleID}
          errors={fieldErrors(status.error)}
          value={value}
          onChange={setValue}
        />
      }
    />
  )
}
