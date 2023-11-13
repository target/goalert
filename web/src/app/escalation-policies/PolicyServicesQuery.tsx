import React from 'react'
import { gql, useQuery } from 'urql'
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

function PolicyServicesQuery(props: { policyID: string }): React.ReactNode {
  const [{ data, fetching, error }] = useQuery({
    query,
    variables: { id: props.policyID },
  })

  if (!data && fetching) {
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
