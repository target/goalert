import React from 'react'
import gql from 'graphql-tag'
import { Switch, Route } from 'react-router-dom'

import SimpleListPage from '../lists/SimpleListPage'
import ServiceDetails from './ServiceDetails'
import ServiceLabelList from './ServiceLabelList'
import IntegrationKeyList from './IntegrationKeyList'

import { PageNotFound } from '../error-pages/Errors'

import ServiceAlerts from './components/ServiceAlerts'

import ServiceCreateDialog from './ServiceCreateDialog'
import HeartbeatMonitorList from './HeartbeatMonitorList'

const query = gql`
  query servicesQuery($input: ServiceSearchOptions) {
    data: services(input: $input) {
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

export default class ServiceRouter extends React.PureComponent {
  renderList = () => (
    <SimpleListPage
      query={query}
      variables={{ input: { favoritesFirst: true } }}
      mapDataNode={n => ({
        title: n.name,
        subText: n.description,
        url: n.id,
        isFavorite: n.isFavorite,
      })}
      createForm={<ServiceCreateDialog />}
    />
  )

  renderDetails = ({ match }) => (
    <ServiceDetails serviceID={match.params.serviceID} />
  )
  renderAlerts = ({ match }) => (
    <ServiceAlerts serviceID={match.params.serviceID} />
  )
  renderKeys = ({ match }) => (
    <IntegrationKeyList serviceID={match.params.serviceID} />
  )
  renderHeartbeatMonitors = ({ match }) => (
    <HeartbeatMonitorList serviceID={match.params.serviceID} />
  )
  renderLabels = ({ match }) => (
    <ServiceLabelList serviceID={match.params.serviceID} />
  )

  render() {
    return (
      <Switch>
        <Route exact path='/services' render={this.renderList} />
        <Route
          exact
          path='/services/:serviceID/alerts'
          render={this.renderAlerts}
        />
        <Route exact path='/services/:serviceID' render={this.renderDetails} />
        <Route
          exact
          path='/services/:serviceID/integration-keys'
          render={this.renderKeys}
        />
        <Route
          exact
          path='/services/:serviceID/heartbeat-monitors'
          render={this.renderHeartbeatMonitors}
        />
        <Route
          exact
          path='/services/:serviceID/labels'
          render={this.renderLabels}
        />

        <Route component={PageNotFound} />
      </Switch>
    )
  }
}
