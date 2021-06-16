import React, { useState } from 'react'
import CreateFAB from '../../lists/CreateFAB'
import ScheduleOnCallNotificationsFormDialog from './ScheduleOnCallNotificationsFormDialog'

export default function ScheduleOnCallNotificationsCreateFab(): JSX.Element {
  const [showCreate, setShowCreate] = useState(false)

  return (
    <React.Fragment>
      <CreateFAB
        title='Create Notification Rule'
        onClick={() => setShowCreate(true)}
      />
      {showCreate && (
        <ScheduleOnCallNotificationsFormDialog
          onClose={() => setShowCreate(false)}
        />
      )}
    </React.Fragment>
  )
}
