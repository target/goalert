import React, { Component } from 'react'
import { PropTypes as p } from 'prop-types'
import AlertsList from '../../alerts/components/AlertsList'
import gql from 'graphql-tag'
import Options from '../../util/Options'
import ConfirmationDialog from '../../dialogs/components/ConfirmationDialog'
import PageActions from '../../util/PageActions'
import AlertsListFilter from '../../alerts/components/AlertsListFilter'
import Search from '../../util/Search'
import { sendGAEvent } from '../../util/GoogleAnalytics'

const updateAllAlertsMutation = gql`
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

export default class ServiceAlerts extends Component {
  static propTypes = {
    serviceID: p.string.isRequired,
  }

  state = {
    alertStatus: '',
    showUpdateConfirmation: false,
  }

  handleClickAckAll = () => {
    this.setState({ alertStatus: 'acknowledge', showUpdateConfirmation: true })
  }

  handleClickCloseAll = () => {
    this.setState({ alertStatus: 'close', showUpdateConfirmation: true })
  }

  handleShowForm = (key, bool) => {
    this.setState({
      [key]: bool,
    })
  }

  handleAllAlertsSuccess = () => {
    sendGAEvent({
      category: 'Service',
      action: this.state.alertStatus + ' All Action Completed',
    })
  }

  getMenuOptions = () => {
    return [
      {
        text: 'Acknowledge All Alerts',
        onClick: this.handleClickAckAll,
      },
      {
        text: 'Close All Alerts',
        onClick: this.handleClickCloseAll,
      },
    ]
  }

  render() {
    const { serviceID } = this.props

    return (
      <React.Fragment>
        <PageActions key='actions'>
          <Search key='search' />
          <AlertsListFilter key='filter' serviceID={serviceID} />
          <Options key='options' options={this.getMenuOptions()} />
        </PageActions>
        <ConfirmationDialog
          key='update-alerts-form'
          mutation={updateAllAlertsMutation}
          refetchQueries={['alerts']}
          mutationVariables={{
            input: {
              service_id: serviceID,
              status: this.state.alertStatus + 'd', // closed or acknowledged
            },
          }}
          onMutationSuccess={this.handleAllAlertsSuccess}
          onRequestClose={() =>
            this.handleShowForm('showUpdateConfirmation', false)
          }
          open={this.state.showUpdateConfirmation}
          message={
            'This will ' +
            this.state.alertStatus +
            ' all the alerts for this service.'
          }
          warning='This will stop all notifications from being sent out for all alerts with this service.'
        />
        <AlertsList serviceID={serviceID} />
      </React.Fragment>
    )
  }
}
