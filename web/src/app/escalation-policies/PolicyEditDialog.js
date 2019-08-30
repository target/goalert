import React, { PureComponent } from 'react'
import p from 'prop-types'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import Query from '../util/Query'
import FormDialog from '../dialogs/FormDialog'
import PolicyForm from './PolicyForm'

const query = gql`
  query($id: ID!) {
    escalationPolicy(id: $id) {
      id
      name
      description
      repeat
    }
  }
`

const mutation = gql`
  mutation($input: UpdateEscalationPolicyInput!) {
    updateEscalationPolicy(input: $input)
  }
`

export default class PolicyEditDialog extends PureComponent {
  static propTypes = {
    escalationPolicyID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    value: null,
    errors: [],
  }

  renderMutation = defaultValue => {
    return (
      <Mutation
        mutation={mutation}
        onCompleted={this.props.onClose}
        refetchQueries={() => [
          {
            query,
            variables: { id: this.props.escalationPolicyID },
          },
        ]}
      >
        {(commit, status) => this.renderDialog(defaultValue, commit, status)}
      </Mutation>
    )
  }

  renderDialog(defaultValue, commit, status) {
    const { loading, error } = status
    const { value } = this.state
    const fieldErrs = fieldErrors(error)

    return (
      <FormDialog
        title='Edit Escalation Policy'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: {
                id: this.props.escalationPolicyID,
                name: (value && value.name) || defaultValue.name,
                description:
                  (value && value.description) || defaultValue.description,
                repeat:
                  (value && value.repeat.value) || defaultValue.repeat.value,
              },
            },
          })
        }}
        form={
          <PolicyForm
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
    return (
      <Query
        query={query}
        variables={{ id: this.props.escalationPolicyID }}
        noPoll
        render={({ data }) => {
          const { id, name, description, repeat } = data.escalationPolicy || {}
          return this.renderMutation({
            id,
            name,
            description,
            repeat: { label: repeat.toString(), value: repeat.toString() },
          })
        }}
      />
    )
  }
}
