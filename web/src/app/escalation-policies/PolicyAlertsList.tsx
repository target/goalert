import React from 'react'
import AlertsList from '../alerts/AlertsList'
import AlertsListControls from '../alerts/components/AlertsListControls'

export default function ServiceAlerts(props: {
  policyID: string
}): JSX.Element {
  return (
    <AlertsList policyID={props.policyID} cardHeader={<AlertsListControls />} />
  )
}
