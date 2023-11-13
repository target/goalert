import React, { useState } from 'react'
import { useMutation } from '@apollo/client'
import { gql } from 'urql'
import { nonFieldErrors } from '../util/errutil'
import UserForm from './UserForm'
import FormDialog from '../dialogs/FormDialog'

interface RotationAddUserDialogProps {
  rotationID: string
  userIDs: string[]
  onClose: () => void
}

interface Value {
  users: string[]
}

const mutation = gql`
  mutation ($input: UpdateRotationInput!) {
    updateRotation(input: $input)
  }
`

const RotationAddUserDialog = (
  props: RotationAddUserDialogProps,
): React.ReactNode => {
  const { rotationID, userIDs, onClose } = props
  const [value, setValue] = useState<Value>({
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

  return (
    <FormDialog
      title='Add User'
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={onClose}
      onSubmit={() => updateRotationMutation()}
      form={
        <UserForm
          errors={nonFieldErrors(error)}
          disabled={loading}
          value={value}
          onChange={(value: Value) => setValue(value)}
        />
      }
    />
  )
}

export default RotationAddUserDialog
