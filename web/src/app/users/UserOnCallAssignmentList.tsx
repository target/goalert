import React from 'react'
import { gql, useQuery } from 'urql'
import { Card } from '@mui/material'
import FlatList from '../lists/FlatList'
import { sortBy, values } from 'lodash'
import { GenericError, ObjectNotFound } from '../error-pages'
import { OnCallServiceAssignment, User } from '../../schema'

const query = gql`
  query userInfo($id: ID!) {
    user(id: $id) {
      id
      name
      onCallOverview {
        serviceAssignments {
          serviceID
          serviceName
          escalationPolicyName
          stepNumber
        }
      }
    }
  }
`

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
function services(onCallSteps: OnCallServiceAssignment[] = []): Service[] {
  const svcs: { [index: string]: Service } = {}
  ;(onCallSteps || []).forEach((s) => {
    if (!svcs[s.serviceID]) {
      svcs[s.serviceID] = {
        id: s.serviceID,
        name: s.serviceName,
        policyName: s.escalationPolicyName,
        policySteps: [s.stepNumber],
      }
    } else {
      svcs[s.serviceID].policySteps.push(s.stepNumber)
    }
  })

  let result = values(svcs)
  result = sortBy(result, 'name')

  return result
}

export default function UserOnCallAssignmentList(props: {
  userID: string
  currentUser?: boolean
}): React.JSX.Element {
  const userID = props.userID
  const [{ data, error }] = useQuery<{ user: User }>({
    query,
    variables: { id: userID },
  })

  if (error) {
    return <GenericError error={error.message} />
  }
  if (!data?.user) {
    return <ObjectNotFound />
  }

  const user = data.user

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
        items={services(user.onCallOverview.serviceAssignments).map((svc) => ({
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
