import React from 'react'
import p from 'prop-types'
import { connect } from 'react-redux'
import { urlParamSelector } from '../selectors'
import { setURLParam } from '../actions'
import Query from '../util/Query'
import gql from 'graphql-tag'
import { FormControlLabel, Switch } from '@material-ui/core'
import { oneOfShape } from '../util/propTypes'

const tzQuery = gql`
  query($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
    }
  }
`

@connect(
  state => ({ zone: urlParamSelector(state)('tz', 'local') }),
  dispatch => ({
    setZone: value => dispatch(setURLParam('tz', value, 'local')),
  }),
)
export class ScheduleTZFilter extends React.PureComponent {
  static propTypes = {
    label: p.func,

    // one of scheduleID or scheduleTimeZone must be specified
    _tz: oneOfShape({
      scheduleID: p.string,
      scheduleTimeZone: p.string,
    }),

    // provided by connect
    zone: p.string,
    setZone: p.func,
  }
  render() {
    const { scheduleID, scheduleTimeZone } = this.props
    if (scheduleTimeZone) return this.renderControl(scheduleTimeZone)

    return (
      <Query
        variables={{ id: scheduleID }}
        query={tzQuery}
        noPoll
        render={({ data }) => this.renderControl(data.schedule.timeZone)}
      />
    )
  }

  renderControl(tz) {
    const { zone, label, setZone } = this.props
    return (
      <FormControlLabel
        control={
          <Switch
            checked={zone !== 'local'}
            onChange={e => setZone(e.target.checked ? tz : 'local')}
            value={tz}
          />
        }
        label={label ? label(tz) : `Show times in ${tz}`}
      />
    )
  }
}
