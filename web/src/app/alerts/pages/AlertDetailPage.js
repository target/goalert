import { gql } from '@apollo/client'
import React, { Component } from 'react'
import { GenericError, ObjectNotFound } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import { Query } from '@apollo/client/react/components'
import AlertDetails from '../components/AlertDetails'
import { POLL_ERROR_INTERVAL, POLL_INTERVAL } from '../../config'

const query = gql`
  query AlertDetailsPageQuery($id: Int!) {
    alert(id: $id) {
      id
      alertID
      status
      summary
      details
      createdAt
      service {
        id
        name
        escalationPolicy {
          id
          steps {
            delayMinutes
            targets {
              id
              type
              name
            }
          }
        }
      }
      state {
        lastEscalation
        stepNumber
        repeatCount
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
          if (!data && loading) return <Spinner />
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
