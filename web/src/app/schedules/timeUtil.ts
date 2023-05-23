import { DateTime } from 'luxon'

export interface Spanable {
  start: string
  end: string
}

export function ensureInterval<T extends Spanable, N extends Spanable>(
  value: T,
  newValue: N,
): N {
  const oldDur = DateTime.fromISO(value.end).diff(DateTime.fromISO(value.start))
  if (value.start !== newValue.start) {
    const oldEnd = DateTime.fromISO(value.end)
    const newStart = DateTime.fromISO(newValue.start)
    if (newStart >= oldEnd) {
      // if start time is put after end time, move end time forward
      newValue.end = newStart.plus(oldDur).toUTC().toISO()
    }
  } else if (value.end !== newValue.end) {
    const newEnd = DateTime.fromISO(newValue.end)
    const start = DateTime.fromISO(newValue.start)
    if (newEnd <= start) {
      // if end time is put before start time, move start time back
      newValue.start = newEnd.minus(oldDur).toUTC().toISO()
    }
  }

  return newValue
}
