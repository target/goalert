import {
  formatTimestamp,
  TimeFormatOpts,
  toRelativePrecise,
} from './timeFormat'
import { Duration } from 'luxon'

describe('formatTimestamp', () => {
  const check = (opts: TimeFormatOpts, exp: string): void => {
    it(`${opts.time} === ${exp}`, () => {
      expect(formatTimestamp(opts)).toBe(exp)
    })
  }

  check({ time: '2020-01-01T00:00:00Z', zone: 'UTC' }, 'Jan 1, 2020, 12:00 AM')
  check(
    { time: '2020-01-01T00:00:00Z', zone: 'UTC', format: 'default' },
    'Jan 1, 2020, 12:00 AM',
  )
  check(
    { time: '2020-01-01T00:00:00Z', zone: 'UTC', format: 'clock' },
    '12:00 AM',
  )
  check(
    { time: '2020-01-01T00:00:00Z', zone: 'UTC', format: 'weekday-clock' },
    'Wed 12:00 AM',
  )
  check(
    {
      time: '2020-01-01T00:00:00Z',
      zone: 'UTC',
      format: 'relative-date',
      from: '2020-01-02T00:00:00Z',
    },
    'Yesterday, January 1',
  )

  check(
    {
      time: '2020-01-01T00:00:00Z',
      zone: 'UTC',
      format: 'relative',
      from: '2020-01-02T00:00:00Z',
    },
    '1 day ago',
  )

  check(
    {
      time: '2020-01-02T00:00:00Z',
      zone: 'UTC',
      format: 'relative',
      from: '2020-01-01T00:00:00Z',
    },
    'in 1 day',
  )
})

describe('toRelativePrecise', () => {
  const check = (dur: Duration, exp: string): void => {
    it(`${dur.toFormat('dDays h:m:s')} === ${exp}`, () => {
      expect(toRelativePrecise(dur)).toBe(exp)
    })
  }

  check(Duration.fromObject({ minutes: -1 }), '1 minute ago')
  check(Duration.fromObject({ minutes: 1 }), 'in 1 minute')
  check(Duration.fromObject({ hours: 1.5 }), 'in 1 hour 30 minutes')
})
