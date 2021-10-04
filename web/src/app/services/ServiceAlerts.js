import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import { PropTypes as p } from 'prop-types'
import Button from '@mui/material/Button'
import ButtonGroup from '@mui/material/ButtonGroup'
import Grid from '@mui/material/Grid'
import makeStyles from '@mui/styles/makeStyles';

import AlertsList from '../alerts/AlertsList'
import FormDialog from '../dialogs/FormDialog'
import AlertsListFilter from '../alerts/components/AlertsListFilter'

const mutation = gql`
  mutation UpdateAlertsByServiceMutation($input: UpdateAlertsByServiceInput!) {
    updateAlertsByService(input: $input)
  }
`

const useStyles = makeStyles({
  filter: {
    width: 'fit-content',
  },
})

export default function ServiceAlerts(props) {
  const { serviceID } = props
  const classes = useStyles()

  const [alertStatus, setAlertStatus] = useState('')
  const [showDialog, setShowDialog] = useState(false)
  const [mutate, mutationStatus] = useMutation(mutation, {
    variables: {
      input: {
        serviceID: serviceID,
        newStatus: alertStatus,
      },
    },
    onCompleted: () => setShowDialog(false),
  })

  const { loading } = mutationStatus

  const handleClickAckAll = () => {
    setAlertStatus('StatusAcknowledged')
    setShowDialog(true)
  }

  const handleClickCloseAll = () => {
    setAlertStatus('StatusClosed')
    setShowDialog(true)
  }

  const getStatusText = () => {
    if (alertStatus === 'StatusAcknowledged') {
      return 'acknowledge'
    }

    return 'close'
  }

  const secondaryActions = (
    <Grid className={classes.filter} container spacing={2} alignItems='center'>
      <Grid item>
        <ButtonGroup color='secondary' variant='outlined'>
          <Button onClick={handleClickAckAll}>Acknowledge All</Button>
          <Button onClick={handleClickCloseAll}>Close All</Button>
        </ButtonGroup>
      </Grid>
      <Grid item>
        <AlertsListFilter serviceID={serviceID} />
      </Grid>
    </Grid>
  )

  return (
    <React.Fragment>
      {showDialog && (
        <FormDialog
          title='Are you sure?'
          confirm
          subTitle={`This will ${getStatusText()} all the alerts for this service.`}
          caption='This will stop all notifications from being sent out for all alerts with this service.'
          onSubmit={() => mutate()}
          loading={loading}
          onClose={() => setShowDialog(false)}
        />
      )}
      <AlertsList serviceID={serviceID} secondaryActions={secondaryActions} />
    </React.Fragment>
  )
}

ServiceAlerts.propTypes = {
  serviceID: p.string.isRequired,
}
