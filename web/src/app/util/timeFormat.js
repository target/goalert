import { Interval, DateTime } from 'luxon'

export function formatTimeSince(_since, _now = DateTime.utc()) {
  if (!_since) return ''
  const since = _since instanceof DateTime ? _since : DateTime.fromISO(_since)
  const now = _now instanceof DateTime ? _now : DateTime.fromISO(_now)
  const diff = now.diff(since)

  if (diff.as('minutes') < 1) {
    return `< 1m ago`
  }

  if (diff.as('hours') < 1) {
    return `${Math.floor(diff.as('minutes'))}m ago`
  }

  if (diff.as('days') < 1) {
    return `${Math.floor(diff.as('hours'))}h ago`
  }

  if (diff.as('months') < 1) {
    return `${Math.floor(diff.as('days'))}d ago`
  }

  if (diff.as('years') < 1) {
    return `${Math.floor(diff.as('months'))}mo ago`
  }

  return `${Math.floor(diff.as('years'))}y ago`
}

export function relativeDate(_to, _from = DateTime.utc()) {
  const to = _to instanceof DateTime ? _to : DateTime.fromISO(_to)
  const from = (_from instanceof DateTime ? _from : DateTime.fromISO(_from))
    .setZone(to.zoneName)
    .startOf('day')

  const fmt = {
    month: 'long',
    day: 'numeric',
  }
  const build = (prefix = '', opts = {}) =>
    `${prefix} ${to.toLocaleString({ ...fmt, ...opts })}`.trim()

  if (Interval.after(from, { days: 1 }).contains(to)) return build('Today,')

  if (from.year !== to.year) fmt.year = 'numeric'

  if (Interval.before(from, { days: 1 }).contains(to))
    return build('Yesterday,')
  if (Interval.before(from, { weeks: 1 }).contains(to))
    return build('Last', { weekday: 'long' })
  if (Interval.after(from, { days: 2 }).contains(to)) return build('Tomorrow,')
  if (Interval.after(from, { weeks: 1 }).contains(to))
    return build('This', { weekday: 'long' })
  if (Interval.after(from, { weeks: 2 }).contains(to))
    return build('Next', { weekday: 'long' })

  return build('', { weekday: 'long' })
}

export function logTimeFormat(_to, _from) {
  const to = DateTime.fromISO(_to)
  if (Interval.after(_from, { days: 1 }).contains(to))
    return 'Today at ' + to.toFormat('h:mm a')
  if (Interval.before(_from, { days: 1 }).contains(to))
    return 'Yesterday at ' + to.toFormat('h:mm a')
  if (Interval.before(_from, { weeks: 1 }).contains(to))
    return 'Last ' + to.weekdayLong + ' at ' + to.toFormat('h:mm a')
  return to.toFormat('MM/dd/yyyy')
}

const fixed = DateTime.fromObject({
  month: 1,
  day: 2,
  hour: 15, // 3pm
  minute: 4,
  second: 5,
  year: 2006,
  millisecond: 99,
}).setZone('America/Boise', { keepLocalTime: true })

const localeKeys = [
  ['yyyy', 'yy', 'y'],
  ['LL', 'L', 'LLLL', 'LLL', 'LLLLL'],
  ['dd', 'd'],
  ['u', 'S'],
  ['HH', 'H'],
  ['hh', 'h'],
  ['mm', 'm'],
  ['ss', 's'],
  ['ZZZZZ', 'ZZZZ', 'ZZZ', 'ZZ', 'Z'],
  'z',
  'a',
  ['cccc', 'ccc', 'ccccc'],
]

// getFormatMask will return an input mask for a given time format.
export const getFormatMask = format =>
  fixed.toFormat(format).replace(/[0-9APM]/g, '_')

// getPaddedLocaleFormatString will return a padded format string
// corresponding to the current locale.
export const getPaddedLocaleFormatString = opts =>
  getPaddedFormatString(fixed.toLocaleString(opts))

// getPaddedFormatString will return a format string with all values padded
// that corresponds to the provided time string.
//
// The string should represent the time: Monday, Jan 2, 2006 at 3:04:05.099 PM, MST-7
export const getPaddedFormatString = _s => {
  let s = _s
  localeKeys.forEach(_keys => {
    const keys = Array.isArray(_keys) ? _keys : [_keys]
    keys.some(k => {
      const old = s
      s = s.replace(fixed.toFormat(k), k)
      return s !== old
    })
  })
  return (
    s
      // ensure we always use the padded versions
      .replace(/H+/, 'HH')
      .replace(/h+/, 'hh')
      .replace(/m+/, 'mm')
      .replace(/s+/, 'ss')
      .replace(/d+/, 'dd')
      .replace(/\bL\b/, 'LL')
      .replace(/\bM\b/, 'MM')
      .replace(/\by\b/, 'yy')
  )
}
