import { gql, useQuery } from 'urql'
import { DateTime } from 'luxon'

const schedTZQuery = gql`
  query SchedZone($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
    }
  }
`

interface ScheduleTZResult {
  q: { loading: false } // for compatability, until loading logic is removed (in favor of suspense)
  // zone is schedule time zone name if ready; else empty string
  zone: string
  // isLocalZone is true if schedule and system time zone are equal
  isLocalZone: boolean
  // zoneAbbr is schedule time zone abbreviation if ready; else empty string
  zoneAbbr: string
}

export function useScheduleTZ(scheduleID: string): ScheduleTZResult {
  const [q] = useQuery({
    query: schedTZQuery,
    variables: { id: scheduleID },
  })
  const zone = q.data?.schedule?.timeZone ?? 'local'
  const isLocalZone = zone === DateTime.local().zoneName
  const zoneAbbr = zone ? DateTime.local({ zone }).toFormat('ZZZZ') : ''

  if (q.error) {
    console.error(
      `useScheduleTZ: issue getting timezone for schedule ${scheduleID}: ${q.error.message}`,
    )
  }

  return { q: { loading: false }, zone, isLocalZone, zoneAbbr }
}
