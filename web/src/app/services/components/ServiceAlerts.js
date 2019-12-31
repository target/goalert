import React, { useState } from 'react'
import { PropTypes as p } from 'prop-types'
import AlertsList from '../../alerts/components/AlertsList'
import gql from 'graphql-tag'
import Options from '../../util/Options'
import ConfirmationDialog from '../../dialogs/components/ConfirmationDialog'
import PageActions from '../../util/PageActions'
import AlertsListFilter from '../../alerts/components/AlertsListFilter'
import Search from '../../util/Search'
import { sendGAEvent } from '../../util/GoogleAnalytics'

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

  const handleClickAckAll = () => {
    setAlertStatus('acknowledge')
    setShowDialog(true)
  }

  const handleClickCloseAll = () => {
    setAlertStatus('close')
    setShowDialog(true)
  }

  const handleAllAlertsSuccess = () => {
    sendGAEvent({
      category: 'Service',
      action: alertStatus + ' All Action Completed',
    })
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
        <Options key='options' options={getMenuOptions()} />
      </PageActions>
      <ConfirmationDialog
        key='update-alerts-form'
        mutation={mutation}
        refetchQueries={['alerts']}
        mutationVariables={{
          input: {
            service_id: serviceID,
            status: alertStatus + 'd', // closed or acknowledged
          },
        }}
        onMutationSuccess={handleAllAlertsSuccess}
        onRequestClose={() => setShowDialog(false)}
        open={showDialog}
        message={`This will ${alertStatus} all the alerts for this service.`}
        warning='This will stop all notifications from being sent out for all alerts with this service.'
      />
      <AlertsList serviceID={serviceID} />
    </React.Fragment>
  )
}

ServiceAlerts.propTypes = {
  serviceID: p.string.isRequired,
}
