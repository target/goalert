import React from 'react'
import { gql, useMutation } from 'urql'
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
}): JSX.Element {
  const { ruleID, ...rest } = props

  const [{ fetching, error }, deleteNotification] = useMutation(mutation)

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      loading={fetching}
      errors={nonFieldErrors(error)}
      subTitle='This will delete the notification rule.'
      onSubmit={() =>
        deleteNotification(
          { id: ruleID },
          { additionalTypenames: ['UserNotificationRule'] },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }
      {...rest}
    />
  )
}
