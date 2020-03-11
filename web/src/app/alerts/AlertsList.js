import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import {Hidden, Typography, makeStyles, isWidthDown} from '@material-ui/core'
import { useSelector } from 'react-redux'
import { absURLSelector, urlParamSelector } from '../selectors'
import QueryList from '../lists/QueryList'
import gql from 'graphql-tag'

import AlertsListFilter from './components/AlertsListFilter'
import AlertsListControls from './components/AlertsListControls'
import statusStyles from '../util/statusStyles'
import { formatTimeSince } from '../util/timeFormat'
import {useMutation, useQuery} from '@apollo/react-hooks'
import UpdateAlertsSnackbar from './components/UpdateAlertsSnackbar'
import {
  ArrowUpward as EscalateIcon,
  Check as AcknowledgeIcon,
  Close as CloseIcon,
} from '@material-ui/icons'
import Snackbar from "@material-ui/core/Snackbar"
import SnackbarContent from "@material-ui/core/SnackbarContent"
import InfoIcon from "@material-ui/icons/Info"
import CreateAlertFab from "./CreateAlertFab"
import useWidth from "../util/useWidth"

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

const updateMutation = gql`
  mutation UpdateAlertsMutation($input: UpdateAlertsInput!) {
    updateAlerts(input: $input) {
      alertID
      id
    }
  }
`

const escalateMutation = gql`
  mutation EscalateAlertsMutation($input: [Int!]) {
    escalateAlerts(input: $input) {
      alertID
      id
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

  const [errorMessage, setErrorMessage] = useState('')
  const [updateMessage, setUpdateMessage] = useState('')
  const [updateComplete, setUpdateComplete] = useState(false)

  // get redux vars
  const absURL = useSelector(absURLSelector)
  const params = useSelector(urlParamSelector)
  const allServices = params('allServices')
  const filter = params('filter', 'active')
  const isFirstLogin = params('isFirstLogin')

  // always open unless clicked away from or there are services present
  const [_showNoFavoritesWarning, setShowNoFavoritesWarning] = useState(true)

  // query to see if the current user has any favorited services
  // if allServices is not true
  const favoritesQueryStatus = useQuery(
    gql`
      query($input: ServiceSearchOptions) {
        services(input: $input) {
          nodes {
            id
          }
        }
      }
    `,
    {
      variables: {
        favoritesOnly: true,
        first: 1,
      },
    },
  )

  const noFavorites = !favoritesQueryStatus.data?.nodes?.length
  const showNoFavoritesWarning =
    _showNoFavoritesWarning &&
    !allServices &&
    !props.serviceID &&
    !isFirstLogin &&
    noFavorites

  function handleCloseNoFavoritesWarning(event, reason) {
    if (reason === 'clickaway') {
      setShowNoFavoritesWarning(false)
    }
  }

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

  const [ackAlerts] = useMutation(updateMutation)
  const [closeAlerts] = useMutation(updateMutation)
  const [escalateAlerts] = useMutation(escalateMutation)

  function onCompleted(numUpdated, checkedItems) {
    console.log('on complete')
    setUpdateComplete(true) // for create fab transition
    setUpdateMessage(`${numUpdated} of ${checkedItems.length} alerts updated`)
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

  /*
   * Passes the proper actions to ListControls depending
   * on which tab is currently filtering the alerts list
   */
  function getActions() {
    const actions = []
    // todo: don't show actions with nothing checked
    // if (!checkedItems.length) return actions

    if (filter !== 'closed' && filter !== 'acknowledged') {
      actions.push({
        icon: <AcknowledgeIcon />,
        label: 'Acknowledge',
        onClick: checkedItems => ackAlerts({
          variables: {
            input: {
              alertIDs: checkedItems,
              newStatus: 'StatusAcknowledged',
            },
          },
          onError: err => {
            setUpdateComplete(true)
            setErrorMessage(err.message)
          },
          onCompleted: data => onCompleted(data?.updateAlerts?.length ?? 0, checkedItems)
        }),
        // ariaLabel: 'Acknowledge Selected Alerts',
        // dataCy: 'acknowledge',
      })
    }

    if (filter !== 'closed') {
      actions.push(
        {
          icon: <CloseIcon />,
          label: 'Close',
          onClick: checkedItems => closeAlerts({
            variables: {
              input: {
                alertIDs: checkedItems,
                newStatus: 'StatusClosed',
              },
            },
            onError: err => {
              setUpdateComplete(true)
              setErrorMessage(err.message)
            },
            onCompleted: data => onCompleted(data?.updateAlerts?.length ?? 0, checkedItems),
          }),
          // ariaLabel: 'Close Selected Alerts',
          // dataCy: 'close',
        },
        {
          icon: <EscalateIcon />,
          label: 'Escalate',
          onClick: checkedItems => escalateAlerts({
            variables: {
              input: checkedItems,
            },
            onError: err => {
              setUpdateComplete(true)
              setErrorMessage(err.message)
            },
            onCompleted: data => onCompleted(data?.escalateAlerts?.length ?? 0, checkedItems),
          }),
          // ariaLabel: 'Escalate Selected Alerts',
          // dataCy: 'escalate',
        },
      )
    }

    return actions
  }

  return (
    <React.Fragment>
      <QueryList
        query={alertsListQuery}
        infiniteScroll
        mapDataNode={a => ({
          id: a.id,
          className: getStatusClassName(a.status),
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
          selectable: a.status !== 'StatusClosed'
        })}
        variables={variables}
        filter={<AlertsListFilter />}
        cardHeader={
          <Hidden mdDown>
            <AlertsListControls />
          </Hidden>
        }
        checkboxActions={getActions()}
      />

      <CreateAlertFab
        serviceID={props.serviceID}
        showFavoritesWarning={showNoFavoritesWarning}
        transition={isFullScreen && (showNoFavoritesWarning || updateComplete)}
      />

      {/* Update message after using checkbox actions */}
      <UpdateAlertsSnackbar
        errorMessage={errorMessage}
        onClose={() => setUpdateComplete(false)}
        onExited={() => {
          setErrorMessage('')
          setUpdateMessage('')
        }}
        open={updateComplete}
        updateMessage={updateMessage}
      />

      {/* No favorites warning when viewing alerts */}
      <Snackbar
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'left',
        }}
        open={showNoFavoritesWarning}
        onClose={handleCloseNoFavoritesWarning}
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
    </React.Fragment>
  )
}

AlertsList.propTypes = {
  serviceID: p.string,
}
