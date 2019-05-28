import React from 'react'

import gql from 'graphql-tag'
import QuerySelect from './QuerySelect'

const query = gql`
  query($input: ServiceSearchOptions) {
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
  query($id: ID!) {
    service(id: $id) {
      id
      name
      isFavorite
    }
  }
`
export class ServiceSelect extends React.PureComponent {
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
