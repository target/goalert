import { gql } from 'urql'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query ($input: TimeZoneSearchOptions) {
    timeZones(input: $input) {
      nodes {
        id
      }
    }
  }
`

export const TimeZoneSelect = makeQuerySelect('TimeZoneSelect', {
  query,
  mapDataNode: (n: { id: string }) => ({ label: n.id, value: n.id }),
})
