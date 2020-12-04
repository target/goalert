import { DateTime } from 'luxon'

// calcNewActiveIndex returns the newActiveIndex for a swap operation
// -1 will be returned if there was no change
export function calcNewActiveIndex(oldActiveIndex, oldIndex, newIndex) {
  if (oldIndex === newIndex) {
    return -1
  }
  if (oldActiveIndex === oldIndex) {
    return newIndex
  }

  if (oldIndex > oldActiveIndex && newIndex <= oldActiveIndex) {
    return oldActiveIndex + 1
  }

  if (oldIndex < oldActiveIndex && newIndex >= oldActiveIndex) {
    return oldActiveIndex - 1
  }
  return -1
}

// formatTime returns the formatted time with the timezone (if different than local timezone)
export function formatTime(time, tz) {
  const schedTime = DateTime.fromISO(time, { zone: tz }).toLocaleString(
    DateTime.TIME_SIMPLE,
  )

  const localTime = DateTime.fromISO(time).toLocaleString(DateTime.TIME_SIMPLE)

  if (schedTime === localTime) {
    return `${schedTime} ${tz}`
  }

  return `${schedTime} ${tz} (${localTime} local)`
}

// formatDay returns the day given a time and timezone
export function formatDay(time, tz) {
  const day = DateTime.fromISO(time, { zone: tz }).weekdayLong
  const localDay = DateTime.fromISO(time).weekdayLong

  if (day === localDay) {
    return `${day}`
  }

  return `${day} (${localDay})`
}

// formatWeeklySummary returns the summary for a weekly rotation
// taking into consideration extra formatting needed if timezone does not match with local timezone
export function formatWeeklySummary(shiftLength, start, tz) {
  let details = ''
  const day = DateTime.fromISO(start, { zone: tz }).weekdayLong
  const schedTime = DateTime.fromISO(start, { zone: tz }).toLocaleString(
    DateTime.TIME_SIMPLE,
  )
  const localDay = DateTime.fromISO(start).weekdayLong
  const localTime = DateTime.fromISO(start).toLocaleString(DateTime.TIME_SIMPLE)

  details += 'Hands off '
  details += shiftLength === 1 ? 'weekly on' : `every ${shiftLength} weeks on`
  details += ` ${day}` + ' at ' + schedTime + ' ' + tz

  if (day !== localDay || schedTime !== localTime) {
    details += ' (' + localDay + ' at ' + localTime + ' local time)'
  }

  details += '.'

  return details
}

// handoffSummary returns the summary description for the rotation
export function handoffSummary(rotation) {
  const tz = rotation.timeZone

  if (!tz) return 'Loading handoff information...'

  let details = ''
  switch (rotation.type) {
    case 'hourly':
      details += 'First hand off time at ' + formatTime(rotation.start, tz)
      details +=
        ', hands off every ' +
        (rotation.shiftLength === 1
          ? 'hour'
          : rotation.shiftLength + ' hours') +
        '.'
      break
    case 'daily':
      details += 'Hands off '
      details +=
        rotation.shiftLength === 1
          ? 'daily at'
          : `every ${rotation.shiftLength} days at`
      details += ' ' + formatTime(rotation.start, tz) + '.'
      break
    case 'weekly':
      details += formatWeeklySummary(rotation.shiftLength, rotation.start, tz)
      break
  }

  return details
}

// reorderList will move an item from the oldIndex to the newIndex, preserving order
// returning the result as a new array.
export function reorderList(_items, oldIndex, newIndex) {
  const items = _items.slice()
  items.splice(oldIndex, 1) // remove 1 element from oldIndex position
  items.splice(newIndex, 0, _items[oldIndex]) // add dest to newIndex position
  return items
}
