import React from 'react'
import gql from 'graphql-tag'
import Query from '../util/Query'
import ScheduleCalendar from './ScheduleCalendar'
import { urlParamSelector } from '../selectors'
import { connect } from 'react-redux'
import withWidth, { isWidthDown } from '@material-ui/core/withWidth/index'
import { getLuxonStartOfWeek } from '../util/luxon-helpers'
import { DateTime } from 'luxon'

const query = gql`
  query scheduleCalendarShifts(
    $id: ID!
    $start: ISOTimestamp!
    $end: ISOTimestamp!
  ) {
    schedule(id: $id) {
      id
      shifts(start: $start, end: $end) {
        user {
          id
          name
        }
        start
        end
        truncated
      }
    }
  }
`

const mapStateToProps = state => {
  // false: monthly, true: weekly
  const weekly = urlParamSelector(state)('weekly', false)
  const start = urlParamSelector(state)(
    'start',
    weekly
      ? getLuxonStartOfWeek()
          .toUTC()
          .toISO()
      : DateTime.local()
          .startOf('month')
          .toUTC()
          .toISO(),
  )

  const unitToAdd = weekly ? { weeks: 1 } : { months: 1 }
  const end = DateTime.fromISO(start)
    .plus(unitToAdd)
    .toUTC()
    .toISO()

  return {
    start,
    end,
  }
}

@withWidth()
@connect(mapStateToProps, null)
export default class ScheduleCalendarQuery extends React.PureComponent {
  render() {
    if (isWidthDown('sm', this.props.width)) return null

    return (
      <Query
        query={query}
        variables={{
          id: this.props.scheduleID,
          start: this.props.start,
          end: this.props.end,
        }}
        render={({ data }) => (
          <ScheduleCalendar
            scheduleID={data.schedule.id}
            shifts={data.schedule.shifts}
          />
        )}
      />
    )
  }
}
