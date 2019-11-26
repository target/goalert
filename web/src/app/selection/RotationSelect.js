import gql from 'graphql-tag'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query($input: RotationSearchOptions) {
    rotations(input: $input) {
      nodes {
        id
        name
        isFavorite
      }
    }
  }
`

const valueQuery = gql`
  query($id: ID!) {
    rotation(id: $id) {
      id
      name
      isFavorite
    }
  }
`
export const RotationSelect = makeQuerySelect('RotationSelect', {
  variables: { favoritesFirst: true },
  defaultQueryVariables: { favoritesOnly: true },
  query,
  valueQuery,
})
