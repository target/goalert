import React from 'react'
import { gql } from '@apollo/client'
import { Switch, Route } from 'react-router-dom'

import SimpleListPage from '../lists/SimpleListPage'
import ServiceDetails from './ServiceDetails'
import ServiceLabelList from './ServiceLabelList'
import IntegrationKeyList from './IntegrationKeyList'
import { PageNotFound } from '../error-pages/Errors'
import ServiceAlerts from './ServiceAlerts'
import ServiceCreateDialog from './ServiceCreateDialog'
import HeartbeatMonitorList from './HeartbeatMonitorList'
import { useURLParam } from '../actions'
import ServiceLabelFilterContainer from './ServiceLabelFilterContainer'
import getServiceLabel from '../util/getServiceLabel'
import AlertMetrics from './AlertMetrics/AlertMetrics'

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

export default function ServiceRouter() {
  const [searchParam, setSearchParam] = useURLParam('search', '')
  const { labelKey, labelValue } = getServiceLabel(searchParam)

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
        createForm={<ServiceCreateDialog />}
        createLabel='Service'
        searchAdornment={
          <ServiceLabelFilterContainer
            value={{ labelKey, labelValue }}
            onChange={({ labelKey, labelValue }) =>
              setSearchParam(labelKey ? labelKey + '=' + labelValue : '')
            }
            onReset={() => setSearchParam()}
          />
        }
      />
    )
  }

  function renderDetails({ match }) {
    return <ServiceDetails serviceID={match.params.serviceID} />
  }

  function renderAlerts({ match }) {
    return <ServiceAlerts serviceID={match.params.serviceID} />
  }

  function renderKeys({ match }) {
    return <IntegrationKeyList serviceID={match.params.serviceID} />
  }

  function renderHeartbeatMonitors({ match }) {
    return <HeartbeatMonitorList serviceID={match.params.serviceID} />
  }

  function renderLabels({ match }) {
    return <ServiceLabelList serviceID={match.params.serviceID} />
  }

  function renderAlertMetrics({ match }) {
    return <AlertMetrics serviceID={match.params.serviceID} />
  }

  return (
    <Switch>
      <Route exact path='/services' render={renderList} />
      <Route exact path='/services/:serviceID/alerts' render={renderAlerts} />
      <Route exact path='/services/:serviceID' render={renderDetails} />
      <Route
        exact
        path='/services/:serviceID/integration-keys'
        render={renderKeys}
      />
      <Route
        exact
        path='/services/:serviceID/heartbeat-monitors'
        render={renderHeartbeatMonitors}
      />
      <Route exact path='/services/:serviceID/labels' render={renderLabels} />
      <Route
        exact
        path='/services/:serviceID/alert-metrics'
        render={renderAlertMetrics}
      />
      <Route component={PageNotFound} />
    </Switch>
  )
}
