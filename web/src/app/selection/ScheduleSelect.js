import gql from 'graphql-tag'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query($input: ScheduleSearchOptions) {
    schedules(input: $input) {
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
    schedule(id: $id) {
      id
      name
      isFavorite
    }
  }
`

export const ScheduleSelect = makeQuerySelect('ScheduleSelect', {
  variables: { favoritesFirst: true },
  defaultQueryVariables: { favoritesFirst: true },
  query,
  valueQuery,
})
