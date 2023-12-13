import React from 'react'
import _ from 'lodash'
import { DateTime, Interval } from 'luxon'

import {
  FlatListListItem,
  FlatListNotice,
  FlatListSub,
} from '../../lists/FlatList'
import { ExplicitZone, splitAtMidnight } from '../../util/luxon-helpers'
import { parseInterval } from '../../util/shifts'
import { Shift } from './sharedUtils'
import { Tooltip } from '@mui/material'
import { fmtLocal, fmtTime } from '../../util/timeFormat'

export type Sortable<T> = T & {
  // at is the earliest point in time for a list item
  at: DateTime
  // itemType categorizes a list item
  itemType:
    | 'subheader'
    | 'gap'
    | 'shift'
    | 'start'
    | 'end'
    | 'outOfBounds'
    | 'active'
}

export function getSubheaderItems(
  schedInterval: Interval,
  shifts: Shift[],
  zone: ExplicitZone,
): Sortable<FlatListSub>[] {
  if (!schedInterval.isValid) {
    return []
  }

  // get earliest and farthest out start/end times
  const lowerBound = DateTime.min(
    schedInterval.start,
    ...shifts.map((s) => DateTime.fromISO(s.start, { zone })),
  )

  const upperBound = DateTime.max(
    schedInterval.end,
    ...shifts.map((s) => DateTime.fromISO(s.end, { zone })),
  )

  const dayInvs = splitAtMidnight(
    Interval.fromDateTimes(lowerBound, upperBound),
  )

  return dayInvs.map((day) => {
    const at = day.start.startOf('day')
    return {
      id: 'header_' + at.toISO(),
      subHeader: day.start.toFormat('cccc, LLLL d'),
      at,
      itemType: 'subheader',
    }
  })
}

export function getOutOfBoundsItems(
  schedInterval: Interval,
  shifts: Shift[],
  zone: ExplicitZone,
): Sortable<FlatListNotice>[] {
  if (!schedInterval.isValid) {
    return []
  }

  // get earliest and farthest out start/end times
  const lowerBound = DateTime.min(
    schedInterval.start,
    ...shifts.map((s) => DateTime.fromISO(s.start, { zone })),
  )

  const upperBound = DateTime.max(
    schedInterval.end,
    ...shifts.map((s) => DateTime.fromISO(s.end, { zone })),
  )

  const beforeStart = Interval.fromDateTimes(
    lowerBound,
    schedInterval.start,
  ).mapEndpoints((e) => e.startOf('day')) // ensure sched start date is not included

  const afterEnd = Interval.fromDateTimes(
    schedInterval.end,
    upperBound,
  ).mapEndpoints((e) => e.plus({ day: 1 }).startOf('day')) // ensure sched end date is not included

  const daysBeforeStart = splitAtMidnight(beforeStart)
  const daysAfterEnd = splitAtMidnight(afterEnd)
  const intervals = daysBeforeStart.concat(daysAfterEnd)

  let details = ''
  return intervals.map((interval) => {
    if (interval.end <= schedInterval.start) {
      details = 'This day is before the set start date.'
    } else if (interval.start >= schedInterval.end) {
      details = 'This day is after the set end date.'
    }

    return {
      id: 'day-out-of-bounds_' + interval.start.toISO(),
      type: 'INFO',
      message: '',
      details,
      at: interval.start.startOf('day'),
      itemType: 'outOfBounds',
    }
  })
}

export function getCoverageGapItems(
  schedInterval: Interval,
  shifts: Shift[],
  zone: ExplicitZone,
  handleCoverageClick?: (coverageGap: Interval) => void,
): Sortable<FlatListNotice>[] {
  if (!schedInterval.isValid) {
    return []
  }
  const shiftIntervals = shifts.map((s) => parseInterval(s, zone))
  const gapIntervals = _.flatMap(
    schedInterval.difference(...shiftIntervals),
    (inv) => splitAtMidnight(inv),
  )
  const isLocalZone = zone === DateTime.local().zoneName
  return gapIntervals.map((gap) => {
    let details = 'No coverage'
    let title = 'No coverage'
    if (gap.length('hours') === 24) {
      // nothing to do
      title = ''
    } else if (gap.start.equals(gap.start.startOf('day'))) {
      details += ` until ${fmtTime(gap.end, zone, false)}`
      title += ` until ${fmtLocal(gap.end)}`
    } else if (gap.end.equals(gap.start.plus({ day: 1 }).startOf('day'))) {
      details += ` after ${fmtTime(gap.start, zone, false)}`
      title += ` after ${fmtLocal(gap.start)}`
    } else {
      details += ` from ${fmtTime(gap.start, zone, false)} to ${fmtTime(
        gap.end,
        zone,
        false,
      )}`
      title += ` from ${fmtLocal(gap.start)} to ${fmtLocal(gap.end)}`
    }

    return {
      'data-cy': 'day-no-coverage',
      id: 'day-no-coverage_' + gap.start.toISO(),
      type: 'WARNING',
      message: '',
      details: (
        <Tooltip title={!isLocalZone ? title : ''} placement='right'>
          <span>{details}</span>
        </Tooltip>
      ),
      at: gap.start,
      itemType: 'gap',
      handleOnClick: () => {
        if (handleCoverageClick) {
          handleCoverageClick(gap)
        }
      },
    }
  })
}

export function sortItems(
  items: Sortable<FlatListListItem>[],
): Sortable<FlatListListItem>[] {
  return items.sort((a, b) => {
    if (a.at < b.at) return -1
    if (a.at > b.at) return 1

    // a and b are at same time; use item type priority instead
    // currently-active notice first
    if (a.itemType === 'active') return -1
    if (b.itemType === 'active') return 1
    // then subheaders
    if (a.itemType === 'subheader') return -1
    if (b.itemType === 'subheader') return 1
    // then out of bounds
    if (a.itemType === 'outOfBounds') return -1
    if (b.itemType === 'outOfBounds') return 1
    // then start notice
    if (a.itemType === 'start') return -1
    if (b.itemType === 'start') return 1
    // then gaps
    if (a.itemType === 'gap') return -1
    if (b.itemType === 'gap') return 1
    // then shifts
    if (
      // both shifts
      a.itemType === 'shift' &&
      b.itemType === 'shift' &&
      // typescript hints
      'title' in a &&
      'title' in b &&
      a.title &&
      b.title
    ) {
      return a.title < b.title ? -1 : 1
    }
    if (a.itemType === 'shift') return -1
    if (b.itemType === 'shift') return 1
    // then end notice
    if (a.itemType === 'end') return -1
    if (b.itemType === 'end') return 1

    // identical items; should never get to this point
    return 0
  })
}
