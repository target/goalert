import React from 'react'
import { gql, useQuery } from 'urql'
import PolicyStepsCard from './PolicyStepsCard'
import { GenericError, ObjectNotFound } from '../error-pages'

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
            ... on DestinationDisplayInfo {
              text
              iconURL
              iconAltText
              linkURL
            }
            ... on DestinationDisplayInfoError {
              error
            }
          }
        }
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
  const [{ data, error }] = useQuery({
    query: policyStepsQueryDest,
    variables: { id: props.escalationPolicyID },
  })

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
