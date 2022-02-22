import React from 'react'
import { gql, useQuery } from '@apollo/client'
import { useParams } from 'react-router-dom'
import PolicyServicesCard from './PolicyServicesCard'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'

const query = gql`
  query ($id: ID!) {
    escalationPolicy(id: $id) {
      id
      assignedTo {
        id
        name
      }
    }
  }
`

function PolicyServicesQuery() {
  const { escalationPolicyID } = useParams()
  const { data, loading, error } = useQuery(query, {
    variables: { id: escalationPolicyID },
  })

  if (!data && loading) {
    return <Spinner />
  }

  if (error) {
    return <GenericError error={error.message} />
  }

  if (!data.escalationPolicy) {
    return <ObjectNotFound />
  }

  return (
    <PolicyServicesCard services={data.escalationPolicy.assignedTo || []} />
  )
}

export default PolicyServicesQuery
