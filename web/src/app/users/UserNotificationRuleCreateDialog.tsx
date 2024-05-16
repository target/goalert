import React, { useState } from 'react'
import { gql, useMutation } from 'urql'
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
}): JSX.Element {
  const [value, setValue] = useState({ contactMethodID: '', delayMinutes: 0 })

  const [{ fetching, error }, createNotification] = useMutation(mutation)

  const fieldErrs = fieldErrors(error)

  return (
    <FormDialog
      title='Create New Notification Rule'
      loading={fetching}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() => {
        createNotification(
          {
            input: { ...value, userID: props.userID },
          },
          {
            additionalTypenames: [
              'UserNotificationRule',
              'UserContactMethod',
              'User',
            ],
          },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }}
      form={
        <UserNotificationRuleForm
          userID={props.userID}
          errors={fieldErrs}
          disabled={fetching}
          value={value}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
