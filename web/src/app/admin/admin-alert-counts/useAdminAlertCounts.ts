import { DateTime, Duration, Interval } from 'luxon'
import _ from 'lodash'
import { Alert } from '../../../schema'

export type AlertCountSeries = {
  serviceName: string
  dailyCounts: AlertCountDataPoint[]
  id: string
  total: number
  max: number
  avg: number
}
export type AlertCountDataPoint = {
  date: string
  dayTotal: number
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
    let svcTotal = 0
    let svcMax = 0
    Interval.fromISO(opts.int)
      .splitBy(Duration.fromISO(opts.dur))
      .map((i) => {
        const bucket = alerts.filter((a) => {
          return i.contains(DateTime.fromISO(a.createdAt as string))
        })
        alertCounts.push({
          date: i.start.toUTC().toISO(),
          dayTotal: bucket.length,
        })
        svcTotal += bucket.length
        if (bucket.length > svcMax) svcMax = bucket.length
      })
    return {
      serviceName: alerts[0].service?.name as string,
      id: alerts[0].service?.id as string,
      dailyCounts: alertCounts,
      total: svcTotal,
      max: svcMax,
      avg: Math.round(svcTotal / alertCounts.length),
    }
  })
}
