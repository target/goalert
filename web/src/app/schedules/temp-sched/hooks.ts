import { gql, QueryResult, useQuery } from '@apollo/client'
import { DateTime } from 'luxon'

const schedTZQuery = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      id
      timeZone
    }
  }
`

interface ScheduleTZResult {
  // q is the Apollo query status
  q: QueryResult
  // zone is schedule time zone name if ready; else empty string
  zone: string
  // isLocalZone is true if schedule and system time zone are equal
  isLocalZone: boolean
  // zoneAbbr is schedule time zone abbreviation if ready; else empty string
  zoneAbbr: string
}

export function useScheduleTZ(scheduleID: string): ScheduleTZResult {
  const q = useQuery(schedTZQuery, {
    variables: { id: scheduleID },
  })
  const zone = q.data?.schedule?.timeZone ?? ''
  const isLocalZone = zone === DateTime.local().zoneName
  const zoneAbbr = zone ? DateTime.fromObject({ zone }).toFormat('ZZZZ') : ''

  if (q.error) {
    console.error(
      `useScheduleTZ: issue getting timezone for schedule ${scheduleID}: ${q.error.message}`,
    )
  }

  return { q, zone, isLocalZone, zoneAbbr }
}
