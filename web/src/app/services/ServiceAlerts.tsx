import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
import Button from '@mui/material/Button'
import ButtonGroup from '@mui/material/ButtonGroup'
import Grid from '@mui/material/Grid'
import makeStyles from '@mui/styles/makeStyles'

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

export default function ServiceAlerts(props: {
  serviceID: string
}): React.JSX.Element {
  const classes = useStyles()

  const [alertStatus, setAlertStatus] = useState('')
  const [showDialog, setShowDialog] = useState(false)
  const [mutationStatus, mutate] = useMutation(mutation)

  const handleClickAckAll = (): void => {
    setAlertStatus('StatusAcknowledged')
    setShowDialog(true)
  }

  const handleClickCloseAll = (): void => {
    setAlertStatus('StatusClosed')
    setShowDialog(true)
  }

  const getStatusText = (): string => {
    if (alertStatus === 'StatusAcknowledged') {
      return 'acknowledge'
    }

    return 'close'
  }

  const secondaryActions = (
    <Grid className={classes.filter} container spacing={2} alignItems='center'>
      <Grid item>
        <AlertsListFilter serviceID={props.serviceID} />
      </Grid>
      <Grid item>
        <ButtonGroup variant='outlined'>
          <Button onClick={handleClickAckAll}>Acknowledge All</Button>
          <Button onClick={handleClickCloseAll}>Close All</Button>
        </ButtonGroup>
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
          onSubmit={() =>
            mutate({
              input: {
                serviceID: props.serviceID,
                newStatus: alertStatus,
              },
            }).then((res) => {
              if (res.error) return
              setShowDialog(false)
            })
          }
          loading={mutationStatus.fetching}
          onClose={() => setShowDialog(false)}
        />
      )}
      <AlertsList
        serviceID={props.serviceID}
        secondaryActions={secondaryActions}
      />
    </React.Fragment>
  )
}
