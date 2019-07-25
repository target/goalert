import React, { useState } from 'react'
import p from 'prop-types'
import HeartbeatDeleteDialog from './HeartbeatDeleteDialog'
import OtherActions from '../util/OtherActions'

export default function HeartbeatMonitorListItem(props) {
  return (
    <React.Fragment>
      <div>
        Sends an alert if no heartbeat is received {props.timeoutMinutes}{' '}
        minutes after the last reported time.
      </div>
      <div>Last known state: {props.lastState}</div>
      <div>Last report time: {props.lastHeartbeatTime}</div>
    </React.Fragment>
  )
}

HeartbeatMonitorListItem.propTypes = {
  timeoutMinutes: p.number.isRequired,
  lastState: p.string.isRequired,
  lastHeartbeatTime: p.string.isRequired,
}

export function HeartbeatMonitorListItemActions(props) {
  // const [showEditDialog, setShowEditDialog] = useState(false)
  const [showDeleteDialog, setShowDeleteDialog] = useState(false)

  return (
    <React.Fragment>
      <OtherActions
        actions={[
          // {
          //   label: 'Edit',
          //   onClick: () => setShowEditDialog(true),
          // },
          {
            label: 'Delete',
            onClick: () => setShowDeleteDialog(true),
          },
        ]}
      />
      {showDeleteDialog && (
        <HeartbeatDeleteDialog
          heartbeatID={props.monitorID}
          onClose={() => setShowDeleteDialog(false)}
        />
      )}
    </React.Fragment>
  )
}

HeartbeatMonitorListItemActions.propTypes = {
  monitorID: p.string.isRequired,
  refetchQueries: p.arrayOf(p.string),
}
