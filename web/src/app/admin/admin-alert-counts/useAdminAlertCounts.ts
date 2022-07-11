import { DateTime, Duration, Interval } from 'luxon'
import _ from 'lodash'
import { Alert } from '../../../schema'

export type AlertCountSeries = {
  serviceName: string
  id: string
  data: AlertCountDataPoint[]
}
export type AlertCountDataPoint = {
  date: string
  label: string
  total: number
}
export type AlertCountOpts = {
  int: string // iso-formatted interval
  dur: string // iso-formatted duration
  alerts: Alert[]
}

export function useAdminAlertCounts(opts: AlertCountOpts): AlertCountSeries[] {
  const alerts = opts.alerts
  const groupBySvcTest = _.groupBy(alerts, 'service.id')
  return Object.values(groupBySvcTest).map((alerts) => {
    const alertCounts: AlertCountDataPoint[] = []
    Interval.fromISO(opts.int)
      .splitBy(Duration.fromISO(opts.dur))
      .map((i) => {
        const date = i.start.toLocaleString({
          month: 'short',
          day: 'numeric',
        })
        const label = i.start.toLocaleString({
          month: 'short',
          day: 'numeric',
          year: 'numeric',
        })

        const bucket = alerts.filter((a) => {
          return i.contains(DateTime.fromISO(a.createdAt as string))
        })
        alertCounts.push({
          date,
          label,
          total: bucket.length,
        })
      })
    return {
      serviceName: alerts[0].service?.name as string,
      id: alerts[0].service?.id as string,
      data: alertCounts,
    }
  })
}
