import { DateTime, Duration } from 'luxon'
import { useURLParams } from '../../actions'

export type IntervalOption = {
  label: string
  value: string
}

export const INTERVAL_OPTIONS: Array<IntervalOption> = [
  { year: 1 },
  { month: 1 },
  { week: 1 },
  { day: 1 },
  { hour: 8 },
  { hour: 1 },
  { minute: 15 },
  { minute: 5 },
  { minute: 1 },
].map((o) => ({
  label: Duration.fromObject(o).toHuman(),
  value: Duration.fromObject(o).toISO(),
}))

// getValidIntervals returns all intervals that are valid for the given span.
//
// A valid interval is one that will return between 4 and 1000 data points.
export function getValidIntervals(s: {
  start: string
  end: string
}): Array<IntervalOption> {
  const start = DateTime.fromISO(s.start)
  const end = DateTime.fromISO(s.end)
  const duration = end.diff(start).as('milliseconds')

  return INTERVAL_OPTIONS.filter((o) => {
    const interval = Duration.fromISO(o.value).as('milliseconds')
    const points = duration / interval
    return points >= 4 && points <= 1000
  })
}

// getClosestInterval returns the closest valid interval to the given graph interval.
export function getClosestInterval(s: {
  start: string
  end: string
  graphInterval: string
}): string {
  const intervals = getValidIntervals(s)
  const graphInterval = Duration.fromISO(s.graphInterval).as('milliseconds')

  let closest = intervals[0]
  let closestDiff = Math.abs(
    Duration.fromISO(closest.value).as('milliseconds') - graphInterval,
  )

  for (const interval of intervals) {
    const diff = Math.abs(
      Duration.fromISO(interval.value).as('milliseconds') - graphInterval,
    )
    if (diff < closestDiff) {
      closest = interval
      closestDiff = diff
    }
  }

  return closest.value
}

export type MessageLogsParams = {
  search: string
  start: string
  end: string
  graphInterval: string
  segmentBy: string
}

export function useMessageLogsParams(): [
  MessageLogsParams,
  (params: Partial<MessageLogsParams>) => void,
] {
  const end = DateTime.now().startOf('hour').plus({ hour: 1 })
  const start = end.plus({ days: -1 })

  const [params, setParams] = useURLParams<MessageLogsParams>({
    search: '',
    start: start.toISO(),
    end: end.toISO(),
    graphInterval: getValidIntervals({
      start: start.toISO(),
      end: end.toISO(),
    })[0].value,
    segmentBy: '',
  })

  return [
    params,
    (newParams: Partial<MessageLogsParams>) => {
      const s = { ...params, ...newParams }
      s.start = s.start || end.plus({ days: -1 }).toISO()
      s.end = s.end || end.toISO()
      s.graphInterval = getClosestInterval(s)
      setParams(s)
    },
  ]
}
