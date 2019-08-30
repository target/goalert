import React from 'react'
import { PropTypes as p } from 'prop-types'
import gql from 'graphql-tag'
import { Mutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import UserForm from './UserForm'
import FormDialog from '../dialogs/FormDialog'

const mutation = gql`
  mutation($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`

export default class RotationAddUserDialog extends React.Component {
  static propTypes = {
    rotationID: p.string.isRequired,
    userIDs: p.array.isRequired,
    onClose: p.func.isRequired,
  }

  state = {
    value: null,
    errors: [],
  }

  render() {
    const defaultValue = {
      users: [],
    }

    return (
      <Mutation mutation={mutation} refetchQueries={() => ['rotationUsers']}>
        {(commit, status) => this.renderDialog(defaultValue, commit, status)}
      </Mutation>
    )
  }

  renderDialog(defaultValue, commit, status) {
    const { value } = this.state
    const { loading, error } = status
    const fieldErrs = fieldErrors(error)

    // append to users array from selected users
    let users = []
    const userIDs = (value && value.users) || defaultValue.users

    this.props.userIDs.forEach(u => users.push(u))
    userIDs.forEach(u => users.push(u))
    return (
      <FormDialog
        title='Add User'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={this.props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: {
                id: this.props.rotationID,
                userIDs: users,
              },
            },
          }).then(() => this.props.onClose())
        }}
        form={
          <UserForm
            errors={fieldErrs}
            disabled={loading}
            value={this.state.value || defaultValue}
            onChange={value => this.setState({ value })}
          />
        }
      />
    )
  }
}
