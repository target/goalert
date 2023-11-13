import React, { useEffect, useState } from 'react'
import { ApolloError, gql, useMutation, useQuery } from '@apollo/client'
import Spinner from '../loading/components/Spinner'
import FormDialog from '../dialogs/FormDialog'
import { useConfigValue, useSessionInfo } from '../util/RequireConfig'
import { FieldError, fieldErrors, nonFieldErrors } from '../util/errutil'
import UserEditForm, { Value } from './UserEditForm'
import { Query } from '../../schema'

const userAuthQuery = gql`
  query ($id: ID!) {
    user(id: $id) {
      authSubjects {
        providerID
        subjectID
      }
    }
  }
`

const updateUserMutation = gql`
  mutation ($input: UpdateUserInput!) {
    updateUser(input: $input)
  }
`

const updateBasicAuthMutation = gql`
  mutation ($input: UpdateBasicAuthInput!) {
    updateBasicAuth(input: $input)
  }
`

const createBasicAuthMutation = gql`
  mutation ($input: CreateBasicAuthInput!) {
    createBasicAuth(input: $input)
  }
`

interface UserEditDialogProps {
  userID: string
  role: string
  onClose: () => void
}

function UserEditDialog(props: UserEditDialogProps): React.ReactNode {
  const [authDisableBasic] = useConfigValue('Auth.DisableBasic')
  const defaultValue: Value = {
    username: '',
    oldPassword: '',
    password: '',
    confirmNewPassword: '',
    isAdmin: props.role === 'admin',
  }

  const {
    ready: isSessionReady,
    userID: currentUserID,
    isAdmin: currentUserAdmin,
  } = useSessionInfo()

  const [value, setValue] = useState(defaultValue)
  const [errors, setErrors] = useState<FieldError[]>([])

  const { loading, data } = useQuery<Pick<Query, 'user'>>(userAuthQuery, {
    variables: { id: props.userID },
  })
  useEffect(() => {
    if (!data?.user?.authSubjects) return

    const basicAuth = data.user.authSubjects.find(
      (s) => s.providerID === 'basic',
    )
    if (!basicAuth) return

    if (basicAuth.subjectID === value.username) return
    setValue({ ...value, username: basicAuth.subjectID })
  }, [data?.user?.authSubjects])

  const [createBasicAuth, createBasicAuthStatus] = useMutation(
    createBasicAuthMutation,
    {
      variables: {
        input: {
          userID: props.userID,
          username: value.username || null,
          password: value.password || null,
        },
      },
    },
  )

  const [editBasicAuth, editBasicAuthStatus] = useMutation(
    updateBasicAuthMutation,
    {
      variables: {
        input: {
          userID: props.userID,
          oldPassword: value.oldPassword || null,
          password: value.password || null,
        },
      },
    },
  )

  const [editUser, editUserStatus] = useMutation(updateUserMutation, {
    variables: {
      input: {
        id: props.userID,
        role:
          defaultValue.isAdmin !== value.isAdmin
            ? value.isAdmin
              ? 'admin'
              : 'user'
            : null,
      },
    },
  })

  const userHasBasicAuth = (data?.user?.authSubjects ?? []).some(
    (s) => s.providerID === 'basic',
  )

  // Checks if any of the password fields are used. Used to skip any unnecessary updateUserMutation
  function passwordChanged(): boolean {
    return Boolean(
      value.oldPassword || value.password || value.confirmNewPassword,
    )
  }

  // Validates inputs to the newPassword and confirmNewPassword fields
  function handleValidation(): FieldError[] {
    let err: FieldError[] = []
    if (!passwordChanged()) return err
    if (value.password !== value.confirmNewPassword) {
      err = [
        ...err,
        {
          field: 'confirmNewPassword',
          message: 'Passwords do not match',
        } as FieldError,
      ]
    }
    if (!userHasBasicAuth && !value.username) {
      err = [
        ...err,
        {
          field: 'username',
          message: 'Username required',
        } as FieldError,
      ]
    }
    return err
  }

  // wrapper function to handle errors caught while executing useMutation Promises
  function errorHandler(
    caughtError: unknown,
    errorList: FieldError[],
  ): FieldError[] {
    if (caughtError instanceof ApolloError) {
      errorList = [...errorList, ...fieldErrors(caughtError)]
    } else {
      console.error(caughtError)
    }
    return errorList
  }

  // async wrapper function in order to await for useMutation Promises
  async function submitHandler(): Promise<void> {
    let errorList: FieldError[] = []
    errorList = [...errorList, ...handleValidation()]

    if (!errorList?.length && passwordChanged() && userHasBasicAuth) {
      try {
        await editBasicAuth()
      } catch (err) {
        console.error(err)
        errorList = errorHandler(err, errorList)
      }
    }

    if (!errorList?.length && passwordChanged() && !userHasBasicAuth) {
      try {
        await createBasicAuth()
      } catch (err) {
        console.error(err)
        errorList = errorHandler(err, errorList)
      }
    }

    if (!errorList?.length && defaultValue.isAdmin !== value.isAdmin) {
      try {
        await editUser()
      } catch (err) {
        console.error(err)
        errorList = errorHandler(err, errorList)
      }
    }

    setErrors(errorList)
    if (!errorList?.length) {
      props.onClose()
    }
  }

  const notices: object[] = []
  if (
    props.role === 'admin' &&
    props.userID === currentUserID &&
    !value.isAdmin
  ) {
    notices.push({
      type: 'WARNING',
      message: 'Updating role to User',
      details:
        'If you remove your admin privileges you will need to log in as a different admin to restore them.',
    })
  }
  if (authDisableBasic) {
    notices.push({
      type: 'WARNING',
      message: 'Basic Auth is Disabled',
      details: 'Password authentication is currently disabled.',
    })
  }

  if (!isSessionReady) return <Spinner />

  return (
    <FormDialog
      title='Edit User Access'
      loading={editBasicAuthStatus.loading || editUserStatus.loading || loading}
      errors={[
        ...nonFieldErrors(editBasicAuthStatus.error),
        ...nonFieldErrors(editUserStatus.error),
        ...nonFieldErrors(createBasicAuthStatus.error),
      ]}
      onClose={props.onClose}
      onSubmit={submitHandler}
      notices={notices}
      form={
        <UserEditForm
          value={value}
          errors={errors}
          isAdmin={currentUserAdmin}
          disabled={!!authDisableBasic && !currentUserAdmin}
          requireOldPassword={
            props.userID === currentUserID && userHasBasicAuth
          }
          hasUsername={userHasBasicAuth}
          onChange={(value) => {
            setValue(value)
          }}
        />
      }
    />
  )
}

export default UserEditDialog
