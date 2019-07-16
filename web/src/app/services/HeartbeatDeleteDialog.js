import React from 'react'
import p from 'prop-types'

import { graphql2Client } from '../apollo'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { nonFieldErrors } from '../util/errutil'
import Query from '../util/Query'

import FormDialog from '../dialogs/FormDialog'

const query = gql`
  query($id: ID!) {
    heartbeat(id: $id) {
      id
      name
    }
  }
`
const updateQuery = gql`
  query($id: ID!) {
    service(id: $id) {
      id
      heartbeats {
        id
        name
      }
    }
  }
`

const mutation = gql`
  mutation($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default class HearbeatDeleteDialog extends React.PureComponent {
  static propTypes = {
    heartbeatID: p.number.isRequired,
    onClose: p.func,
  }

  renderQuery() {
    return (
      <Query
        noPoll
        query={query}
        variables={{ id: this.props.heartbeatID }}
        render={({ data }) => this.renderMutation(data.heartbeat)}
      />
    )
  }

  renderMutation(data) {
    return (
      <Mutation
        client={graphql2Client}
        mutation={mutation}
        onCompleted={this.props.onClose}
        update={cache => {
          const { service } = cache.readQuery({
            query: updateQuery,
            variables: { id: data.serviceID },
          })
          cache.writeQuery({
            query: updateQuery,
            variables: { id: data.serviceID },
            data: {
              service: {
                ...service,
                heartbeats: (service.heartbeats || []).filter(
                  beat => beat.id !== this.props.heartbeatID,
                ),
              },
            },
          })
        }}
      >
        {(commit, status) => this.renderDialog(data, commit, status)}
      </Mutation>
    )
  }

  renderDialog(data, commit, mutStatus) {
    const { loading, error } = mutStatus

    return (
      <FormDialog
        title='Are you sure?'
        confirm
        subTitle={`This will delete the heartbeat: ${data.name}`}
        caption=''
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          const input = [
            {
              type: 'hearbeat',
              id: this.props.hearbeatID,
            },
          ]
          return commit({
            variables: {
              input,
            },
          })
        }}
      />
    )
  }

  render() {
    return this.renderQuery()
  }
}
