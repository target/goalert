import React from 'react'
import { gql, useQuery } from '@apollo/client'
import { PropTypes as p } from 'prop-types'
import PolicyStepsCard from './PolicyStepsCard'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'

export const policyStepsQuery = gql`
  query stepsQuery($id: ID!) {
    escalationPolicy(id: $id) {
      id
      repeat
      steps {
        id
        delayMinutes
        targets {
          id
          name
          type
        }
      }
    }
  }
`

function PolicyStepsQuery(props) {
  const { data, loading, error } = useQuery(policyStepsQuery, {
    variables: { id: props.escalationPolicyID },
  })

  if (!data && loading) {
    return <Spinner />
  }
  if (error) {
    return <GenericError error={error.message} />
  }
  if (!data?.escalationPolicy) {
    return <ObjectNotFound />
  }

  return (
    <PolicyStepsCard
      escalationPolicyID={props.escalationPolicyID}
      repeat={data.escalationPolicy.repeat}
      steps={data.escalationPolicy.steps || []}
    />
  )
}

PolicyStepsQuery.propTypes = {
  escalationPolicyID: p.string.isRequired,
}

export default PolicyStepsQuery
