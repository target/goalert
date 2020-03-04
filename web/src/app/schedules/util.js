import { DateTime } from 'luxon'

export const days = [
  'Sunday',
  'Monday',
  'Tuesday',
  'Wednesday',
  'Thursday',
  'Friday',
  'Saturday',
]

// Shifts a weekdayFilter so that it matches the luxon day n
//
// Default is 7 (Sunday)
export function alignWeekdayFilter(n, filter) {
  if (n === 7) return filter

  return filter.slice(7 - n).concat(filter.slice(0, 7 - n))
}

export function mapRuleTZ(fromTZ, toTZ, rule) {
  const start = parseClock(rule.start, fromTZ).setZone(toTZ)
  const end = parseClock(rule.end, fromTZ).setZone(toTZ)
  return {
    ...rule,
    start: formatClock(start),
    end: formatClock(end),
    weekdayFilter: alignWeekdayFilter(start.weekday, rule.weekdayFilter),
  }
}

// gqlClockTimeToISO will return an ISO timestamp representing
// the given GraphQL ClockTime value at the current date in the
// provided time zone.
export function gqlClockTimeToISO(time, zone) {
  return DateTime.fromFormat(time, 'HH:mm', { zone })
    .toUTC()
    .toISO()
}

// isoToGQLClockTime will return a GraphQL ClockTime value for
// the given ISO timestamp, with respect to the provided time zone.
export function isoToGQLClockTime(timestamp, zone) {
  return DateTime.fromISO(timestamp, { zone }).toFormat('HH:mm')
}

export function weekdaySummary(filter) {
  const bin = filter.map(f => (f ? '1' : '0')).join('')
  switch (bin) {
    case '1000001':
      return 'Weekends'
    case '0000000':
      return 'Never'
    case '0111110':
      return 'M—F'
    case '0111111':
      return 'M—F and Sat'
    case '1111110':
      return 'M—F and Sun'
    case '1111111':
      return 'Everyday'
  }

  const d = []
  let chain = []
  const flush = () => {
    if (chain.length < 3) {
      chain.forEach(day => d.push(day.slice(0, 3)))
      chain = []
      return
    }

    d.push(chain[0].slice(0, 3) + '—' + chain[chain.length - 1].slice(0, 3))
    chain = []
  }
  days.forEach((day, idx) => {
    if (filter[idx]) {
      chain.push(day)
      return
    }
    flush()
  })
  flush()
  return d.join(', ')
}

export function parseClock(s, zone) {
  const dt = DateTime.fromObject({
    hours: parseInt(s.split(':')[0], 10),
    minutes: parseInt(s.split(':')[1], 10),
    weekday: 7, // sunday
    zone,
  })

  // backtrack if we jumped over DST
  if (dt.offset !== DateTime.utc().setZone('zone').offset)
    return dt.minus({ weeks: 1 })

  return dt
}

export function formatClock(dt) {
  return `${dt.hour
    .toString()
    .padStart(2, '0')}:${dt.minute.toString().padStart(2, '0')}`
}

export function ruleSummary(rules, scheduleZone, displayZone) {
  const everyDay = r => !r.weekdayFilter.some(w => !w) && r.start === r.end

  rules = rules.filter(r => r.weekdayFilter.some(w => w)) // ignore disabled
  if (rules.length === 0) return 'Never'
  if (rules.some(everyDay)) return 'Always'

  const getTime = str => parseClock(str, scheduleZone).setZone(displayZone)

  return rules
    .map(r => {
      const start = getTime(r.start)
      const weekdayFilter = alignWeekdayFilter(start.weekday, r.weekdayFilter)
      return `${weekdaySummary(weekdayFilter)} from ${start.toLocaleString(
        DateTime.TIME_SIMPLE,
      )} to ${getTime(r.end).toLocaleString(DateTime.TIME_SIMPLE)} `
    })
    .join('\n')
}

export function formatOverrideTime(_start, _end, zone) {
  const start =
    _start instanceof DateTime
      ? _start.setZone(zone)
      : DateTime.fromISO(_start, { zone })
  const end =
    _end instanceof DateTime
      ? _end.setZone(zone)
      : DateTime.fromISO(_end, { zone })
  const sameDay = start.startOf('day').equals(end.startOf('day'))
  return `${start.toLocaleString(
    DateTime.DATETIME_MED,
  )} to ${end.toLocaleString(
    sameDay ? DateTime.TIME_SIMPLE : DateTime.DATETIME_MED,
  )}`
}

export function mapOverrideUserError(conflictingOverride, value, zone) {
  if (!conflictingOverride) return []

  const errs = []
  const isReplace =
    conflictingOverride.addUser && conflictingOverride.removeUser

  const replaceMsg = add =>
    add
      ? `replacing ${conflictingOverride.removeUser.name}`
      : `replaced by ${conflictingOverride.addUser.name}`

  const time = formatOverrideTime(
    conflictingOverride.start,
    conflictingOverride.end,
    zone,
  )

  const check = (valueField, errField) => {
    if (!conflictingOverride[errField]) return
    const verb = errField === 'addUser' ? 'added' : 'removed'
    if (value[valueField] === conflictingOverride[errField].id) {
      errs.push({
        field: valueField,
        message: `Already ${
          isReplace ? replaceMsg(errField === 'addUser') : verb
        } from ${time}`,
      })
    }
  }
  check('addUserID', 'addUser')
  check('addUserID', 'removeUser')
  check('removeUserID', 'addUser')
  check('removeUserID', 'removeUser')

  return errs
}
