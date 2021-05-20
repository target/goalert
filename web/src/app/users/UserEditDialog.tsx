import React from 'react'
import { gql, useMutation } from '@apollo/client'
import { Checkbox, FormControlLabel } from '@material-ui/core'
import Spinner from '../loading/components/Spinner'
import FormDialog from '../dialogs/FormDialog'
import { useSessionInfo } from '../util/RequireConfig'
import { nonFieldErrors } from '../util/errutil'

const mutation = gql`
  mutation ($input: UpdateUserInput!) {
    updateUser(input: $input)
  }
`

interface UserEditDialogProps {
  userID: string
  role: string
  onClose: () => void
}

function UserEditDialog(props: UserEditDialogProps): JSX.Element {
  const { ready: isSessionReady, userID: currentUserID } = useSessionInfo()

  const [adminChecked, setAdminChecked] = React.useState(props.role === 'admin')

  const [editUser, editUserStatus] = useMutation(mutation, {
    onCompleted: props.onClose,
  })

  function handleChange(e: React.ChangeEvent<HTMLInputElement>): void {
    setAdminChecked(e.target.checked)
  }

  if (!isSessionReady) return <Spinner />

  return (
    <FormDialog
      title='Edit User Role'
      confirm
      errors={nonFieldErrors(editUserStatus.error)}
      onClose={props.onClose}
      onSubmit={() =>
        editUser({
          variables: {
            input: {
              id: props.userID,
              role: adminChecked ? 'admin' : 'user',
            },
          },
        })
      }
      notices={
        props.role === 'admin' &&
        adminChecked === false &&
        props.userID === currentUserID
          ? [
              {
                type: 'WARNING',
                message: 'Updating role to User',
                details:
                  'If you remove your admin privileges you will need to log in as a different admin to restore them.',
              },
            ]
          : []
      }
      form={
        <FormControlLabel
          label='Admin'
          control={
            <Checkbox
              checked={adminChecked}
              onChange={handleChange}
              name='isAdmin'
            />
          }
        />
      }
    />
  )
}

export default UserEditDialog
