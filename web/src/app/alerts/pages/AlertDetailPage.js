import React, { Component } from 'react'
import { GenericError, ObjectNotFound } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import gql from 'graphql-tag'
import { Query } from 'react-apollo'
import AlertDetails from '../components/AlertDetails'
import { POLL_ERROR_INTERVAL, POLL_INTERVAL } from '../../util/poll_intervals'

const query = gql`
  query AlertDetailsPageQuery($id: Int!) {
    alert(id: $id) {
      number: _id
      id
      status: status_2
      escalation_level
      description
      details
      summary
      service_id
      source
      assignments {
        id
        name
      }
      service {
        id
        name
        escalation_policy_id
      }
      logs_2 {
        event
        message
        timestamp
      }
      escalation_policy_snapshot {
        repeat
        current_level
        last_escalation
        steps {
          delay_minutes
          users {
            id
            name
          }
          schedules {
            id
            name
          }
        }
      }
    }
  }
`

export default class AlertDetailPage extends Component {
  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.match.params.alertID }}
        pollInterval={POLL_INTERVAL}
      >
        {({ loading, error, data, startPolling }) => {
          if (loading) return <Spinner />
          if (error) {
            startPolling(POLL_ERROR_INTERVAL)
            return <GenericError error={error.message} />
          }

          if (!data.alert) return <ObjectNotFound type='alert' />
          startPolling(POLL_INTERVAL)
          return <AlertDetails data={data.alert} />
        }}
      </Query>
    )
  }
}
