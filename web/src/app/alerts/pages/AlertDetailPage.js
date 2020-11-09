import { gql, useQuery } from '@apollo/client'
import React from 'react'
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

function AlertDetailPage(props) {
  const { loading, error, data } = useQuery(query, {
    variables: { id: props.match.params.alertID },
  })

  if (!data && loading) return <Spinner />

  if (error) return <GenericError error={error.message} />

  if (!data.alert) return <ObjectNotFound type='alert' />

  return <AlertDetails data={data.alert} />
}

export default AlertDetailPage
