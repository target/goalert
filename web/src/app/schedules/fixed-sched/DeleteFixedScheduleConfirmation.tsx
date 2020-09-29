import React from 'react'
import gql from 'graphql-tag'
import { useMutation } from '@apollo/react-hooks'
import FormDialog from '../../dialogs/FormDialog'
import { fmt, Value } from './sharedUtils'

const mutation = gql`
  mutation($input: ResetScheduleShiftsInput!) {
    resetScheduleShifts(input: $input)
  }
`

type DeleteFixedScheduleConfirmationProps = {
  scheduleID: string
  onClose: () => void
  value: Value
}

export default function DeleteFixedScheduleConfirmation({
  scheduleID,
  onClose,
  value,
}: DeleteFixedScheduleConfirmationProps) {
  const [deleteFixedSchedule, { loading, error }] = useMutation(mutation, {
    onCompleted: () => onClose(),
    variables: {
      input: {
        scheduleID: scheduleID,
        start: value.start,
        end: value.end,
      },
    },
  })

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`Deleting this fixed schedule will remove all fixed shifts from ${fmt(
        value.start,
      )} to ${fmt(value.end)}.`}
      loading={loading}
      errors={error ? [error] : []}
      onClose={onClose}
      onSubmit={() => deleteFixedSchedule()}
    />
  )
}
