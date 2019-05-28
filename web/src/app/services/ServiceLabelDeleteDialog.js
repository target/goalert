import React from 'react'
import p from 'prop-types'

import { graphql2Client } from '../apollo'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'

const updateQuery = gql`
  query($id: ID!) {
    service(id: $id) {
      id
      labels {
        key
        value
      }
    }
  }
`

const mutation = gql`
  mutation($input: SetLabelInput!) {
    setLabel(input: $input)
  }
`

export default class ServiceLabelDeleteDialog extends React.PureComponent {
  static propTypes = {
    serviceID: p.string.isRequired,
    labelKey: p.string.isRequired,
    onClose: p.func,
  }

  renderMutation() {
    return (
      <Mutation
        client={graphql2Client}
        mutation={mutation}
        onCompleted={this.props.onClose}
        update={cache => {
          const { service } = cache.readQuery({
            query: updateQuery,
            variables: { id: this.props.serviceID },
          })
          cache.writeQuery({
            query: updateQuery,
            variables: { id: this.props.serviceID },
            data: {
              service: {
                ...service,
                labels: (service.labels || []).filter(
                  l => l.key !== this.props.labelKey,
                ),
              },
            },
          })
        }}
      >
        {(commit, status) => this.renderDialog(commit, status)}
      </Mutation>
    )
  }

  renderDialog(commit, mutStatus) {
    const { loading, error } = mutStatus

    return (
      <FormDialog
        title='Are you sure?'
        confirm
        subTitle={`This will delete the label: ${this.props.labelKey}`}
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          const input = {
            key: this.props.labelKey,
            value: '',
            target: {
              type: 'service',
              id: this.props.serviceID,
            },
          }
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
    return this.renderMutation()
  }
}
