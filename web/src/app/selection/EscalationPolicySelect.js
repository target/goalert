import gql from 'graphql-tag'
import { makeQuerySelect } from './QuerySelect'

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

export const EscalationPolicySelect = makeQuerySelect(
  'EscalationPolicySelect',
  { query, valueQuery },
)
