import React, { useState } from 'react'
import { useMutation } from '@apollo/client'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'
import { setMutation } from './ScheduleOnCallNotifications'

interface ScheduleOnCallNotificationFormProps {
  id?: string
  scheduleID: string
  rules?: Array<string>
  onClose: () => void
}

export default function ScheduleOnCallNotificationFormDialog(
  p: ScheduleOnCallNotificationFormProps,
): JSX.Element {
  const [rules, setRules] = useState(p?.rules ?? [])
  const [mutate, mutationStatus] = useMutation(setMutation, {
    variables: {
      scheduleID: p.scheduleID,
      rules,
    },
  })

  return (
    <FormDialog
      title={(p.id ? 'Edit ' : 'Create ') + 'Notification Rule'}
      errors={nonFieldErrors(mutationStatus.error)}
      onClose={() => p.onClose()}
      onSubmit={() => mutate()}
    />
  )
}
