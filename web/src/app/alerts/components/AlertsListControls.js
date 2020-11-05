import React from 'react'
import Tabs from '@material-ui/core/Tabs'
import Tab from '@material-ui/core/Tab'
import { useDispatch, useSelector } from 'react-redux'
import { setAlertsStatusFilter } from '../../actions'
import { alertFilterSelector } from '../../selectors/url'

const tabs = ['active', 'unacknowledged', 'acknowledged', 'closed', 'all']

function AlertsListControls() {
  const dispatch = useDispatch()
  const filter = useSelector(alertFilterSelector)

  let currTab = tabs.indexOf(filter)
  if (currTab === -1) currTab = 0 // handle jargon input from url params

  return (
    <Tabs
      value={currTab}
      onChange={(e, idx) => dispatch(setAlertsStatusFilter(tabs[idx]))}
      centered
      indicatorColor='primary'
      textColor='primary'
    >
      <Tab label='Active' />
      <Tab label='Unacknowledged' />
      <Tab label='Acknowledged' />
      <Tab label='Closed' />
      <Tab label='All' />
    </Tabs>
  )
}

export default AlertsListControls
