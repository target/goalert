import React, { useState } from 'react'
import OtherActions from '../../util/OtherActions'
import ScheduleOnCallNotificationFormDialog from './ScheduleOnCallNotificationFormDialog'
import ScheduleOnCallNotificationDeleteDialog from './ScheduleOnCallNotificationDeleteDialog'
import { WeekdayFilter } from '../../../schema'

interface ScheduleOnCallNotificationActionProps {
  rule: Rule
  scheduleID: string
}

export type Rule = {
  id: string
  target: {
    id: string
    type: string
    name: string
  }
  time: string
  weekdayFilter: WeekdayFilter
}

export default function ScheduleOnCallNotificationAction(
  p: ScheduleOnCallNotificationActionProps,
): JSX.Element {
  const [showEdit, setShowEdit] = useState(false)
  const [showDelete, setShowDelete] = useState(false)

  return (
    <React.Fragment>
      <OtherActions
        actions={[
          {
            label: 'Edit',
            onClick: () => setShowEdit(true),
          },
          { label: 'Delete', onClick: () => setShowDelete(true) },
        ]}
      />
      {showEdit && (
        <ScheduleOnCallNotificationFormDialog
          scheduleID={p.scheduleID}
          onClose={() => setShowEdit(false)}
        />
      )}
      {showDelete && (
        <ScheduleOnCallNotificationDeleteDialog
          id={p.rule.id}
          scheduleID={p.scheduleID}
          onClose={() => setShowDelete(false)}
        />
      )}
    </React.Fragment>
  )
}
