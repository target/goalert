import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import p from 'prop-types'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import UserForm from './UserForm'
import FormDialog from '../dialogs/FormDialog'

const mutation = gql`
  mutation($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`

const RotationAddUserDialog = (props) => {
  const { rotationID, userIDs, onClose } = props
  const [value, setValue] = useState({
    users: [],
  })
  // append to users array from selected users
  const users = []
  const uIDs = value.users
  userIDs.forEach((u) => users.push(u))
  uIDs.forEach((u) => users.push(u))

  const [updateRotationMutation, { loading, error }] = useMutation(mutation, {
    variables: {
      input: {
        id: rotationID,
        userIDs: users,
      },
    },
    onCompleted: onClose,
  })

  return (
    <FormDialog
      title='Add User'
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={onClose}
      onSubmit={() => updateRotationMutation()}
      form={
        <UserForm
          errors={fieldErrors(error)}
          disabled={loading}
          value={value}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}

RotationAddUserDialog.propTypes = {
  rotationID: p.string.isRequired,
  userIDs: p.arrayOf(p.string).isRequired,
  onClose: p.func.isRequired,
}

export default RotationAddUserDialog
