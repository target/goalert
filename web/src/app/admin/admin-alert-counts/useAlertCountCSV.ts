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
    let max = 0
    let total = 0
    for (let i = 0; i < svc.data.length; i++) {
      if (svc.data[i].total > max) max = svc.data[i].total
      total += svc.data[i].total
    }
    return (
      [
        svc.serviceName || '',
        `${urlPrefix}/services/${svc.id}`,
        `${total}`,
        `${max}`,
        `${total / svc.data.length}`,
      ]
        .map(formatCSVField)
        .join(',') + '\r\n'
    )
  })
  return data + rows.join('')
}
