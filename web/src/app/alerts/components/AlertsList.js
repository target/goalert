/* eslint @typescript-eslint/camelcase: 0 */
import React, { Component } from 'react'
import Card from '@material-ui/core/Card'
import InfoIcon from '@material-ui/icons/Info'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import Snackbar from '@material-ui/core/Snackbar'
import SnackbarContent from '@material-ui/core/SnackbarContent'
import withStyles from '@material-ui/core/styles/withStyles'
import isFullScreen from '@material-ui/core/withMobileDialog'
import { debounce } from 'lodash-es'
import { graphql } from 'react-apollo'
import InfiniteScroll from 'react-infinite-scroll-component'
import { styles as globalStyles } from '../../styles/materialStyles'
import { getParameterByName } from '../../util/query_param'
import CreateAlertFab from '../CreateAlertFab'
import AlertsListDataWrapper from './AlertsListDataWrapper'
import { alertsQuery } from '../queries/AlertsListQuery'
import { connect } from 'react-redux'
import CheckedAlertsFormControl from './CheckedAlertsFormControl'
import { GenericError } from '../../error-pages'
import { withRouter } from 'react-router-dom'
import Hidden from '@material-ui/core/Hidden'
import {
  searchSelector,
  alertAllServicesSelector,
  alertFilterSelector,
} from '../../selectors/url'
import AlertsListControls from '../components/AlertsListControls'
import { LegacyGraphQLClient } from '../../apollo'

const LIMIT = 25

const styles = theme => ({
  ...globalStyles(theme),
  snackbar: {
    backgroundColor: theme.palette.primary['500'],
    height: '6.75em',
    width: '20em', // only triggers on desktop, 100% on mobile devices
  },
  snackbarIcon: {
    fontSize: 20,
    opacity: 0.9,
    marginRight: theme.spacing(1),
  },
  snackbarMessage: {
    display: 'flex',
    alignItems: 'center',
  },
})

const filterToOmit = filter => {
  switch (filter) {
    case 'acknowledged':
      return { omit_active: false, omit_triggered: true, omit_closed: true }
    case 'unacknowledged':
      return { omit_active: true, omit_triggered: false, omit_closed: true }
    case 'closed':
      return { omit_active: true, omit_triggered: true, omit_closed: false }
    case 'all':
      return { omit_active: false, omit_triggered: false, omit_closed: false }
  }

  // active (default)
  return { omit_active: false, omit_triggered: false, omit_closed: true }
}

/*
 * Returns true if the first specified array contains all elements
 * from the second one. False otherwise.
 */
function arrayContainsArray(superset, subset) {
  return subset.every(value => superset.indexOf(value) >= 0)
}

const mapStateToProps = state => {
  return {
    actionComplete: state.alerts.actionComplete,
    allServices: alertAllServicesSelector(state),
    isFirstLogin: state.main.isFirstLogin,
    searchParam: searchSelector(state),
    filter: alertFilterSelector(state),
  }
}

@connect(mapStateToProps) // must connect to redux before calling graphql
@graphql(alertsQuery, {
  options: props => {
    return {
      client: LegacyGraphQLClient,
      variables: {
        favorite_services_only: props.serviceID ? false : !props.allServices,
        service_id: props.serviceID || '',
        search: props.searchParam,
        sort_desc: getParameterByName('sortDesc') === 'true',
        limit: LIMIT,
        offset: 0,
        ...filterToOmit(props.filter),
        sort_by: getParameterByName('sortBy') || 'status', // status, id, created_at, summary, or service
        favorites_first: true,
        favorites_only: !props.serviceID,
        services_limit: 1,
        services_search: '',
      },
      notifyOnNetworkStatusChange: true, // updates data.loading bool for refetching, and fetching more
      fetchPolicy: 'cache-and-network',
    }
  },
  props: props => {
    return {
      data: props.data,
      loadMore: queryVariables => {
        return props.data.fetchMore({
          variables: queryVariables,
          updateQuery(previousResult, { fetchMoreResult }) {
            if (!fetchMoreResult) return previousResult

            const p = previousResult.alerts2 // previous
            const n = fetchMoreResult.alerts2 // next

            const pIDs = p.items.map(i => i.id)
            const nIDs = n.items.map(i => i.id)

            // return previous result if the whole next result is duplicate data
            // IDs will always be unique
            if (arrayContainsArray(pIDs, nIDs)) return previousResult

            return Object.assign({}, previousResult, {
              // append the new alerts results to the old one
              alerts2: {
                __typename: n.__typename,
                items: [...p.items, ...n.items],
                total_count: n.total_count,
              },
            })
          },
        })
      },
    }
  },
})
@withStyles(styles)
@isFullScreen()
@withRouter
export default class AlertsList extends Component {
  state = {
    snackbarOpen: true, // always open unless clicked away from or there are services present
  }

  // TODO: Temp fix until apollo cache updated after all relevant mutations affecting this component
  componentDidMount() {
    this.refetch()
  }

  componentWillUnmount() {
    this.refetch.cancel()
  }

  /*
   * Display current data until new data loads in when refetching
   * i.e. only show loading placeholders on first page load
   */
  shouldComponentUpdate(nextProps) {
    return !(
      this.props.data.alerts2 &&
      (nextProps.data.networkStatus === 2 || nextProps.data.networkStatus === 4)
    )
  }

  handleCloseSnackbar = (event, reason) => {
    if (reason === 'clickaway') {
      this.setState({ snackbarOpen: false })
    }
  }

  /*
   * Refetch from scratch after a filter is changed
   */
  refetch = debounce(extraProps => {
    const offset = 0
    this.props.data.refetch(this.getQueryData(offset, extraProps))
  }, 100)

  getQueryData = (offset, extraProps) => {
    return {
      favorite_services_only: this.props.serviceID
        ? false
        : !this.props.allServices,
      service_id: this.props.serviceID || '', // TODO: adding the || "" ensures we get the same cache key as elsewhere, let's find a better way to normalize...
      search: this.props.searchParam,
      sort_desc: getParameterByName('sortDesc') === 'true',
      offset: offset,
      limit: LIMIT,
      ...filterToOmit(this.props.filter),
      sort_by: getParameterByName('sortBy') || 'status',
      favorites_first: true,
      favorites_only: !this.props.serviceID,
      services_limit: 1,
      services_search: '',
      ...extraProps,
    }
  }

  renderLoading = () => {
    const style = {
      color: 'lightgrey',
      background: 'lightgrey',
      height: '0.875em',
    }

    const loadingItems = []
    for (let i = 0; i < 5; i++) {
      loadingItems.push(
        <ListItem key={i} style={{ display: 'block' }}>
          <ListItemText style={{ ...style, width: '50%' }} />
          <ListItemText
            style={{ ...style, width: '35%', margin: '5px 0 5px 0' }}
          />
          <ListItemText style={{ ...style, width: '65%' }} />
        </ListItem>,
      )
    }

    return loadingItems
  }

  renderError = data => {
    return (
      <ListItem style={{ justifyContent: 'center' }}>
        <GenericError error={data.error.message} />
      </ListItem>
    )
  }

  renderNoResults = () => {
    return (
      <ListItem>
        <ListItemText primary='No results' />
      </ListItem>
    )
  }

  render() {
    const {
      actionComplete,
      allServices,
      classes,
      data,
      fullScreen,
      onServicePage,
      isFirstLogin,
      loadMore,
      serviceID,
    } = this.props
    const { snackbarOpen } = this.state

    // status 2: setting variables (occurs when refetching)
    // status 4: refetching
    const net = data.networkStatus
    const isLoading = net === 2 || net === 4

    let offset = 0
    let len = 0
    let hasMore = true

    if (data.alerts2) {
      offset = len = data.alerts2.items.length
      hasMore = offset < data.alerts2.total_count && !data.error
      if (len <= LIMIT) this.props.data.startPolling(3500)
      else this.props.data.stopPolling()
    }

    // Scrollable infinite list should be sorted by id, can be adjusted with status filters/search in appbar
    const noFavorites =
      data.services2 &&
      data.services2.items &&
      data.services2.items.length === 0 &&
      !serviceID

    let content = null
    if (data.error) content = this.renderError(data)
    else if (isLoading) content = this.renderLoading()
    else if (data.alerts2 && !isLoading && !data.alerts2.items.length)
      content = this.renderNoResults()

    const dataToShow = data.alerts2 ? data.alerts2.items : []
    let hasData = false
    if (!content) {
      hasData = true
      content = dataToShow.map(alert => (
        <AlertsListDataWrapper
          key={alert.id}
          alert={alert}
          onServicePage={onServicePage}
        />
      ))
    }

    const showFavoritesWarning =
      snackbarOpen && noFavorites && !allServices && !serviceID && !isFirstLogin

    return (
      <React.Fragment>
        <Snackbar
          anchorOrigin={{
            vertical: 'bottom',
            horizontal: 'left',
          }}
          open={showFavoritesWarning}
          onClose={this.handleCloseSnackbar}
        >
          <SnackbarContent
            className={classes.snackbar}
            aria-describedby='client-snackbar'
            message={
              <span id='client-snackbar' className={classes.snackbarMessage}>
                <InfoIcon className={classes.snackbarIcon} />
                It looks like you have no favorited services. Visit your most
                used services to set them as a favorite, or enable the filter to
                view alerts for all services.
              </span>
            }
          />
        </Snackbar>
        <CreateAlertFab
          serviceID={serviceID}
          showFavoritesWarning={showFavoritesWarning}
          transition={fullScreen && (showFavoritesWarning || actionComplete)}
        />
        <div>
          <CheckedAlertsFormControl data={data} refetch={this.refetch} />
          <Card style={{ width: '100%' }}>
            <Hidden mdDown>
              <AlertsListControls />
            </Hidden>
            {!hasData && (
              <List
                id='alerts-list'
                style={{ padding: 0 }}
                data-cy='alerts-list-no-data'
              >
                {content}
              </List>
            )}
            {hasData && (
              <List
                id='alerts-list'
                style={{ padding: 0 }}
                data-cy='alerts-list'
              >
                <InfiniteScroll
                  scrollableTarget='content'
                  next={() => loadMore(this.getQueryData(offset))}
                  dataLength={len}
                  hasMore={hasMore}
                  loader={null}
                  scrollThreshold={(len - 20) / len}
                  style={{ overflow: 'hidden' }}
                >
                  {content}
                </InfiniteScroll>
              </List>
            )}
          </Card>
        </div>
      </React.Fragment>
    )
  }
}
