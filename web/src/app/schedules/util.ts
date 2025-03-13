import { DateTime } from 'luxon'
import { ScheduleRule, User, UserOverride, WeekdayFilter } from '../../schema'
import { FieldError } from '../util/errutil'

export const days = [
  'Sunday',
  'Monday',
  'Tuesday',
  'Wednesday',
  'Thursday',
  'Friday',
  'Saturday',
]

// dstWeekOffset will return dt forward or backward a week
// if `dt.offset` does not match the `expectedOffset`.
function dstWeekOffset(expectedOffset: number, dt: DateTime) {
  if (dt.offset === expectedOffset) return dt

  dt = dt.minus({ weeks: 1 })
  if (dt.offset === expectedOffset) return dt

  return dt.plus({ weeks: 2 })
}

export function parseClock(s: string, zone: string) {
  const dt = DateTime.fromObject(
    {
      hour: parseInt(s.split(':')[0], 10),
      minute: parseInt(s.split(':')[1], 10),
      weekday: 7, // sunday
    },
    { zone },
  )

  return dstWeekOffset(DateTime.utc().setZone(zone).offset, dt)
}

export function formatClock(dt: DateTime) {
  return `${dt.hour.toString().padStart(2, '0')}:${dt.minute
    .toString()
    .padStart(2, '0')}`
}

// Shifts a weekdayFilter so that it matches the luxon day n
//
// Default is 7 (Sunday)
export function alignWeekdayFilter(n: number, filter: WeekdayFilter) {
  if (n === 7) return filter

  return filter.slice(7 - n).concat(filter.slice(0, 7 - n))
}

// gqlClockTimeToISO will return an ISO timestamp representing
// the given GraphQL ClockTime value at the current date in the
// provided time zone.
export function gqlClockTimeToISO(time: string, zone: string) {
  return DateTime.fromFormat(time, 'HH:mm', { zone }).toUTC().toISO()
}

// isoToGQLClockTime will return a GraphQL ClockTime value for
// the given ISO timestamp, with respect to the provided time zone.
export function isoToGQLClockTime(timestamp: string, zone: string) {
  return DateTime.fromISO(timestamp, { zone }).toFormat('HH:mm')
}

export function weekdaySummary(filter: boolean[]) {
  const bin = filter.map((f) => (f ? '1' : '0')).join('')
  switch (bin) {
    case '1000001':
      return 'weekends'
    case '0000000':
      return 'never'
    case '0111110':
      return 'M—F'
    case '0111111':
      return 'M—F and Sat'
    case '1111110':
      return 'M—F and Sun'
    case '1111111':
      return 'every day'
  }

  const d: string[] = []
  let chain: string[] = []
  const flush = () => {
    if (chain.length < 3) {
      chain.forEach((day) => d.push(day.slice(0, 3)))
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

export function ruleSummary(rules: ScheduleRule[], scheduleZone: string, displayZone: string) {
  const everyDay = (r: ScheduleRule) => !r.weekdayFilter.some((w) => !w) && r.start === r.end

  rules = rules.filter((r) => r.weekdayFilter.some((w) => w)) // ignore disabled
  if (rules.length === 0) return 'Never'
  if (rules.some(everyDay)) return 'Always'

  const getTime = (str: string) => parseClock(str, scheduleZone).setZone(displayZone)

  return rules
    .map((r: ScheduleRule) => {
      const start = getTime(r.start)
      const weekdayFilter = alignWeekdayFilter(start.weekday, r.weekdayFilter)
      let summary = weekdaySummary(weekdayFilter)
      summary = summary[0].toUpperCase() + summary.slice(1)
      return `${summary} from ${start.toLocaleString(
        DateTime.TIME_SIMPLE,
      )} to ${getTime(r.end).toLocaleString(DateTime.TIME_SIMPLE)} `
    })
    .join('\n')
}

export function formatOverrideTime(_start: DateTime | string, _end: DateTime | string, zone: string) {
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

export function mapOverrideUserError(
  conflictingOverride: UserOverride,
  value: UserOverride,
  zone: string
): FieldError[] {
  if (!conflictingOverride) return []

  const errs: FieldError[] = []
  const isReplace =
    conflictingOverride.addUser && conflictingOverride.removeUser

  const replaceMsg = (add: boolean) =>
    add
      ? `replacing ${conflictingOverride.removeUser?.name}`
      : `replaced by ${conflictingOverride.addUser?.name}`

  const time = formatOverrideTime(
    conflictingOverride.start,
    conflictingOverride.end,
    zone
  )

  const check = (
    valueField: keyof UserOverride,
    errField: keyof UserOverride
  ) => {
    if (!conflictingOverride[errField]) return
    const verb = errField === 'addUser' ? 'added' : 'removed'
    if (value[valueField] === (conflictingOverride[errField] as User).id) {
      errs.push({
        field: valueField,
        message: `Already ${isReplace ? replaceMsg(errField === 'addUser') : verb} from ${time}`,
      } as FieldError)
    }
  };

  check("addUserID", "addUser")
  check("addUserID", "removeUser")
  check("removeUserID", "addUser")
  check("removeUserID", "removeUser")

  return errs
}
