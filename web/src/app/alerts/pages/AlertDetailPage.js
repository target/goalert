import React from 'react'
import { gql, useQuery } from '@apollo/client'
import { useParams } from 'react-router-dom'
import { GenericError, ObjectNotFound } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import AlertDetails from '../components/AlertDetails'

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
          repeat
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
      pendingNotifications {
        destination
      }
    }
  }
`

function AlertDetailPage() {
  const { alertID } = useParams()
  const { loading, error, data } = useQuery(query, {
    variables: { id: alertID },
  })

  if (!data && loading) return <Spinner />
  if (error) return <GenericError error={error.message} />
  if (!data.alert) return <ObjectNotFound type='alert' />

  return <AlertDetails data={data.alert} />
}

export default AlertDetailPage
