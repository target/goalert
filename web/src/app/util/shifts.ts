import { DateTime, Interval } from 'luxon'
import _ from 'lodash'

export interface SpanISO {
  start: string
  end: string
}

export function parseInterval(s: SpanISO): Interval {
  return Interval.fromDateTimes(
    DateTime.fromISO(s.start),
    DateTime.fromISO(s.end),
  )
}

export function trimSpans<T extends SpanISO>(
  spans: T[],
  ...intervals: Interval[]
): T[] {
  intervals = Interval.merge(intervals)

  return _.flatten(
    spans.map((s) => {
      const ivl = parseInterval(s)

      return ivl.difference(...intervals).map((ivl) => ({
        ...s,
        start: ivl.start.toISO(),
        end: ivl.end.toISO(),
      }))
    }),
  )
}

// isISOAfter
// Compares two ISO timestamps, returning true if `a` occurs after `b`.
export function isISOAfter(a: string, b: string): boolean {
  return DateTime.fromISO(a) > DateTime.fromISO(b)
}

// isISOBefore
// Compares two ISO timestamps, returning true if `a` occurs before `b`.
export function isISOBefore(a: string, b: string): boolean {
  return DateTime.fromISO(a) < DateTime.fromISO(b)
}
