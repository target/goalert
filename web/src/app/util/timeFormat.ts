import {
  Interval,
  DateTime,
  Duration,
  DurationLikeObject,
  DateTimeFormatOptions,
} from 'luxon'
import { ExplicitZone } from './luxon-helpers'

export const getDT = (
  t: string | DateTime | number,
  z?: ExplicitZone,
): DateTime => {
  if (DateTime.isDateTime(t)) return t.setZone(z || 'local')
  if (typeof t === 'number') return DateTime.fromMillis(t, { zone: z })

  return DateTime.fromISO(t, { zone: z })
}

export const getDur = (d: string | DurationLikeObject | Duration): Duration => {
  if (typeof d === 'string') return Duration.fromISO(d)
  if (Duration.isDuration(d)) return d
  return Duration.fromObject(d)
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

export type FormatRelativeArg = {
  dur: Duration | string | DurationLikeObject
  noQualifier?: boolean
  units?: ReadonlyArray<keyof DurationLikeObject>
  min?: Duration | string | DurationLikeObject
  precise?: boolean
}

export function formatRelative({
  dur: durArg,
  noQualifier,
  units = ['days', 'hours', 'minutes'],
  min: minArg,
  precise,
}: FormatRelativeArg): string {
  const parts = []
  let dur = getDur(durArg)
  let min = minArg ? getDur(minArg) : undefined
  if (!min) {
    // default to lowest unit
    min = Duration.fromObject({ [units[units.length - 1]]: 1 })
  }
  const neg = dur.valueOf() < 0
  let prefix = noQualifier || neg ? '' : 'in '
  const suffix = noQualifier || !neg ? '' : ' ago'
  if (dur.valueOf() < 0) dur = dur.negate()
  if (min && min.valueOf() > dur.valueOf()) {
    dur = min
    prefix = neg ? '< ' : '> '
  }

  for (const unit of units) {
    const val = Math.floor(dur.as(unit))
    if (val === 0) continue
    const part = Duration.fromObject({ [unit]: val })
    dur = dur.minus(part)

    parts.push(part.toHuman({ unitDisplay: 'short' }))
    if (!precise) break
  }

  return prefix + parts.join(', ') + suffix
}

function formatGuard(fmt: never): never {
  throw new Error('invalid time format ' + fmt)
}

export type FormatTimestampArg = {
  time: string | DateTime
  zone?: string
} & (
  | {
      format: 'relative'
      from?: string | DateTime

      // If true, the 'relative' format will include multiple units.
      precise?: boolean
      min?: string | DurationLikeObject | Duration
      units?: ReadonlyArray<keyof DurationLikeObject>
    }
  | {
      format: 'relative-date'
      from?: string | DateTime
    }
  | {
      format?: 'clock' | 'default' | 'weekday-clock'
    }
)

export function formatTimestamp(arg: FormatTimestampArg): string {
  const { zone = 'local' } = arg
  const dt = getDT(arg.time, zone)
  const from = getDT(
    'from' in arg && arg.from ? arg.from : DateTime.utc(),
    zone,
  )

  if (!arg.format || arg.format === 'default')
    return dt.toLocaleString(DateTime.DATETIME_MED)
  if (arg.format === 'clock') return dt.toLocaleString(DateTime.TIME_SIMPLE)
  if (arg.format === 'relative-date') return relativeDate(dt, from)

  if (arg.format === 'weekday-clock')
    return dt.toLocaleString({
      hour: 'numeric',
      minute: 'numeric',
      weekday: 'short',
    })

  if (arg.format === 'relative')
    return formatRelative({
      dur: dt.diff(from, [...(arg.units ?? ['days', 'hours', 'minutes'])]),
      precise: arg.precise,
      min: arg.min,
      units: arg.units,
    })

  // Create a type error if we add a new format and forget to handle it.
  formatGuard(arg.format)
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
