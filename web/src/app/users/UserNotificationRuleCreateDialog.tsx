import React, { useState } from 'react'
import { useMutation } from '@apollo/client'
import { gql } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserNotificationRuleForm from './UserNotificationRuleForm'

const mutation = gql`
  mutation ($input: CreateUserNotificationRuleInput!) {
    createUserNotificationRule(input: $input) {
      id
    }
  }
`

export default function UserNotificationRuleCreateDialog(props: {
  onClose: () => void
  userID: string
}): React.ReactNode {
  const [value, setValue] = useState({ contactMethodID: '', delayMinutes: 0 })

  const [createNotification, { loading, error }] = useMutation(mutation, {
    onCompleted: props.onClose,
  })

  const fieldErrs = fieldErrors(error)

  return (
    <FormDialog
      title='Create New Notification Rule'
      loading={loading}
      errors={nonFieldErrors(error)}
      onSubmit={() => {
        return createNotification({
          variables: {
            input: { ...value, userID: props.userID },
          },
        })
      }}
      form={
        <UserNotificationRuleForm
          userID={props.userID}
          errors={fieldErrs}
          disabled={loading}
          value={value}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
