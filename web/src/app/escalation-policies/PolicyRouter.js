import React from 'react'
import { gql } from '@apollo/client'
import { Switch, Route } from 'react-router-dom'
import PolicyCreateDialog from './PolicyCreateDialog'
import PolicyDetails from './PolicyDetails'
import PolicyServicesQuery from './PolicyServicesQuery'
import { PageNotFound } from '../error-pages/Errors'
import SimpleListPage from '../lists/SimpleListPage'

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

export default function PolicyRouter() {
  function renderList() {
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

  return (
    <Switch>
      <Route exact path='/escalation-policies' render={renderList} />
      <Route
        exact
        path='/escalation-policies/:escalationPolicyID'
        component={PolicyDetails}
      />
      <Route
        exact
        path='/escalation-policies/:escalationPolicyID/services'
        component={PolicyServicesQuery}
      />
      <Route component={PageNotFound} />
    </Switch>
  )
}
