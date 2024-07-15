import React, { useEffect, useState } from 'react'
import FormDialog from '../../dialogs/FormDialog'
import ScheduleOnCallNotificationsForm, {
  Value,
} from './ScheduleOnCallNotificationsForm'
import { NO_DAY } from './util'
import { CombinedError, gql, useMutation, useQuery } from 'urql'
import {
  Schedule,
  SetScheduleOnCallNotificationRulesInput,
} from '../../../schema'
import { DateTime } from 'luxon'
import { useErrorConsumer } from '../../util/ErrorConsumer'

const getRulesQuery = gql`
  query GetRules($scheduleID: ID!) {
    schedule(id: $scheduleID) {
      id
      timeZone
      onCallNotificationRules {
        id
        weekdayFilter
        time
        dest {
          type
          args
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

interface ScheduleOnCallNotificationsEditDialogProps {
  onClose: () => void
  scheduleID: string
  ruleID: string
  disablePortal?: boolean
}

export default function ScheduleOnCallNotificationsEditDialog(
  props: ScheduleOnCallNotificationsEditDialogProps,
): JSX.Element {
  const { onClose, scheduleID } = props
  const [err, setErr] = useState<CombinedError | null>(null)
  const [q] = useQuery<{ schedule: Schedule }>({
    query: getRulesQuery,
    variables: { scheduleID },
  })
  if (q.error) throw q.error
  const sched = q.data?.schedule
  if (!sched) throw new Error('no data for schedule ' + scheduleID)
  const rule = sched.onCallNotificationRules.find((r) => r.id === props.ruleID)
  if (!rule) throw new Error('no rule for id ' + props.ruleID)

  const [value, setValue] = useState<Value>({
    time: rule.time || null,
    weekdayFilter: rule.weekdayFilter || NO_DAY,
    dest: {
      type: rule.dest.type,
      args: rule.dest.args,
    },
  })
  useEffect(() => {
    setErr(null)
  }, [value])

  const [m, commit] = useMutation(setRulesMut)
  useEffect(() => {
    setErr(m.error || null)
  }, [m.error])

  const errs = useErrorConsumer(err)

  const form = (
    <ScheduleOnCallNotificationsForm
      scheduleID={scheduleID}
      value={value}
      onChange={setValue}
      disabled={m.fetching}
      destTypeError={errs.getErrorByPath(
        /setScheduleOnCallNotificationRules.input.rules.+.dest.type/,
      )}
      destFieldErrors={errs.getErrorMap(
        /setScheduleOnCallNotificationRules.input.rules.+.dest(.args)?/,
      )}
    />
  )

  return (
    <FormDialog
      title='Edit Notification Rule'
      errors={errs.remainingLegacyCallback()}
      disablePortal={props.disablePortal}
      loading={m.fetching}
      onClose={onClose}
      onSubmit={() =>
        commit(
          {
            input: {
              scheduleID,
              rules: [
                ...sched.onCallNotificationRules.filter((r) => r !== rule),
                value.time
                  ? {
                      id: rule.id,
                      ...value,
                      time: DateTime.fromISO(value.time)
                        .setZone(sched.timeZone)
                        .toFormat('HH:mm'),
                    }
                  : { id: rule.id, dest: value.dest },
              ],
            } satisfies SetScheduleOnCallNotificationRulesInput,
          },
          { additionalTypenames: ['Schedule'] },
        ).then((res) => {
          if (res.error) {
            setErr(res.error)
            return
          }

          onClose()
        })
      }
      form={form}
    />
  )
}
