import { Alert } from '../../../schema'
import {
  ackDistribution,
  durationMs,
  formatDuration,
  formatPercent,
  percentile,
  responseByService,
  summarizeResponse,
} from './util'

function alert(
  svc: string,
  status: Alert['status'],
  timeToAck?: string,
  timeToClose?: string,
  escalated = false,
): Alert {
  return {
    status,
    service: { id: svc, name: `Service ${svc}` },
    metrics: timeToAck ? { timeToAck, timeToClose, escalated } : null,
  } as unknown as Alert
}

const alerts: Alert[] = [
  alert('a', 'StatusClosed', 'PT30S', 'PT10M'),
  alert('a', 'StatusClosed', 'PT12M', 'PT2H', true),
  alert('b', 'StatusClosed', 'PT5H', 'PT6H', true),
  alert('a', 'StatusAcknowledged'),
  alert('b', 'StatusUnacknowledged'),
]

test('durationMs', () => {
  expect(durationMs('PT30S')).toEqual(30000)
  expect(durationMs(null)).toBeNull()
  expect(durationMs('nope')).toBeNull()
})

test('percentile', () => {
  expect(percentile([], 50)).toBeNull()
  expect(percentile([1, 2, 3, 4], 50)).toEqual(2.5)
  expect(percentile([10, 1, 5], 50)).toEqual(5)
  expect(percentile([1, 2, 3, 4, 5, 6, 7, 8, 9, 10], 90)).toBeCloseTo(9.1)
})

test('summarizeResponse', () => {
  expect(summarizeResponse(alerts)).toMatchObject({
    total: 5,
    open: 2,
    unacknowledged: 1,
    closed: 3,
    escalated: 2,
    medianAckMs: 12 * 60000,
    medianCloseMs: 2 * 3600000,
  })
  expect(summarizeResponse([])).toMatchObject({ total: 0, medianAckMs: null })
})

test('ackDistribution', () => {
  const counts = ackDistribution(alerts).map((b) => b.count)
  expect(counts).toEqual([1, 0, 1, 0, 0, 0, 1])
})

test('responseByService', () => {
  const svcs = responseByService(alerts)
  expect(svcs.map((s) => s.id)).toEqual(['a', 'b'])
  expect(svcs[0]).toMatchObject({ total: 3, closed: 2, escalated: 1 })
  expect(svcs[0].medianAckMs).toEqual((30000 + 12 * 60000) / 2)
  expect(svcs[1].medianAckMs).toEqual(5 * 3600000)
})

test('formatDuration', () => {
  expect(formatDuration(null)).toEqual('—')
  expect(formatDuration(45000)).toEqual('45s')
  expect(formatDuration(12 * 60000)).toEqual('12m')
  expect(formatDuration(3 * 3600000 + 5 * 60000)).toEqual('3h 5m')
  expect(formatDuration(52 * 3600000)).toEqual('2d 4h')
})

test('formatPercent', () => {
  expect(formatPercent(1, 3)).toEqual('33%')
  expect(formatPercent(0, 0)).toEqual('—')
})
