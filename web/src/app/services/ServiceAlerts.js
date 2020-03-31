/* eslint @typescript-eslint/camelcase: 0 */
import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { useMutation } from 'react-apollo'
import AlertsList from '../alerts/AlertsList'
import gql from 'graphql-tag'
import Options from '../util/Options'
import PageActions from '../util/PageActions'
import AlertsListFilter from '../alerts/components/AlertsListFilter'
import Search from '../util/Search'
import FormDialog from '../dialogs/FormDialog'

const mutation = gql`
  mutation UpdateAlertStatusByServiceMutation(
    $input: UpdateAlertStatusByServiceInput!
  ) {
    updateAlertStatusByService(input: $input)
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

  const getStatusTxt = () => {
    if (alertStatus === 'StatusAcknowledged') {
      return 'acknowledge'
    }

    return 'close'
  }

  const getMenuOptions = () => {
    return [
      {
        text: 'Acknowledge All Alerts',
        onClick: handleClickAckAll,
      },
      {
        text: 'Close All Alerts',
        onClick: handleClickCloseAll,
      },
    ]
  }

  return (
    <React.Fragment>
      <PageActions key='actions'>
        <Search key='search' />
        <AlertsListFilter key='filter' serviceID={serviceID} />
        <Options key='options' options={getMenuOptions()} legacyClient />
      </PageActions>
      {showDialog && (
        <FormDialog
          title='Are you sure?'
          confirm
          subTitle={`This will ${getStatusTxt()} all the alerts for this service.`}
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
