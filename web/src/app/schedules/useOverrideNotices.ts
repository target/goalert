import { Notice, TemporarySchedule } from '../../schema'
import { parseInterval, SpanISO } from '../util/shifts'
import { useQuery, gql } from '@apollo/client'
import { Interval } from 'luxon'

const scheduleQuery = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      id
      name
      timeZone
      temporarySchedules {
        start
        end
      }
    }
  }
`
export default function useOverrideNotices(
  scheduleID: string,
  value: SpanISO,
): Notice[] {
  const { data, loading } = useQuery(scheduleQuery, {
    variables: {
      id: scheduleID,
    },
    pollInterval: 0,
  })
  if (loading) {
    return []
  }
  const tempSchedules = data?.schedule?.temporarySchedules
  const zone = data?.schedule?.timeZone
  const valueInterval = parseInterval(value, zone)
  const doesOverlap = tempSchedules
    .map((t: TemporarySchedule) => parseInterval(t, zone))
    .some((invl: Interval) => invl.overlaps(valueInterval))

  if (!doesOverlap) {
    return []
  }
  return [
    {
      type: 'WARNING',
      message: 'This override overlaps with one or more temporary schedules',
      details: 'Overrides do not take effect during temporary schedules',
    },
  ]
}
