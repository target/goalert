import React, { PureComponent } from 'react'
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

export default class PolicyRouter extends PureComponent {
  renderList = () => (
    <SimpleListPage
      query={query}
      mapDataNode={n => ({
        title: n.name,
        subText: n.description,
        url: n.id,
      })}
      createForm={<PolicyCreateDialog />}
    />
  )

  renderDetails = ({ match }) => (
    <PolicyDetails escalationPolicyID={match.params.escalationPolicyID} />
  )

  renderServices = ({ match }) => (
    <PolicyServicesQuery escalationPolicyID={match.params.escalationPolicyID} />
  )

  render() {
    return (
      <Switch>
        <Route exact path='/escalation-policies' component={this.renderList} />
        <Route
          exact
          path='/escalation-policies/:escalationPolicyID'
          component={this.renderDetails}
        />
        <Route
          exact
          path='/escalation-policies/:escalationPolicyID/services'
          component={this.renderServices}
        />
        <Route component={PageNotFound} />
      </Switch>
    )
  }
}
