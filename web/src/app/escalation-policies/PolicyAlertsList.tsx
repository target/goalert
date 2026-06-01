import React from 'react'
import AlertsList from '../alerts/AlertsList'

export default function ServiceAlerts(props: {
  policyID: string
}): JSX.Element {
  return <AlertsList policyID={props.policyID} />
}
