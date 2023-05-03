import { DateTime, DateTimeFormatOptions, Duration, Interval } from 'luxon'
import { DebugMessage } from '../../../schema'

interface MessageLogGraphData {
  date: string
  label: string
  count: number
}

interface MessageLogGraphDataOptions {
  start: string
  end: string
  duration: string
  logs: DebugMessage[]
}

export function useMessageLogGraphData(
  opts: MessageLogGraphDataOptions,
): MessageLogGraphData[] {
  const { start, end, duration, logs } = opts
  const ttlInterval = Interval.fromDateTimes(
    DateTime.fromISO(
      start || DateTime.now().minus({ hours: 8 }).toISO(), // if no start set, show past 8 hours
    ),
    DateTime.fromISO(end || DateTime.now().toISO()),
  )

  const intervals = ttlInterval?.splitBy(Duration.fromISO(duration)) ?? []

  return intervals.map((interval) => {
    const locale: DateTimeFormatOptions = {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: 'numeric',
      minute: 'numeric',
    }

    const date = interval.start.toLocaleString({
      month: 'short',
      day: 'numeric',
      hour: 'numeric',
      minute: 'numeric',
    })
    const label =
      interval.start.toLocaleString(locale) +
      ' - ' +
      interval.end.toLocaleString(locale)

    const intervalLogs = logs.filter((log: DebugMessage) =>
      interval.contains(DateTime.fromISO(log.createdAt)),
    )

    return {
      date,
      label,
      count: intervalLogs.length,
    } as MessageLogGraphData
  })
}
