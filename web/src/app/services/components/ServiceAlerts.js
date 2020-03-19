/* eslint @typescript-eslint/camelcase: 0 */
import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import { useMutation } from 'react-apollo'
import AlertsList from '../../alerts/components/AlertsList'
import gql from 'graphql-tag'
import Options from '../../util/Options'
import PageActions from '../../util/PageActions'
import AlertsListFilter from '../../alerts/components/AlertsListFilter'
import Search from '../../util/Search'
import FormDialog from '../../dialogs/FormDialog'
import { LegacyGraphQLClient } from '../../apollo'

const mutation = gql`
  mutation UpdateAlertStatusByServiceMutation(
    $input: UpdateAlertStatusByServiceInput!
  ) {
    updateAlertStatusByService(input: $input) {
      number: _id
      id
      status: status_2
      created_at
      summary
      service {
        id
        name
      }
    }
  }
`

export default function ServiceAlerts(props) {
  const { serviceID } = props

  const [alertStatus, setAlertStatus] = useState('')
  const [showDialog, setShowDialog] = useState(false)
  const [mutate, mutationStatus] = useMutation(mutation, {
    client: LegacyGraphQLClient,
    variables: {
      input: {
        service_id: serviceID,
        status: alertStatus + 'd', // closed or acknowledged
      },
    },
    onCompleted: () => setShowDialog(false),
  })

  const { loading } = mutationStatus

  const handleClickAckAll = () => {
    setAlertStatus('acknowledge')
    setShowDialog(true)
  }

  const handleClickCloseAll = () => {
    setAlertStatus('close')
    setShowDialog(true)
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
          subTitle={`This will ${alertStatus} all the alerts for this service.`}
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
