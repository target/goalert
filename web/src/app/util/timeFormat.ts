import {
  Interval,
  DateTime,
  Duration,
  DurationLikeObject,
  DateTimeFormatOptions,
} from 'luxon'
import { ExplicitZone } from './luxon-helpers'

export type TimeFormat =
  | 'relative'
  | 'relative-date'
  | 'clock'
  | 'locale'
  | 'weekday-clock'

export type TimeFormatOpts = {
  time: string
  zone?: string

  // omitSameDate will omit the date if it is the same as the provided date.
  //
  // Has no effect if format is 'relative', 'since', or 'clock'.
  omitSameDate?: string

  format?: TimeFormat

  // now is the current time to use for relative time calculations, defaults to
  // DateTime.utc().toISO()
  now?: string

  // If true, the 'relative' format will include multiple units.
  precise?: boolean
}

const isSameDay = (a: DateTime, b: DateTime): boolean => {
  return a.hasSame(b, 'day') && a.hasSame(b, 'month') && a.hasSame(b, 'year')
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

export function toRelativePrecise(
  dur: Duration,
  units: ReadonlyArray<keyof DurationLikeObject> = ['days', 'hours', 'minutes'],
): string {
  const parts = []
  const prefix = dur.valueOf() > 0 ? 'in ' : ''
  const suffix = dur.valueOf() > 0 ? '' : ' ago'
  if (dur.valueOf() < 0) dur = dur.negate()

  for (const unit of units) {
    const val = Math.floor(dur.as(unit))
    if (val === 0) continue
    const part = Duration.fromObject({ [unit]: val })
    dur = dur.minus(part)

    parts.push(part.toHuman())
  }

  return prefix + parts.join(' ') + suffix
}

export function formatTimestamp(opts: TimeFormatOpts): string {
  const {
    time,
    zone = 'local',
    omitSameDate,
    format = 'locale',

    now = DateTime.utc().toISO(),
  } = opts

  const dt = DateTime.fromISO(time, { zone })
  const omit = omitSameDate && DateTime.fromISO(omitSameDate, { zone })

  let formatted: string
  switch (format) {
    case 'relative-date':
      formatted = relativeDate(dt, DateTime.fromISO(now))
      break
    case 'relative':
      if (opts.precise) {
        formatted = toRelativePrecise(dt.diff(DateTime.fromISO(now)), [
          'days',
          'hours',
          'minutes',
        ])
        break
      }

      formatted =
        dt.toRelative({
          style: 'short',
          base: DateTime.fromISO(now),
        }) || ''
      break
    case 'weekday-clock':
      formatted = dt.toLocaleString({
        hour: 'numeric',
        minute: 'numeric',
        weekday: 'short',
      })
      break
    case 'clock':
      formatted = dt.toLocaleString(DateTime.TIME_SIMPLE)
      break
    case 'locale':
      if (omit && isSameDay(dt, omit)) {
        formatted = dt.toLocaleString(DateTime.TIME_SIMPLE)
        break
      }

      formatted = dt.toLocaleString(DateTime.DATETIME_MED)
      break
    default:
      throw new Error('invalid format ' + format)
  }
  if (!formatted) throw new Error('invalid time ' + time)

  return formatted
}

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
