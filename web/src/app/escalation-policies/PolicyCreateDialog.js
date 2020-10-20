import { gql } from '@apollo/client'
import React, { PureComponent } from 'react'
import p from 'prop-types'
import { Mutation } from '@apollo/client/react/components'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import { Redirect } from 'react-router-dom'
import FormDialog from '../dialogs/FormDialog'
import PolicyForm from './PolicyForm'

const mutation = gql`
  mutation($input: CreateEscalationPolicyInput!) {
    createEscalationPolicy(input: $input) {
      id
    }
  }
`

export default class PolicyCreateDialog extends PureComponent {
  static propTypes = {
    onClose: p.func,
  }

  state = {
    value: null,
    errors: [],
  }

  renderDialog(commit, status) {
    const { loading, data, error } = status
    const { value } = this.state

    if (data && data.createEscalationPolicy) {
      return (
        <Redirect
          push
          to={`/escalation-policies/${data.createEscalationPolicy.id}`}
        />
      )
    }

    const fieldErrs = fieldErrors(error)
    const defaultValue = {
      name: '',
      description: '',
      repeat: { label: '3', value: '3' },
    }

    return (
      <FormDialog
        title='Create Escalation Policy'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: {
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
            onChange={(value) => this.setState({ value })}
          />
        }
      />
    )
  }

  render() {
    return (
      <Mutation mutation={mutation} onCompleted={this.props.onClose}>
        {(commit, status) => this.renderDialog(commit, status)}
      </Mutation>
    )
  }
}
