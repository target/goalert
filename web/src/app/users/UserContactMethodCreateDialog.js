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
    cmCreated: false,
  }

  // cmCreated is false by default until verification is complete
  onComplete() {
    if (!this.state.cmCreated) {
      this.setState({ cmCreated: true })
    } else {
      this.props.onClose()
    }
  }
  render() {
    return (
      <Mutation
        client={graphql2Client}
        mutation={createMutation}
        awaitRefetchQueries
        refetchQueries={['nrList', 'cmList']}
        onCompleted={this.onComplete}
      >
        {(commit, status) => this.renderDialog(commit, status)}
      </Mutation>
    )
  }

  renderDialog(commit, status) {
    const { loading, error } = status

    return (
      <FormDialog
        title='Create New Contact Method'
        subtitle={
          this.state.cmCreated ? 'Verify contact menthod to continue' : null
        }
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
        form={this.renderForm(status)}
      />
    )
  }

  renderForm(status) {
    const { loading, error } = status
    const fieldErrs = fieldErrors(error)

    if (this.state.cmCreated) {
      // return
      return console.log('submitted')
    } else {
      return (
        <UserContactMethodForm
          errors={fieldErrs}
          disabled={loading}
          value={this.state.value}
          onChange={value => this.setState({ value })}
        />
      )
    }
  }
}
