import gql from 'graphql-tag'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query($input: UserSearchOptions) {
    users(input: $input) {
      nodes {
        id
        name
      }
    }
  }
`

const valueQuery = gql`
  query($id: ID!) {
    user(id: $id) {
      id
      name
    }
  }
`

export const UserSelect = makeQuerySelect('UserSelect', {
  query,
  valueQuery,
})
