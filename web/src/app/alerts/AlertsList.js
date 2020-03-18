import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import gql from 'graphql-tag'
import {
  Hidden,
  Snackbar,
  SnackbarContent,
  Typography,
  makeStyles,
  isWidthDown,
} from '@material-ui/core'
import {
  ArrowUpward as EscalateIcon,
  Check as AcknowledgeIcon,
  Close as CloseIcon,
  Info as InfoIcon,
} from '@material-ui/icons'
import { useSelector } from 'react-redux'

import AlertsListFilter from './components/AlertsListFilter'
import AlertsListControls from './components/AlertsListControls'
import CreateAlertFab from './CreateAlertFab'
import UpdateAlertsSnackbar from './components/UpdateAlertsSnackbar'

import { formatTimeSince } from '../util/timeFormat'
import { urlParamSelector } from '../selectors'
import { useMutation, useQuery } from '@apollo/react-hooks'
import QueryList from '../lists/QueryList'
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
        service {
          id
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
    updateAlerts(inputa: $input) {
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
      return [] // empty array returns all statuses
    // active is the default tab
    default:
      return ['StatusAcknowledged', 'StatusUnacknowledged']
  }
}

export default function AlertsList(props) {
  const classes = useStyles()
  const width = useWidth()
  const isFullScreen = isWidthDown('md', width)

  // state initialization
  const [errorMessage, setErrorMessage] = useState('')
  const [updateMessage, setUpdateMessage] = useState('')

  // used if user dismisses snackbar before the auto-close timer finishes
  const [actionCompleteDismissed, setActionCompleteDismissed] = useState(
    true,
  )

  // defaults to open unless favorited services are present or warning is dismissed
  const [favoritesWarningDismissed, setFavoritesWarningDismissed] = useState(
    false,
  )

  // get redux url vars
  const params = useSelector(urlParamSelector)
  const allServices = params('allServices')
  const filter = params('filter', 'active')
  const isFirstLogin = params('isFirstLogin')

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
        input: {
          favoritesOnly: true,
          first: 1,
        },
      },
    },
  )

  // checks to show no favorites warning
  const noFavorites =
    !favoritesQueryStatus.data?.services?.nodes?.length &&
    !favoritesQueryStatus.loading
  const showNoFavoritesWarning =
    !favoritesWarningDismissed && // has not been dismissed
    !allServices &&               // all services aren't being queries
    !props.serviceID &&           // not viewing alerts from services page
    !isFirstLogin &&              // don't show two pop-ups at the same time
    noFavorites                   // and lastly, user has no favorited services

  /*
   * Closes the no favorites warning snackbar only if clicking
   * away to lose focus
   */
  function handleCloseNoFavoritesWarning(event, reason) {
    if (reason === 'clickaway') {
      setFavoritesWarningDismissed(false)
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

  // checkbox action mutations
  const [ackAlerts, a] = useMutation(updateMutation)
  const [closeAlerts, c] = useMutation(updateMutation)
  const [escalateAlerts, e] = useMutation(escalateMutation)

  const actionComplete = (a.called && !a.loading) || (c.called && !c.loading) || (e.called && !e.loading)

  /*
   * Called on the successful completion of a checkbox action.
   * Sets showing the update snackbar and sets the message to be
   * displayed
   */
  function onCompleted(numUpdated, checkedItems) {
    setActionCompleteDismissed(false) // for create fab transition
    setUpdateMessage(`${numUpdated} of ${checkedItems.length} alerts updated`)
  }

  /*
   * Adds border of color depending on each alert's status
   * on left side of each list item
   */
  function getListItemStatus(s) {
    switch (s) {
      case 'StatusAcknowledged':
        return 'warn'
      case 'StatusUnacknowledged':
        return 'err'
      case 'StatusClosed':
        return 'ok'
    }
  }

  /*
   * Passes the proper actions to ListControls depending
   * on which tab is currently filtering the alerts list
   */
  function getActions() {
    const actions = []

    if (filter !== 'closed' && filter !== 'acknowledged') {
      actions.push({
        icon: <AcknowledgeIcon />,
        label: 'Acknowledge',
        onClick: checkedItems => {
          ackAlerts({
            variables: {
              input: {
                alertIDs: checkedItems,
                newStatus: 'StatusAcknowledged',
              },
            },
          })
            .then(res =>
              onCompleted(res?.data?.updateAlerts?.length ?? 0, checkedItems),
            )
            .catch(err => {
              setActionCompleteDismissed(false)
              setErrorMessage(err.message)
            })
        },
      })
    }

    if (filter !== 'closed') {
      actions.push(
        {
          icon: <CloseIcon />,
          label: 'Close',
          onClick: checkedItems => {
            closeAlerts({
              variables: {
                input: {
                  alertIDs: checkedItems,
                  newStatus: 'StatusClosed',
                },
              },
            })
              .then(res =>
                onCompleted(res?.data?.updateAlerts?.length ?? 0, checkedItems),
              )
              .catch(err => {
                setActionCompleteDismissed(false)
                setErrorMessage(err.message)
              })
          },
        },
        {
          icon: <EscalateIcon />,
          label: 'Escalate',
          onClick: checkedItems => {
            escalateAlerts({
              variables: {
                input: checkedItems,
              },
            })
              .then(res =>
                onCompleted(
                  res?.data?.escalateAlerts?.length ?? 0,
                  checkedItems,
                ),
              )
              .catch(err => {
                setActionCompleteDismissed(false)
                setErrorMessage(err.message)
              })
          },
        },
      )
    }

    return actions
  }

  // render
  return (
    <React.Fragment>
      <QueryList
        query={alertsListQuery}
        infiniteScroll
        mapDataNode={a => ({
          id: a.id,
          status: getListItemStatus(a.status),
          title: `${a.alertID}: ${a.status
            .toUpperCase()
            .replace('STATUS', '')}`,
          subText: (props.serviceID ? '' : a.service.name + ': ') + a.summary,
          action: (
            <Typography variant='caption'>
              {formatTimeSince(a.createdAt)}
            </Typography>
          ),
          url: `/alerts/${a.id}`,
          selectable: a.status !== 'StatusClosed',
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
        transition={
          isFullScreen && (showNoFavoritesWarning || actionCompleteDismissed)
        }
      />

      {/* Update message after using checkbox actions */}
      <UpdateAlertsSnackbar
        errorMessage={errorMessage}
        onClose={() => setActionCompleteDismissed(true)}
        onExited={() => {
          setErrorMessage('')
          setUpdateMessage('')
        }}
        open={actionComplete && !actionCompleteDismissed}
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
