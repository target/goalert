import React, { ReactElement, useState, useContext, useEffect } from 'react'
import { useMutation } from '@apollo/client'
import { useQuery, gql } from 'urql'
import { Grid, Hidden, ListItemText } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import {
  ArrowUpward as EscalateIcon,
  Check as AcknowledgeIcon,
  Close as CloseIcon,
} from '@mui/icons-material'

import AlertsListFilter from './components/AlertsListFilter'
import AlertsListControls from './components/AlertsListControls'
import QueryList from '../lists/QueryList'
import CreateAlertDialog from './CreateAlertDialog/CreateAlertDialog'
import { useURLParam } from '../actions'
import { ControlledPaginatedListAction } from '../lists/ControlledPaginatedList'
import ServiceNotices from '../services/ServiceNotices'
import { Time } from '../util/Time'
import { NotificationContext } from '../main/SnackbarNotification'
import ReactGA from 'react-ga4'
import { useConfigValue } from '../util/RequireConfig'

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

  const [event, setEvent] = useState('')
  const [analyticsID] = useConfigValue('General.GoogleAnalyticsID') as [string]
  const [selectedCount, setSelectedCount] = useState(0)
  const [checkedCount, setCheckedCount] = useState(0)

  const [allServices] = useURLParam('allServices', false)
  const [fullTime] = useURLParam('fullTime', false)
  const [filter] = useURLParam<string>('filter', 'active')

  useEffect(() => {
    if (analyticsID && event.length)
      ReactGA.event({ category: 'Bulk Alert Action', action: event })
  }, [event, analyticsID])

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

  const { setNotification } = useContext(NotificationContext)

  const [mutate] = useMutation(updateMutation, {
    onCompleted: (data) => {
      const numUpdated =
        data.updateAlerts?.length || data.escalateAlerts?.length || 0

      const msg = `${numUpdated} of ${checkedCount} alert${
        checkedCount === 1 ? '' : 's'
      } updated`

      setNotification({
        message: msg,
        severity: 'info',
      })
    },
    onError: (error) => {
      setNotification({
        message: error.message,
        severity: 'error',
      })
    },
  })

  const makeUpdateAlerts =
    (newStatus: string) => (alertIDs: (string | number)[]) => {
      setCheckedCount(alertIDs.length)

      let mutation = updateMutation
      let variables: MutationVariables | StatusUnacknowledgedVariables = {
        input: { newStatus, alertIDs },
      }

      switch (newStatus) {
        case 'StatusUnacknowledged':
          mutation = escalateMutation
          variables = { input: alertIDs }
          setEvent('alertlist_escalated')
          break
        case 'StatusAcknowledged':
          setEvent('alertlist_acknowledged')
          break
        case 'StatusClosed':
          setEvent('alertlist_closed')
          break
      }

      mutate({ mutation, variables })
    }

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
      actions.push({
        icon: <CloseIcon />,
        label: 'Close',
        onClick: makeUpdateAlerts('StatusClosed'),
      })

      if (selectedCount === 1) {
        actions.push({
          icon: <EscalateIcon />,
          label: 'Escalate',
          onClick: makeUpdateAlerts('StatusUnacknowledged'),
        })
      }
    }

    return actions
  }

  return (
    <React.Fragment>
      <Grid container direction='column' spacing={2}>
        <ServiceNotices serviceID={props.serviceID} />
        <Grid item>
          <QueryList
            query={alertsListQuery}
            infiniteScroll
            onSelectionChange={(selected) => setSelectedCount(selected.length)}
            headerNote={getHeaderNote()}
            mapDataNode={(a) => ({
              id: a.id,
              status: getListItemStatus(a.status),
              title: `${a.alertID}: ${a.status
                .toUpperCase()
                .replace('STATUS', '')}`,
              subText:
                (props.serviceID ? '' : a.service.name + ': ') + a.summary,
              action: (
                <ListItemText
                  className={classes.alertTimeContainer}
                  secondary={
                    <Time
                      time={a.createdAt}
                      format={fullTime ? 'default' : 'relative'}
                    />
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
            renderCreateDialog={(onClose) => (
              <CreateAlertDialog
                serviceID={props.serviceID}
                onClose={onClose}
              />
            )}
            createLabel='Alert'
            cardHeader={
              <Hidden lgDown>
                <AlertsListControls />
              </Hidden>
            }
            checkboxActions={getActions()}
          />
        </Grid>
      </Grid>
    </React.Fragment>
  )
}
