import React, { useState } from 'react'
import FormDialog from '../../dialogs/FormDialog'
import ScheduleOnCallNotificationsForm from './ScheduleOnCallNotificationsForm'
import { useOnCallRulesData, useSetOnCallRulesSubmit } from './hooks'
import { NO_DAY, Value, mapOnCallErrors } from './util'
import { useSchedOnCallNotifyTypes } from '../../util/useDestinationTypes'

interface ScheduleOnCallNotificationsCreateDialogProps {
  onClose: () => void
  scheduleID: string
}

export default function ScheduleOnCallNotificationsCreateDialog(
  props: ScheduleOnCallNotificationsCreateDialogProps,
): JSX.Element {
  const { onClose, scheduleID } = props
  const destTypes = useSchedOnCallNotifyTypes()
  const [value, setValue] = useState<Value>({
    time: null,
    weekdayFilter: NO_DAY,
    dest: {
      type: destTypes[0].type,
      values: [],
    },
  })

  const { q, zone, rules } = useOnCallRulesData(scheduleID)

  const { m, submit } = useSetOnCallRulesSubmit(
    scheduleID,
    zone,
    value,
    ...rules,
  )

  const [dialogErrors, fieldErrors] = mapOnCallErrors(m.error, q.error)
  const busy = (q.loading && !zone) || m.loading

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
