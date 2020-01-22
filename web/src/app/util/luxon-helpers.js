import { DateTime } from 'luxon'

// getLuxonStartOfWeek
// Given a JS Date object, this yields a DateTime object set to the previous Sunday at 12am.
// If no date is provided, default to now
// This is excluded from Luxon core since certain countries start the week on Saturday or Sunday
// Conventionally, weeks start on a Monday
export function getLuxonStartOfWeek(JSDate = new Date()) {
  // 1-based i.e. [Mon=1, Tues=2, ... Sat=6, Sun=7]
  const _weekdayIndex = DateTime.fromJSDate(JSDate).toLocal().weekday
  const weekdayIndex = _weekdayIndex === 7 ? 0 : _weekdayIndex

  return DateTime.fromJSDate(JSDate)
    .toLocal()
    .minus({
      days: weekdayIndex,
    })
    .startOf('day')
}

// getLuxonEndOfWeek
export function getLuxonEndOfWeek(JSDate = new Date()) {
  // 1-based i.e. [Mon=1, Tues=2, ... Sat=6, Sun=7]
  const _weekdayIndex = DateTime.fromJSDate(JSDate).toLocal().weekday
  const weekdayIndex = _weekdayIndex === 7 ? 0 : _weekdayIndex

  return DateTime.fromJSDate(JSDate)
    .toLocal()
    .plus({
      days: 6 - weekdayIndex,
    })
    .endOf('day')
}
