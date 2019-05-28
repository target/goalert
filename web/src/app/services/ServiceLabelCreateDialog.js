import React from 'react'
import p from 'prop-types'

import { graphql2Client } from '../apollo'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import ServiceLabelForm from './ServiceLabelForm'

const mutation = gql`
  mutation($input: SetLabelInput!) {
    setLabel(input: $input)
  }
`
const query = gql`
  query($serviceID: ID!) {
    service(id: $serviceID) {
      id
      labels {
        key
        value
      }
    }
  }
`

export default class ServiceLabelCreateDialog extends React.PureComponent {
  static propTypes = {
    serviceID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    value: { key: '', value: '' },
    errors: [],
  }

  renderMutation() {
    return (
      <Mutation
        client={graphql2Client}
        mutation={mutation}
        onCompleted={this.props.onClose}
        update={cache => {
          const { service } = cache.readQuery({
            query,
            variables: { serviceID: this.props.serviceID },
          })
          const labels = (service.labels || []).filter(
            l => l.key !== this.state.value.key,
          )
          if (this.state.value.value) {
            labels.push({ ...this.state.value, __typename: 'Label' })
          }
          cache.writeData({
            query,
            variables: { serviceID: this.props.serviceID },
            data: {
              service: {
                ...service,
                labels,
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
        title='Set Label Value'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: {
                ...this.state.value,
                target: { type: 'service', id: this.props.serviceID },
              },
            },
          })
        }}
        form={
          <ServiceLabelForm
            errors={fieldErrors(error)}
            disabled={loading}
            value={this.state.value}
            onChange={value => this.setState({ value })}
          />
        }
      />
    )
  }

  render() {
    return this.renderMutation()
  }
}
