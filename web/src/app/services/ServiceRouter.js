import React from 'react'
import { useDispatch, useSelector } from 'react-redux'
import gql from 'graphql-tag'
import { Switch, Route } from 'react-router-dom'
import Grid from '@material-ui/core/Grid'

import SimpleListPage from '../lists/SimpleListPage'
import ServiceDetails from './ServiceDetails'
import ServiceLabelList from './ServiceLabelList'
import IntegrationKeyList from './IntegrationKeyList'
import { LabelKeySelect } from '../selection/LabelKeySelect'
import { LabelValueSelect } from '../selection/LabelValueSelect'
import { PageNotFound } from '../error-pages/Errors'
import ServiceAlerts from './components/ServiceAlerts'
import ServiceCreateDialog from './ServiceCreateDialog'
import HeartbeatMonitorList from './HeartbeatMonitorList'
import FilterContainer from '../util/FilterContainer'
import { searchSelector } from '../selectors'
import { setURLParam } from '../actions'

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
  const searchParam = useSelector(searchSelector) // current total search string on page load
  const dispatch = useDispatch()
  const setSearchParam = value => dispatch(setURLParam('search', value))

  // grab key and value from the search param, if at all
  let key = null
  let value = null
  if (searchParam.includes('=')) {
    const searchSplit = searchParam.split(/(!=|=)/)
    key = searchSplit[0]
    value = searchSplit[2] // [1] being != or =
  }

  function renderList() {
    return (
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
        createLabel='Service'
        searchFilters={renderSearchFilters()}
      />
    )
  }

  function renderSearchFilters() {
    return (
      <FilterContainer
        iconButtonProps={{
          color: 'default',
          'aria-label': 'Show Label Filters',
        }}
        onReset={() => setSearchParam()}
      >
        <Grid item xs={12}>
          <LabelKeySelect
            label='Select Label Key'
            value={key}
            onChange={onKeyChange}
          />
        </Grid>
        <Grid item xs={12}>
          <LabelValueSelect
            label='Select Label Value'
            value={value}
            onChange={onValueChange}
            disabled={!key}
          />
        </Grid>
      </FilterContainer>
    )
  }

  function onKeyChange(newKey) {
    if (newKey) {
      setSearchParam(newKey + '=')
    } else {
      setSearchParam() // clear search if clearing key
    }
  }

  function onValueChange(newValue) {
    // should be disabled if empty, but just in case :)
    if (!key) return

    if (newValue) {
      setSearchParam(key + '=' + newValue)
    } else {
      setSearchParam(key + '=')
    }
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

      <Route component={PageNotFound} />
    </Switch>
  )
}
