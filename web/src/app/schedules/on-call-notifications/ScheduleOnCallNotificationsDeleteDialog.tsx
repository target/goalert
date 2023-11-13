import React from 'react'

import FormDialog from '../../dialogs/FormDialog'
import { useDeleteOnCallRule } from './hooks'

interface ScheduleOnCallNotificationsDeleteDialogProps {
  scheduleID: string
  ruleID: string
  onClose: () => void
}

export default function ScheduleOnCallNotificationsDeleteDialog(
  p: ScheduleOnCallNotificationsDeleteDialogProps,
): React.ReactNode {
  const update = useDeleteOnCallRule(p.scheduleID, p.ruleID)

  if (!update.busy && !update.rule) {
    return (
      <FormDialog
        alert
        title='No longer exists'
        onClose={() => p.onClose()}
        subTitle='That notification rule does not exist or is already deleted.'
      />
    )
  }

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      loading={update.busy}
      errors={update.dialogErrors}
      subTitle={`${update.rule?.target?.name} will no longer be notified ${update.ruleSummary}.`}
      onSubmit={() => update.submit().then(p.onClose)}
      onClose={() => p.onClose()}
    />
  )
}
