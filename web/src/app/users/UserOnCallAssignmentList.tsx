import React from 'react'
import { gql, useQuery } from '@apollo/client'
import { Card } from '@mui/material'
import FlatList from '../lists/FlatList'
import { sortBy, values } from 'lodash'
import Spinner from '../loading/components/Spinner'
import { GenericError, ObjectNotFound } from '../error-pages'

const query = gql`
  query userInfo($id: ID!) {
    user(id: $id) {
      id
      name
      onCallSteps {
        id
        stepNumber
        escalationPolicy {
          id
          name
          assignedTo {
            id
            name
          }
        }
      }
    }
  }
`
interface QueryResult {
  user: {
    id: string
    name: string
    onCallSteps: OnCallStep[]
  }
}
interface OnCallStep {
  id: string
  stepNumber: number
  escalationPolicy: {
    id: string
    name: string
    assignedTo: {
      id: string
      name: string
    }[]
  }
}

interface Service {
  id: string
  name: string
  policyName: string
  policySteps: number[]
}

// services returns a sorted list of services
// with policy information mapped in the following structure:
// {id: 'svc id', name: 'svc name', policyName: 'policy name', policySteps: [0,1,2]}
//
function services(onCallSteps: OnCallStep[] = []): Service[] {
  const svcs: { [index: string]: Service } = {}
  ;(onCallSteps || []).forEach((s) =>
    (s.escalationPolicy.assignedTo || []).forEach((svc) => {
      if (!svcs[svc.id]) {
        svcs[svc.id] = {
          id: svc.id,
          name: svc.name,
          policyName: s.escalationPolicy.name,
          policySteps: [s.stepNumber],
        }
      } else {
        svcs[svc.id].policySteps.push(s.stepNumber)
      }
    }),
  )

  let result = values(svcs)
  result = sortBy(result, 'name')

  return result
}

export default function UserOnCallAssignmentList(props: {
  userID: string
  currentUser?: boolean
}): React.ReactNode {
  const userID = props.userID
  const { data, loading, error } = useQuery(query, {
    variables: { id: userID },
  })

  if (!data && loading) {
    return <Spinner />
  }
  if (error) {
    return <GenericError error={error.message} />
  }
  if (!data?.user) {
    return <ObjectNotFound />
  }

  const user = (data as QueryResult).user

  return (
    <Card>
      <FlatList
        headerNote={
          props.currentUser
            ? 'Showing your current on-call assignments.'
            : `Showing current on-call assignments for ${user.name}`
        }
        emptyMessage={
          props.currentUser
            ? 'You are not currently on-call.'
            : `${user.name} is not currently on-call.`
        }
        items={services(user.onCallSteps).map((svc) => ({
          title: svc.name,
          url: '/services/' + svc.id,
          subText: `${svc.policyName}: ${svc.policySteps
            .map((n) => `Step ${n + 1}`)
            .join(', ')}`,
        }))}
      />
    </Card>
  )
}
