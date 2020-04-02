import React from 'react'
import p from 'prop-types'

import { Redirect } from 'react-router-dom'

import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import ServiceForm from './ServiceForm'

const createMutation = gql`
  mutation createService($input: CreateServiceInput!) {
    createService(input: $input) {
      id
      name
      description
      escalationPolicyID
    }
  }
`

const inputVars = ({ name, description, escalationPolicyID }, attempt = 0) => {
  const vars = {
    name,
    description,
    escalationPolicyID,
    favorite: true,
  }
  if (!vars.escalationPolicyID) {
    vars.newEscalationPolicy = {
      name: attempt ? `${name} Policy ${attempt}` : name + ' Policy',
      description: 'Auto-generated policy for ' + name,
      steps: [
        {
          delayMinutes: 5,
          targets: [
            {
              type: 'user',
              id: '__current_user',
            },
          ],
        },
      ],
    }
  }

  return vars
}

export default class ServiceCreateDialog extends React.PureComponent {
  static propTypes = {
    onClose: p.func,
  }

  state = {
    value: null,
    errors: [],
  }

  renderMutation(defaultValue) {
    return (
      <Mutation mutation={createMutation}>
        {(commit, status) => this.renderDialog(defaultValue, commit, status)}
      </Mutation>
    )
  }

  renderDialog(
    defaultValue = { name: '', description: '', escalationPolicyID: '' },
    commit,
    status,
  ) {
    const { loading, data, error } = status
    if (data && data.createService) {
      return <Redirect push to={`/services/${data.createService.id}`} />
    }

    const fieldErrs = fieldErrors(error).filter(
      (e) => !e.field.startsWith('newEscalationPolicy.'),
    )

    return (
      <FormDialog
        title='Create New Service'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          let n = 1
          const onErr = (err) => {
            // retry if it's a policy name conflict
            if (
              err.graphQLErrors &&
              err.graphQLErrors[0].extensions &&
              err.graphQLErrors[0].extensions.isFieldError &&
              err.graphQLErrors[0].extensions.fieldName ===
                'newEscalationPolicy.Name'
            ) {
              n++
              return commit({
                variables: {
                  input: inputVars(this.state.value || defaultValue, n),
                },
              }).then(null, onErr)
            }
          }

          return commit({
            variables: {
              input: inputVars(this.state.value || defaultValue),
            },
          }).then(null, onErr)
        }}
        form={
          <ServiceForm
            errors={fieldErrs}
            disabled={loading}
            value={this.state.value || defaultValue}
            onChange={(value) => this.setState({ value })}
          />
        }
      />
    )
  }

  render() {
    return this.renderMutation()
  }
}
