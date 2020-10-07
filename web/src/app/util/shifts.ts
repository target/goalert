import { DateTime, Interval } from 'luxon'
import _ from 'lodash-es'

interface SpanISO {
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
