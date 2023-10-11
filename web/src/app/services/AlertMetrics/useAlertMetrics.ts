import { DateTime, Duration, Interval } from 'luxon'
import { Alert } from '../../../schema'

export type AlertMetricPoint = {
  date: string
  label: string
  count: number
  nonEscalatedCount: number
  escalatedCount: number
  noiseCount: number
  avgTimeToClose: number
  avgTimeToAck: number
}

export type AlertMetricsOpts = {
  int: string // iso-formatted interval
  dur: string // iso-formatted duration
  alerts: Alert[]
}

export function useAlertMetrics(opts: AlertMetricsOpts): AlertMetricPoint[] {
  let alerts = opts.alerts
  return Interval.fromISO(opts.int)
    .splitBy(Duration.fromISO(opts.dur))
    .map((i) => {
      const date = i.start.toLocaleString({
        month: 'short',
        day: 'numeric',
      })
      const label =
        i.start.toLocaleString({
          month: 'short',
          day: 'numeric',
          year: 'numeric',
        }) +
        ' - ' +
        i.end.toLocaleString({
          month: 'short',
          day: 'numeric',
          year: 'numeric',
        })

      const nextIvl = alerts.findIndex(
        (a) => !i.contains(DateTime.fromISO(a.metrics?.closedAt as string)),
      )
      const bucket = nextIvl === -1 ? alerts : alerts.slice(0, nextIvl)
      alerts = alerts.slice(nextIvl)

      const escalatedCount = bucket.filter((a) => a.metrics?.escalated).length
      const noiseCount = bucket.filter((a) => Boolean(a.noiseReason)).length

      return {
        date,
        label,
        count: bucket.length,
        nonEscalatedCount: bucket.length - escalatedCount,
        escalatedCount,
        noiseCount,

        // get average of a.metrics.timeToClose values
        avgTimeToClose: bucket.length
          ? bucket.reduce((acc, a) => {
              if (!a.metrics?.timeToClose) return acc
              const timeToClose = Duration.fromISO(a.metrics.timeToClose)
              return acc + Math.ceil(timeToClose.as('minutes'))
            }, 0) / bucket.length
          : 0,

        avgTimeToAck: bucket.length
          ? bucket.reduce((acc, a) => {
              if (!a.metrics?.timeToAck) return acc
              const timeToAck = Duration.fromISO(a.metrics.timeToAck)
              return acc + Math.ceil(timeToAck.as('minutes'))
            }, 0) / bucket.length
          : 0,
      }
    })
}
