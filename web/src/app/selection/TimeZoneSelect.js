import { gql } from '@apollo/client'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query($input: TimeZoneSearchOptions) {
    timeZones(input: $input) {
      nodes {
        id
      }
    }
  }
`

export const TimeZoneSelect = makeQuerySelect('TimeZoneSelect', {
  query,
  mapDataNode: ({ id }) => ({ label: id, value: id }),
})
