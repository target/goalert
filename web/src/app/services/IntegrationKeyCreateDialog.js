import { gql } from '@apollo/client'
import React from 'react'

import p from 'prop-types'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import IntegrationKeyForm from './IntegrationKeyForm'

const mutation = gql`
  mutation($input: CreateIntegrationKeyInput!) {
    createIntegrationKey(input: $input) {
      id
      name
      type
      href
    }
  }
`
const query = gql`
  query($serviceID: ID!) {
    service(id: $serviceID) {
      id
      integrationKeys {
        id
        name
        type
        href
      }
    }
  }
`

export default class IntegrationKeyCreateDialog extends React.PureComponent {
  static propTypes = {
    serviceID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    value: { name: '', type: 'generic' },
    errors: [],
  }

  render() {
    return (
      <Mutation
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
        title='Create New Integration Key'
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
          <IntegrationKeyForm
            errors={fieldErrors(error)}
            disabled={loading}
            value={this.state.value}
            onChange={(value) => this.setState({ value })}
          />
        }
      />
    )
  }
}
