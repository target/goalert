import React from 'react'
import { gql } from 'urql'
import QueryList from '../lists/QueryList'
import PolicyCreateDialog from './PolicyCreateDialog'

const query = gql`
  query epsQuery($input: EscalationPolicySearchOptions) {
    data: escalationPolicies(input: $input) {
      nodes {
        id
        name
        description
        isFavorite
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`

export default function PolicyList(): React.ReactNode {
  return (
    <QueryList
      query={query}
      variables={{ input: { favoritesFirst: true } }}
      mapDataNode={(n) => ({
        title: n.name,
        subText: n.description,
        url: n.id,
        isFavorite: n.isFavorite,
      })}
      renderCreateDialog={(onClose) => <PolicyCreateDialog onClose={onClose} />}
      createLabel='Escalation Policy'
    />
  )
}
