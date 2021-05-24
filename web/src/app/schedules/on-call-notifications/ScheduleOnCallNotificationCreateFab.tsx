import React, { useState } from 'react'
import CreateFAB from '../../lists/CreateFAB'
import ScheduleOnCallNotificationFormDialog from './ScheduleOnCallNotificationFormDialog'

interface ScheduleOnCallNotificationCreateFabProps {
  scheduleID: string
}

export default function ScheduleOnCallNotificationCreateFab(
  p: ScheduleOnCallNotificationCreateFabProps,
): JSX.Element {
  const [showCreate, setShowCreate] = useState(false)

  return (
    <React.Fragment>
      <CreateFAB
        title='Create Notification Rule'
        onClick={() => setShowCreate(true)}
      />
      {showCreate && (
        <ScheduleOnCallNotificationFormDialog
          scheduleID={p.scheduleID}
          onClose={() => setShowCreate(false)}
        />
      )}
    </React.Fragment>
  )
}
