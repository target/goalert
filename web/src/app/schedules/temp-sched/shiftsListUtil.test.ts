import { DateTime, Interval } from 'luxon'
import { Shift } from './sharedUtils'
import { getCoverageGapItems, getSubheaderItems } from './shiftsListUtil'
import { Chance } from 'chance'
import * as _ from 'lodash'

const c = new Chance()
const chicago = 'America/Chicago'
const newyork = 'America/New_York'

describe('getSubheaderItems', () => {
  function check(
    name: string,
    schedInterval: Interval,
    shifts: Shift[],
    expected: string[],
    zone = chicago,
  ): void {
    it(name, () => {
      const result = getSubheaderItems(schedInterval, shifts, zone)

      expect(result).toHaveLength(expected.length)
      expect(_.uniq(result.map((r) => r.id))).toHaveLength(expected.length)

      result.forEach((r, i) => {
        expect(r.at.zoneName).toEqual(zone)
        expect(r.at).toEqual(r.at.startOf('day'))
        expect(r.subHeader).toBe(expected[i])
      })
    })
  }

  check(
    '0 hr sched interval; no shifts',
    Interval.fromISO(
      `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-13T00:00:00.000-05:00'}`,
    ),
    [],
    [],
  )

  check(
    '1 hr sched interval; no shifts',
    Interval.fromISO(
      `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-13T01:00:00.000-05:00'}`,
    ),
    [],
    ['Friday, August 13'],
  )

  check(
    '1 hr sched interval; no shifts; alternate zone',
    Interval.fromDateTimes(
      DateTime.fromISO('2021-08-13T00:00:00.000-05:00', { zone: newyork }),
      DateTime.fromISO('2021-08-13T01:00:00.000-05:00', { zone: newyork }),
    ),
    [],
    ['Friday, August 13'],
    newyork,
  )

  check(
    '24 hr sched interval; no shifts',
    Interval.fromISO(
      `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T00:00:00.000-05:00'}`,
    ),
    [],
    ['Friday, August 13'],
  )

  check(
    '25 hr sched interval; no shifts',
    Interval.fromISO(
      `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T01:00:00.000-05:00'}`,
    ),
    [],
    ['Friday, August 13', 'Saturday, August 14'],
  )

  check(
    '50 hr sched interval; no shifts',
    Interval.fromISO(
      `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-15T02:00:00.000-05:00'}`,
    ),
    [],
    ['Friday, August 13', 'Saturday, August 14', 'Sunday, August 15'],
  )

  check(
    '24 hr sched interval; 1 shift before sched start',
    Interval.fromISO(
      `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T00:00:00.000-05:00'}`,
    ),
    [
      {
        userID: c.guid(),
        start: '2021-08-12T00:00:00.000-05:00',
        end: '2021-08-13T05:00:00.000-05:00',
      },
    ],
    ['Thursday, August 12', 'Friday, August 13'],
  )

  check(
    '24 hr sched interval; 1 shift inside sched interval',
    Interval.fromISO(
      `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T00:00:00.000-05:00'}`,
    ),
    [
      {
        userID: c.guid(),
        start: '2021-08-13T02:00:00.000-05:00',
        end: '2021-08-13T03:00:00.000-05:00',
      },
    ],
    ['Friday, August 13'],
  )

  check(
    '24 hr sched interval; 1 shift after sched interval',
    Interval.fromISO(
      `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T00:00:00.000-05:00'}`,
    ),
    [
      {
        userID: c.guid(),
        start: '2021-08-15T02:00:00.000-05:00',
        end: '2021-08-16T04:00:00.000-05:00',
      },
    ],
    [
      'Friday, August 13',
      'Saturday, August 14',
      'Sunday, August 15',
      'Monday, August 16',
    ],
  )

  check(
    '30 hr sched interval; 3 random shifts',
    Interval.fromISO(
      `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T06:00:00.000-05:00'}`,
    ),
    [
      {
        userID: c.guid(),
        start: '2021-08-13T01:00:00.000-05:00',
        end: '2021-08-13T03:00:00.000-05:00',
      },
      {
        userID: c.guid(),
        start: '2021-08-13T02:00:00.000-05:00',
        end: '2021-08-13T04:00:00.000-05:00',
      },
      {
        userID: c.guid(),
        start: '2021-08-15T02:00:00.000-05:00',
        end: '2021-08-15T08:00:00.000-05:00',
      },
    ],
    ['Friday, August 13', 'Saturday, August 14', 'Sunday, August 15'],
  )
})

describe('getCoverageGapItems', () => {
  function check(
    name: string,
    schedInterval: Interval,
    shifts: Shift[],
    // expected is an array of start times for each coverage gap
    expected: string[],
    zone = chicago,
  ): void {
    it(name, () => {
      const result = getCoverageGapItems(schedInterval, shifts, zone)

      expect(result).toHaveLength(expected.length)
      expect(_.uniq(result.map((r) => r.id))).toHaveLength(expected.length)

      result.forEach((r, i) => {
        expect(r.at.zoneName).toEqual(zone)
        expect(r.at).toEqual(DateTime.fromISO(expected[i], { zone }))
      })
    })
  }

  check(
    '0 hr sched interval; no shifts',
    Interval.fromISO(
      `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-13T00:00:00.000-05:00'}`,
    ),
    [],
    [],
  )

  check(
    '1 hr sched interval; no shifts; alternate zone',
    Interval.fromDateTimes(
      DateTime.fromISO('2021-08-13T00:00:00.000-05:00', { zone: newyork }),
      DateTime.fromISO('2021-08-13T01:00:00.000-05:00', { zone: newyork }),
    ),
    [],
    ['2021-08-13T00:00:00.000-05:00'],
    newyork,
  )

  check(
    '3 hr sched interval; 1 shift; 2 gaps',
    Interval.fromDateTimes(
      DateTime.fromISO('2021-08-13T00:00:00.000-05:00', { zone: newyork }),
      DateTime.fromISO('2021-08-13T03:00:00.000-05:00', { zone: newyork }),
    ),
    [
      {
        userID: c.guid(),
        start: '2021-08-13T01:00:00.000-05:00',
        end: '2021-08-13T02:00:00.000-05:00',
      },
    ],
    ['2021-08-13T00:00:00.000-05:00', '2021-08-13T02:00:00.000-05:00'],
    newyork,
  )

  check(
    '3 hr sched interval; 1 shift; 1 gap before',
    Interval.fromDateTimes(
      DateTime.fromISO('2021-08-13T00:00:00.000-05:00', { zone: newyork }),
      DateTime.fromISO('2021-08-13T03:00:00.000-05:00', { zone: newyork }),
    ),
    [
      {
        userID: c.guid(),
        start: '2021-08-13T01:00:00.000-05:00',
        end: '2021-08-13T03:00:00.000-05:00',
      },
    ],
    ['2021-08-13T00:00:00.000-05:00'],
    newyork,
  )

  check(
    '3 hr sched interval; 1 shift; 1 gap after',
    Interval.fromDateTimes(
      DateTime.fromISO('2021-08-13T00:00:00.000-05:00', { zone: newyork }),
      DateTime.fromISO('2021-08-13T03:00:00.000-05:00', { zone: newyork }),
    ),
    [
      {
        userID: c.guid(),
        start: '2021-08-13T00:00:00.000-05:00',
        end: '2021-08-13T01:00:00.000-05:00',
      },
    ],
    ['2021-08-13T01:00:00.000-05:00'],
    newyork,
  )
})
