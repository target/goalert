import React from 'react'
import { PropTypes as p } from 'prop-types'
import Card from '@material-ui/core/Card'
import Typography from '@material-ui/core/Typography'
import withStyles from '@material-ui/core/styles/withStyles'
import { connect } from 'react-redux'
import moment from 'moment'
import BigCalendar from 'react-big-calendar'
import '../../node_modules/react-big-calendar/lib/css/react-big-calendar.css'
import CalendarEventWrapper from './CalendarEventWrapper'
import CalendarToolbar from './CalendarToolbar'
import ScheduleOverrideCreateDialog from './ScheduleOverrideCreateDialog'
import { resetURLParams, setURLParam } from '../actions'
import { urlParamSelector } from '../selectors'
import { DateTime, Interval } from 'luxon'

const localizer = BigCalendar.momentLocalizer(moment)

const styles = {
  calendarContainer: {
    padding: '1em',
  },
}

const mapStateToProps = state => {
  // false: monthly, true: weekly
  const weekly = urlParamSelector(state)('weekly', false)
  let start = urlParamSelector(state)(
    'start',
    weekly
      ? moment()
          .startOf('week')
          .toISOString()
      : moment()
          .startOf('month')
          .toISOString(),
  )

  let end = moment(start)
    .add(1, weekly ? 'week' : 'month')
    .toISOString()

  return {
    start,
    end,
    weekly,
    activeOnly: urlParamSelector(state)('activeOnly', false),
    userFilter: urlParamSelector(state)('userFilter', []),
  }
}

const mapDispatchToProps = dispatch => {
  return {
    setWeekly: value => dispatch(setURLParam('weekly', value)),
    setStart: value => dispatch(setURLParam('start', value)),
    resetFilter: () =>
      dispatch(
        resetURLParams('userFilter', 'start', 'activeOnly', 'tz', 'weekly'),
      ),
  }
}

@withStyles(styles)
@connect(
  mapStateToProps,
  mapDispatchToProps,
)
export default class ScheduleCalendar extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,
    shifts: p.array.isRequired,
  }

  state = {
    /*
     * overrideDialog should be either an object of
     * the dialog properties to use, or null to close
     * the dialog.
     */
    overrideDialog: null,
  }

  /*
   * Offsets the calendar forward or backwards
   * a week or month, depending on the current
   * view type.
   */
  onNavigate = nextDate => {
    if (this.props.weekly) {
      this.props.setStart(
        moment(nextDate)
          .startOf('week')
          .startOf('day')
          .toISOString(),
      )
    } else {
      this.props.setStart(
        moment(nextDate)
          .startOf('month')
          .startOf('day')
          .toISOString(),
      )
    }
  }

  /*
   * Resets the start date to the beginning of the month
   * when switching views.
   *
   * e.g. Monthly: February -> Weekly: Start at the week
   * of February 1st
   *
   * e.g. Weekly: February 17-23 -> Monthly: Start at the
   * beginning of February
   *
   * If viewing the current month however, show the current
   * week.
   */
  onView = nextView => {
    const start = this.props.start
    const prevStartMonth = moment(start).month()
    const currMonth = moment().month()

    // if viewing the current month, show the current week
    if (nextView === 'week' && prevStartMonth === currMonth) {
      this.props.setWeekly(true)
      this.props.setStart(
        moment()
          .startOf('week')
          .toISOString(),
      )

      // if not on the current month, show the first week of the month
    } else if (nextView === 'week' && prevStartMonth !== currMonth) {
      this.props.setWeekly(true)
      this.props.setStart(
        moment(this.props.start)
          .startOf('month')
          .startOf('week')
          .toISOString(),
      )

      // go from week to monthly view
      // e.g. if navigating to an overlap of two months such as
      // Jan 27 - Feb 2, show the latter month (February)
    } else {
      this.props.setWeekly(false)

      this.props.setStart(
        moment(start)
          .endOf('week')
          .startOf('month')
          .toISOString(),
      )
    }
  }

  /*
   * Return a GoAlert dog red color for the events, and a slightly
   * darker version of that red if selected
   */
  eventStyleGetter = (event, start, end, isSelected) => {
    return {
      style: {
        backgroundColor: isSelected ? '#8f1022' : '#cd1831',
        borderColor: '#8f1022',
      },
    }
  }

  /*
   * Return a light red shade of the current date instead of
   * the default light blue
   */
  dayPropGetter = date => {
    if (moment(date).isSame(moment(), 'd')) {
      return {
        style: {
          backgroundColor: '#FFECEC',
        },
      }
    }
  }

  render() {
    const { classes, shifts, start, weekly } = this.props

    // fill available doesn't work in weekly view
    const height = weekly ? '100%' : '-webkit-fill-available'

    return (
      <React.Fragment>
        <Typography variant='caption' color='textSecondary'>
          <i>
            Times shown are in{' '}
            {Intl.DateTimeFormat().resolvedOptions().timeZone}
          </i>
        </Typography>
        <Card>
          <div className={classes.calendarContainer}>
            <BigCalendar
              date={new Date(start)}
              localizer={localizer}
              events={this.getCalEvents(shifts)}
              style={{ height, font: '-webkit-control' }}
              tooltipAccessor={() => null}
              views={['month', 'week']}
              view={weekly ? 'week' : 'month'}
              popup
              eventPropGetter={this.eventStyleGetter}
              dayPropGetter={this.dayPropGetter}
              onNavigate={this.onNavigate}
              onView={this.onView}
              components={{
                eventWrapper: props => (
                  <CalendarEventWrapper
                    onOverrideClick={overrideDialog =>
                      this.setState({ overrideDialog })
                    }
                    {...props}
                  />
                ),
                toolbar: props => (
                  <CalendarToolbar
                    onOverrideClick={() =>
                      this.setState({ overrideDialog: { variant: 'add' } })
                    }
                    {...props}
                  />
                ),
              }}
            />
          </div>
        </Card>
        {Boolean(this.state.overrideDialog) && (
          <ScheduleOverrideCreateDialog
            defaultValue={this.state.overrideDialog.defaultValue}
            variant={this.state.overrideDialog.variant}
            scheduleID={this.props.scheduleID}
            onClose={() => this.setState({ overrideDialog: null })}
          />
        )}
      </React.Fragment>
    )
  }

  getCalEvents = shifts => {
    // if any users in users array, only show the ids present
    let filteredShifts = shifts.slice()
    if (this.props.userFilter.length > 0) {
      filteredShifts = filteredShifts.filter(shift =>
        this.props.userFilter.includes(shift.user.id),
      )
    }

    if (this.props.activeOnly) {
      filteredShifts = filteredShifts.filter(shift =>
        Interval.fromDateTimes(
          DateTime.fromISO(shift.start),
          DateTime.fromISO(shift.end),
        ).contains(DateTime.local()),
      )
    }

    return filteredShifts.map(shift => {
      return {
        title: shift.user.name,
        userID: shift.user.id,
        start: new Date(shift.start),
        end: new Date(shift.end),
      }
    })
  }
}
