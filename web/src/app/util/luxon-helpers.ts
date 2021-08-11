import { DateTime, Interval } from 'luxon'

// getStartOfWeek returns the current or previous sunday at 00:00:00
// In GoAlert, weeks begin on Sunday
export function getStartOfWeek(dt = DateTime.now()): DateTime {
  // sunday
  if (dt.weekday === 7) {
    return dt.startOf('day')
  }

  return dt.startOf('week').minus({ day: 1 })
}

// getEndOfWeek returns the current or next saturday at 23:59:59:999
// In GoAlert, weeks end on Saturday
export function getEndOfWeek(dt = DateTime.now()): DateTime {
  // sunday
  if (dt.weekday === 7) {
    return dt.plus({ day: 1 }).endOf('week').minus({ day: 1 })
  }

  // saturday
  if (dt.weekday === 6) {
    return dt.endOf('day')
  }

  return dt.endOf('week').minus({ day: 1 })
}

// getNextWeekday returns the first instance of the given `weekday` after `since`.
// 1 is Monday and 7 is Sunday.
// Because the weekday depends on one's physical location, `zone` is an explicit parameter
export function getNextWeekday(
  weekday: number,
  since: DateTime,
  zone: string,
): DateTime {
  if (weekday < 1 || weekday > 7) {
    console.error(
      `getNextWeekday: out of bounds. got ${weekday}; want 1 <= weekday <= 7`,
    )
  }
  const start = since.setZone(zone)

  if (start.weekday === weekday) {
    return start.plus({ week: 1 }).startOf('day')
  }

  if (start.weekday < weekday) {
    return start.plus({ days: weekday - start.weekday }).startOf('day')
  }

  return start.plus({ days: weekday + (7 - start.weekday) }).startOf('day')
}

// splitAtMidnight divides an interval at each midnight between interval's start and end
//
// same day -> [inv]
// 2 days -> [inv.start -> midnight, midnight -> inv.end]
// 3 days -> [inv.start -> midnight, midnight -> midnight, midnight -> inv.end]
export function splitAtMidnight(inv: Interval): Interval[] {
  // dummy interval shifted forward 1 day
  const dummy = inv.mapEndpoints((e) => e.plus({ day: 1 }))

  const midnights: DateTime[] = []
  let iter = dummy.start
  while (iter < dummy.end) {
    midnights.push(iter.startOf('day'))
    iter = iter.plus({ day: 1 })
  }

  return inv.splitAt(...midnights)
}
