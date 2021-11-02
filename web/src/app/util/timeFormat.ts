import { Interval, DateTime, DateTimeFormatOptions } from 'luxon'
import { ExplicitZone } from './luxon-helpers'

export function formatTimeSince(
  _since: DateTime | string,
  _now = DateTime.utc(),
): string {
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
    return `> ${Math.floor(diff.as('months'))}mo ago`
  }

  return `> ${Math.floor(diff.as('years'))}y ago`
}

export function relativeDate(
  _to: DateTime | string,
  _from = DateTime.utc(),
): string {
  const to = _to instanceof DateTime ? _to : DateTime.fromISO(_to)
  const from = (_from instanceof DateTime ? _from : DateTime.fromISO(_from))
    .setZone(to.zoneName)
    .startOf('day')

  const fmt: DateTimeFormatOptions = {
    month: 'long',
    day: 'numeric',
  }
  const build = (prefix = '', opts = {}): string =>
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

// fmtTime returns simple string for ISO string or DateTime object.
// If `withZoneAbbr` is not specified, zone info will only be provided for non-local times.
// Only 12-hour if the locale is.
// e.g. '9:30 AM', '9:30 PM', '9:30 AM CDT'
export function fmtTime(
  time: DateTime | string,
  zone: ExplicitZone,
  withZoneAbbr: boolean | null = null,
): string {
  if (!time) return ''
  if (typeof time === 'string') {
    time = DateTime.fromISO(time, { zone })
  } else {
    time = time.setZone(zone)
  }

  const prefix = time.toLocaleString(DateTime.TIME_SIMPLE)
  const suffix = time.toFormat('ZZZZ')

  if (withZoneAbbr === true) return prefix + ' ' + suffix
  if (withZoneAbbr === false) return prefix

  if (zone === DateTime.local().zoneName) return prefix
  return prefix + ' ' + suffix
}

// fmtLocal is like fmtTime but uses the system zone and displays zone info by default.
export function fmtLocal(
  time: DateTime | string,
  withZoneAbbr: boolean | null = true,
): string {
  return fmtTime(time, 'local', withZoneAbbr)
}
