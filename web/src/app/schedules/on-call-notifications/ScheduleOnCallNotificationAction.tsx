import React from 'react'
import ScheduleOnCallNotificationFormDialog from './ScheduleOnCallNotificationFormDialog'
import ScheduleOnCallNotificationDeleteDialog from './ScheduleOnCallNotificationDeleteDialog'
import { Rule } from './util'

interface ScheduleOnCallNotificationActionProps {
  scheduleID: string
  editRule?: Rule
  deleteRule?: Rule
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
          scheduleID={p.scheduleID}
          onClose={p.handleOnCloseEdit}
        />
      )}
      {p.deleteRule && (
        <ScheduleOnCallNotificationDeleteDialog
          rule={p.deleteRule}
          scheduleID={p.scheduleID}
          onClose={p.handleOnCloseDelete}
        />
      )}
    </React.Fragment>
  )
}
