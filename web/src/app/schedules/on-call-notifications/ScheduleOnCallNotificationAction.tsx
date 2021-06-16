import React from 'react'
import ScheduleOnCallNotificationFormDialog from './ScheduleOnCallNotificationFormDialog'
import ScheduleOnCallNotificationDeleteDialog from './ScheduleOnCallNotificationDeleteDialog'
import { Rule } from './util'

interface ScheduleOnCallNotificationActionProps {
  editRule: Rule | null
  deleteRule: Rule | null
  handleOnCloseEdit: () => void
  handleOnCloseDelete: () => void
}

export default function ScheduleOnCallNotificationAction(
  p: ScheduleOnCallNotificationActionProps,
): JSX.Element {
  return (
    <React.Fragment>
      {p.editRule && (
        <ScheduleOnCallNotificationFormDialog
          rule={p.editRule}
          onClose={p.handleOnCloseEdit}
        />
      )}
      {p.deleteRule && (
        <ScheduleOnCallNotificationDeleteDialog
          rule={p.deleteRule}
          onClose={p.handleOnCloseDelete}
        />
      )}
    </React.Fragment>
  )
}
