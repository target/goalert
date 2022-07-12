import { AlertCountSeries } from './useAdminAlertCounts'

function formatCSVField(data: string): string {
  if (!/[,"\r\n]/.test(data)) return data

  return `"${data.replace(/"/g, '""')}"`
}

export type useAlertCSVOpts = {
  alertCounts: AlertCountSeries[]
  urlPrefix: string
}

export function useAlertCountCSV({
  urlPrefix,
  alertCounts,
}: useAlertCSVOpts): string {
  let data = ''
  const cols = [`Service Name`, `Service URL`, `Total`, `Max`, `Average`]
  data += cols.map(formatCSVField).join(',') + '\r\n'

  const rows = alertCounts.map((svc) => {
    return (
      [
        svc.serviceName || '',
        `${urlPrefix}/services/${svc.id}`,
        `${svc.total}`,
        `${svc.max}`,
        `${svc.avg}`,
      ]
        .map(formatCSVField)
        .join(',') + '\r\n'
    )
  })
  return data + rows.join('')
}
