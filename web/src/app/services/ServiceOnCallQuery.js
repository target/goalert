import React from 'react'
import p from 'prop-types'
import ServiceOnCallList from './ServiceOnCallList'
import gql from 'graphql-tag'
import Query from '../util/Query'

const query = gql`
  query onCallQuery($id: ID!) {
    service(id: $id) {
      id
      onCallUsers {
        userID
        userName
        stepNumber
      }
    }
  }
`

export default class ServiceOnCallQuery extends React.PureComponent {
  static propTypes = {
    serviceID: p.string.isRequired,
  }

  render() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.serviceID }}
        render={({ data }) => (
          <ServiceOnCallList items={data.service.onCallUsers} />
        )}
      />
    )
  }
}
