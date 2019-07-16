import React from 'react'
import p from 'prop-types'

import { graphql2Client } from '../apollo'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import HearbeatForm from './HeartbeatForm'

const mutation = gql`
  mutation($input: CreateHeartbeatInput!) {
    createHeartbeat(input: $input) {
      id
      serviceID
      name
      heartbeatInterval
      lastState
    }
  }
`
const query = gql`
  query($serviceID: ID!) {
    service(id: $serviceID) {
      id
      heartbeats {
        id
        name
        heartbeatInterval
        lastState
      }
    }
  }
`

export default class HeartbeatCreateDialog extends React.PureComponent {
  static propTypes = {
    serviceID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    value: { name: '' },
    errors: [],
  }

  render() {
    return (
      <Mutation
        client={graphql2Client}
        mutation={mutation}
        onCompleted={this.props.onClose}
        update={(cache, { data: { createIntegrationKey } }) => {
          const { service } = cache.readQuery({
            query,
            variables: { serviceID: this.props.serviceID },
          })
          cache.writeData({
            query,
            variables: { serviceID: this.props.serviceID },
            data: {
              service: {
                ...service,
                integrationKeys: (service.integrationKeys || []).concat(
                  createIntegrationKey,
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

  renderDialog(commit, status) {
    const { loading, error } = status
    return (
      <FormDialog
        maxWidth='sm'
        title='Create New Heartbeat'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: { ...this.state.value, serviceID: this.props.serviceID },
            },
          })
        }}
        form={
          <HearbeatForm
            errors={fieldErrors(error)}
            disabled={loading}
            value={this.state.value}
            onChange={value => this.setState({ value })}
          />
        }
      />
    )
  }
}
