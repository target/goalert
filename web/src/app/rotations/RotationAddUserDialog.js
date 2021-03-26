import React, { useState } from 'react'
import { gql } from '@apollo/client'
import p from 'prop-types'
import { Mutation } from '@apollo/client/react/components'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import UserForm from './UserForm'
import FormDialog from '../dialogs/FormDialog'

const mutation = gql`
  mutation($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`

const RotationAddUserDialog = (props) => {
  const [value, setValue] = useState(null)

  const defaultValue = {
    users: [],
  }

  const renderDialog = (defaultValue, commit, status) => {
    const { loading, error } = status
    const fieldErrs = fieldErrors(error)

    // append to users array from selected users
    const users = []
    const userIDs = (value && value.users) || defaultValue.users

    props.userIDs.forEach((u) => users.push(u))
    userIDs.forEach((u) => users.push(u))

    return (
      <FormDialog
        title='Add User'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={props.onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: {
                id: props.rotationID,
                userIDs: users,
              },
            },
          }).then(() => props.onClose())
        }}
        form={
          <UserForm
            errors={fieldErrs}
            disabled={loading}
            value={value || defaultValue}
            onChange={(value) => setValue(value)}
          />
        }
      />
    )
  }

  return (
    <Mutation mutation={mutation}>
      {(commit, status) => renderDialog(defaultValue, commit, status)}
    </Mutation>
  )
}

RotationAddUserDialog.propTypes = {
  rotationID: p.string.isRequired,
  userIDs: p.array.isRequired,
  onClose: p.func.isRequired,
}

export default RotationAddUserDialog
