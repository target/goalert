import { DateTime, Duration } from 'luxon'
import { Alert } from '../../../schema'

function formatCSVField(data: string): string {
  if (!/[,"\r\n]/.test(data)) return data

  return `"${data.replace(/"/g, '""')}"`
}

export type useAlertCSVOpts = {
  alerts: Alert[]
  urlPrefix: string
}

export function useAlertCSV({ urlPrefix, alerts }: useAlertCSVOpts): string {
  let data = ''
  const zoneAbbr = DateTime.local().toFormat('ZZZZ Z')
  const cols = [
    `Created At (${zoneAbbr})`,
    `Closed At (${zoneAbbr})`,
    `Time to Ack`,
    `Time to Close`,
    `Alert ID`,
    `Escalated`,
    `Noise Reason`,
    `Status`,
    `Summary`,
    `Details`,
    `Service Name`,
    `Service URL`,
  ]
  data += cols.map(formatCSVField).join(',') + '\r\n'

  const rows = alerts.map(
    (a) =>
      [
        DateTime.fromISO(a.createdAt).toLocal().toSQL({
          includeOffset: false,
        }),
        DateTime.fromISO(a.metrics?.closedAt as string)
          .toLocal()
          .toSQL({
            includeOffset: false,
          }),
        Duration.fromISO(a.metrics?.timeToAck as string).toFormat('hh:mm:ss'),
        Duration.fromISO(a.metrics?.timeToClose as string).toFormat('hh:mm:ss'),
        a.alertID.toString(),
        (a.metrics?.escalated as boolean).toString(),
        a.noiseReason || '',
        a.status.replace('Status', ''),
        a.summary,
        a.details,
        a.service?.name || '',
        `${urlPrefix}/services/${a.service?.id}`,
      ]
        .map(formatCSVField)
        .join(',') + '\r\n',
  )

  return data + rows.join('')
}
