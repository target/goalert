import React, { ReactNode } from 'react'
import IconButton from '@material-ui/core/IconButton'
import Typography from '@material-ui/core/Typography'
import makeStyles from '@material-ui/core/styles/makeStyles'
import Tooltip from '@material-ui/core/Tooltip/Tooltip'
import Alert from '@material-ui/lab/Alert'
import AlertTitle from '@material-ui/lab/AlertTitle'
import ScheduleIcon from '@material-ui/icons/Schedule'
import Delete from '@material-ui/icons/Delete'
import Error from '@material-ui/icons/Error'
import _ from 'lodash-es'

import FlatList from '../../lists/FlatList'
import { fmt, Shift } from './sharedUtils'
import { UserAvatar } from '../../util/avatars'
import { useUserInfo, WithUserInfo } from '../../util/useUserInfo'
import { DateTime, Interval } from 'luxon'
import { useURLParam } from '../../actions'
import { relativeDate } from '../../util/timeFormat'
import { styles } from '../../styles/materialStyles'
import { parseInterval } from '../../util/shifts'

const useStyles = makeStyles((theme) => {
  return {
    alert: {
      margin: '8px 0 8px 0',
    },
    secondaryActionWrapper: {
      display: 'flex',
      alignItems: 'center',
    },
    secondaryActionError: {
      color: styles(theme).error.color,
    },
  }
})

type TempSchedShiftsListProps = {
  value: Shift[]
  onRemove: (shift: Shift) => void

  start: string
  end: string
}

type FlatListSub = {
  subHeader: string
}

type FlatListItem = {
  title?: string
  subText?: string
  icon?: JSX.Element
  secondaryAction?: JSX.Element | null
  render?: (item: FlatListItem) => ReactNode
}

type ShiftItem = {
  shift: Shift & WithUserInfo
  added: boolean
  start: DateTime
  end: DateTime
  interval: Interval
  isValid: boolean
}

type FlatListListItem = FlatListSub | FlatListItem

export default function TempSchedShiftsList({
  start,
  end,
  value,
  onRemove,
}: TempSchedShiftsListProps): JSX.Element {
  const classes = useStyles()
  const _shifts = useUserInfo(value)
  const [zone] = useURLParam('tz', 'local')
  const schedInterval = parseInterval({ start, end })

  function shiftsByDay(sortedShifts: ShiftItem[]): FlatListListItem[] {
    const result: FlatListListItem[] = []
    const lastShift = sortedShifts.reduce(
      (result, candidate) => (candidate.end > result.end ? candidate : result),
      sortedShifts[0],
    )

    const firstShiftStart = sortedShifts[0].start
    const lastShiftEnd = lastShift.end

    const displaySpan = Interval.fromDateTimes(
      DateTime.min(schedInterval.start, firstShiftStart).startOf('day'),
      DateTime.max(schedInterval.end, lastShiftEnd).endOf('day'),
    )

    const days = displaySpan.splitBy({ days: 1 })
    days.forEach((day, dayIdx) => {
      const dayShifts = sortedShifts.filter((s) => day.overlaps(s.interval))

      // render subheader for each day
      result.push({
        subHeader: relativeDate(day.start),
      })

      // render no coverage if no shifts for the day
      if (!dayShifts.length) {
        result.push({
          render: () => (
            <Alert
              key={day.start.toISO() + '-no-coverage'}
              className={classes.alert}
              severity='warning'
            >
              No coverage
            </Alert>
          ),
        })
        return
      }

      // checkCoverage will determine if there is a gap of 1 minute or more between the given datetimes
      const checkCoverage = (s: DateTime, e: DateTime): boolean => {
        return Interval.fromDateTimes(s, e).length('minutes') > 1
      }

      // craft user friendly shift string
      dayShifts.forEach((s, shiftIdx) => {
        // check start of day coverage for the first shift
        // if on the first day, temp sched start is used
        const _s = dayIdx === 0 ? DateTime.fromISO(start) : day.start
        if (shiftIdx === 0 && checkCoverage(_s, s.start)) {
          result.push({
            render: () => (
              <Alert
                key={_s.toISO() + '-no-start-coverage'}
                className={classes.alert}
                severity='warning'
              >
                No coverage until {s.start.setZone(zone).toFormat('hh:mm a')}
              </Alert>
            ),
          })
        }

        let shiftDetails = ''
        const startTime = s.start.toLocaleString({
          hour: 'numeric',
          minute: 'numeric',
        })
        const endTime = s.end.toLocaleString({
          hour: 'numeric',
          minute: 'numeric',
        })

        if (s.interval.engulfs(day)) {
          // shift (s.interval) spans all day
          shiftDetails = 'All day'
        } else if (day.engulfs(s.interval)) {
          // shift is inside the day
          shiftDetails = `From ${startTime} to ${endTime}`
        } else if (day.contains(s.end)) {
          shiftDetails = `Active until ${endTime}`
        } else {
          // shift starts and continues on for the rest of the day
          shiftDetails = `Active starting at ${startTime}`
        }

        result.push({
          title: s.shift.user.name,
          subText: shiftDetails,
          icon: <UserAvatar userID={s.shift.userID} />,
          secondaryAction: s.added ? null : (
            <div className={classes.secondaryActionWrapper}>
              {!s.isValid && (
                <Tooltip
                  title='This shift extends beyond the start and/or end of this temporary schedule'
                  placement='left'
                >
                  <Error className={classes.secondaryActionError} />
                </Tooltip>
              )}
              <IconButton onClick={() => onRemove(s.shift)}>
                <Delete />
              </IconButton>
            </div>
          ),
        })

        // prevents actions from rendering on each item if it's for the same shift
        s.added = true

        // check coverage until next shift within the current day, if exists
        if (
          shiftIdx < dayShifts.length - 1 &&
          checkCoverage(s.end, dayShifts[shiftIdx + 1].start)
        ) {
          result.push({
            render: () => (
              <Alert
                key={s.end.toISO() + '-no-middle-coverage'}
                className={classes.alert}
                severity='warning'
              >
                No coverage from {fmt(s.end.toISO(), zone)} to{' '}
                {fmt(dayShifts[shiftIdx + 1].start.toISO(), zone)}
              </Alert>
            ),
          })
        }

        // check end of day/temp sched coverage
        // if on the last day, temp sched end is used
        const _e = dayIdx === days.length - 1 ? DateTime.fromISO(end) : day.end
        if (shiftIdx === dayShifts.length - 1 && checkCoverage(s.end, _e)) {
          result.push({
            render: () => (
              <Alert
                key={_e.toISO() + '-no-end-coverage'}
                className={classes.alert}
                severity='warning'
              >
                No coverage after {s.end.setZone(zone).toFormat('hh:mm a')}
              </Alert>
            ),
          })
        }
      })
    })

    return result
  }

  function items(): FlatListListItem[] {
    // sort shifts and add some properties
    const shifts = _.sortBy(_shifts, 'start').map((s) => ({
      shift: s,
      added: false,
      start: DateTime.fromISO(s.start, { zone }),
      end: DateTime.fromISO(s.end, { zone }),
      interval: Interval.fromDateTimes(
        DateTime.fromISO(s.start, { zone }),
        DateTime.fromISO(s.end, { zone }),
      ),
      isValid: schedInterval.engulfs(parseInterval(s)),
    }))

    let result: FlatListListItem[] = []

    // add start time of temp schedule to top of list
    result.push({
      render: () => (
        <Alert
          key='start'
          className={classes.alert}
          severity='success'
          icon={<ScheduleIcon />}
        >
          Starts on {fmt(start, zone)}
        </Alert>
      ),
    })

    if (shifts.length > 0) {
      result = result.concat(shiftsByDay(shifts))
    } else {
      result.push({
        render: () => (
          <Alert key='no-coverage' className={classes.alert} severity='info'>
            <AlertTitle>No coverage</AlertTitle> Add a shift to get started
          </Alert>
        ),
      })
    }

    // add end time of temp schedule to top of list
    result.push({
      render: () => (
        <Alert
          key='end'
          className={classes.alert}
          severity='success'
          icon={<ScheduleIcon />}
        >
          Ends on {fmt(end, zone)}
        </Alert>
      ),
    })

    return result
  }

  return (
    <React.Fragment>
      <Typography variant='subtitle1' component='h3'>
        Shifts
      </Typography>
      <FlatList
        items={items()}
        emptyMessage='Add a user to the left to get started.'
        dense
      />
    </React.Fragment>
  )
}
