import { DateTime, Interval } from 'luxon'
import { Shift } from './sharedUtils'
import {
  getCoverageGapItems,
  getSubheaderItems,
  getOutOfBoundsItems,
} from './shiftsListUtil'
import { Chance } from 'chance'
import * as _ from 'lodash'

const c = new Chance()
const chicago = 'America/Chicago'
const newYork = 'America/New_York'

interface TestConfig {
  name: string
  schedIntervalISO: string
  shifts: Shift[]
  // expected is an array of start times for each coverage gap
  expected: string[]
  zone: string
}

describe('getSubheaderItems', () => {
  function check(tc: TestConfig): void {
    it(tc.name, () => {
      const schedInterval = Interval.fromISO(tc.schedIntervalISO, {
        zone: tc.zone,
      })
      const result = getSubheaderItems(schedInterval, tc.shifts, tc.zone)

      expect(result).toHaveLength(tc.expected.length)
      expect(_.uniq(result.map((r) => r.id))).toHaveLength(tc.expected.length)

      result.forEach((r, i) => {
        expect(r.at.zoneName).toEqual(tc.zone)
        expect(r.at).toEqual(r.at.startOf('day'))
        expect(r.subHeader).toBe(tc.expected[i])
      })
    })
  }

  check({
    name: '0 hr sched interval; no shifts',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-13T00:00:00.000-05:00'}`,
    shifts: [],
    expected: [],
    zone: chicago,
  })

  check({
    name: '1 hr sched interval; no shifts',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-13T01:00:00.000-05:00'}`,
    shifts: [],
    expected: ['Friday, August 13'],
    zone: chicago,
  })

  check({
    name: '1 hr sched interval; no shifts; alternate zone',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-13T01:00:00.000-05:00'}`,
    shifts: [],
    expected: ['Friday, August 13'],
    zone: newYork,
  })

  check({
    name: '24 hr sched interval; no shifts',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T00:00:00.000-05:00'}`,
    shifts: [],
    expected: ['Friday, August 13'],
    zone: chicago,
  })

  check({
    name: '25 hr sched interval; no shifts',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T01:00:00.000-05:00'}`,
    shifts: [],
    expected: ['Friday, August 13', 'Saturday, August 14'],
    zone: chicago,
  })

  check({
    name: '50 hr sched interval; no shifts',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-15T02:00:00.000-05:00'}`,
    shifts: [],
    expected: ['Friday, August 13', 'Saturday, August 14', 'Sunday, August 15'],
    zone: chicago,
  })

  check({
    name: '24 hr sched interval; 1 shift before sched start',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T00:00:00.000-05:00'}`,
    shifts: [
      {
        userID: c.guid(),
        start: '2021-08-12T00:00:00.000-05:00',
        end: '2021-08-13T05:00:00.000-05:00',
        truncated: false,
      },
    ],
    expected: ['Thursday, August 12', 'Friday, August 13'],
    zone: chicago,
  })

  check({
    name: '24 hr sched interval; 1 shift inside sched interval',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T00:00:00.000-05:00'}`,
    shifts: [
      {
        userID: c.guid(),
        start: '2021-08-13T02:00:00.000-05:00',
        end: '2021-08-13T03:00:00.000-05:00',
        truncated: false,
      },
    ],
    expected: ['Friday, August 13'],
    zone: chicago,
  })

  check({
    name: '24 hr sched interval; 1 shift after sched interval',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T00:00:00.000-05:00'}`,
    shifts: [
      {
        userID: c.guid(),
        start: '2021-08-15T02:00:00.000-05:00',
        end: '2021-08-16T04:00:00.000-05:00',
        truncated: false,
      },
    ],
    expected: [
      'Friday, August 13',
      'Saturday, August 14',
      'Sunday, August 15',
      'Monday, August 16',
    ],
    zone: chicago,
  })

  check({
    name: '30 hr sched interval; 3 random shifts',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T06:00:00.000-05:00'}`,
    shifts: [
      {
        userID: c.guid(),
        start: '2021-08-13T01:00:00.000-05:00',
        end: '2021-08-13T03:00:00.000-05:00',
        truncated: false,
      },
      {
        userID: c.guid(),
        start: '2021-08-13T02:00:00.000-05:00',
        end: '2021-08-13T04:00:00.000-05:00',
        truncated: false,
      },
      {
        userID: c.guid(),
        start: '2021-08-15T02:00:00.000-05:00',
        end: '2021-08-15T08:00:00.000-05:00',
        truncated: false,
      },
    ],
    expected: ['Friday, August 13', 'Saturday, August 14', 'Sunday, August 15'],
    zone: chicago,
  })
})

describe('getCoverageGapItems', () => {
  function check(tc: TestConfig): void {
    it(tc.name, () => {
      const schedInterval = Interval.fromISO(tc.schedIntervalISO, {
        zone: tc.zone,
      })
      const result = getCoverageGapItems(
        schedInterval,
        tc.shifts,
        tc.zone,
        () => {},
      )

      expect(result).toHaveLength(tc.expected.length)
      expect(_.uniq(result.map((r) => r.id))).toHaveLength(tc.expected.length)

      result.forEach((r, i) => {
        expect(r.at.zoneName).toEqual(tc.zone)
        expect(r.at).toEqual(
          DateTime.fromISO(tc.expected[i], { zone: tc.zone }),
        )
      })
    })
  }

  check({
    name: '0 hr sched interval; no shifts',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-13T00:00:00.000-05:00'}`,
    shifts: [],
    expected: [],
    zone: chicago,
  })

  check({
    name: '1 hr sched interval; no shifts; alternate zone',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-13T01:00:00.000-05:00'}`,
    shifts: [],
    expected: ['2021-08-13T00:00:00.000-05:00'],
    zone: newYork,
  })

  check({
    name: '3 hr sched interval; 1 shift; 2 gaps',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-13T03:00:00.000-05:00'}`,
    shifts: [
      {
        userID: c.guid(),
        start: '2021-08-13T01:00:00.000-05:00',
        end: '2021-08-13T02:00:00.000-05:00',
        truncated: false,
      },
    ],
    expected: [
      '2021-08-13T00:00:00.000-05:00',
      '2021-08-13T02:00:00.000-05:00',
    ],
    zone: chicago,
  })

  check({
    name: '3 hr sched interval; 1 shift; 1 gap before',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-13T03:00:00.000-05:00'}`,
    shifts: [
      {
        userID: c.guid(),
        start: '2021-08-13T01:00:00.000-05:00',
        end: '2021-08-13T03:00:00.000-05:00',
        truncated: false,
      },
    ],
    expected: ['2021-08-13T00:00:00.000-05:00'],
    zone: chicago,
  })

  check({
    name: '3 hr sched interval; 1 shift; 1 gap after',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-13T03:00:00.000-05:00'}`,
    shifts: [
      {
        userID: c.guid(),
        start: '2021-08-13T00:00:00.000-05:00',
        end: '2021-08-13T01:00:00.000-05:00',
        truncated: false,
      },
    ],
    expected: ['2021-08-13T01:00:00.000-05:00'],
    zone: chicago,
  })
})

describe('getOutOfBoundsItems', () => {
  function check(tc: TestConfig): void {
    it(tc.name, () => {
      const schedInterval = Interval.fromISO(tc.schedIntervalISO, {
        zone: tc.zone,
      })
      const result = getOutOfBoundsItems(schedInterval, tc.shifts, tc.zone)

      expect(result).toHaveLength(tc.expected.length)
      expect(_.uniq(result.map((r) => r.id))).toHaveLength(tc.expected.length)

      // expect to see start time of each out of bounds day
      result.forEach((r, i) => {
        expect(r.at.zoneName).toEqual(tc.zone)
        expect(r.at.toISO()).toEqual(tc.expected[i])
      })
    })
  }

  check({
    name: '0 hr sched interval; no shifts',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-13T00:00:00.000-05:00'}`,
    shifts: [],
    expected: [],
    zone: chicago,
  })

  check({
    name: '1 hr sched interval; no shifts',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-13T01:00:00.000-05:00'}`,
    shifts: [],
    expected: [],
    zone: chicago,
  })

  check({
    name: '24 hr sched interval; shift 1 day before sched start',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T00:00:00.000-05:00'}`,
    shifts: [
      {
        userID: c.guid(),
        start: '2021-08-12T00:00:00.000-05:00',
        end: '2021-08-13T05:00:00.000-05:00',
        truncated: false,
      },
    ],
    expected: ['2021-08-12T00:00:00.000-05:00'],
    zone: chicago,
  })

  check({
    name: '24 hr sched interval; shift the day the sched ends',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T00:00:00.000-05:00'}`,
    shifts: [
      {
        userID: c.guid(),
        start: '2021-08-14T00:00:00.000-05:00',
        end: '2021-08-14T05:00:00.000-05:00',
        truncated: false,
      },
    ],
    expected: [],
    zone: chicago,
  })

  check({
    name: '24 hr sched interval; shift 3 days before sched start',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T00:00:00.000-05:00'}`,
    shifts: [
      {
        userID: c.guid(),
        start: '2021-08-10T00:00:00.000-05:00',
        end: '2021-08-10T05:00:00.000-05:00',
        truncated: false,
      },
    ],
    expected: [
      '2021-08-10T00:00:00.000-05:00',
      '2021-08-11T00:00:00.000-05:00',
      '2021-08-12T00:00:00.000-05:00',
    ],
    zone: chicago,
  })

  check({
    name: '24 hr sched interval; shift 3 days after sched end',
    schedIntervalISO: `${'2021-08-13T00:00:00.000-05:00'}/${'2021-08-14T00:00:00.000-05:00'}`,
    shifts: [
      {
        userID: c.guid(),
        start: '2021-08-16T00:00:00.000-05:00',
        end: '2021-08-16T05:00:00.000-05:00',
        truncated: false,
      },
    ],
    expected: [
      '2021-08-15T00:00:00.000-05:00',
      '2021-08-16T00:00:00.000-05:00',
    ],
    zone: chicago,
  })

  check({
    name: 'sched interval ends at midnight; 1 shift starts and ends on same day as schedule start',
    schedIntervalISO: `${'2021-10-12T01:00:00.000-05:00'}/${'2021-10-13T00:00:00.000-05:00'}`,
    shifts: [
      {
        userID: c.guid(),
        start: '2021-10-12T00:00:00.000-05:00',
        end: '2021-10-12T06:00:00.000-05:00',
        truncated: false,
      },
    ],
    expected: [],
    zone: chicago,
  })

  check({
    name: 'sched interval ends at 1am; 1 shift starts and ends on same day as schedule end',
    schedIntervalISO: `${'2021-10-12T00:00:00.000-05:00'}/${'2021-10-13T01:00:00.000-05:00'}`,
    shifts: [
      {
        userID: c.guid(),
        start: '2021-10-13T00:00:00.000-05:00',
        end: '2021-10-13T06:00:00.000-05:00',
        truncated: false,
      },
    ],
    expected: [],
    zone: chicago,
  })
})
