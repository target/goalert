import React, { useState } from 'react'
import { mapOnCallErrors, NO_DAY, Value } from './util'
import FormDialog from '../../dialogs/FormDialog'
import ScheduleOnCallNotificationsForm from './ScheduleOnCallNotificationsForm'
import { useOnCallRulesData, useSetOnCallRulesSubmit } from './hooks'

interface ScheduleOnCallNotificationsCreateDialogProps {
  onClose: () => void
  scheduleID: string
}

export default function ScheduleOnCallNotificationsCreateDialog(
  props: ScheduleOnCallNotificationsCreateDialogProps,
): JSX.Element {
  const { onClose, scheduleID } = props
  const [value, setValue] = useState<Value | null>(null)
  const [slackType, setSlackType] = useState('channel')

  const { q, zone, rules } = useOnCallRulesData(scheduleID)

  const newValue: Value = value || {
    time: null,
    weekdayFilter: NO_DAY,
    slackChannelID: null,
    slackUserGroup: null,
  }
  if (!newValue.slackChannelID) delete newValue.slackChannelID
  if (!newValue.slackUserGroup) delete newValue.slackUserGroup
  const { m, submit } = useSetOnCallRulesSubmit(
    scheduleID,
    zone,
    newValue,
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
          value={newValue}
          onChange={(value) => setValue(value)}
          slackType={slackType}
          setSlackType={setSlackType}
        />
      }
    />
  )
}
