import React from 'react'
import Tabs from '@material-ui/core/Tabs'
import Tab from '@material-ui/core/Tab'
import { connect } from 'react-redux'
import { setAlertsStatusFilter } from '../../actions'
import { alertFilterSelector } from '../../selectors/url'

const mapStateToProps = state => ({
  filter: alertFilterSelector(state),
})

const mapDispatchToProps = dispatch => ({
  setAlertsStatusFilter: value => dispatch(setAlertsStatusFilter(value)),
})

const tabs = ['active', 'unacknowledged', 'acknowledged', 'closed', 'all']

@connect(
  mapStateToProps,
  mapDispatchToProps,
)
export default class AlertsListControls extends React.PureComponent {
  render() {
    let currTab = tabs.indexOf(this.props.filter)
    if (currTab === -1) currTab = 0 // handle jargon input from url params

    return (
      <Tabs
        value={currTab}
        onChange={(e, idx) => this.props.setAlertsStatusFilter(tabs[idx])}
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
}
