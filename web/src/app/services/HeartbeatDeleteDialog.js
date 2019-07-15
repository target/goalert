import React from 'react'
import p from 'prop-types'

// import { graphql2Client } from '../apollo'
// import gql from 'graphql-tag'
// import { Mutation } from 'react-apollo'
import { nonFieldErrors } from '../util/errutil'
// import Query from '../util/Query'

import FormDialog from '../dialogs/FormDialog'

/* const query = gql`
  query($id: ID!) {
    integrationKey(id: $id) {
      id
      name
      serviceID
    }
  }
`
const updateQuery = gql`
  query($id: ID!) {
    service(id: $id) {
      id
      integrationKeys {
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
` */

export default class HearbeatDeleteDialog extends React.PureComponent {
  static propTypes = {
    integrationKeyID: p.string.isRequired,
    onClose: p.func,
  }

  /* renderQuery() {
    return (
      <Query
        noPoll
        query={query}
        variables={{ id: this.props.heartbeatID }}
        render={({ data }) => this.renderMutation(data.heartbeat)}
      />
    )
  } */

  /* renderMutation(data) {
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
                integrationKeys: (service.integrationKeys || []).filter(
                  key => key.id !== this.props.integrationKeyID,
                ),
              },
            },
          })
        }}
      >
        {(commit, status) => this.renderDialog(data, commit, status)}
      </Mutation>
    )
  } */

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
