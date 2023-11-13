import React from 'react'
import { gql, useMutation } from '@apollo/client'
import FormDialog from '../dialogs/FormDialog'
import { nonFieldErrors } from '../util/errutil'

const mutation = gql`
  mutation ($id: ID!) {
    deleteAll(input: [{ id: $id, type: notificationRule }])
  }
`
export default function UserNotificationRuleDeleteDialog(props: {
  ruleID: string
  onClose: () => void
}): React.ReactNode {
  const { ruleID, ...rest } = props

  const [deleteNotification, { loading, error }] = useMutation(mutation, {
    onCompleted: props.onClose,
  })

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      loading={loading}
      errors={nonFieldErrors(error)}
      subTitle='This will delete the notification rule.'
      onSubmit={() => deleteNotification({ variables: { id: ruleID } })}
      {...rest}
    />
  )
}
