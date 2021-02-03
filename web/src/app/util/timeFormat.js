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
    return `> ${Math.floor(diff.as('months'))}mo ago`
  }

  return `> ${Math.floor(diff.as('years'))}y ago`
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
