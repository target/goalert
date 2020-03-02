import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { Hidden, Typography, makeStyles } from '@material-ui/core'
import { useDispatch, useSelector } from 'react-redux'
import { absURLSelector, urlParamSelector } from '../selectors'
import QueryList from '../lists/QueryList'
import gql from 'graphql-tag'

import AlertsListFilter from './components/AlertsListFilter'
import AlertsListControls from './components/AlertsListControls'
import AlertsListFloatingItems from './AlertsListFloatingItems'
import statusStyles from '../util/statusStyles'
import {
  setAlertsActionComplete as _setAlertsActionComplete,
  setCheckedItems as _setCheckedItems,
} from '../actions'
import { formatTimeSince } from '../util/timeFormat'
import { useMutation } from '@apollo/react-hooks'
import UpdateAlertsSnackbar from './components/UpdateAlertsSnackbar'
import {
  ArrowUpward as EscalateIcon,
  Check as AcknowledgeIcon,
  Close as CloseIcon,
} from '@material-ui/icons'

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

  const [errorMessage, setErrorMessage] = useState('')
  const [updateMessage, setUpdateMessage] = useState('')

  // get redux vars
  const absURL = useSelector(absURLSelector)
  const params = useSelector(urlParamSelector)
  const allServices = params('allServices')
  const filter = params('filter', 'active')

  const dispatch = useDispatch()

  const actionComplete = useSelector(state => state.alerts.actionComplete)
  const setAlertsActionComplete = bool =>
    dispatch(_setAlertsActionComplete(bool))

  const checkedItems = useSelector(state => state.list.checkedItems)
  const setCheckedItems = array => dispatch(_setCheckedItems(array))

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

  function onCompleted(numUpdated) {
    setAlertsActionComplete(true) // for create fab transition
    setUpdateMessage(`${numUpdated} of ${checkedItems.length} alerts updated`)
    setCheckedItems([])
  }

  // ack mutation
  const [ackAlerts] = useMutation(updateMutation, {
    variables: {
      input: {
        alertIDs: checkedItems,
        newStatus: 'StatusAcknowledged',
      },
    },
    onError: err => {
      setAlertsActionComplete(true)
      setErrorMessage(err.message)
    },
    onCompleted: data => onCompleted(data?.updateAlerts?.length ?? 0),
  })

  // close mutation
  const [closeAlerts] = useMutation(updateMutation, {
    variables: {
      input: {
        alertIDs: checkedItems,
        newStatus: 'StatusClosed',
      },
    },
    onError: err => {
      setAlertsActionComplete(true)
      setErrorMessage(err.message)
    },
    onCompleted: data => onCompleted(data?.updateAlerts?.length ?? 0),
  })

  // escalate mutation
  const [escalateAlerts] = useMutation(escalateMutation, {
    variables: {
      input: checkedItems,
    },
    onError: err => {
      setAlertsActionComplete(true)
      setErrorMessage(err.message)
    },
    onCompleted: data => onCompleted(data?.escalateAlerts?.length ?? 0),
  })

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
    if (!checkedItems.length) return actions

    if (filter !== 'closed' && filter !== 'acknowledged') {
      actions.push({
        icon: <AcknowledgeIcon />,
        tooltip: 'Acknowledge',
        onClick: ackAlerts,
        ariaLabel: 'Acknowledge Selected Alerts',
        dataCy: 'acknowledge',
      })
    }

    if (filter !== 'closed') {
      actions.push(
        {
          icon: <CloseIcon />,
          tooltip: 'Close',
          onClick: closeAlerts,
          ariaLabel: 'Close Selected Alerts',
          dataCy: 'close',
        },
        {
          icon: <EscalateIcon />,
          tooltip: 'Escalate',
          onClick: escalateAlerts,
          ariaLabel: 'Escalate Selected Alerts',
          dataCy: 'escalate',
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
        mapDataNode={a => {
          let item = {
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
          }

          if (a.status === 'StatusClosed') {
            item.CheckboxProps = {
              disabled: true,
              onClick: () => {},
            }
          }

          return item
        }}
        variables={variables}
        filter={<AlertsListFilter />}
        cardHeader={
          <Hidden mdDown>
            <AlertsListControls />
          </Hidden>
        }
        actions={getActions()}
      />
      <UpdateAlertsSnackbar
        errorMessage={errorMessage}
        numberChecked={checkedItems.length}
        onClose={() => setAlertsActionComplete(false)}
        onExited={() => {
          setErrorMessage('')
          setUpdateMessage('')
        }}
        open={actionComplete}
        updateMessage={updateMessage}
      />
      <AlertsListFloatingItems serviceID={props.serviceID} />
    </React.Fragment>
  )
}

AlertsList.propTypes = {
  serviceID: p.string,
}
