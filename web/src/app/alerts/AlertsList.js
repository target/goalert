import React from 'react'
import { PropTypes as p } from 'prop-types'
import { Grid, Hidden, makeStyles } from '@material-ui/core'
import { useSelector } from 'react-redux'
import { urlParamSelector } from '../selectors'
import QueryList from '../lists/QueryList'
import gql from 'graphql-tag'
import CreateAlertFab from './CreateAlertFab'
import AlertsListFilter from './components/AlertsListFilter'
import AlertsListControls from './components/AlertsListControls'
import CheckedAlertsFormControl from './AlertsCheckboxControls'
import { useApolloClient } from '@apollo/react-hooks'
import statusStyles from '../util/statusStyles'

export const alertsListQuery = gql`
  query alertsList($input: AlertSearchOptions) {
    alerts(input: $input) {
      nodes {
        id
        alertID
        status
        summary
        details
        createdAt
        serviceID
        service {
          name
        }
      }

      pageInfo {
        endCursor
        hasNextPage
      }
    }
  }
`

// todo: which global styles do we need
const useStyles = makeStyles(theme => ({
  // snackbar: {
  //   backgroundColor: theme.palette.primary['500'],
  //   height: '6.75em',
  //   width: '20em', // only triggers on desktop, 100% on mobile devices
  // },
  // snackbarIcon: {
  //   fontSize: 20,
  //   opacity: 0.9,
  //   marginRight: theme.spacing(1),
  // },
  // snackbarMessage: {
  //   display: 'flex',
  //   alignItems: 'center',
  // },
  ...statusStyles,
}))

function getStatusFilter(s) {
  switch (s) {
    case 'acknowledged':
      return ['StatusAcknowledged']
    case 'unacknowledged':
      return ['StatusUnacknowledged']
    case 'closed':
      return ['StatusClosed']
    case 'all':
      return ['StatusAcknowledged', 'StatusUnacknowledged', 'StatusClosed']
    // active is the default tab
    default:
      return ['StatusAcknowledged', 'StatusUnacknowledged']
  }
}

export default function AlertsList(props) {
  const classes = useStyles()

  const params = useSelector(urlParamSelector)
  // const actionComplete = useSelector(state => state.alerts.actionComplete)
  const allServices = params('allServices')
  const filter = params('filter', 'active')
  // const isFirstLogin = params('isFirstLogin')

  const variables = {
    input: {
      filterByStatus: getStatusFilter(filter),
      first: 25,
      // default to favorites only, unless viewing alerts from a service's page
      favoritesOnly: !props.serviceID && !allServices,
    },
  }

  function getStatusClassName(s) {
    switch (s) {
      case 'StatusAcknowledged':
        return classes.statusWarning
      case 'StatusUnacknowledged':
        return classes.statusError
      default:
        return classes.noStatus
    }
  }

  return (
    <React.Fragment>
      <QueryList
        query={alertsListQuery}
        mapDataNode={a => ({
          className: getStatusClassName(a.status),
          title: `${a.alertID}: ${a.status
            .toUpperCase()
            .replace('STATUS', '')}`,
          subText: (props.serviceID ? '' : a.service.name + ': ') + a.summary,
        })}
        variables={variables}
        controls={
          <React.Fragment>
            <Grid item>
              <CheckedAlertsFormControl variables={variables} />
            </Grid>
            <Grid item>
              <AlertsListFilter />
            </Grid>
          </React.Fragment>
        }
        cardHeader={
          <Hidden mdDown>
            <AlertsListControls />
          </Hidden>
        }
      />
      <CreateAlertFab
        serviceID={props.serviceID}
        // showFavoritesWarning={showFavoritesWarning}
        // transition={fullScreen && (showFavoritesWarning || actionComplete)}
      />
    </React.Fragment>
  )
}

AlertsList.propTypes = {
  serviceID: p.string,
}
