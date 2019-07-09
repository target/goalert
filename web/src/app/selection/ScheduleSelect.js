import React from 'react'

import gql from 'graphql-tag'
import QuerySelect from './QuerySelect'

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
export class ScheduleSelect extends React.PureComponent {
  render() {
    return (
      <QuerySelect
        {...this.props}
        variables={{ input: { favoritesFirst: true } }}
        defaultQueryVariables={{ input: { favoritesOnly: true } }}
        query={query}
        valueQuery={valueQuery}
      />
    )
  }
}
