import React from 'react'
import { PropTypes as p } from 'prop-types'
import { Card, Button } from '@material-ui/core'
import Grid from '@material-ui/core/Grid'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import Switch from '@material-ui/core/Switch'
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
import { getStartOfWeek, getEndOfWeek } from '../util/luxon-helpers'
import LuxonLocalizer from '../util/LuxonLocalizer'
import { parseInterval, trimSpans } from '../util/shifts'
import _ from 'lodash'
import GroupAdd from '@material-ui/icons/GroupAdd'
import FilterContainer from '../util/FilterContainer'
import { UserSelect } from '../selection'

const localizer = LuxonLocalizer(DateTime, { firstDayOfWeek: 0 })

const styles = (theme) => ({
  calendarContainer: {
    padding: '1em',
  },
  card: {
    marginTop: 4,
  },
  filterBtn: {
    marginRight: theme.spacing(1.75),
  },
  tempSchedBtn: {
    marginLeft: theme.spacing(1.75),
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
    handleSetActiveOnly: (value) => dispatch(setURLParam('activeOnly', value)),
    handleSetUserFilter: (value) => dispatch(setURLParam('userFilter', value)),
    handleResetFilter: () =>
      dispatch(
        resetURLParams('userFilter', 'start', 'activeOnly', 'tz', 'duration'),
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

  /*
   * Offsets the calendar forward or backwards
   * a week or month, depending on the current
   * view type.
   */
  handleCalNavigate = (nextDate) => {
    if (this.props.weekly) {
      this.props.setStart(
        getStartOfWeek(DateTime.fromJSDate(nextDate)).toUTC().toISO(),
      )
    } else {
      this.props.setStart(
        DateTime.fromJSDate(nextDate)
          .toLocal()
          .startOf('month')
          .toUTC()
          .toISO(),
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
  handleViewChange = (nextView) => {
    const start = this.props.start
    const prevStartMonth = DateTime.fromISO(start).toLocal().month
    const currMonth = DateTime.local().month

    // if viewing the current month, show the current week
    if (nextView === 'week' && prevStartMonth === currMonth) {
      this.props.setWeekly(true)
      this.props.setStart(getStartOfWeek().toUTC().toISO())

      // if not on the current month, show the first week of the month
    } else if (nextView === 'week' && prevStartMonth !== currMonth) {
      this.props.setWeekly(true)
      this.props.setStart(
        DateTime.fromISO(this.props.start)
          .toLocal()
          .startOf('month')
          .toUTC()
          .toISO(),
      )

      // go from week to monthly view
      // e.g. if navigating to an overlap of two months such as
      // Jan 27 - Feb 2, show the latter month (February)
    } else {
      this.props.setWeekly(false)

      this.props.setStart(
        getEndOfWeek(DateTime.fromJSDate(new Date(start)))
          .toLocal()
          .startOf('month')
          .toUTC()
          .toISO(),
      )
    }
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
      activeOnly,
      handleSetActiveOnly,
      userFilter,
      handleSetUserFilter,
      handleResetFilter,
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
              onNavigate={this.handleCalNavigate}
              onView={this.handleViewChange}
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
                toolbar: (props) => (
                  <CalendarToolbar
                    date={props.date}
                    label={props.label}
                    onNavigate={props.onNavigate}
                    onView={props.onView}
                    view={props.view}
                    startAdornment={
                      <FilterContainer
                        onReset={handleResetFilter}
                        iconButtonProps={{
                          size: 'small',
                          className: classes.filterBtn,
                        }}
                      >
                        <Grid item xs={12}>
                          <FormControlLabel
                            control={
                              <Switch
                                checked={activeOnly}
                                onChange={(e) =>
                                  handleSetActiveOnly(e.target.checked)
                                }
                                value='activeOnly'
                              />
                            }
                            label='Active shifts only'
                          />
                        </Grid>
                        <Grid item xs={12}>
                          <UserSelect
                            label='Filter users...'
                            multiple
                            value={userFilter}
                            onChange={handleSetUserFilter}
                          />
                        </Grid>
                      </FilterContainer>
                    }
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
