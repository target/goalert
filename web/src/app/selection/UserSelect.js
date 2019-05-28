import React from 'react'

import gql from 'graphql-tag'
import QuerySelect from './QuerySelect'

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
export class UserSelect extends React.PureComponent {
  render() {
    return <QuerySelect {...this.props} query={query} valueQuery={valueQuery} />
  }
}
