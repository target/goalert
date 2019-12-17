import {
  formatTimeSince,
  logTimeFormat,
  getPaddedFormatString,
  getFormatMask,
} from './timeFormat'
import { DateTime, Duration } from 'luxon'

describe('formatTimeSince', () => {
  const check = (time, exp) => {
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
  check({ months: 3 }, '3mo ago')
  check({ months: 20, seconds: 1 }, '1y ago')
  check({ months: 200, seconds: 1 }, '16y ago')
})
describe('logTimeFormat', () => {
  const check = (to, from, exp) => {
    it(`alert log time format`, () => {
      expect(logTimeFormat(to, from)).toBe(exp)
    })
  }
  const to = '2019-05-25'
  let from = DateTime.local(2019, 5, 25)
  check(to, from, 'Today at 12:00 AM')
  from = from.plus({ days: 1 })
  check(to, from, 'Yesterday at 12:00 AM')
  from = from.plus({ days: 6 })
  check(to, from, 'Last Saturday at 12:00 AM')
  from = from.plus({ days: 7 })
  check(to, from, '05/25/2019')
})

describe('getPaddedFormatString', () => {
  const check = (input, expected) => {
    it('format string', () =>
      expect(getPaddedFormatString(input)).toEqual(expected))
  }

  check('2006', 'yyyy')
  check('06', 'yy')
  check('1/2/2006 3:04:05.099 PM', 'LL/dd/yyyy hh:mm:ss.u a')
  check('Mon, Jan 2', 'ccc, LLL dd')
})

describe('getFormatMask', () => {
  const check = (input, expected) => {
    it('format string', () => expect(getFormatMask(input)).toEqual(expected))
  }

  check('yyyy', '____')

  check('LL/dd/yyyy hh:mm:ss.u a', '__/__/____ __:__:__.___ __')
})
