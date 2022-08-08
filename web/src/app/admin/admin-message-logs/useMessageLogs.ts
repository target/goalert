import { DateTime, DateTimeFormatOptions, Duration, Interval } from 'luxon'
import { DebugMessage } from '../../../schema'

interface MessageLogData {
  filteredData: DebugMessage[]
  graphData: { date: string; label: string; count: number }[]
}

export type Options = {
  data: DebugMessage[]
  start?: string
  end?: string
  search?: string
  duration: string
}

export function useMessageLogs(opts: Options): MessageLogData {
  const startDT = opts.start ? DateTime.fromISO(opts.start) : null
  const endDT = opts.end ? DateTime.fromISO(opts.end) : null

  const filteredData: DebugMessage[] = opts.data
    .filter((msg: DebugMessage) => {
      const createdAtDT = DateTime.fromISO(msg.createdAt)
      if (opts.search) {
        if (
          opts.search === msg.alertID?.toString() ||
          opts.search === msg.createdAt ||
          opts.search === msg.destination ||
          opts.search === msg.serviceID ||
          opts.search === msg.serviceName ||
          opts.search === msg.userID ||
          opts.search === msg.userName
        ) {
          return true
        }
        return false
      }
      if (startDT && startDT > createdAtDT) return false
      if (endDT && endDT < createdAtDT) return false
      return true
    })
    .sort((_a: DebugMessage, _b: DebugMessage) => {
      const a = DateTime.fromISO(_a.createdAt)
      const b = DateTime.fromISO(_b.createdAt)
      if (a < b) return 1
      if (a > b) return -1
      return 0
    })

  const hasData = filteredData?.length > 0
  const s = hasData
    ? startDT ||
      DateTime.fromISO(filteredData[filteredData.length - 1].createdAt).startOf(
        'day',
      )
    : null
  const e = hasData
    ? endDT || DateTime.fromISO(filteredData[0].createdAt).endOf('day')
    : null
  let ttlInterval: Interval | null = null
  if (s && e && hasData) {
    ttlInterval = Interval.fromDateTimes(s, e)
  }

  const logs = opts.data
  const intervals = ttlInterval?.splitBy(Duration.fromISO(opts.duration)) ?? []

  const graphData = intervals.map((interval) => {
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

    const intervalLogs = logs.filter((log) => {
      return interval.contains(DateTime.fromISO(log.createdAt))
    })

    return {
      date,
      label,
      count: intervalLogs.length,
    }
  })

  return {
    graphData,
    filteredData,
  }
}
