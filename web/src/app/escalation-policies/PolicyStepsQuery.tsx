import React from 'react'
import { gql, useQuery } from '@apollo/client'
import PolicyStepsCard from './PolicyStepsCard'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'
import { useExpFlag } from '../util/useExpFlag'

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

export const policyStepsQueryDest = gql`
  query stepsQueryDest($id: ID!) {
    escalationPolicy(id: $id) {
      id
      repeat
      steps {
        id
        delayMinutes
        stepNumber
        actions {
          type
          displayInfo {
            text
            iconURL
            iconAltText
            linkURL
          }
        }
      }
    }
  }
`

function PolicyStepsQuery(props: { escalationPolicyID: string }): JSX.Element {
  const hasDestTypesFlag = useExpFlag('dest-types')

  const { data, loading, error } = useQuery(
    hasDestTypesFlag ? policyStepsQueryDest : policyStepsQuery,
    {
      variables: { id: props.escalationPolicyID },
    },
  )

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
