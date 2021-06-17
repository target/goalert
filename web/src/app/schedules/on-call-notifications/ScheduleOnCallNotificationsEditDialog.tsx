import React, { useState } from 'react'

import { Value, useEditRule } from './util'
import FormDialog from '../../dialogs/FormDialog'
import ScheduleOnCallNotificationsForm from './ScheduleOnCallNotificationsForm'

interface ScheduleOnCallNotificationsEditDialogProps {
  onClose: () => void

  scheduleID: string
  ruleID: string
}

export default function ScheduleOnCallNotificationsEditDialog(
  p: ScheduleOnCallNotificationsEditDialogProps,
): JSX.Element {
  const [_value, setValue] = useState<Value | null>(null)
  const update = useEditRule(p.scheduleID, p.ruleID, _value)

  return (
    <FormDialog
      title='Edit Notification Rule'
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
