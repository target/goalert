import React from 'react'
import p from 'prop-types'

import { graphql2Client } from '../apollo'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'

const createMutation = gql`
  mutation($input: CreateUserContactMethodInput!) {
    createUserContactMethod(input: $input) {
      id
    }
  }
`

export default class UserContactMethodCreateDialog extends React.PureComponent {
  static propTypes = {
    userID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    value: {
      name: '',
      type: 'SMS',
      value: '+1',
    },
    errors: [],
  }

  render() {
    return (
      <Mutation
        client={graphql2Client}
        mutation={createMutation}
        awaitRefetchQueries
        refetchQueries={['nrList', 'cmList']}
        onCompleted={this.props.onClose}
      >
        {(commit, status) => this.renderDialog(commit, status)}
      </Mutation>
    )
  }

  renderDialog(commit, status) {
    const { loading, error } = status

    const fieldErrs = fieldErrors(error)

    return (
      <FormDialog
        title='Create New Contact Method'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: {
                ...this.state.value,
                userID: this.props.userID,
                newUserNotificationRule: {
                  delayMinutes: 0,
                },
              },
            },
          })
        }}
        form={
          <UserContactMethodForm
            errors={fieldErrs}
            disabled={loading}
            value={this.state.value}
            onChange={value => this.setState({ value })}
          />
        }
      />
    )
  }
}
