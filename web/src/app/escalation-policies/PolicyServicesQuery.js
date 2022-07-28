import React from 'react'
import { gql, useQuery } from '@apollo/client'
import PolicyServicesCard from './PolicyServicesCard'
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

function PolicyServicesQuery({ policyID }) {
  const { data, loading, error } = useQuery(query, {
    variables: { id: policyID },
  })

  if (error) {
    return <GenericError error={error.message} />
  }

  if (!loading && !data?.escalationPolicy) {
    return <ObjectNotFound />
  }

  return (
    <PolicyServicesCard
      services={data?.escalationPolicy?.assignedTo ?? []}
      isLoading={loading}
    />
  )
}

export default PolicyServicesQuery
