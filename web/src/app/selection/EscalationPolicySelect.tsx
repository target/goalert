import { gql } from 'urql'
import { makeQuerySelect } from './QuerySelect'

const query = gql`
  query ($input: EscalationPolicySearchOptions) {
    escalationPolicies(input: $input) {
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
    escalationPolicy(id: $id) {
      id
      name
      isFavorite
    }
  }
`

export const EscalationPolicySelect = makeQuerySelect(
  'EscalationPolicySelect',
  {
    variables: { favoritesFirst: true },
    defaultQueryVariables: { favoritesFirst: true },
    query,
    valueQuery,
  },
)
