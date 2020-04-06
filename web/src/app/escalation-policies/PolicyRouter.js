import React from 'react'
import gql from 'graphql-tag'
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
      }
      pageInfo {
        hasNextPage
        endCursor
      }
    }
  }
`

export default function PolicyRouter() {
  const renderList = () => (
    <SimpleListPage
      query={query}
      mapDataNode={(n) => ({
        title: n.name,
        subText: n.description,
        url: n.id,
      })}
      createForm={<PolicyCreateDialog />}
      createLabel='Escalation Policy'
    />
  )

  const renderDetails = ({ match }) => (
    <PolicyDetails escalationPolicyID={match.params.escalationPolicyID} />
  )

  const renderServices = ({ match }) => (
    <PolicyServicesQuery escalationPolicyID={match.params.escalationPolicyID} />
  )

  return (
    <Switch>
      <Route exact path='/escalation-policies' render={renderList} />
      <Route
        exact
        path='/escalation-policies/:escalationPolicyID'
        render={renderDetails}
      />
      <Route
        exact
        path='/escalation-policies/:escalationPolicyID/services'
        render={renderServices}
      />
      <Route component={PageNotFound} />
    </Switch>
  )
}
