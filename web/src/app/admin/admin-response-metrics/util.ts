import { Duration } from 'luxon'
import _ from 'lodash'
import { Alert } from '../../../schema'

type Counts = {
  total: number
  closed: number
  escalated: number
  medianAckMs: number | null
  medianCloseMs: number | null
}
export type ResponseSummary = Counts & {
  open: number
  unacknowledged: number
  p90AckMs: number | null
  p90CloseMs: number | null
}
export type ServiceResponse = Counts & { id: string; name: string }

// Upper bounds (in minutes, exclusive) for the time-to-acknowledge histogram.
const ACK_BUCKETS: { label: string; max: number }[] = [
  { label: '< 1m', max: 1 },
  { label: '1–5m', max: 5 },
  { label: '5–15m', max: 15 },
  { label: '15–30m', max: 30 },
  { label: '30–60m', max: 60 },
  { label: '1–4h', max: 240 },
  { label: '> 4h', max: Infinity },
]

// durationMs converts an ISO-8601 duration (e.g. "PT3M48.6S") to milliseconds,
// returning null for missing or invalid values.
export function durationMs(iso?: string | null): number | null {
  if (!iso) return null
  const dur = Duration.fromISO(iso)
  return dur.isValid ? dur.as('milliseconds') : null
}

// percentile returns the p-th percentile (0-100) of values using linear
// interpolation between closest ranks, or null if there are no values.
export function percentile(values: number[], p: number): number | null {
  if (!values.length) return null
  const sorted = [...values].sort((a, b) => a - b)
  const rank = (p / 100) * (sorted.length - 1)
  const lo = Math.floor(rank)
  return sorted[lo] + (sorted[Math.ceil(rank)] - sorted[lo]) * (rank - lo)
}

// times returns a metric in ms for each alert that has one (closed alerts only).
function times(alerts: Alert[], key: 'timeToAck' | 'timeToClose'): number[] {
  return alerts
    .map((a) => durationMs(a.metrics?.[key]))
    .filter((v): v is number => v !== null)
}

const countClosed = (alerts: Alert[]): number =>
  alerts.filter((a) => a.status === 'StatusClosed').length
const countEscalated = (alerts: Alert[]): number =>
  alerts.filter((a) => a.metrics?.escalated).length

// summarizeResponse computes headline response metrics.
export function summarizeResponse(alerts: Alert[]): ResponseSummary {
  const ack = times(alerts, 'timeToAck')
  const close = times(alerts, 'timeToClose')
  const closed = countClosed(alerts)
  return {
    total: alerts.length,
    open: alerts.length - closed,
    unacknowledged: alerts.filter((a) => a.status === 'StatusUnacknowledged')
      .length,
    closed,
    escalated: countEscalated(alerts),
    medianAckMs: percentile(ack, 50),
    p90AckMs: percentile(ack, 90),
    medianCloseMs: percentile(close, 50),
    p90CloseMs: percentile(close, 90),
  }
}

// ackDistribution groups closed alerts by how long they took to be acknowledged.
export function ackDistribution(
  alerts: Alert[],
): { label: string; count: number }[] {
  const buckets = ACK_BUCKETS.map((b) => ({ label: b.label, count: 0 }))
  times(alerts, 'timeToAck').forEach((ms) => {
    buckets[ACK_BUCKETS.findIndex((b) => ms / 60000 < b.max)].count++
  })
  return buckets
}

// responseByService computes per-service response metrics, busiest first.
export function responseByService(alerts: Alert[]): ServiceResponse[] {
  return Object.values(_.groupBy(alerts, (a) => a.service?.id ?? ''))
    .map((svcAlerts) => ({
      id: svcAlerts[0].service?.id ?? '',
      name: svcAlerts[0].service?.name ?? '',
      total: svcAlerts.length,
      closed: countClosed(svcAlerts),
      escalated: countEscalated(svcAlerts),
      medianAckMs: percentile(times(svcAlerts, 'timeToAck'), 50),
      medianCloseMs: percentile(times(svcAlerts, 'timeToClose'), 50),
    }))
    .sort((a, b) => b.total - a.total || a.name.localeCompare(b.name))
}

// formatDuration renders milliseconds as a short human-readable duration
// (e.g. "45s", "12m", "3h 5m", "2d 4h"), or an em dash for null.
export function formatDuration(ms: number | null): string {
  if (ms === null) return '—'
  const sec = Math.round(ms / 1000)
  if (sec < 60) return `${sec}s`
  const min = Math.round(sec / 60)
  if (min < 60) return `${min}m`
  const hours = Math.floor(min / 60)
  if (hours < 24) return min % 60 ? `${hours}h ${min % 60}m` : `${hours}h`
  const days = Math.floor(hours / 24)
  return hours % 24 ? `${days}d ${hours % 24}h` : `${days}d`
}

// formatPercent renders part/whole as a whole-number percentage.
export function formatPercent(part: number, whole: number): string {
  return whole ? `${Math.round((part / whole) * 100)}%` : '—'
}
