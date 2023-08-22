import React from 'react'
import { gql, useQuery } from '@apollo/client'
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
        stepNumber
        targets {
          id
          name
          type
        }
      }
    }
  }
`

function PolicyStepsQuery(props: { escalationPolicyID: string }): JSX.Element {
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

export default PolicyStepsQuery
