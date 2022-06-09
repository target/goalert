import React from 'react'
import { gql } from 'urql'
import SimpleListPage from '../lists/SimpleListPage'
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

export default function PolicyList(): JSX.Element {
  return (
    <SimpleListPage
      query={query}
      variables={{ input: { favoritesFirst: true } }}
      mapDataNode={(n) => ({
        title: n.name,
        subText: n.description,
        url: n.id,
        isFavorite: n.isFavorite,
      })}
      createForm={<PolicyCreateDialog />}
      createLabel='Escalation Policy'
    />
  )
}
