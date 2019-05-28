import React from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import Query from '../util/Query'
import { Card } from '@material-ui/core'
import FlatList from '../lists/FlatList'
import { sortBy, values } from 'lodash-es'

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
// services returns a sorted list of services
// with policy information mapped in the following structure:
// {id: 'svc id', name: 'svc name', policyName: 'policy name', policySteps: [0,1,2]}
//
function services(onCallSteps = []) {
  const svcs = {}
  ;(onCallSteps || []).forEach(s =>
    (s.escalationPolicy.assignedTo || []).forEach(svc => {
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

export default class UserOnCallAssignmentList extends React.PureComponent {
  static propTypes = {
    userID: p.string.isRequired,
    currentUser: p.bool,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.userID }}
        render={({ data }) => this.renderData(data.user)}
      />
    )
  }

  renderData = user => {
    const current = this.props.currentUser
    return (
      <Card>
        <FlatList
          headerNote={
            current
              ? 'Showing your current on-call assignments.'
              : `Showing current on-call assignments for ${user.name}`
          }
          emptyMessage={
            current
              ? 'You are not currently on-call.'
              : `${user.name} is not currently on-call.`
          }
          items={services(user.onCallSteps).map(svc => ({
            title: svc.name,
            url: '/services/' + svc.id,
            subText: `${svc.policyName}: ${svc.policySteps
              .map(n => `Step ${n + 1}`)
              .join(', ')}`,
          }))}
        />
      </Card>
    )
  }
}
