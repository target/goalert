import { gql } from 'urql'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query ($input: UserSearchOptions) {
    users(input: $input) {
      nodes {
        id
        name
        email
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
      email
      isFavorite
    }
  }
`

export const UserSelect = makeQuerySelect('UserSelect', {
  variables: { favoritesFirst: true },
  defaultQueryVariables: { favoritesFirst: true },
  query,
  valueQuery,
  mapDataNode: (u) => ({
    value: u.id,
    label: u.name,
    subText: u.email,
    isFavorite: u.isFavorite,
  }),
})
