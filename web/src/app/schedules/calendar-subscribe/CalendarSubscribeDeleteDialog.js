import React from 'react'
import { useMutation, gql } from '@apollo/client'
import { PropTypes as p } from 'prop-types'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'

const mutation = gql`
  mutation($id: ID!) {
    deleteAll(input: [{ id: $id, type: calendarSubscription }])
  }
`

export default function CalendarSubscribeDeleteDialog(props) {
  const [deleteSubscription, status] = useMutation(mutation, {
    variables: {
      id: props.calSubscriptionID,
    },
    onCompleted: props.onClose,
  })

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      loading={status.loading}
      errors={nonFieldErrors(status.error)}
      subTitle='This will delete the calendar subscription.'
      onSubmit={deleteSubscription}
      onClose={props.onClose}
    />
  )
}

CalendarSubscribeDeleteDialog.propTypes = {
  calSubscriptionID: p.string.isRequired,
  onClose: p.func.isRequired,
}
