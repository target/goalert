import React, { PureComponent } from 'react'
import { PropTypes as p } from 'prop-types'
import Query from '../util/Query'
import gql from 'graphql-tag'
import PolicyStepsCard from './PolicyStepsCard'

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

export default class PolicyStepsQuery extends PureComponent {
  static propTypes = {
    escalationPolicyID: p.string.isRequired,
  }

  render() {
    return (
      <Query
        query={policyStepsQuery}
        render={({ data }) => {
          return (
            <PolicyStepsCard
              escalationPolicyID={this.props.escalationPolicyID}
              repeat={data.escalationPolicy.repeat}
              steps={data.escalationPolicy.steps || []}
            />
          )
        }}
        variables={{ id: this.props.escalationPolicyID }}
      />
    )
  }
}
