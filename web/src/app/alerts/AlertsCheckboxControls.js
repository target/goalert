import React, { useEffect, useState } from 'react'
import { useQuery, useMutation } from '@apollo/react-hooks'
import { useDispatch, useSelector } from 'react-redux'
import { PropTypes as p } from 'prop-types'
import {
  setCheckedAlerts as _setCheckedAlerts,
  setAlertsActionComplete as _setAlertsActionComplete,
} from '../actions'
import {
  Checkbox,
  Grid,
  Icon,
  IconButton,
  Tooltip,
  makeStyles,
} from '@material-ui/core'
import {
  ArrowDropDown,
  Check as AcknowledgeIcon,
  Close as CloseIcon,
  ArrowUpward as EscalateIcon,
} from '@material-ui/icons'
import OtherActions from '../util/OtherActions'
import classnames from 'classnames'
import gql from 'graphql-tag'
import UpdateAlertsSnackbar from './components/UpdateAlertsSnackbar'
import { urlParamSelector } from '../selectors'
import { alertsListQuery } from './AlertsList'
import { GenericError } from '../error-pages'

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

const useStyles = makeStyles({
  checkboxGridContainer: {
    paddingLeft: '1em',
  },
  hover: {
    '&:hover': {
      cursor: 'pointer',
    },
  },
  icon: {
    alignItems: 'center',
    display: 'flex',
  },
  popper: {
    opacity: 1,
  },
})

export default function AlertsCheckboxControlsQuery(props) {
  const { loading, error, data } = useQuery(alertsListQuery, {
    variables: props.variables,
  })

  if (loading && !data) return null
  // todo: test what this looks like
  if (error) return <GenericError error={error.message} />

  return <AlertsCheckboxControls alerts={data.alerts.nodes} />
}

AlertsCheckboxControlsQuery.propTypes = {
  variables: p.object.isRequired,
}

export function AlertsCheckboxControls(props) {
  const classes = useStyles()

  const [errorMessage, setErrorMessage] = useState('')
  const [updateMessage, setUpdateMessage] = useState('')

  const params = useSelector(urlParamSelector)
  const filter = params('filter', 'active')
  const alerts = useSelector(state => state.alerts.alerts).filter(
    a => a.status !== 'StatusClosed',
  )
  const actionComplete = useSelector(state => state.alerts.actionComplete)
  const checkedAlerts = useSelector(state => state.alerts.checkedAlerts)

  const dispatch = useDispatch()
  const setCheckedAlerts = arr => dispatch(_setCheckedAlerts(arr))
  const setAlertsActionComplete = bool =>
    dispatch(_setAlertsActionComplete(bool))

  useEffect(() => {
    return () => setNone()
  }, [])

  const [ackAlerts] = useMutation(updateMutation, {
    variables: {
      input: {
        alertIDs: checkedAlerts,
        newStatus: 'StatusAcknowledged',
      },
    },
    onError: err => {
      setAlertsActionComplete(true)
      setErrorMessage(err.message)
    },
    update: (cache, { data }) => onUpdate(data?.updateAlerts?.length ?? 0),
  })

  const [closeAlerts] = useMutation(updateMutation, {
    variables: {
      input: {
        alertIDs: checkedAlerts,
        newStatus: 'StatusClosed',
      },
    },
    onError: err => {
      setAlertsActionComplete(true)
      setErrorMessage(err.message)
    },
    update: (cache, { data }) => onUpdate(data?.updateAlerts?.length ?? 0),
  })

  const [escalateAlerts] = useMutation(escalateMutation, {
    variables: {
      input: checkedAlerts,
    },
    onError: err => {
      setAlertsActionComplete(true)
      setErrorMessage(err.message)
    },
    update: (cache, { data }) => onUpdate(data?.escalateAlerts?.length ?? 0),
  })

  function areNoneChecked() {
    return checkedAlerts.length === 0
  }

  function setAll() {
    setCheckedAlerts(alerts.map(a => a.id))
  }

  function setNone() {
    return setCheckedAlerts([])
  }

  function handleToggleSelectAll() {
    if (areNoneChecked()) {
      return setAll()
    }
    return setNone()
  }

  function onUpdate(numUpdated) {
    setAlertsActionComplete(true) // for create fab transition
    setUpdateMessage(`${numUpdated} of ${checkedAlerts.length} alerts updated`)
    setNone()
  }

  return (
    <React.Fragment>
      <UpdateAlertsSnackbar
        errorMessage={errorMessage}
        numberChecked={checkedAlerts.length}
        onClose={() => setAlertsActionComplete(false)}
        onExited={() => {
          setErrorMessage('')
          setUpdateMessage('')
        }}
        open={actionComplete}
        updateMessage={updateMessage}
      />
      <Grid container className={classes.checkboxGridContainer}>
        <Grid item>
          <Checkbox
            checked={
              alerts.length === checkedAlerts.length && alerts.length !== 0
            }
            data-cy='select-all'
            indeterminate={
              checkedAlerts.length > 0 && alerts.length !== checkedAlerts.length
            }
            tabIndex={-1}
            onChange={handleToggleSelectAll}
          />
        </Grid>
        <Grid
          item
          className={classnames(classes.hover, classes.icon)}
          data-cy='checkboxes-menu'
        >
          <OtherActions
            icon={
              <Icon>
                <ArrowDropDown />
              </Icon>
            }
            actions={[
              {
                label: 'All',
                onClick: setAll,
              },
              {
                label: 'None',
                onClick: setNone,
              },
            ]}
            placement='right'
          />
        </Grid>
        {renderActionButtons()}
      </Grid>
    </React.Fragment>
  )

  function renderActionButtons() {
    if (!checkedAlerts.length) return null

    let ack = null
    let close = null
    let escalate = null
    if (
      filter === 'active' ||
      filter === 'unacknowledged' ||
      filter === 'all'
    ) {
      ack = (
        <Tooltip
          title='Acknowledge'
          placement='bottom'
          classes={{ popper: classes.popper }}
        >
          <IconButton
            aria-label='Acknowledge Selected Alerts'
            data-cy='acknowledge'
            onClick={ackAlerts}
          >
            <AcknowledgeIcon />
          </IconButton>
        </Tooltip>
      )
    }

    if (filter !== 'closed') {
      close = (
        <Tooltip
          title='Close'
          placement='bottom'
          classes={{ popper: classes.popper }}
        >
          <IconButton
            aria-label='Close Selected Alerts'
            data-cy='close'
            onClick={closeAlerts}
          >
            <CloseIcon />
          </IconButton>
        </Tooltip>
      )
      escalate = (
        <Tooltip
          title='Escalate'
          placement='bottom'
          classes={{ popper: classes.popper }}
        >
          <IconButton
            aria-label='Escalate Selected Alerts'
            data-cy='escalate'
            onClick={escalateAlerts}
          >
            <EscalateIcon />
          </IconButton>
        </Tooltip>
      )
    }

    return (
      <Grid item className={classes.icon}>
        {ack}
        {close}
        {escalate}
      </Grid>
    )
  }
}

AlertsCheckboxControls.propTypes = {
  alerts: p.array.isRequired,
}
