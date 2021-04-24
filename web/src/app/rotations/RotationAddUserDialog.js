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
  const { rotationID, onClose } = props
  const [value, setValue] = useState(null)
  const defaultValue = {
    users: [],
  }
  // append to users array from selected users
  const users = []
  const userIDs = (value && value.users) || defaultValue.users
  props.userIDs.forEach((u) => users.push(u))
  userIDs.forEach((u) => users.push(u))

  const [updateRotationMutation, updateRotationMutationStatus] = useMutation(
    mutation,
    {
      variables: {
        input: {
          id: rotationID,
          userIDs: users,
        },
      },
      onCompleted: onClose,
    },
  )

  return (
    <FormDialog
      title='Add User'
      loading={updateRotationMutationStatus.loading}
      errors={nonFieldErrors(updateRotationMutationStatus.error)}
      onClose={onClose}
      onSubmit={() => updateRotationMutation()}
      form={
        <UserForm
          errors={fieldErrors(updateRotationMutationStatus.error)}
          disabled={updateRotationMutationStatus.loading}
          value={value || defaultValue}
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
