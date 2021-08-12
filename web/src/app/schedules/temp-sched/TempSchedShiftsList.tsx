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
import { styles } from '../../styles/materialStyles'
import { parseInterval } from '../../util/shifts'
import { useScheduleTZ } from './hooks'
import Spinner from '../../loading/components/Spinner'
import { splitAtMidnight } from '../../util/luxon-helpers'

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

type SortableItem = FlatListListItem & {
  // at is the earliest point in time for a given list item
  at: DateTime
  // itemType categorizes list item
  itemType: 'subheader' | 'gap' | 'shift' | 'start' | 'end'
}

export function sortItems(items: SortableItem[]): FlatListListItem[] {
  return items.sort((a, b) => {
    if (a.at < b.at) return -1
    if (a.at > b.at) return 1

    // a and b are at same time; use item type priority instead
    // subheaders first
    if (a.itemType === 'subheader') return -1
    if (b.itemType === 'subheader') return 1
    // then start notice
    if (a.itemType === 'start') return -1
    if (b.itemType === 'start') return 1
    // then gaps
    if (a.itemType === 'gap') return -1
    if (b.itemType === 'gap') return 1
    // then shifts
    if (a.itemType === 'shift') return -1
    if (b.itemType === 'shift') return 1
    // then end notice
    if (a.itemType === 'end') return -1
    if (b.itemType === 'end') return 1

    // identical items; should never get to this point
    return 0
  })
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
  const { q, zone } = useScheduleTZ(scheduleID)
  let shifts = useUserInfo(value)
  if (edit) {
    shifts = shifts.filter(
      (s) => DateTime.fromISO(s.start, { zone }) > DateTime.now().setZone(zone),
    )
  }

  const fmtTime = (dt: DateTime): string =>
    dt.toLocaleString(DateTime.TIME_SIMPLE)

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

    const subheaderItems = (() => {
      let lowerBound = schedInterval.start
      let upperBound = schedInterval.end

      // loop once to set timespan
      for (const s of shifts) {
        lowerBound = DateTime.min(
          lowerBound,
          DateTime.fromISO(s.start, { zone }),
        )
        upperBound = DateTime.max(upperBound, DateTime.fromISO(s.end, { zone }))
      }

      const dayInvs = splitAtMidnight(
        Interval.fromDateTimes(lowerBound, upperBound),
      )

      return dayInvs.map((day) => {
        const at = day.start
        return {
          id: 'header_' + at.toISO(),
          subHeader: day.start.toFormat('cccc, LLLL d'),
          at,
          itemType: 'subheader',
        } as SortableItem
      })
    })()

    const coverageGapItems = (() => {
      const shiftIntervals = shifts.map((s) => parseInterval(s, zone))
      const gapIntervals = _.flatMap(
        schedInterval.difference(...shiftIntervals),
        (inv) => splitAtMidnight(inv),
      )
      return gapIntervals.map((gap) => {
        return {
          id: 'day-no-coverage_' + gap.start.toISO(),
          type: 'WARNING',
          message: '',
          details: 'No coverage',
          at: gap.start,
          itemType: 'gap',
        } as SortableItem
      })
    })()

    const shiftItems = (() => {
      return _.flatMap(shifts, (s) => {
        const shiftInv = parseInterval(s, zone)
        const isValid = schedInterval.engulfs(shiftInv)
        const dayInvs = splitAtMidnight(shiftInv)

        return dayInvs.map((inv, index) => {
          const startTime = fmtTime(inv.start)
          const endTime = fmtTime(inv.end)

          let subText = ''
          if (inv.length('hours') === 24) {
            // shift spans all day
            subText = 'All day'
          } else if (inv.engulfs(shiftInv)) {
            // shift is inside the day
            subText = `From ${startTime} to ${endTime}`
          } else if (inv.end === shiftInv.end) {
            subText = `Active until ${endTime}`
          } else {
            // shift starts and continues on for the rest of the day
            subText = `Active starting at ${startTime}\n`
          }

          return {
            scrollIntoView: true,
            id: s.start + s.userID,
            title: s.user.name,
            subText,
            userID: s.userID,
            icon: <UserAvatar userID={s.userID} />,
            secondaryAction:
              index === 0 ? (
                <div className={classes.secondaryActionWrapper}>
                  {!isValid && (
                    <Tooltip
                      title='This shift extends beyond the start and/or end of this temporary schedule'
                      placement='left'
                    >
                      <Error color='error' />
                    </Tooltip>
                  )}
                  <IconButton
                    aria-label='delete shift'
                    onClick={() => onRemove(s)}
                  >
                    <Delete />
                  </IconButton>
                </div>
              ) : null,
            at: inv.start,
            itemType: 'shift',
          } as SortableItem
        })
      })
    })()

    const startItem = (() => {
      let details = `Starts at ${fmtTime(DateTime.fromISO(start, { zone }))}`
      let message = ''

      if (
        edit &&
        DateTime.fromISO(start, { zone }) < DateTime.now().setZone(zone)
      ) {
        message = 'Currently active'
        details = 'Historical shifts will not be displayed'
      }

      return {
        id: 'day-start_' + start,
        type: 'OK',
        icon: <ScheduleIcon />,
        message,
        details,
        at: DateTime.fromISO(start, { zone }),
        itemType: 'start',
      } as SortableItem
    })()

    const endItem: SortableItem = {
      id: 'ends-at_' + end,
      type: 'OK',
      icon: <ScheduleIcon />,
      message: '',
      details: `Ends at ${fmtTime(DateTime.fromISO(end, { zone }))}`,
      at: DateTime.fromISO(end, { zone }),
      itemType: 'end',
    }

    return sortItems([
      ...shiftItems,
      ...coverageGapItems,
      ...subheaderItems,
      startItem,
      endItem,
    ])
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
