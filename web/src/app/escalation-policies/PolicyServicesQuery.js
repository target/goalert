import React from 'react'
import { PropTypes as p } from 'prop-types'
import gql from 'graphql-tag'
import PolicyServicesCard from './PolicyServicesCard'
import { useQuery } from '@apollo/react-hooks'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'

const query = gql`
  query($id: ID!) {
    escalationPolicy(id: $id) {
      id
      assignedTo {
        id
        name
      }
    }
  }
`

function PolicyServicesQuery(props) {
  const { data, loading, error } = useQuery(query, {
    variables: { id: props.escalationPolicyID },
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

PolicyServicesQuery.propTypes = {
  escalationPolicyID: p.string.isRequired,
}

export default PolicyServicesQuery
