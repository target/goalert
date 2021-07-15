import { gql } from '@apollo/client'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query ($input: UserSearchOptions) {
    users(input: $input) {
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
    user(id: $id) {
      id
      name
      isFavorite
    }
  }
`

export const UserSelect = makeQuerySelect('UserSelect', {
  variables: { favoritesFirst: true },
  defaultQueryVariables: { favoritesFirst: true },
  query,
  valueQuery,
})
