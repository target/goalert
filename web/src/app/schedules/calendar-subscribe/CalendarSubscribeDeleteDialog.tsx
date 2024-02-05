import React, { ReactNode } from 'react'
import { useMutation, gql } from 'urql'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'

const mutation = gql`
  mutation ($id: ID!) {
    deleteAll(input: [{ id: $id, type: calendarSubscription }])
  }
`

interface CalendarSubscribeDeleteDialogProps {
  calSubscriptionID: string
  onClose: () => void
}

export default function CalendarSubscribeDeleteDialog(
  props: CalendarSubscribeDeleteDialogProps,
): ReactNode {
  const [status, commit] = useMutation(mutation)

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      loading={status.fetching}
      errors={nonFieldErrors(status.error)}
      subTitle='This will delete the calendar subscription.'
      onSubmit={() =>
        commit(
          {
            id: props.calSubscriptionID,
          },
          { additionalTypenames: ['User'] },
        ).then((result) => {
          if (!result.error) props.onClose()
        })
      }
      onClose={props.onClose}
    />
  )
}
