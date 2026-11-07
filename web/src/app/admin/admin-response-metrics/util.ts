import { DateTime, Duration } from 'luxon'
import _ from 'lodash'
import { Alert } from '../../../schema'

export type ResponseSummary = {
  total: number
  open: number
  unacknowledged: number
  closed: number
  escalated: number
  medianAckMs: number | null
  p90AckMs: number | null
  medianCloseMs: number | null
  p90CloseMs: number | null
}

export type AckBucket = {
  label: string
  count: number
}

export type HourCount = {
  hour: number
  label: string
  count: number
}

export type ServiceResponse = {
  id: string
  name: string
  total: number
  closed: number
  escalated: number
  medianAckMs: number | null
  medianCloseMs: number | null
}

// Upper bounds (in minutes, exclusive) for the time-to-acknowledge histogram.
const ACK_BUCKETS: { label: string; max: number }[] = [
  { label: '< 1 min', max: 1 },
  { label: '1–5 min', max: 5 },
  { label: '5–15 min', max: 15 },
  { label: '15–30 min', max: 30 },
  { label: '30–60 min', max: 60 },
  { label: '1–4 hours', max: 240 },
  { label: '> 4 hours', max: Infinity },
]

// durationMs converts an ISO-8601 duration (e.g. "PT3M48.6S") to milliseconds,
// returning null for missing or invalid values.
export function durationMs(iso?: string | null): number | null {
  if (!iso) return null
  const dur = Duration.fromISO(iso)
  if (!dur.isValid) return null
  return dur.as('milliseconds')
}

// percentile returns the p-th percentile (0-100) of values using linear
// interpolation between closest ranks, or null if there are no values.
export function percentile(values: number[], p: number): number | null {
  if (!values.length) return null
  const sorted = [...values].sort((a, b) => a - b)
  const rank = (p / 100) * (sorted.length - 1)
  const lo = Math.floor(rank)
  const hi = Math.ceil(rank)
  return sorted[lo] + (sorted[hi] - sorted[lo]) * (rank - lo)
}

function ackTimes(alerts: Alert[]): number[] {
  return alerts
    .map((a) => durationMs(a.metrics?.timeToAck))
    .filter((v): v is number => v !== null)
}

function closeTimes(alerts: Alert[]): number[] {
  return alerts
    .map((a) => durationMs(a.metrics?.timeToClose))
    .filter((v): v is number => v !== null)
}

// summarizeResponse computes headline response metrics. Metrics are only
// available for closed alerts, so time-based values ignore open alerts.
export function summarizeResponse(alerts: Alert[]): ResponseSummary {
  const ack = ackTimes(alerts)
  const close = closeTimes(alerts)
  return {
    total: alerts.length,
    open: alerts.filter((a) => a.status !== 'StatusClosed').length,
    unacknowledged: alerts.filter((a) => a.status === 'StatusUnacknowledged')
      .length,
    closed: alerts.filter((a) => a.status === 'StatusClosed').length,
    escalated: alerts.filter((a) => a.metrics?.escalated).length,
    medianAckMs: percentile(ack, 50),
    p90AckMs: percentile(ack, 90),
    medianCloseMs: percentile(close, 50),
    p90CloseMs: percentile(close, 90),
  }
}

// ackDistribution groups closed alerts by how long they took to be acknowledged.
export function ackDistribution(alerts: Alert[]): AckBucket[] {
  const buckets = ACK_BUCKETS.map((b) => ({ label: b.label, count: 0 }))
  ackTimes(alerts).forEach((ms) => {
    const min = ms / 60000
    const idx = ACK_BUCKETS.findIndex((b) => min < b.max)
    buckets[idx].count++
  })
  return buckets
}

// alertsByHour counts alerts by the hour of day they were created in the
// given time zone (defaults to the browser's local zone).
export function alertsByHour(alerts: Alert[], zone = 'local'): HourCount[] {
  const hours: HourCount[] = _.range(24).map((hour) => ({
    hour,
    label: DateTime.fromObject({ hour }).toFormat('HH:mm'),
    count: 0,
  }))
  alerts.forEach((a) => {
    const dt = DateTime.fromISO(a.createdAt, { zone })
    if (dt.isValid) hours[dt.hour].count++
  })
  return hours
}

// responseByService computes per-service response metrics, busiest first.
export function responseByService(alerts: Alert[]): ServiceResponse[] {
  return Object.values(_.groupBy(alerts, (a) => a.service?.id ?? ''))
    .map((svcAlerts) => ({
      id: svcAlerts[0].service?.id ?? '',
      name: svcAlerts[0].service?.name ?? '',
      total: svcAlerts.length,
      closed: svcAlerts.filter((a) => a.status === 'StatusClosed').length,
      escalated: svcAlerts.filter((a) => a.metrics?.escalated).length,
      medianAckMs: percentile(ackTimes(svcAlerts), 50),
      medianCloseMs: percentile(closeTimes(svcAlerts), 50),
    }))
    .sort((a, b) => b.total - a.total || a.name.localeCompare(b.name))
}

// formatDuration renders milliseconds as a short human-readable duration
// (e.g. "45s", "12m", "3h 5m", "2d 4h"), or an em dash for null.
export function formatDuration(ms: number | null): string {
  if (ms === null) return '—'
  const totalSec = Math.round(ms / 1000)
  if (totalSec < 60) return `${totalSec}s`
  const totalMin = Math.round(totalSec / 60)
  if (totalMin < 60) return `${totalMin}m`
  const hours = Math.floor(totalMin / 60)
  if (hours < 24) {
    const min = totalMin % 60
    return min ? `${hours}h ${min}m` : `${hours}h`
  }
  const days = Math.floor(hours / 24)
  const hrs = hours % 24
  return hrs ? `${days}d ${hrs}h` : `${days}d`
}

// formatPercent renders part/whole as a whole-number percentage.
export function formatPercent(part: number, whole: number): string {
  if (!whole) return '—'
  return `${Math.round((part / whole) * 100)}%`
}
