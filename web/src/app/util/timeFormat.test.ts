import {
  formatTimeSince,
  formatTimestamp,
  TimeFormatOpts,
  toRelativePrecise,
} from './timeFormat'
import { DateTime, Duration, DurationLikeObject } from 'luxon'

describe('formatTimestamp', () => {
  const check = (opts: TimeFormatOpts, exp: string): void => {
    it(`${opts.time} === ${exp}`, () => {
      expect(formatTimestamp(opts)).toBe(exp)
    })
  }

  check({ time: '2020-01-01T00:00:00Z', zone: 'UTC' }, 'Jan 1, 2020, 12:00 AM')
  check(
    { time: '2020-01-01T00:00:00Z', zone: 'UTC', format: 'clock' },
    '12:00 AM',
  )

  check(
    {
      time: '2020-01-01T00:00:00Z',
      zone: 'UTC',
      format: 'relative',
      now: '2020-01-02T00:00:00Z',
    },
    '1 day ago',
  )

  check(
    {
      time: '2020-01-02T00:00:00Z',
      zone: 'UTC',
      format: 'relative',
      now: '2020-01-01T00:00:00Z',
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

describe('formatTimeSince', () => {
  const check = (time: DurationLikeObject, exp: string): void => {
    const dur = Duration.fromObject(time)
    it(`${dur.toFormat('dDays h:m:s')} === ${exp}`, () => {
      const since = DateTime.utc()
      expect(formatTimeSince(since, since.plus(dur))).toBe(exp)
    })
  }
  check({ seconds: -1 }, '< 1m ago')
  check({ seconds: 1 }, '< 1m ago')
  check({ seconds: 59 }, '< 1m ago')
  check({ minutes: 1 }, '1m ago')
  check({ minutes: 1, seconds: 1 }, '1m ago')
  check({ hours: 1 }, '1h ago')
  check({ hours: 1, seconds: 1 }, '1h ago')
  check({ hours: 3, seconds: 1 }, '3h ago')
  check({ days: 1, seconds: 1 }, '1d ago')
  check({ days: 20, seconds: 1 }, '20d ago')
  check({ months: 3, days: 5 }, '> 3mo ago')
  check({ months: 20, seconds: 1 }, '> 1y ago')
  check({ months: 200, seconds: 1 }, '> 16y ago')
})
