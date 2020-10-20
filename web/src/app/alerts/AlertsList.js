import { useMutation, useQuery, gql } from '@apollo/client'
import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { Hidden, ListItemText, isWidthDown } from '@material-ui/core'
import {
  ArrowUpward as EscalateIcon,
  Check as AcknowledgeIcon,
  Close as CloseIcon,
} from '@material-ui/icons'
import { useSelector } from 'react-redux'

import AlertsListFilter from './components/AlertsListFilter'
import AlertsListControls from './components/AlertsListControls'
import CreateAlertFab from './CreateAlertFab'
import UpdateAlertsSnackbar from './components/UpdateAlertsSnackbar'

import { formatTimeSince } from '../util/timeFormat'
import { urlParamSelector } from '../selectors'
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
    updateAlerts(input: $input) {
      status
      id
    }
  }
`

const escalateMutation = gql`
  mutation EscalateAlertsMutation($input: [Int!]) {
    escalateAlerts(input: $input) {
      status
      id
    }
  }
`

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
  const width = useWidth()
  const isMobileScreenSize = isWidthDown('md', width)

  const [checkedCount, setCheckedCount] = useState(0)

  // used if user dismisses snackbar before the auto-close timer finishes
  const [actionCompleteDismissed, setActionCompleteDismissed] = useState(true)

  // get redux url vars
  const params = useSelector(urlParamSelector)
  const allServices = params('allServices')
  const filter = params('filter', 'active')

  // query for current service name if props.serviceID is provided
  const serviceNameQuery = useQuery(
    gql`
      query($id: ID!) {
        service(id: $id) {
          id
          name
        }
      }
    `,
    {
      variables: { id: props.serviceID || '' },
      skip: !props.serviceID,
    },
  )

  // alerts list query variables
  const variables = {
    input: {
      filterByStatus: getStatusFilter(filter),
      first: 25,
      // default to favorites only, unless viewing alerts from a service's page
      favoritesOnly: !props.serviceID && !allServices,
      includeNotified: !props.serviceID, // keep service list alerts specific to that service
    },
  }

  if (props.serviceID) {
    variables.input.filterByServiceID = [props.serviceID]
  }

  const [mutate, status] = useMutation(updateMutation)

  const makeUpdateAlerts = (newStatus) => (alertIDs) => {
    setCheckedCount(alertIDs.length)
    setActionCompleteDismissed(false)

    let mutation = updateMutation
    let variables = { input: { newStatus, alertIDs } }

    if (newStatus === 'StatusUnacknowledged') {
      mutation = escalateMutation
      variables = { input: alertIDs }
    }

    mutate({ mutation, variables })
  }

  let updateMessage, errorMessage
  if (status.error && !status.loading) {
    errorMessage = status.error.message
  }

  if (status.data && !status.loading) {
    const numUpdated =
      status.data.updateAlerts?.length ||
      status.data.escalateAlerts?.length ||
      0

    updateMessage = `${numUpdated} of ${checkedCount} alert${
      checkedCount === 1 ? '' : 's'
    } updated`
  }

  const showAlertActionSnackbar = Boolean(
    !actionCompleteDismissed && (errorMessage || updateMessage),
  )

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
   * Gets the header to display above the list to give a quick overview
   * on if they are viewing alerts for all services or only their
   * favorited services.
   *
   * Possibilities:
   *   - Home page, showing alerts for all services
   *   - Home page, showing alerts for any favorited services and notified alerts
   *   - Services page, alerts for that service
   */
  function getHeaderNote() {
    const { favoritesOnly, includeNotified } = variables.input

    if (includeNotified && favoritesOnly) {
      return `Showing ${filter} alerts you are on-call for and from any services you have favorited.`
    }

    if (allServices) {
      return `Showing ${filter} alerts for all services.`
    }

    if (props.serviceID && serviceNameQuery.data?.service?.name) {
      return `Showing ${filter} alerts for the service ${serviceNameQuery.data.service.name}.`
    }

    return null
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
        onClick: makeUpdateAlerts('StatusAcknowledged'),
      })
    }

    if (filter !== 'closed') {
      actions.push(
        {
          icon: <CloseIcon />,
          label: 'Close',
          onClick: makeUpdateAlerts('StatusClosed'),
        },
        {
          icon: <EscalateIcon />,
          label: 'Escalate',
          onClick: makeUpdateAlerts('StatusUnacknowledged'),
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
        headerNote={getHeaderNote()}
        mapDataNode={(a) => ({
          id: a.id,
          status: getListItemStatus(a.status),
          title: `${a.alertID}: ${a.status
            .toUpperCase()
            .replace('STATUS', '')}`,
          subText: (props.serviceID ? '' : a.service.name + ': ') + a.summary,
          action: <ListItemText secondary={formatTimeSince(a.createdAt)} />,
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
        transition={isMobileScreenSize && showAlertActionSnackbar}
      />

      {/* Update message after using checkbox actions */}
      <UpdateAlertsSnackbar
        errorMessage={errorMessage}
        onClose={() => setActionCompleteDismissed(true)}
        open={showAlertActionSnackbar}
        updateMessage={updateMessage}
      />
    </React.Fragment>
  )
}

AlertsList.propTypes = {
  serviceID: p.string,
}
