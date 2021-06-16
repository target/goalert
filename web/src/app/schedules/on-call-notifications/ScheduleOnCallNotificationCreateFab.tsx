import React, { useState } from 'react'
import CreateFAB from '../../lists/CreateFAB'
import ScheduleOnCallNotificationFormDialog from './ScheduleOnCallNotificationFormDialog'

export default function ScheduleOnCallNotificationCreateFab(): JSX.Element {
  const [showCreate, setShowCreate] = useState(false)

  return (
    <React.Fragment>
      <CreateFAB
        title='Create Notification Rule'
        onClick={() => setShowCreate(true)}
      />
      {showCreate && (
        <ScheduleOnCallNotificationFormDialog
          onClose={() => setShowCreate(false)}
        />
      )}
    </React.Fragment>
  )
}
