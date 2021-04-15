import { DateTime } from 'luxon'

// getStartOfWeek
// Given a luxon DateTime, this yields a luxon DateTime set to the previous Sunday at 12am.
// If no date is provided, default to now
// This is excluded from Luxon core since certain countries start the week on Saturday or Sunday
// (Conventionally, weeks start on a Monday)
export function getStartOfWeek(luxonDateTime = DateTime.local()): DateTime {
  // 1-based i.e. [Mon=1, Tues=2, ... Sat=6, Sun=7]
  const _weekdayIndex = luxonDateTime.toLocal().weekday
  const weekdayIndex = _weekdayIndex === 7 ? 0 : _weekdayIndex

  return luxonDateTime
    .toLocal()
    .minus({
      days: weekdayIndex,
    })
    .startOf('day')
}

// getEndOfWeek
export function getEndOfWeek(luxonDateTime = DateTime.local()): DateTime {
  // 1-based i.e. [Mon=1, Tues=2, ... Sat=6, Sun=7]
  const _weekdayIndex = luxonDateTime.toLocal().weekday
  const weekdayIndex = _weekdayIndex === 7 ? 0 : _weekdayIndex

  return luxonDateTime
    .toLocal()
    .plus({
      days: 6 - weekdayIndex,
    })
    .endOf('day')
}
