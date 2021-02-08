import React from 'react'
import gql from 'graphql-tag'
import { useMutation } from '@apollo/client'
import FormDialog from '../../dialogs/FormDialog'
import { fmt, Value } from './sharedUtils'

const mutation = gql`
  mutation($input: ClearTemporarySchedulesInput!) {
    clearTemporarySchedules(input: $input)
  }
`

type TempSchedDeleteConfirmationProps = {
  scheduleID: string
  onClose: () => void
  value: Value
}

export default function TempSchedDeleteConfirmation({
  scheduleID,
  onClose,
  value,
}: TempSchedDeleteConfirmationProps): JSX.Element {
  const [deleteTempSchedule, { loading, error }] = useMutation(mutation, {
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
      subTitle={`This will clear all temporary schedule data from ${fmt(
        value.start,
      )} to ${fmt(value.end)}.`}
      loading={loading}
      errors={error ? [error] : []}
      onClose={onClose}
      onSubmit={() => deleteTempSchedule()}
    />
  )
}
