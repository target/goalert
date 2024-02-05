import React, { useState } from 'react'
import FormDialog from '../../dialogs/FormDialog'
import ScheduleOnCallNotificationsForm from './ScheduleOnCallNotificationsForm2'
import { useOnCallRulesData, useSetOnCallRulesSubmit } from './hooks'
import { NO_DAY, Value, mapOnCallErrors } from './util'

interface ScheduleOnCallNotificationsCreateDialog2Props {
  onClose: () => void
  scheduleID: string
}

const defaultValue: Value = {
  time: null,
  weekdayFilter: NO_DAY,
  type: 'slackChannel',
  targetID: null,
}

export default function ScheduleOnCallNotificationsCreateDialog2(
  props: ScheduleOnCallNotificationsCreateDialog2Props,
): JSX.Element {
  const { onClose, scheduleID } = props
  const [value, setValue] = useState<Value>(defaultValue)

  const { q, zone, rules } = useOnCallRulesData(scheduleID)

  const { m, submit } = useSetOnCallRulesSubmit(
    scheduleID,
    zone,
    value,
    ...rules,
  )

  const [dialogErrors, fieldErrors] = mapOnCallErrors(m.error, q.error)
  const busy = (q.fetching && !zone) || m.fetching

  return (
    <FormDialog
      title='Create Notification Rule'
      errors={dialogErrors}
      loading={busy}
      onClose={onClose}
      onSubmit={() => submit().then(onClose)}
      form={
        <ScheduleOnCallNotificationsForm
          scheduleID={scheduleID}
          errors={fieldErrors}
          value={value}
          onChange={setValue}
        />
      }
    />
  )
}
