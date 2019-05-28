import React from 'react'

import gql from 'graphql-tag'
import QuerySelect from './QuerySelect'

const query = gql`
  query($input: RotationSearchOptions) {
    rotations(input: $input) {
      nodes {
        id
        name
      }
    }
  }
`

const valueQuery = gql`
  query($id: ID!) {
    rotation(id: $id) {
      id
      name
    }
  }
`
export class RotationSelect extends React.PureComponent {
  render() {
    return <QuerySelect {...this.props} query={query} valueQuery={valueQuery} />
  }
}
