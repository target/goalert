import { DateTime } from 'luxon'

// selectedDaysSinceTimestamp takes a timestamp and returns the number of days until the timestamp, based on the selected options.
//
// If there are no matching options, it returns 0.
export function selectedDaysUntilTimestamp(
  ts: string,
  dayOptions: number[],
  _from: string = DateTime.utc().toISO(),
): number {
  const dt = DateTime.fromISO(ts)
  const days = Math.round(dt.diff(DateTime.fromISO(_from), 'days').days)

  if (dayOptions.includes(days)) {
    return days
  }

  return 0
}
