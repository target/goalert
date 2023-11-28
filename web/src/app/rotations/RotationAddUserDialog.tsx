import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
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
): JSX.Element => {
  const { rotationID, userIDs, onClose } = props
  const [value, setValue] = useState<Value>({
    users: [],
  })
  // append to users array from selected users
  const users: string[] = []
  const uIDs = value.users
  userIDs.forEach((u) => users.push(u))
  uIDs.forEach((u) => users.push(u))

  const [{ error }, commit] = useMutation(mutation)

  return (
    <FormDialog
      title='Add User'
      errors={nonFieldErrors(error)}
      onClose={onClose}
      onSubmit={() =>
        commit({
          input: {
            id: rotationID,
            userIDs: users,
          },
        }).then(() => onClose())
      }
      form={
        <UserForm
          errors={nonFieldErrors(error)}
          value={value}
          onChange={(value: Value) => setValue(value)}
        />
      }
    />
  )
}

export default RotationAddUserDialog
