import React, { PureComponent } from 'react'
import { PropTypes as p } from 'prop-types'
import Query from '../util/Query'
import gql from 'graphql-tag'
import PolicyServicesCard from './PolicyServicesCard'

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

export default class PolicyServicesQuery extends PureComponent {
  static propTypes = {
    escalationPolicyID: p.string.isRequired,
  }

  render() {
    return (
      <Query
        query={query}
        render={({ data }) => (
          <PolicyServicesCard
            services={data.escalationPolicy.assignedTo || []}
          />
        )}
        variables={{ id: this.props.escalationPolicyID }}
      />
    )
  }
}
