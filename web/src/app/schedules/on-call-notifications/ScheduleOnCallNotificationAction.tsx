import React, { useState } from 'react'
import OtherActions from '../../util/OtherActions'
import ScheduleOnCallNotificationFormDialog from './ScheduleOnCallNotificationFormDialog'
import ScheduleOnCallNotificationDeleteDialog from './ScheduleOnCallNotificationDeleteDialog'
import { Rule } from './util'

interface ScheduleOnCallNotificationActionProps {
  rule: Rule
  scheduleID: string
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
          rule={p.rule}
        />
      )}
      {showDelete && (
        <ScheduleOnCallNotificationDeleteDialog
          rule={p.rule}
          scheduleID={p.scheduleID}
          onClose={() => setShowDelete(false)}
        />
      )}
    </React.Fragment>
  )
}
