import React from 'react'
import { PropTypes as p } from 'prop-types'
import {
  Checkbox,
  Grid,
  Hidden,
  Typography,
  makeStyles,
} from '@material-ui/core'
import { useDispatch, useSelector } from 'react-redux'
import { absURLSelector, urlParamSelector } from '../selectors'
import QueryList from '../lists/QueryList'
import gql from 'graphql-tag'

import AlertsListFilter from './components/AlertsListFilter'
import AlertsListControls from './components/AlertsListControls'
import AlertsCheckboxControls from './AlertsCheckboxControls'
import AlertsListFloatingItems from './AlertsListFloatingItems'
import statusStyles from '../util/statusStyles'
import {
  setCheckedAlerts as _setCheckedAlerts,
  setAlerts as _setAlerts,
} from '../actions'
import { formatTimeSince } from '../util/timeFormat'

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

const useStyles = makeStyles(theme => ({
  checkbox: {
    marginRight: 'auto',
  },
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

  // get redux vars
  const absURL = useSelector(absURLSelector)
  const params = useSelector(urlParamSelector)
  const allServices = params('allServices')
  const checkedAlerts = useSelector(state => state.alerts.checkedAlerts)
  const filter = params('filter', 'active')

  // setup redux actions
  const dispatch = useDispatch()
  const setCheckedAlerts = arr => dispatch(_setCheckedAlerts(arr))
  const setAlerts = arr => dispatch(_setAlerts(arr))

  // alerts list query variables
  const variables = {
    input: {
      filterByStatus: getStatusFilter(filter),
      first: 25,
      // default to favorites only, unless viewing alerts from a service's page
      favoritesOnly: !props.serviceID && !allServices,
    },
  }

  if (props.serviceID) {
    variables.input.filterByServiceID = [props.serviceID]
  }

  function toggleChecked(id) {
    const czechedAlerts = checkedAlerts.slice()

    if (czechedAlerts.includes(id)) {
      const idx = czechedAlerts.indexOf(id)
      czechedAlerts.splice(idx, 1)
      setCheckedAlerts(czechedAlerts)
    } else {
      czechedAlerts.push(id)
      setCheckedAlerts(czechedAlerts)
    }
  }

  /*
   * Adds border of color depending on each alert's status
   * on left side of each list item
   */
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
        infiniteScroll
        mapDataNode={a => ({
          className: getStatusClassName(a.status),
          icon: (
            <Checkbox
              checked={checkedAlerts.includes(a.id)}
              disabled={a.status === 'StatusClosed'}
              data-cy={'alert-' + a.id}
              tabIndex={-1}
              onClick={e => {
                e.stopPropagation()
                e.preventDefault()
                toggleChecked(a.id)
              }}
            />
          ),
          title: `${a.alertID}: ${a.status
            .toUpperCase()
            .replace('STATUS', '')}`,
          subText: (props.serviceID ? '' : a.service.name + ': ') + a.summary,
          action: (
            <Typography variant='caption'>
              {formatTimeSince(a.createdAt)}
            </Typography>
          ),
          url: absURL(`/alerts/${a.id}`),
        })}
        variables={variables}
        onDataChange={items => {
          // used for checkbox controls
          setAlerts(items)
        }}
        controls={
          <React.Fragment>
            <Grid item className={classes.checkbox}>
              <AlertsCheckboxControls />
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
      <AlertsListFloatingItems serviceID={props.serviceID} />
    </React.Fragment>
  )
}

AlertsList.propTypes = {
  serviceID: p.string,
}
