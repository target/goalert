import React from 'react'
import p from 'prop-types'

import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import Query from '../util/Query'

import FormDialog from '../dialogs/FormDialog'
import ServiceForm from './ServiceForm'

const query = gql`
  query service($id: ID!) {
    service(id: $id) {
      id
      name
      description
      escalationPolicyID
      escalationPolicy {
        id
        name
      }
    }
  }
`
const mutation = gql`
  mutation updateService($input: UpdateServiceInput!) {
    updateService(input: $input)
  }
`

export default class ServiceEditDialog extends React.PureComponent {
  static propTypes = {
    serviceID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    value: null,
    errors: [],
  }

  renderQuery() {
    return (
      <Query
        query={query}
        variables={{ id: this.props.serviceID }}
        noPoll
        render={({ data }) => {
          const { id, name, description, escalationPolicyID } =
            data.service || {}
          return this.renderMutation({
            id,
            name,
            description,
            escalationPolicyID,
          })
        }}
      />
    )
  }

  renderMutation(defaultValue) {
    return (
      <Mutation
        mutation={mutation}
        onCompleted={this.props.onClose}
        refetchQueries={() => [
          {
            query,
            variables: { id: this.props.serviceID },
          },
        ]}
      >
        {(commit, status) => this.renderDialog(defaultValue, commit, status)}
      </Mutation>
    )
  }

  renderDialog(
    defaultValue = { name: '', description: '', escalationPolicyID: '' },
    commit,
    status,
  ) {
    const { loading, error } = status
    const fieldErrs = fieldErrors(error)

    return (
      <FormDialog
        title='Edit Service'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: this.state.value || defaultValue,
            },
          }).then(() => this.props.onClose())
        }}
        form={
          <ServiceForm
            epRequired
            errors={fieldErrs}
            disabled={loading}
            value={this.state.value || defaultValue}
            onChange={value => this.setState({ value })}
          />
        }
      />
    )
  }

  render() {
    return this.renderQuery()
  }
}
