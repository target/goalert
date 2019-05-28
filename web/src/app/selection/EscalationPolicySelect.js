import React from 'react'

import gql from 'graphql-tag'
import QuerySelect from './QuerySelect'

const query = gql`
  query($input: EscalationPolicySearchOptions) {
    escalationPolicies(input: $input) {
      nodes {
        id
        name
      }
    }
  }
`

const valueQuery = gql`
  query($id: ID!) {
    escalationPolicy(id: $id) {
      id
      name
    }
  }
`
export class EscalationPolicySelect extends React.PureComponent {
  render() {
    return <QuerySelect {...this.props} query={query} valueQuery={valueQuery} />
  }
}
