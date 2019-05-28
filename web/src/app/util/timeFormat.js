import { Interval, DateTime } from 'luxon'

export function relativeDate(_to, _from = DateTime.utc()) {
  const to = _to instanceof DateTime ? _to : DateTime.fromISO(_to)
  const from = (_from instanceof DateTime ? _from : DateTime.fromISO(_from))
    .setZone(to.zoneName)
    .startOf('day')

  let fmt = {
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
