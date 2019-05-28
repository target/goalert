import React from 'react'
import p from 'prop-types'

import { graphql2Client } from '../apollo'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import UserNotificationRuleForm from './UserNotificationRuleForm'

const createMutation = gql`
  mutation($input: CreateUserNotificationRuleInput!) {
    createUserNotificationRule(input: $input) {
      id
    }
  }
`

export default class UserNotificationRuleCreateDialog extends React.PureComponent {
  static propTypes = {
    userID: p.string.isRequired,
    onClose: p.func,
  }

  state = {
    value: {
      contactMethodID: '',
      delayMinutes: '0',
    },
    errors: [],
  }

  render() {
    return (
      <Mutation
        client={graphql2Client}
        mutation={createMutation}
        awaitRefetchQueries
        refetchQueries={['nrList']}
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
        title='Create New Notification Rule'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: { ...this.state.value, userID: this.props.userID },
            },
          })
        }}
        form={
          <UserNotificationRuleForm
            userID={this.props.userID}
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
