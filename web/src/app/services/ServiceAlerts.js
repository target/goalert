import { gql, useMutation } from '@apollo/client'
/* eslint @typescript-eslint/camelcase: 0 */
import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'

import AlertsList from '../alerts/AlertsList'
import PageActions from '../util/PageActions'
import FormDialog from '../dialogs/FormDialog'
import OtherActions from '../util/OtherActions'

const mutation = gql`
  mutation UpdateAlertsByServiceMutation($input: UpdateAlertsByServiceInput!) {
    updateAlertsByService(input: $input)
  }
`

export default function ServiceAlerts(props) {
  const { serviceID } = props

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

  const getMenuOptions = () => {
    return [
      {
        label: 'Acknowledge All Alerts',
        onClick: handleClickAckAll,
      },
      {
        label: 'Close All Alerts',
        onClick: handleClickCloseAll,
      },
    ]
  }

  return (
    <React.Fragment>
      <PageActions key='actions'>
        <OtherActions actions={getMenuOptions()} />
      </PageActions>
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
      <AlertsList serviceID={serviceID} />
    </React.Fragment>
  )
}

ServiceAlerts.propTypes = {
  serviceID: p.string.isRequired,
}
