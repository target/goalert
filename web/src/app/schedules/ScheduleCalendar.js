import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { Card, Button, makeStyles } from '@material-ui/core'
import Grid from '@material-ui/core/Grid'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import Switch from '@material-ui/core/Switch'
import Typography from '@material-ui/core/Typography'
import { Calendar } from 'react-big-calendar'
import 'react-big-calendar/lib/css/react-big-calendar.css'
import CalendarEventWrapper, {
  EventHandlerContext,
} from './CalendarEventWrapper'
import CalendarToolbar from './CalendarToolbar'
import ScheduleOverrideCreateDialog from './ScheduleOverrideCreateDialog'
import { useResetURLParams, useURLParam } from '../actions'
import { DateTime, Interval } from 'luxon'
import { theme } from '../mui'
import LuxonLocalizer from '../util/LuxonLocalizer'
import { parseInterval, trimSpans } from '../util/shifts'
import _ from 'lodash'
import GroupAdd from '@material-ui/icons/GroupAdd'
import FilterContainer from '../util/FilterContainer'
import { UserSelect } from '../selection'
import SpinContainer from '../loading/components/SpinContainer'
import { useCalendarNavigation } from './hooks'

const localizer = LuxonLocalizer(DateTime, { firstDayOfWeek: 0 })

const useStyles = makeStyles((theme) => ({
  card: {
    padding: theme.spacing(2),
  },
  filterBtn: {
    marginRight: theme.spacing(1.75),
  },
  tempSchedBtn: {
    marginLeft: theme.spacing(1.75),
  },
}))

function ScheduleCalendar(props) {
  const classes = useStyles()
  const { weekly, start } = useCalendarNavigation()

  const [overrideDialog, setOverrideDialog] = useState(null)
  const [activeOnly, setActiveOnly] = useURLParam('activeOnly', false)
  const [userFilter, setUserFilter] = useURLParam('userFilter', [])
  const resetFilter = useResetURLParams('userFilter', 'activeOnly')

  const {
    shifts,
    temporarySchedules,
    onNewTempSched,
    onEditTempSched,
    onDeleteTempSched,
  } = props

  const eventStyleGetter = (event, start, end, isSelected) => {
    if (event.fixed) {
      return {
        style: {
          backgroundColor: isSelected ? '#094F13' : '#0C6618',
          borderColor: '#094F13',
        },
      }
    }
  }

  const getCalEvents = (shifts, _tempScheds) => {
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
    if (userFilter.length > 0) {
      filteredShifts = filteredShifts.filter((shift) =>
        userFilter.includes(shift.user.id),
      )
    }

    if (activeOnly) {
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

  return (
    <React.Fragment>
      <Typography variant='caption' color='textSecondary'>
        <i>
          Times shown are in {Intl.DateTimeFormat().resolvedOptions().timeZone}
        </i>
      </Typography>
      <Card className={classes.card} data-cy='calendar'>
        <CalendarToolbar
          filter={
            <FilterContainer
              onReset={resetFilter}
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
                      onChange={(e) => setActiveOnly(e.target.checked)}
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
                  onChange={setUserFilter}
                />
              </Grid>
            </FilterContainer>
          }
          endAdornment={
            <Button
              variant='contained'
              color='primary'
              data-cy='new-temp-sched'
              onClick={onNewTempSched}
              className={classes.tempSchedBtn}
              startIcon={<GroupAdd />}
              title='Make temporary change to schedule'
            >
              Temp Sched
            </Button>
          }
        />
        <SpinContainer loading={props.loading}>
          <EventHandlerContext.Provider
            value={{
              onEditTempSched,
              onDeleteTempSched,
              onOverrideClick: setOverrideDialog,
            }}
          >
            <Calendar
              date={DateTime.fromISO(start).toJSDate()}
              localizer={localizer}
              events={getCalEvents(shifts, temporarySchedules)}
              style={{
                height: weekly ? '100%' : '45rem',
                fontFamily: theme.typography.body2.fontFamily,
                fontSize: theme.typography.body2.fontSize,
              }}
              tooltipAccessor={() => null}
              views={['month', 'week']}
              view={weekly ? 'week' : 'month'}
              showAllEvents
              eventPropGetter={eventStyleGetter}
              onNavigate={() => {}} // stub to hide false console err
              onView={() => {}} // stub to hide false console err
              components={{
                eventWrapper: CalendarEventWrapper,
                toolbar: () => null,
              }}
            />
          </EventHandlerContext.Provider>
        </SpinContainer>
      </Card>
      {Boolean(overrideDialog) && (
        <ScheduleOverrideCreateDialog
          defaultValue={overrideDialog.defaultValue}
          variant={overrideDialog.variant}
          scheduleID={props.scheduleID}
          onClose={() => setOverrideDialog(null)}
          onChooseOverrideType={(override) => setOverrideDialog(override)}
          removeUserReadOnly
        />
      )}
    </React.Fragment>
  )
}

ScheduleCalendar.propTypes = {
  scheduleID: p.string.isRequired,
  shifts: p.array.isRequired,
  temporarySchedules: p.array,
  loading: p.bool,
}

export default ScheduleCalendar
