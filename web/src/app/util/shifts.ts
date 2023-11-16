import { DateTime, Interval } from 'luxon'
import _ from 'lodash'
import { ExplicitZone } from './luxon-helpers'

export interface SpanISO {
  start: string
  end: string
}

export function checkInterval(s: SpanISO): boolean {
  if (!s.start || !s.end) return false
  if (s.end < s.start) return false
  return true
}

export function parseInterval(s: SpanISO, zone: ExplicitZone): Interval {
  return Interval.fromDateTimes(
    DateTime.fromISO(s.start, { zone }),
    DateTime.fromISO(s.end, { zone }),
  )
}

export function trimSpans<T extends SpanISO>(
  spans: T[],
  intervals: Interval[],
  zone: ExplicitZone,
): T[] {
  intervals = Interval.merge(intervals)

  return _.flatten(
    spans.map((s) => {
      const ivl = parseInterval(s, zone)

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
