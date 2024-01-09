import { gql } from 'urql'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query ($input: ServiceSearchOptions) {
    services(input: $input) {
      nodes {
        id
        name
        isFavorite
      }
    }
  }
`

const valueQuery = gql`
  query ($id: ID!) {
    service(id: $id) {
      id
      name
      isFavorite
    }
  }
`
export const ServiceSelect = makeQuerySelect('ServiceSelect', {
  variables: { favoritesFirst: true },
  defaultQueryVariables: { favoritesFirst: true },
  query,
  valueQuery,
})
