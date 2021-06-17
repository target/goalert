import React, { useState } from 'react'

import { Value, useCreateRule } from './util'
import FormDialog from '../../dialogs/FormDialog'
import ScheduleOnCallNotificationsForm from './ScheduleOnCallNotificationsForm'

interface ScheduleOnCallNotificationsCreateDialogProps {
  onClose: () => void

  scheduleID: string
}

export default function ScheduleOnCallNotificationsCreateDialog(
  p: ScheduleOnCallNotificationsCreateDialogProps,
): JSX.Element {
  const [_value, setValue] = useState<Value | null>(null)
  const update = useCreateRule(p.scheduleID, _value)

  return (
    <FormDialog
      title='Create Notification Rule'
      errors={update.dialogErrors}
      loading={update.busy}
      onClose={() => p.onClose()}
      onSubmit={() => update.submit().then(p.onClose)}
      form={
        <ScheduleOnCallNotificationsForm
          scheduleID={p.scheduleID}
          errors={update.fieldErrors}
          value={update.value}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
