import React from 'react'
import Tabs from '@material-ui/core/Tabs'
import Tab from '@material-ui/core/Tab'
import { useURLParam } from '../../actions'

const tabs = ['active', 'unacknowledged', 'acknowledged', 'closed', 'all']

function AlertsListControls() {
  const [filter, setFilter] = useURLParam('filter', 'active')

  let currTab = tabs.indexOf(filter)
  if (currTab === -1) currTab = 0 // handle jargon input from url params

  return (
    <Tabs
      value={currTab}
      onChange={(e, idx) => setFilter(tabs[idx])}
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
