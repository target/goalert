import React from 'react'
import IconButton from '@material-ui/core/IconButton'
import Typography from '@material-ui/core/Typography'
import makeStyles from '@material-ui/core/styles/makeStyles'
import Tooltip from '@material-ui/core/Tooltip/Tooltip'
import ScheduleIcon from '@material-ui/icons/Schedule'
import Delete from '@material-ui/icons/Delete'
import Error from '@material-ui/icons/Error'
import _ from 'lodash'

import { Shift } from './sharedUtils'
import FlatList, { FlatListListItem } from '../../lists/FlatList'
import { UserAvatar } from '../../util/avatars'
import { useUserInfo } from '../../util/useUserInfo'
import { DateTime, Interval } from 'luxon'
import { relativeDate } from '../../util/timeFormat'
import { styles } from '../../styles/materialStyles'
import { parseInterval } from '../../util/shifts'
import { splitAtMidnight } from '../../util/luxon-helpers'
import { useScheduleTZ } from './hooks'
import Spinner from '../../loading/components/Spinner'

const useStyles = makeStyles((theme) => {
  return {
    secondaryActionWrapper: {
      display: 'flex',
      alignItems: 'center',
    },
    secondaryActionError: {
      color: styles(theme).error.color,
    },
    shiftsContainer: {
      paddingRight: '0.5rem',
    },
    listSpinner: {
      marginTop: '20rem',
    },
  }
})

type TempSchedShiftsListProps = {
  value: Shift[]
  onRemove: (shift: Shift) => void

  start: string
  end: string

  edit?: boolean

  scheduleID: string
}

export default function TempSchedShiftsList({
  edit,
  start,
  end,
  value,
  onRemove,
  scheduleID,
}: TempSchedShiftsListProps): JSX.Element {
  const classes = useStyles()
  const _shifts = useUserInfo(value)
  const { q, zone } = useScheduleTZ(scheduleID)
  const timeFmt = 'h:mm a'

  const schedInterval = parseInterval({ start, end }, zone)

  function items(): FlatListListItem[] {
    // render helpful message if interval is invalid
    // shouldn't ever be seen because of our validation checks, but just in case
    if (!schedInterval.isValid) {
      return [
        {
          id: 'invalid',
          type: 'ERROR',
          message: 'Invalid Start/End',
          transition: true,
          details:
            'Oops! There was a problem with the interval selected in step 1. Please try again.',
        },
      ]
    }

    const sortedShifts = _shifts.length
      ? _.sortBy(_shifts, 'start').map((s) => ({
          id: s.start + s.userID,
          shift: s,
          start: DateTime.fromISO(s.start, { zone }),
          end: DateTime.fromISO(s.end, { zone }),
          added: false,
          interval: Interval.fromDateTimes(
            DateTime.fromISO(s.start, { zone }),
            DateTime.fromISO(s.end, { zone }),
          ),
          isValid: schedInterval.engulfs(parseInterval(s)),
        }))
      : []

    const firstShiftStart = sortedShifts.length
      ? sortedShifts[0].start
      : schedInterval.start

    // get farthest out end time
    // although shifts are sorted, the last shift may not necessarily end last
    const lastShiftEnd = sortedShifts.length
      ? sortedShifts.reduce(
          (result, candidate) =>
            candidate.end > result.end ? candidate : result,
          sortedShifts[0],
        ).end
      : schedInterval.end

    let spanStart = DateTime.min(schedInterval.start, firstShiftStart).startOf(
      'day',
    )
    if (edit)
      spanStart = DateTime.max(spanStart, DateTime.utc().startOf('hour'))

    const displaySpan = Interval.fromDateTimes(
      spanStart,
      DateTime.max(schedInterval.end, lastShiftEnd).endOf('day'),
    )

    const result: FlatListListItem[] = []

    const days = splitAtMidnight(displaySpan)
    days.forEach((dayInterval, dayIdx) => {
      const dayShifts = sortedShifts.filter((s) =>
        dayInterval.overlaps(s.interval),
      )

      // render subheader for each day
      result.push({
        id: 'header_' + dayInterval.start,
        subHeader: relativeDate(dayInterval.start),
      })

      let dayStart = dayInterval.start
      if (dayIdx === 0 && firstShiftStart.day === schedInterval.start.day) {
        dayStart = schedInterval.start
      }

      // add start time of temp schedule to top of list
      // for day that it will start on
      if (dayStart.day === schedInterval.start.day) {
        let details = `Starts at ${DateTime.fromISO(start, { zone }).toFormat(
          timeFmt,
        )}`
        let message = ''

        if (edit && DateTime.fromISO(start, { zone }) < DateTime.utc()) {
          message = 'Currently active'
          details = 'Historical shifts will not be displayed'
        }

        result.push({
          id: 'day-start_' + start,
          type: 'OK',
          icon: <ScheduleIcon />,
          message,
          details,
        })
      }

      // render no coverage and continue if no shifts for the given day
      if (dayShifts.length === 0) {
        return result.push({
          id: 'day-no-coverage_' + start,
          type: 'WARNING',
          message: '',
          details: 'No coverage',
        })
      }

      // checkCoverage will determine if there is a gap of 1 minute or more between the given datetimes
      const checkCoverage = (s: DateTime, e: DateTime): boolean => {
        return Interval.fromDateTimes(s, e).length('minutes') >= 1
      }

      // craft list item JSX for each day
      dayShifts.forEach((s, shiftIdx) => {
        if (s.isValid && shiftIdx === 0 && checkCoverage(dayStart, s.start)) {
          result.push({
            id: 'no-coverage-until_' + s.start,
            type: 'WARNING',
            message: '',
            details: `No coverage until ${s.start
              .setZone(zone)
              .toFormat(timeFmt)}`,
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

        if (s.interval.engulfs(dayInterval)) {
          // shift (s.interval) spans all day
          shiftDetails = 'All day'
        } else if (dayInterval.engulfs(s.interval)) {
          // shift is inside the day
          shiftDetails = `From ${startTime} to ${endTime}`
        } else if (dayInterval.contains(s.end)) {
          shiftDetails = `Active until ${endTime}`
        } else {
          // shift starts and continues on for the rest of the day
          shiftDetails = `Active starting at ${startTime}`
        }

        result.push({
          scrollIntoView: true,
          id: s.start + s.shift.userID,
          title: s.shift.user?.name,
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
              <IconButton
                aria-label='delete shift'
                onClick={() => onRemove(s.shift)}
              >
                <Delete />
              </IconButton>
            </div>
          ),
        })

        // signify that a portion of the shift is now added to the list
        // prevents actions from rendering on subsequent list items for the same shift timespan
        s.added = true

        // check coverage until the next shift (if there is one) within the current day
        if (
          s.isValid &&
          shiftIdx < dayShifts.length - 1 &&
          checkCoverage(s.end, dayShifts[shiftIdx + 1].start)
        ) {
          result.push({
            id: 'no-coverage-from_' + s.end,
            type: 'WARNING',
            message: '',
            details: `No coverage from ${s.end
              .setZone(zone)
              .toFormat(timeFmt)} to 
            ${dayShifts[shiftIdx + 1].start.setZone(zone).toFormat(timeFmt)}`,
          })
        }

        // check end of day/temp sched coverage
        // if on the day of temp sched's end, temp sched end is used
        let dayEnd = dayInterval.end
        if (
          dayIdx === days.length - 1 &&
          lastShiftEnd.day === schedInterval.end.day
        ) {
          dayEnd = schedInterval.end
        }
        if (
          s.isValid &&
          shiftIdx === dayShifts.length - 1 &&
          checkCoverage(s.end, dayEnd)
        ) {
          result.push({
            id: 'no-coverage-after_' + s.end,
            type: 'WARNING',
            message: '',
            details: `No coverage after ${s.end
              .setZone(zone)
              .toFormat(timeFmt)}`,
          })
        }
      })
    })

    // add end time of temp schedule to bottom of list
    result.push({
      id: 'ends-at_' + end,
      type: 'OK',
      icon: <ScheduleIcon />,
      message: '',
      details: `Ends at ${DateTime.fromISO(end, { zone }).toFormat(timeFmt)}`,
    })

    return result
  }

  return (
    <div className={classes.shiftsContainer}>
      <Typography variant='subtitle1' component='h3'>
        Shifts
      </Typography>
      {q.loading ? (
        <div className={classes.listSpinner}>
          <Spinner />
        </div>
      ) : (
        <FlatList
          data-cy='shifts-list'
          items={items()}
          emptyMessage='Add a user to the left to get started.'
          dense
          transition
        />
      )}
    </div>
  )
}
