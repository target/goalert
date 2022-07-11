import { DateTime, Duration, Interval } from 'luxon'
import _ from 'lodash'
import { Alert, Service } from '../../../schema'

export type ServiceCounts = {
  service: Service
  total: number
}

export type AlertMetricPoint = {
  date: string
  label: string
  serviceCounts: ServiceCounts[]
}

export type AlertMetricsOpts = {
  int: string // iso-formatted interval
  dur: string // iso-formatted duration
  alerts: Alert[]
}

export type AlertCountSeries = {
  serviceName: string
  data: AlertCountDataPoint[]
}
export type AlertCountDataPoint = {
  date: string
  label: string
  total: number
}

export function useAdminAlertCounts(
  opts: AlertMetricsOpts,
): AlertCountSeries[] {
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
      data: alertCounts,
    }
  })
}
