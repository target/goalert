import { DateTime, Duration } from 'luxon'
import { useURLParams } from '../../actions'

export type MessageLogsParams = {
  search: string
  start: string
  end: string
  graphInterval: string
}

export function useMessageLogsParams(): [
  MessageLogsParams,
  (params: Partial<MessageLogsParams>) => void,
] {
  const end = DateTime.now().startOf('hour').plus({ hour: 1 })

  return useURLParams<MessageLogsParams>({
    search: '',
    start: end.plus({ days: -1 }).toISO(),
    end: end.toISO(),
    graphInterval: Duration.fromObject({ hours: 1 }).toISO(),
  })
}
