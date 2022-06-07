import React, { useState } from 'react'
import { useMutation } from '@apollo/client'
import { gql } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import UserForm from './UserForm'
import FormDialog from '../dialogs/FormDialog'
import { GenericError } from '../error-pages'

interface RotationAddUserDialogProps {
  rotationID: string
  userIDs: string[]
  onClose: () => void
}

const mutation = gql`
  mutation ($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`

const RotationAddUserDialog = (
  props: RotationAddUserDialogProps,
): JSX.Element => {
  const { rotationID, userIDs, onClose } = props
  const [value, setValue] = useState({
    users: [],
  })
  // append to users array from selected users
  const users: string[] = []
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

  if (error) {
    return <GenericError error={error.message} />
  }

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
          onChange={(value: React.SetStateAction<{ users: never[] }>) =>
            setValue(value)
          }
        />
      }
    />
  )
}

export default RotationAddUserDialog
