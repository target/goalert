import { gql } from 'urql'
import { IntegrationKey } from '../../schema'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query ($input: IntegrationKeySearchOptions) {
    integrationKeys(input: $input) {
      nodes {
        id
        name
        type
        serviceID
      }
    }
  }
`

export const IntegrationKeySelect = makeQuerySelect('IntegrationKeySelect', {
  query,
  mapDataNode: (key: IntegrationKey) => ({
    label: key.id,
    subText: key.name,
    value: key.id,
  }),
})
