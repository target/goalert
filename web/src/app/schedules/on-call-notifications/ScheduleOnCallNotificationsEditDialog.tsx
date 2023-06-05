import React, { useState } from 'react'

import { DateTime } from 'luxon'
import FormDialog from '../../dialogs/FormDialog'
import { useOnCallRulesData, useSetOnCallRulesSubmit } from './hooks'
import ScheduleOnCallNotificationsForm from './ScheduleOnCallNotificationsForm'
import { EVERY_DAY, mapOnCallErrors, NO_DAY, Value } from './util'

interface ScheduleOnCallNotificationsEditDialogProps {
  onClose: () => void

  scheduleID: string
  ruleID: string
}

export default function ScheduleOnCallNotificationsEditDialog(
  p: ScheduleOnCallNotificationsEditDialogProps,
): JSX.Element {
  const [value, setValue] = useState<Value | null>(null)

  const { q, zone, rules } = useOnCallRulesData(p.scheduleID)

  const rule = rules.find((r) => r.id === p.ruleID)
  const newValue: Value = value || {
    time: rule?.time
      ? DateTime.fromFormat(rule.time, 'HH:mm', { zone }).toISO()
      : null,
    weekdayFilter: rule?.time ? rule.weekdayFilter || EVERY_DAY : NO_DAY,
    type: rule?.target?.type ?? 'slackChannel',
    targetID: rule?.target?.id ?? null,
  }
  const { m, submit } = useSetOnCallRulesSubmit(
    p.scheduleID,
    zone,
    newValue,
    ...rules.filter((r) => r.id !== p.ruleID),
  )

  const [dialogErrors, fieldErrors] = mapOnCallErrors(m.error, q.error)

  return (
    <FormDialog
      title='Edit Notification Rule'
      errors={dialogErrors}
      loading={(q.loading && !zone) || m.loading}
      onClose={() => p.onClose()}
      onSubmit={() => submit().then(p.onClose)}
      form={
        <ScheduleOnCallNotificationsForm
          scheduleID={p.scheduleID}
          errors={fieldErrors}
          value={newValue}
          onChange={setValue}
        />
      }
    />
  )
}
