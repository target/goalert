import React, { useContext } from 'react'
import { Card, Button, makeStyles } from '@material-ui/core'
import { darken } from '@material-ui/core/styles'
import Grid from '@material-ui/core/Grid'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import Switch from '@material-ui/core/Switch'
import Typography from '@material-ui/core/Typography'
import { Calendar } from 'react-big-calendar'
import 'react-big-calendar/lib/css/react-big-calendar.css'
import ScheduleCalendarEventWrapper, {
  ScheduleCalendarEvent,
} from './ScheduleCalendarEventWrapper'
import ScheduleCalendarToolbar from './ScheduleCalendarToolbar'
import { useResetURLParams, useURLParam } from '../../actions'
import { DateTime, Interval } from 'luxon'
import { theme } from '../../mui'
import LuxonLocalizer from '../../util/LuxonLocalizer'
import { parseInterval, trimSpans } from '../../util/shifts'
import _ from 'lodash'
import GroupAdd from '@material-ui/icons/GroupAdd'
import { AccountSwitch, AccountMinus, AccountPlus } from 'mdi-material-ui'
import FilterContainer from '../../util/FilterContainer'
import { UserSelect } from '../../selection'
import SpinContainer from '../../loading/components/SpinContainer'
import { useCalendarNavigation } from './hooks'
import { ScheduleCalendarContext } from '../ScheduleDetails'
import {
  OnCallShift,
  TemporarySchedule,
  User,
  UserOverride,
} from '../../../schema'

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
  overrideTitleIcon: {
    verticalAlign: 'middle',
    borderRadius: '50%',
    background: theme.palette.secondary.main,
    padding: '3px',
    height: '100%',
    width: '18px',
    marginRight: '0.25rem',
  },
}))

interface CalendarEvent {
  start: string
  end: string
  user?: {
    name?: React.ReactNode
    id?: string
  }
  type: 'tempSched' | 'overrideShift' | 'tempSchedShift' | 'onCallShift'
  // fixed: if is 'tempSched' or 'tempSchedShift'
}

interface TempSchedEvent extends CalendarEvent {
  type: 'tempSched'
  tempSched: TemporarySchedule
}

interface OverrideShiftEvent extends CalendarEvent {
  type: 'overrideShift'
  override: UserOverride
}

interface TempSchedShiftEvent extends CalendarEvent {
  type: 'tempSchedShift'
  tempSched: TemporarySchedule
}

interface OnCallShiftEvent extends CalendarEvent {
  type: 'onCallShift'
  userID: string
  user?: User
  truncated: boolean
}

interface X extends CalendarEvent, TempSchedShiftEvent, OnCallShiftEvent {}

interface ScheduleCalendarProps {
  scheduleID: string
  shifts: OnCallShift[]
  overrides: UserOverride[]
  temporarySchedules: TemporarySchedule[]
  loading: boolean
}

function ScheduleCalendar(props: ScheduleCalendarProps): JSX.Element {
  const classes = useStyles()

  const { setOverrideDialog } = useContext(ScheduleCalendarContext)

  const { weekly, start } = useCalendarNavigation()

  const [activeOnly, setActiveOnly] = useURLParam<boolean>('activeOnly', false)
  const [userFilter, setUserFilter] = useURLParam<string[]>('userFilter', [])
  const resetFilter = useResetURLParams('userFilter', 'activeOnly')

  const { shifts, temporarySchedules } = props

  const eventStyleGetter = (
    event: ScheduleCalendarEvent,
    start: Date,
    end: Date,
    isSelected: boolean,
  ): { className?: string; style?: Object } => {
    const green = '#0C6618'
    const lavender = '#BB7E8C'

    if (event.fixed) {
      return {
        style: {
          backgroundColor: isSelected ? darken(green, 0.3) : green,
          borderColor: darken(green, 0.3),
        },
      }
    }
    if (event.isOverride) {
      return {
        style: {
          backgroundColor: isSelected ? darken(lavender, 0.3) : lavender,
          borderColor: darken(lavender, 0.3),
        },
      }
    }

    return {}
  }

  const getOverrideTitle = (o: UserOverride): JSX.Element => {
    if (o.addUser && o.removeUser) {
      // replace override
      return (
        <div>
          <AccountSwitch
            fontSize='small'
            className={classes.overrideTitleIcon}
            aria-label='Replace Override'
          />
          Override
        </div>
      )
    }
    if (o.addUser) {
      // add override
      return (
        <div>
          <AccountPlus
            fontSize='small'
            className={classes.overrideTitleIcon}
            aria-label='Add Override'
          />
          Override
        </div>
      )
    }
    return (
      // remove override
      <div>
        <AccountMinus
          fontSize='small'
          className={classes.overrideTitleIcon}
          aria-label='Remove Override'
        />
        Override
      </div>
    )
  }

  const getCalEvents = (
    shifts: OnCallShift[],
    _tempScheds: TemporarySchedule[],
    userOverrides: UserOverride[],
  ): ScheduleCalendarEvent[] => {
    const tempSchedules: TempSchedEvent[] = _tempScheds.map((sched) => ({
      type: 'tempSched',
      start: sched.start,
      end: sched.end,
      user: { name: 'Temporary Schedule' },
      tempSched: sched,
    }))

    const overrides: OverrideShiftEvent[] = userOverrides.map((o) => ({
      type: 'overrideShift',
      user: {
        name: getOverrideTitle(o),
      },
      start: o.start,
      end: o.end,
      override: o,
    }))

    // flat list of all fixed shifts, with `fixed` set to true
    const tempSchedShifts: TempSchedShiftEvent[] = _.flatten(
      _tempScheds.map((sched) => {
        return sched.shifts.map((s) => ({
          ...s,
          type: 'tempSchedShift',
          tempSched: sched,
        }))
      }),
    )

    const fixedIntervals = tempSchedules.map((t) => parseInterval(t, 'local'))

    // Remove shifts within a temporary schedule, and trim any that overlap
    const onCallShiftEvents: OnCallShiftEvent[] = trimSpans(
      shifts,
      fixedIntervals,
      'local',
    ).map((res) => ({ ...res, type: 'onCallShift' }))

    let filteredShifts: CalendarEvent[] = [
      ...tempSchedules,
      ...tempSchedShifts,
      ...overrides,
      ...onCallShiftEvents,
    ]

    // if any users in users array, only show the ids present
    if (userFilter.length > 0) {
      filteredShifts = filteredShifts.filter((shift) =>
        shift?.user?.id ? userFilter.includes(shift.user.id) : false,
      )
    }

    if (activeOnly) {
      filteredShifts = filteredShifts.filter(
        (shift) =>
          shift.type === 'tempSched' ||
          shift.type === 'tempSchedShift' ||
          Interval.fromDateTimes(
            DateTime.fromISO(shift.start),
            DateTime.fromISO(shift.end),
          ).contains(DateTime.local()),
      )
    }

    return filteredShifts
  }

  return (
    <React.Fragment>
      <Typography variant='caption' color='textSecondary'>
        <i>
          Times shown are in {Intl.DateTimeFormat().resolvedOptions().timeZone}
        </i>
      </Typography>
      <Card className={classes.card} data-cy='calendar'>
        <ScheduleCalendarToolbar
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
              data-cy='new-override'
              onClick={() =>
                setOverrideDialog({
                  variantOptions: ['replace', 'remove', 'add', 'temp'],
                  removeUserReadOnly: false,
                })
              }
              className={classes.tempSchedBtn}
              startIcon={<GroupAdd />}
              title='Make temporary change to schedule'
            >
              Override
            </Button>
          }
        />
        <SpinContainer loading={props.loading}>
          <Calendar
            date={DateTime.fromISO(start).toJSDate()}
            localizer={localizer}
            events={getCalEvents(shifts, temporarySchedules, props.overrides)}
            style={{
              height: weekly ? '100%' : '45rem',
              fontFamily: theme.typography.body2.fontFamily,
              fontSize: theme.typography.body2.fontSize,
            }}
            // tooltipAccessor={() => undefined}
            views={['month', 'week']}
            view={weekly ? 'week' : 'month'}
            showAllEvents
            // eventPropGetter={eventStyleGetter}
            onNavigate={() => {}} // stub to hide false console err
            onView={() => {}} // stub to hide false console err
            components={{
              eventWrapper: ScheduleCalendarEventWrapper,
              toolbar: () => null,
            }}
          />
        </SpinContainer>
      </Card>
    </React.Fragment>
  )
}

export default ScheduleCalendar
