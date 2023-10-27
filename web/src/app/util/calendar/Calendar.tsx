import React, { useContext } from 'react'
import { Card, Button } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { darken, lighten, useTheme, Theme } from '@mui/material/styles'
import Grid from '@mui/material/Grid'
import FormControlLabel from '@mui/material/FormControlLabel'
import Switch from '@mui/material/Switch'
import Typography from '@mui/material/Typography'
import { Calendar as RBCalendar } from 'react-big-calendar'
import 'react-big-calendar/lib/css/react-big-calendar.css'
import ScheduleCalendarToolbar from './CalendarToolbar'
import { useResetURLParams, useURLParam } from '../../actions'
import { DateTime, Interval } from 'luxon'
import LuxonLocalizer from '../LuxonLocalizer'
import { parseInterval, trimSpans } from '../shifts'
import _ from 'lodash'
import GroupAdd from '@mui/icons-material/GroupAdd'
import { AccountSwitch, AccountMinus, AccountPlus } from 'mdi-material-ui'
import FilterContainer from '../FilterContainer'
import { UserSelect } from '../../selection'
import SpinContainer from '../../loading/components/SpinContainer'
import { useCalendarNavigation } from './hooks'
import { OverrideDialogContext } from '../../schedules/ScheduleDetails'
import {
  OnCallShift,
  TemporarySchedule,
  User,
  UserOverride,
} from '../../../schema'
import ScheduleCalendarEventWrapper from './CalendarEventWrapper'

const localizer = LuxonLocalizer(DateTime, { firstDayOfWeek: 0 })

function getBorder(theme: Theme): string {
  if (theme.palette.mode === 'dark') {
    return '0.5px solid ' + lighten(theme.palette.background.paper, 0.2)
  }

  return '0.5px solid ' + darken(theme.palette.background.paper, 0.2)
}

const useStyles = makeStyles((theme: Theme) => ({
  calendar: {
    height: '45rem',
    fontFamily: theme.typography.body2.fontFamily,
    fontSize: theme.typography.body2.fontSize,
    '& .rbc-month-view, .rbc-header, .rbc-time-view, .rbc-timeslot-group': {
      border: getBorder(theme),
    },
    '& .rbc-month-row, .rbc-time-header, .rbc-time-header-content, .rbc-time-slot, .rbc-time-content, .rbc-events-container':
      {
        border: 'none',
      },
    // weekly current time divider line
    '& .rbc-time-content .rbc-current-time-indicator': {
      backgroundColor: theme.palette.primary.main,
    },
  },
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
  title: React.ReactNode
  start: Date
  end: Date
  user?: {
    name: string
    id: string
  }
  targetName?: string
  targetID?: string
}

export interface OnCallShiftEvent extends CalendarEvent {
  type: 'onCallShift'
  userID: string
  user?: User
  truncated: boolean
}

export interface OverrideEvent extends CalendarEvent {
  type: 'override'
  override: UserOverride
}

export interface TempSchedEvent extends CalendarEvent {
  type: 'tempSched'
  tempSched: TemporarySchedule
}

export interface TempSchedShiftEvent extends CalendarEvent {
  type: 'tempSchedShift'
  tempSched: TemporarySchedule
}

export type ScheduleCalendarEvent =
  | OnCallShiftEvent
  | OverrideEvent
  | TempSchedEvent
  | TempSchedShiftEvent

export interface Shift extends OnCallShift {
  targetName?: string
  targetID?: string
}

interface CalendarProps {
  scheduleID?: string
  shifts: Shift[]
  overrides?: UserOverride[]
  temporarySchedules?: TemporarySchedule[]
  loading: boolean
}

export default function Calendar(props: CalendarProps): JSX.Element {
  const classes = useStyles()
  const theme = useTheme()

  const { scheduleID, shifts, temporarySchedules } = props
  const { weekly, start } = useCalendarNavigation()
  const { setOverrideDialog } = useContext(OverrideDialogContext)

  const [activeOnly, setActiveOnly] = useURLParam<boolean>('activeOnly', false)
  const [userFilter, setUserFilter] = useURLParam<string[]>('userFilter', [])
  const resetFilter = useResetURLParams('userFilter', 'activeOnly')

  function eventStyleGetter(
    event: ScheduleCalendarEvent,
  ): React.HTMLAttributes<HTMLDivElement> {
    if (event.type === 'tempSched' || event.type === 'override') {
      return {
        style: {
          backgroundColor: theme.palette.secondary.main,
          color: theme.palette.getContrastText(theme.palette.secondary.main),
        },
      }
    }

    return {
      style: {
        backgroundColor: theme.palette.primary.main,
        color: theme.palette.getContrastText(theme.palette.primary.main),
      },
    }
  }

  function dayStyleGetter(date: Date): React.HTMLAttributes<HTMLDivElement> {
    const outOfBounds =
      DateTime.fromISO(start).month !== DateTime.fromJSDate(date).month
    const currentDay = DateTime.local().hasSame(
      DateTime.fromJSDate(date),
      'day',
    )

    if (theme.palette.mode === 'dark' && (outOfBounds || currentDay)) {
      return {
        style: {
          backgroundColor: lighten(theme.palette.background.paper, 0.1),
          border: getBorder(theme),
        },
      }
    }

    if (outOfBounds || currentDay) {
      return {
        style: {
          backgroundColor: darken(theme.palette.background.paper, 0.1),
          border: getBorder(theme),
        },
      }
    }

    return {
      style: {
        border: getBorder(theme),
      },
    }
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
    // remove override
    return (
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
    shifts: Shift[] = [],
    _tempScheds: TemporarySchedule[] = [],
    userOverrides: UserOverride[] = [],
  ): ScheduleCalendarEvent[] => {
    const tempSchedules: TempSchedEvent[] = _tempScheds.map((sched) => ({
      type: 'tempSched',
      start: new Date(sched.start),
      end: new Date(sched.end),
      title: 'Temporary Schedule',
      tempSched: sched,
    }))

    const overrides: OverrideEvent[] = userOverrides.map((o) => ({
      type: 'override',
      start: new Date(o.start),
      end: new Date(o.end),
      title: getOverrideTitle(o),
      override: o,
    }))

    const tempSchedShifts: TempSchedShiftEvent[] = _.flatten(
      _tempScheds.map((sched) => {
        return sched.shifts.map((s) => ({
          ...s,
          type: 'tempSchedShift',
          start: new Date(s.start),
          end: new Date(s.end),
          title: s.user?.name ?? '',
          tempSched: sched,
          user: s.user ?? undefined,
        }))
      }),
    )

    const fixedIntervals = tempSchedules.map((t) =>
      parseInterval(
        { start: t.start.toISOString(), end: t.end.toISOString() },
        'local',
      ),
    )

    // Remove shifts within a temporary schedule, and trim any that overlap
    const onCallShiftEvents: OnCallShiftEvent[] = trimSpans(
      shifts,
      fixedIntervals,
      'local',
    ).map((s) => ({
      ...s,
      start: new Date(s.start),
      end: new Date(s.end),
      type: 'onCallShift',
      title: s.user?.name ?? s.targetName ?? '',
      user: s.user ?? undefined,
    }))

    let filteredShifts: ScheduleCalendarEvent[] = [
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
          Interval.fromDateTimes(shift.start, shift.end).contains(
            DateTime.local(),
          ),
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
              {scheduleID && (
                <Grid item xs={12}>
                  <UserSelect
                    label='Filter users...'
                    multiple
                    value={userFilter}
                    onChange={setUserFilter}
                  />
                </Grid>
              )}
            </FilterContainer>
          }
          endAdornment={
            scheduleID && (
              <Button
                variant='contained'
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
            )
          }
        />
        <SpinContainer loading={props.loading}>
          <RBCalendar
            date={DateTime.fromISO(start).toJSDate()}
            localizer={localizer}
            events={getCalEvents(shifts, temporarySchedules, props.overrides)}
            className={classes.calendar}
            tooltipAccessor={() => ''}
            views={['month', 'week']}
            view={weekly ? 'week' : 'month'}
            showAllEvents
            eventPropGetter={eventStyleGetter}
            dayPropGetter={dayStyleGetter}
            onNavigate={() => {}} // stub to hide false console err
            onView={() => {}} // stub to hide false console err
            components={{
              // @ts-expect-error Property 'children' does not exist on type - yes it does
              eventWrapper: ({ children, event }) => (
                <ScheduleCalendarEventWrapper
                  event={event}
                  scheduleID={scheduleID}
                >
                  {children}
                </ScheduleCalendarEventWrapper>
              ),
              toolbar: () => null,
            }}
          />
        </SpinContainer>
      </Card>
    </React.Fragment>
  )
}
