import React, { ReactElement, useState } from 'react'
import { useMutation } from '@apollo/client'
import { useQuery, gql } from 'urql'
import { Alert, Hidden, ListItemText } from '@mui/material'
import Snackbar from '@mui/material/Snackbar'
import makeStyles from '@mui/styles/makeStyles'
import {
  ArrowUpward as EscalateIcon,
  Check as AcknowledgeIcon,
  Close as CloseIcon,
} from '@mui/icons-material'
import { DateTime } from 'luxon'

import AlertsListFilter from './components/AlertsListFilter'
import AlertsListControls from './components/AlertsListControls'
import { formatTimeSince } from '../util/timeFormat'
import QueryList from '../lists/QueryList'
import { useIsWidthDown } from '../util/useWidth'
import CreateFAB from '../lists/CreateFAB'
import CreateAlertDialog from './CreateAlertDialog/CreateAlertDialog'
import { useURLParam } from '../actions'
import { ControlledPaginatedListAction } from '../lists/ControlledPaginatedList'

interface AlertsListProps {
  serviceID: string
  secondaryActions?: ReactElement
}

interface MutationVariables {
  input: MutationVariablesInput
}

interface StatusUnacknowledgedVariables {
  input: (string | number)[]
}

interface MutationVariablesInput {
  newStatus: string
  alertIDs: (string | number)[]
}

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

const useStyles = makeStyles({
  alertTimeContainer: {
    width: 'max-content',
  },
})

function getStatusFilter(s: string): string[] {
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

export default function AlertsList(props: AlertsListProps): JSX.Element {
  const classes = useStyles()
  // transition fab above snackbar when snackbar width overlaps fab placement
  const isXs = useIsWidthDown('sm')

  const [checkedCount, setCheckedCount] = useState(0)
  const [showCreate, setShowCreate] = useState(false)

  // used if user dismisses snackbar before the auto-close timer finishes
  const [actionCompleteDismissed, setActionCompleteDismissed] = useState(true)

  const [allServices] = useURLParam('allServices', false)
  const [fullTime] = useURLParam('fullTime', false)
  const [filter] = useURLParam<string>('filter', 'active')

  // query for current service name if props.serviceID is provided
  const [serviceNameQuery] = useQuery({
    query: gql`
      query ($id: ID!) {
        service(id: $id) {
          id
          name
        }
      }
    `,
    variables: { id: props.serviceID || '' },
    pause: !props.serviceID,
  })

  // alerts list query variables
  const variables = {
    input: {
      filterByStatus: getStatusFilter(filter),
      first: 25,
      // default to favorites only, unless viewing alerts from a service's page
      favoritesOnly: !props.serviceID && !allServices,
      includeNotified: !props.serviceID, // keep service list alerts specific to that service,
      filterByServiceID: props.serviceID ? [props.serviceID] : null,
    },
  }

  const [mutate, status] = useMutation(updateMutation)

  const makeUpdateAlerts =
    (newStatus: string) => (alertIDs: (string | number)[]) => {
      setCheckedCount(alertIDs.length)
      setActionCompleteDismissed(false)

      let mutation = updateMutation
      let variables: MutationVariables | StatusUnacknowledgedVariables = {
        input: { newStatus, alertIDs },
      }

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
  function getListItemStatus(s: string): 'ok' | 'warn' | 'err' | undefined {
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
  function getHeaderNote(): string | undefined {
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
  }

  /*
   * Passes the proper actions to ListControls depending
   * on which tab is currently filtering the alerts list
   */
  function getActions(): ControlledPaginatedListAction[] {
    const actions: ControlledPaginatedListAction[] = []

    if (filter === 'unacknowledged' || filter === 'active') {
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
          action: (
            <ListItemText
              className={classes.alertTimeContainer}
              secondary={
                fullTime
                  ? DateTime.fromISO(a.createdAt).toLocaleString(
                      DateTime.DATETIME_MED,
                    )
                  : formatTimeSince(a.createdAt)
              }
            />
          ),
          url: `/services/${a.service.id}/alerts/${a.id}`,
          selectable: a.status !== 'StatusClosed',
        })}
        variables={variables}
        secondaryActions={
          props?.secondaryActions ?? (
            <AlertsListFilter serviceID={props.serviceID} />
          )
        }
        cardHeader={
          <Hidden lgDown>
            <AlertsListControls />
          </Hidden>
        }
        checkboxActions={getActions()}
      />

      <CreateFAB
        title='Create Alert'
        transition={isXs && showAlertActionSnackbar}
        onClick={() => setShowCreate(true)}
      />
      {showCreate && (
        <CreateAlertDialog
          onClose={() => setShowCreate(false)}
          serviceID={props.serviceID}
        />
      )}

      {/* Update message after using checkbox actions */}
      <Snackbar
        autoHideDuration={errorMessage ? null : 6000}
        onClose={() => setActionCompleteDismissed(true)}
        open={showAlertActionSnackbar}
      >
        <Alert
          severity={errorMessage ? 'error' : 'info'}
          onClose={() => setActionCompleteDismissed(true)}
          variant='filled'
        >
          {errorMessage || updateMessage}
        </Alert>
      </Snackbar>
    </React.Fragment>
  )
}
