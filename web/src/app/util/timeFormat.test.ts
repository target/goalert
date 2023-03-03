import {
  formatTimestamp,
  FormatRelativeArg,
  FormatTimestampArg,
  formatRelative,
} from './timeFormat'

describe('formatTimestamp', () => {
  const check = (opts: FormatTimestampArg, exp: string): void => {
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

  check(
    {
      time: '2020-01-02T00:00:10Z',
      zone: 'UTC',
      format: 'relative',
      from: '2020-01-01T00:00:00Z',
      units: ['hour', 'minute', 'seconds'],
      precise: true,
    },
    'in 24 hr, 10 sec',
  )
})

describe('toRelative', () => {
  const check = (arg: FormatRelativeArg, exp: string): void => {
    it(exp, () => {
      expect(formatRelative(arg)).toBe(exp)
    })
  }

  check({ dur: { minute: -1 } }, '1 min ago')
  check({ dur: { minute: 1 } }, 'in 1 min')
  check({ dur: { hour: 1.5 }, precise: true }, 'in 1 hr, 30 min')
  check({ dur: { seconds: -5 }, precise: true }, '< 1 min ago') // default
  check({ dur: { seconds: -5 }, units: ['hour'], precise: true }, '< 1 hr ago') // default min
  check(
    { dur: { seconds: -5 }, min: { minute: 2 }, precise: true },
    '< 2 min ago',
  )

  check(
    {
      dur: { minutes: -1, seconds: -5 },
      units: ['minutes', 'seconds'],
      precise: true,
    },
    '1 min, 5 sec ago',
  )
})
