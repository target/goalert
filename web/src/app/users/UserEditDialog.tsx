import React, { useState } from 'react'
import { ApolloError, gql, useMutation } from '@apollo/client'
import Spinner from '../loading/components/Spinner'
import FormDialog from '../dialogs/FormDialog'
import { useSessionInfo } from '../util/RequireConfig'
import { FieldError, fieldErrors, nonFieldErrors } from '../util/errutil'
import UserEditForm, { Value } from './UserEditForm'

const updateUserInput = gql`
  mutation ($input: UpdateUserInput!) {
    updateUser(input: $input)
  }
`

const updateUserPassword = gql`
  mutation ($input: UpdateUserPassword!) {
    updateUserPassword(input: $input)
  }
`

interface UserEditDialogProps {
  userID: string
  role: string
  onClose: () => void
}

function UserEditDialog(props: UserEditDialogProps): JSX.Element {
  const defaultValue: Value = {
    oldPassword: '',
    newPassword: '',
    confirmNewPassword: '',
    isAdmin: props.role === 'admin',
  }

  const { ready: isSessionReady, userID: currentUserID } = useSessionInfo()

  const [value, setValue] = useState(defaultValue)
  const [errors, setErrors] = useState<FieldError[]>([])

  const [editUser, editUserStatus] = useMutation(updateUserInput, {
    variables: {
      input: {
        id: props.userID,
        role: value.isAdmin ? 'admin' : 'user',
      },
    },
  })

  const [resetPassword, resetPasswordStatus] = useMutation(updateUserPassword, {
    variables: {
      input: {
        id: props.userID,
        oldPassword: value.oldPassword,
        newPassword: value.newPassword,
      },
    },
  })

  // Checks if any of the password fields are used. Used to skip any unnecessary updateUserPassword mutation
  function passwordChanged(): boolean {
    return !(
      value.oldPassword === '' &&
      value.newPassword === '' &&
      value.confirmNewPassword === ''
    )
  }

  // Ensures that if one password field is used, then the others are also populated.
  // Also validates that there are no errors.
  function canProceed(err: FieldError[]): boolean {
    if (
      passwordChanged() &&
      (value.oldPassword === '' ||
        value.newPassword === '' ||
        value.confirmNewPassword === '')
    ) {
      return false
    }

    return !err?.length
  }

  // Validates inputs to the newPassword and confirmNewPassword fields
  function handleValidation(): FieldError[] {
    let err: FieldError[] = []
    if (value.newPassword.length < 8 || value.newPassword.length > 20) {
      err = [
        ...err,
        {
          field: 'newPassword',
          message: 'Password length must be between 8 - 20',
        } as FieldError,
      ]
    }
    if (value.newPassword !== value.confirmNewPassword) {
      err = [
        ...err,
        {
          field: 'confirmNewPassword',
          message: 'Passwords do not match',
        } as FieldError,
      ]
    }
    return err
  }

  // wrapper function to handle errors caught while executing useMutation Promises
  function errorHandler(
    fname: string,
    caughtError: unknown,
    error: ApolloError | undefined,
    errorList: FieldError[],
  ): FieldError[] {
    if (error) {
      errorList = [...errorList, ...fieldErrors(error)]
    }
    if (caughtError instanceof Error) {
      errorList = [
        ...errorList,
        { field: 'oldPassword', message: caughtError.message } as FieldError,
      ]
    } else {
      console.error(caughtError)
    }
    return errorList
  }

  // async wrapper function in order to await for useMutation Promises
  async function submitHandler(): Promise<void> {
    let errorList: FieldError[] = []

    if (passwordChanged()) {
      errorList = [...errorList, ...handleValidation()]
      if (errorList.length === 0) {
        try {
          await resetPassword()
        } catch (err) {
          errorList = errorHandler(
            'oldPassword',
            err,
            resetPasswordStatus.error,
            errorList,
          )
        }
      }
    }
    if (defaultValue.isAdmin !== value.isAdmin) {
      try {
        await editUser()
      } catch (err) {
        errorList = errorHandler(
          'isAdmin',
          err,
          editUserStatus.error,
          errorList,
        )
      }
    }
    setErrors(errorList)
    if (canProceed(errorList)) {
      props.onClose()
    }
  }

  if (!isSessionReady) return <Spinner />

  return (
    <FormDialog
      title='Edit User Info'
      loading={resetPasswordStatus.loading || editUserStatus.loading}
      errors={[
        ...nonFieldErrors(resetPasswordStatus.error),
        ...nonFieldErrors(editUserStatus.error),
      ]}
      onClose={props.onClose}
      onSubmit={submitHandler}
      notices={
        props.role === 'admin' && props.userID === currentUserID
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
        <UserEditForm
          value={value}
          errors={errors}
          onChange={(value) => {
            setValue(value)
          }}
        />
      }
    />
  )
}

export default UserEditDialog
