import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import {
  Checkbox,
  Grid,
  Hidden,
  Typography,
  makeStyles,
  isWidthDown,
} from '@material-ui/core'
import { useDispatch, useSelector } from 'react-redux'
import { absURLSelector, urlParamSelector } from '../selectors'
import QueryList from '../lists/QueryList'
import gql from 'graphql-tag'

import CreateAlertFab from './CreateAlertFab'
import AlertsListFilter from './components/AlertsListFilter'
import AlertsListControls from './components/AlertsListControls'
import CheckedAlertsFormControl from './AlertsCheckboxControls'
import statusStyles from '../util/statusStyles'
import {
  setCheckedAlerts as _setCheckedAlerts,
  setAlerts as _setAlerts,
} from '../actions'
import { formatTimeSince } from '../util/timeFormat'
import SnackbarContent from '@material-ui/core/SnackbarContent'
import InfoIcon from '@material-ui/icons/Info'
import Snackbar from '@material-ui/core/Snackbar'
import useWidth from '../util/useWidth'

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
  const width = useWidth()
  const isFullScreen = isWidthDown('md', width)

  // always open unless clicked away from or there are services present
  const [snackbarOpen, setSnackbarOpen] = useState(true)

  // get redux vars
  const absURL = useSelector(absURLSelector)
  const params = useSelector(urlParamSelector)
  const actionComplete = useSelector(state => state.alerts.actionComplete)
  const allServices = params('allServices')
  const checkedAlerts = useSelector(state => state.alerts.checkedAlerts)
  const filter = params('filter', 'active')
  const isFirstLogin = params('isFirstLogin')

  // setup redux actions
  const dispatch = useDispatch()
  const setCheckedAlerts = arr => dispatch(_setCheckedAlerts(arr))
  const setAlerts = arr => dispatch(_setAlerts(arr))

  // todo: need noFavorites?
  const showFavoritesWarning =
    snackbarOpen && !allServices && !props.serviceID && !isFirstLogin

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

  function handleCloseSnackbar(event, reason) {
    if (reason === 'clickaway') {
      setSnackbarOpen(false)
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
      <Snackbar
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'left',
        }}
        open={showFavoritesWarning}
        onClose={handleCloseSnackbar}
      >
        <SnackbarContent
          className={classes.snackbar}
          aria-describedby='client-snackbar'
          message={
            <span id='client-snackbar' className={classes.snackbarMessage}>
              <InfoIcon className={classes.snackbarIcon} />
              It looks like you have no favorited services. Visit your most used
              services to set them as a favorite, or enable the filter to view
              alerts for all services.
            </span>
          }
        />
      </Snackbar>
      <CreateAlertFab
        serviceID={props.serviceID}
        showFavoritesWarning={showFavoritesWarning}
        transition={isFullScreen && (showFavoritesWarning || actionComplete)}
      />
    </React.Fragment>
  )
}

AlertsList.propTypes = {
  serviceID: p.string,
}
