import { Notice, TemporarySchedule } from '../../schema'
import { checkInterval, parseInterval, SpanISO } from '../util/shifts'
import { useQuery, gql } from 'urql'

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
  const invalidInterval = value.start > value.end
  const [{ data, fetching }] = useQuery({
    query: scheduleQuery,
    variables: {
      id: scheduleID,
    },
    pause: invalidInterval,
  })

  if (fetching || invalidInterval) {
    return []
  }

  const tempSchedules = data?.schedule?.temporarySchedules
  const zone = data?.schedule?.timeZone
  if (!checkInterval(value)) {
    return []
  }
  const valueInterval = parseInterval(value, zone)
  const doesOverlap = tempSchedules.some((t: TemporarySchedule) => {
    if (!checkInterval(t)) return false
    return parseInterval(t, zone).overlaps(valueInterval)
  })

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
