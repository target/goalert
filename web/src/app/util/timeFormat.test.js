import { formatTimeSince, formatTimeLocale, logTimeFormat } from './timeFormat'
import { DateTime, Duration, LocalZone } from 'luxon'

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

describe('formatTimeLocale', () => {
  const check = (time, exp, type) => {
    it(`${time} -> ${exp}`, () => {
      expect(formatTimeLocale(time, type)).toBe(exp)
    })
  }
  const h = 23
  const d = DateTime.fromISO(`1983-10-14T${h}:59:00.000Z`)
  const zone = d.offsetNameShort
  const offset = d.offset / 60
  const e = h - 12 + offset
  const amPM = h + offset < 12 ? 'AM' : 'PM'
  check(d, `October 14, 1983, ${e}:59 ${amPM} ${zone}`, 'full')
  check(d, `10/14/1983, ${e}:59 ${amPM}`, 'short')
  check(d, `Oct 14, 1983, ${e}:59 ${amPM}`)
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
