import React, { useState } from 'react'
import FormDialog from '../../dialogs/FormDialog'
import ScheduleOnCallNotificationsForm, {
  Value,
  errorPaths,
} from './ScheduleOnCallNotificationsFormDest'
import { NO_DAY } from './util'
import { useSchedOnCallNotifyTypes } from '../../util/RequireConfig'
import { splitErrorsByPath } from '../../util/errutil'
import { gql, useMutation, useQuery } from 'urql'
import {
  Schedule,
  SetScheduleOnCallNotificationRulesInput,
} from '../../../schema'
import { DateTime } from 'luxon'

const getRulesQuery = gql`
  query GetRules($scheduleID: ID!) {
    schedule(id: $scheduleID) {
      id
      onCallNotificationRules {
        id
        weekdayFilter
        time
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

const setRulesMut = gql`
  mutation SetRules($input: SetScheduleOnCallNotificationRulesInput!) {
    setScheduleOnCallNotificationRules(input: $input)
  }
`

interface ScheduleOnCallNotificationsCreateDialogProps {
  onClose: () => void
  scheduleID: string
  disablePortal?: boolean
}

export default function ScheduleOnCallNotificationsCreateDialog(
  props: ScheduleOnCallNotificationsCreateDialogProps,
): JSX.Element {
  const { onClose, scheduleID } = props
  const types = useSchedOnCallNotifyTypes()
  const [value, setValue] = useState<Value>({
    time: null,
    weekdayFilter: NO_DAY,
    dest: {
      type: types[0].type,
      values: [],
    },
  })
  const [q] = useQuery<{ schedule: Schedule }>({
    query: getRulesQuery,
    variables: { scheduleID },
  })
  if (q.error) throw q.error
  const sched = q.data?.schedule
  if (!sched) throw new Error('no data for schedule ' + scheduleID)
  const [m, commit] = useMutation(setRulesMut)

  const [formErrors, otherErrs] = splitErrorsByPath(
    m.error,
    errorPaths('createUserContactMethod.input'),
  )

  return (
    <FormDialog
      title='Create Notification Rule'
      errors={otherErrs}
      disablePortal={props.disablePortal}
      loading={m.fetching}
      onClose={onClose}
      onSubmit={() =>
        commit(
          {
            input: {
              scheduleID,
              rules: [
                ...sched.onCallNotificationRules,
                value.time
                  ? {
                      ...value,
                      time: DateTime.fromISO(value.time).toFormat('HH:mm'),
                    }
                  : { dest: value.dest },
              ],
            } satisfies SetScheduleOnCallNotificationRulesInput,
          },
          { additionalTypenames: ['Schedule'] },
        ).then(onClose)
      }
      form={
        <ScheduleOnCallNotificationsForm
          scheduleID={scheduleID}
          errors={formErrors}
          value={value}
          onChange={setValue}
        />
      }
    />
  )
}
