import { DateTime, Interval } from 'luxon'
import _ from 'lodash-es'

interface ISOSpan {
  start: string
  end: string
}

export function parseInterval(s: ISOSpan): Interval {
  return Interval.fromDateTimes(
    DateTime.fromISO(s.start),
    DateTime.fromISO(s.end),
  )
}

export function trimSpans<T extends ISOSpan>(
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
