import React from 'react'

export default function HeartbeatDetails(props) {
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
