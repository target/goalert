import React from 'react'
import { PropTypes as p } from 'prop-types'
import { Card, Button } from '@material-ui/core'
import Typography from '@material-ui/core/Typography'
import withStyles from '@material-ui/core/styles/withStyles'
import { connect } from 'react-redux'
import { Calendar } from 'react-big-calendar'
import 'react-big-calendar/lib/css/react-big-calendar.css'
import CalendarEventWrapper from './CalendarEventWrapper'
import CalendarToolbar from './CalendarToolbar'
import ScheduleOverrideCreateDialog from './ScheduleOverrideCreateDialog'
import { resetURLParams, setURLParam } from '../actions'
import { urlParamSelector } from '../selectors'
import { DateTime, Interval } from 'luxon'
import { theme } from '../mui'
import { getStartOfWeek } from '../util/luxon-helpers'
import LuxonLocalizer from '../util/LuxonLocalizer'
import { parseInterval, trimSpans } from '../util/shifts'
import _ from 'lodash'
import GroupAdd from '@material-ui/icons/GroupAdd'

const localizer = LuxonLocalizer(DateTime, { firstDayOfWeek: 0 })

const styles = (theme) => ({
  calendarContainer: {
    padding: '1em',
  },
  card: {
    marginTop: 4,
  },
  tempSchedBtn: {
    marginLeft: theme.spacing(1),
  },
})

const mapStateToProps = (state) => {
  // false: monthly, true: weekly
  const weekly = urlParamSelector(state)('weekly', false)
  const start = urlParamSelector(state)(
    'start',
    weekly
      ? getStartOfWeek().toUTC().toISO()
      : DateTime.local().startOf('month').toUTC().toISO(),
  )

  const timeUnit = weekly ? { weeks: 1 } : { months: 1 }

  const end = DateTime.fromISO(start).toLocal().plus(timeUnit).toUTC().toISO()

  return {
    start,
    end,
    weekly,
    activeOnly: urlParamSelector(state)('activeOnly', false),
    userFilter: urlParamSelector(state)('userFilter', []),
  }
}

const mapDispatchToProps = (dispatch) => {
  return {
    setWeekly: (value) => dispatch(setURLParam('weekly', value)),
    setStart: (value) => dispatch(setURLParam('start', value)),
    resetFilter: () =>
      dispatch(
        resetURLParams('userFilter', 'start', 'activeOnly', 'tz', 'weekly'),
      ),
  }
}

@withStyles(styles)
@connect(mapStateToProps, mapDispatchToProps)
export default class ScheduleCalendar extends React.PureComponent {
  static propTypes = {
    scheduleID: p.string.isRequired,
    shifts: p.array.isRequired,
    temporarySchedules: p.array,
    CardProps: p.object, // todo: use CardProps from types once TS
  }

  state = {
    /*
     * overrideDialog should be either an object of
     * the dialog properties to use, or null to close
     * the dialog.
     */
    overrideDialog: null,
  }

  eventStyleGetter = (event, start, end, isSelected) => {
    if (event.fixed) {
      return {
        style: {
          backgroundColor: isSelected ? '#094F13' : '#0C6618',
          borderColor: '#094F13',
        },
      }
    }
  }

  render() {
    const {
      classes,
      shifts,
      temporarySchedules,
      start,
      weekly,
      CardProps,
      onNewTempSched,
      onEditTempSched,
      onDeleteTempSched,
    } = this.props

    return (
      <React.Fragment>
        <Typography variant='caption' color='textSecondary'>
          <i>
            Times shown are in{' '}
            {Intl.DateTimeFormat().resolvedOptions().timeZone}
          </i>
        </Typography>
        <Card className={classes.card} {...CardProps}>
          <div data-cy='calendar' className={classes.calendarContainer}>
            <CalendarToolbar
              endAdornment={
                <Button
                  variant='contained'
                  size='small'
                  color='primary'
                  data-cy='new-temp-sched'
                  onClick={onNewTempSched}
                  className={classes.tempSchedBtn}
                  startIcon={<GroupAdd />}
                  title='Make temporary change to this schedule'
                >
                  Temp Sched
                </Button>
              }
            />
            <Calendar
              date={new Date(start)}
              localizer={localizer}
              events={this.getCalEvents(shifts, temporarySchedules)}
              style={{
                height: weekly ? '100%' : '45rem',
                fontFamily: theme.typography.body2.fontFamily,
                fontSize: theme.typography.body2.fontSize,
              }}
              tooltipAccessor={() => null}
              views={['month', 'week']}
              view={weekly ? 'week' : 'month'}
              showAllEvents
              eventPropGetter={this.eventStyleGetter}
              onNavigate={() => {}} // stub to hide false console err
              onView={() => {}} // stub to hide false console err
              components={{
                eventWrapper: (props) => (
                  <CalendarEventWrapper
                    onOverrideClick={(overrideDialog) =>
                      this.setState({ overrideDialog })
                    }
                    onEditTempSched={onEditTempSched}
                    onDeleteTempSched={onDeleteTempSched}
                    {...props}
                  />
                ),
                toolbar: () => null,
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
            removeUserReadOnly
          />
        )}
      </React.Fragment>
    )
  }

  getCalEvents = (shifts, _tempScheds) => {
    const tempSchedules = _tempScheds.map((sched) => ({
      start: sched.start,
      end: sched.end,
      user: { name: 'Temporary Schedule' },
      tempSched: sched,
      fixed: true,
    }))

    // flat list of all fixed shifts, with `fixed` set to true
    const fixedShifts = _.flatten(
      _tempScheds.map((sched) => {
        return sched.shifts.map((s) => ({
          ...s,
          tempSched: sched,
          fixed: true,
          isTempSchedShift: true,
        }))
      }),
    )

    const fixedIntervals = tempSchedules.map(parseInterval)
    let filteredShifts = [
      ...tempSchedules,
      ...fixedShifts,

      // Remove shifts within a temporary schedule, and trim any that overlap
      ...trimSpans(shifts, ...fixedIntervals),
    ]

    // if any users in users array, only show the ids present
    if (this.props.userFilter.length > 0) {
      filteredShifts = filteredShifts.filter((shift) =>
        this.props.userFilter.includes(shift.user.id),
      )
    }

    if (this.props.activeOnly) {
      filteredShifts = filteredShifts.filter(
        (shift) =>
          shift.TempSched ||
          Interval.fromDateTimes(
            DateTime.fromISO(shift.start),
            DateTime.fromISO(shift.end),
          ).contains(DateTime.local()),
      )
    }

    return filteredShifts.map((shift) => {
      return {
        title: shift.user.name,
        userID: shift.user.id,
        start: new Date(shift.start),
        end: new Date(shift.end),
        fixed: shift.fixed,
        isTempSchedShift: shift.isTempSchedShift,
        tempSched: shift.tempSched,
      }
    })
  }
}
