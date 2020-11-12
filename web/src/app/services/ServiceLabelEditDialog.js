import { gql } from '@apollo/client'
import React from 'react'

import p from 'prop-types'
import { Mutation } from '@apollo/client/react/components'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import Query from '../util/Query'

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
    labelKey: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    value: null,
    errors: [],
  }

  render() {
    return this.renderQuery()
  }

  renderQuery() {
    return (
      <Query
        noPoll
        query={query}
        variables={{ serviceID: this.props.serviceID }}
        render={({ data }) =>
          this.renderMutation(
            data.service.labels.find((l) => l.key === this.props.labelKey),
          )
        }
      />
    )
  }

  renderMutation(data) {
    return (
      <Mutation
        mutation={mutation}
        onCompleted={this.props.onClose}
        update={(cache) => {
          const { service } = cache.readQuery({
            query,
            variables: { serviceID: this.props.serviceID },
          })
          const labels = (service.labels || []).filter(
            (l) => l.key !== this.state.value.key,
          )
          if (this.state.value.value) {
            labels.push({ ...this.state.value, __typename: 'Label' })
          }
          cache.writeQuery({
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
        {(commit, status) => this.renderDialog(data, commit, status)}
      </Mutation>
    )
  }

  renderDialog(data, commit, status) {
    const { loading, error } = status
    return (
      <FormDialog
        title='Update Label Value'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          if (!this.state.value) {
            return this.props.onClose()
          }
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
            editValueOnly
            disabled={loading}
            value={this.state.value || { key: data.key, value: data.value }}
            onChange={(value) => this.setState({ value })}
          />
        }
      />
    )
  }
}
