import { Alert } from '../../../schema'
import {
  ackDistribution,
  alertsByHour,
  durationMs,
  formatDuration,
  formatPercent,
  percentile,
  responseByService,
  summarizeResponse,
} from './util'

function alert(
  id: number,
  createdAt: string,
  svc: string,
  status: Alert['status'],
  metrics?: { timeToAck: string; timeToClose: string; escalated: boolean },
): Alert {
  return {
    id: String(id),
    alertID: id,
    createdAt,
    status,
    summary: `alert ${id}`,
    service: { id: svc, name: `Service ${svc}` },
    metrics: metrics ? { ...metrics, closedAt: createdAt } : null,
  } as unknown as Alert
}

const alerts: Alert[] = [
  alert(1, '2024-01-01T00:30:00Z', 'a', 'StatusClosed', {
    timeToAck: 'PT30S',
    timeToClose: 'PT10M',
    escalated: false,
  }),
  alert(2, '2024-01-01T09:15:00Z', 'a', 'StatusClosed', {
    timeToAck: 'PT12M',
    timeToClose: 'PT2H',
    escalated: true,
  }),
  alert(3, '2024-01-01T09:45:00Z', 'b', 'StatusClosed', {
    timeToAck: 'PT5H',
    timeToClose: 'PT6H',
    escalated: true,
  }),
  alert(4, '2024-01-02T23:59:00Z', 'a', 'StatusAcknowledged'),
  alert(5, '2024-01-03T09:00:00Z', 'b', 'StatusUnacknowledged'),
]

describe('durationMs', () => {
  it('parses ISO durations', () => {
    expect(durationMs('PT30S')).toEqual(30000)
    expect(durationMs('PT3M48.5S')).toEqual(228500)
    expect(durationMs('P1DT1H')).toEqual(25 * 3600000)
  })
  it('returns null for missing or invalid values', () => {
    expect(durationMs(null)).toBeNull()
    expect(durationMs(undefined)).toBeNull()
    expect(durationMs('nope')).toBeNull()
  })
})

describe('percentile', () => {
  it('returns null for no values', () => {
    expect(percentile([], 50)).toBeNull()
  })
  it('interpolates between ranks', () => {
    expect(percentile([1, 2, 3, 4], 50)).toEqual(2.5)
    expect(percentile([10, 1, 5], 50)).toEqual(5)
    expect(percentile([1, 2, 3, 4, 5, 6, 7, 8, 9, 10], 90)).toBeCloseTo(9.1)
    expect(percentile([7], 90)).toEqual(7)
  })
})

describe('summarizeResponse', () => {
  it('computes headline metrics from closed alerts', () => {
    const s = summarizeResponse(alerts)
    expect(s.total).toEqual(5)
    expect(s.open).toEqual(2)
    expect(s.unacknowledged).toEqual(1)
    expect(s.closed).toEqual(3)
    expect(s.escalated).toEqual(2)
    expect(s.medianAckMs).toEqual(12 * 60000)
    expect(s.medianCloseMs).toEqual(2 * 3600000)
  })
  it('handles no alerts', () => {
    const s = summarizeResponse([])
    expect(s.total).toEqual(0)
    expect(s.medianAckMs).toBeNull()
    expect(s.p90CloseMs).toBeNull()
  })
})

describe('ackDistribution', () => {
  it('buckets closed alerts by time to acknowledge', () => {
    const counts = Object.fromEntries(
      ackDistribution(alerts).map((b) => [b.label, b.count]),
    )
    expect(counts['< 1 min']).toEqual(1)
    expect(counts['5–15 min']).toEqual(1)
    expect(counts['> 4 hours']).toEqual(1)
    expect(Object.values(counts).reduce((a, b) => a + b, 0)).toEqual(3)
  })
})

describe('alertsByHour', () => {
  it('counts alerts by hour of creation in the given zone', () => {
    const hours = alertsByHour(alerts, 'UTC')
    expect(hours).toHaveLength(24)
    expect(hours[0].count).toEqual(1)
    expect(hours[9].count).toEqual(3)
    expect(hours[23].count).toEqual(1)
    expect(hours[9].label).toEqual('09:00')
  })
  it('respects the time zone', () => {
    const hours = alertsByHour(alerts, 'UTC+2')
    expect(hours[2].count).toEqual(1)
    expect(hours[11].count).toEqual(3)
  })
})

describe('responseByService', () => {
  it('groups by service, busiest first', () => {
    const svcs = responseByService(alerts)
    expect(svcs.map((s) => s.id)).toEqual(['a', 'b'])
    expect(svcs[0]).toMatchObject({
      name: 'Service a',
      total: 3,
      closed: 2,
      escalated: 1,
    })
    expect(svcs[0].medianAckMs).toEqual((30000 + 12 * 60000) / 2)
    expect(svcs[1].medianAckMs).toEqual(5 * 3600000)
  })
})

describe('formatDuration', () => {
  it('formats durations compactly', () => {
    expect(formatDuration(null)).toEqual('—')
    expect(formatDuration(45000)).toEqual('45s')
    expect(formatDuration(12 * 60000)).toEqual('12m')
    expect(formatDuration(3 * 3600000 + 5 * 60000)).toEqual('3h 5m')
    expect(formatDuration(2 * 3600000)).toEqual('2h')
    expect(formatDuration(52 * 3600000)).toEqual('2d 4h')
  })
})

describe('formatPercent', () => {
  it('formats percentages', () => {
    expect(formatPercent(1, 3)).toEqual('33%')
    expect(formatPercent(0, 0)).toEqual('—')
  })
})
