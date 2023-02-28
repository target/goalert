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
} & (
  | {
      format: 'relative'

      // now is the current time to use for relative time calculations, defaults to
      // DateTime.utc().toISO()
      now?: string

      // If true, the 'relative' format will include multiple units.
      precise?: boolean
    }
  | {
      format: 'relative-date'

      // now is the current time to use for relative time calculations, defaults to
      // DateTime.utc().toISO()
      now?: string
    }
  | {
      format: 'clock' | 'default' | 'weekday-clock'
    }
)

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

function formatGuard(fmt: never): never {
  throw new Error('invalid time format ' + fmt)
}

export function formatTimestamp(opts: TimeFormatOpts): string {
  const { time, zone = 'local' } = opts

  const dt = DateTime.fromISO(time, { zone })
  const now = DateTime.utc().toISO()

  if (opts.format === 'default') return dt.toLocaleString(DateTime.DATETIME_MED)
  if (opts.format === 'clock') return dt.toLocaleString(DateTime.TIME_SIMPLE)

  if (opts.format === 'relative-date')
    return relativeDate(dt, DateTime.fromISO(opts.now || now))

  if (opts.format === 'weekday-clock')
    return dt.toLocaleString({
      hour: 'numeric',
      minute: 'numeric',
      weekday: 'short',
    })

  if (opts.format === 'relative' && opts.precise)
    return toRelativePrecise(dt.diff(DateTime.fromISO(now)), [
      'days',
      'hours',
      'minutes',
    ])

  if (opts.format === 'relative')
    return (
      dt.toRelative({
        style: 'short',
        base: DateTime.fromISO(opts.now || now),
      }) || ''
    )

  // Create a type error if we add a new format and forget to handle it.
  formatGuard(opts.format)
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
