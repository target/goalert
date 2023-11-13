import React from 'react'
import { useMutation, gql } from '@apollo/client'
import FormDialog from '../../dialogs/FormDialog'
import { fmt, TempSchedValue } from './sharedUtils'
import { DateTime } from 'luxon'

const mutation = gql`
  mutation ($input: ClearTemporarySchedulesInput!) {
    clearTemporarySchedules(input: $input)
  }
`

type TempSchedDeleteConfirmationProps = {
  scheduleID: string
  onClose: () => void
  value: TempSchedValue
}

export default function TempSchedDeleteConfirmation({
  scheduleID,
  onClose,
  value,
}: TempSchedDeleteConfirmationProps): React.ReactNode {
  const [deleteTempSchedule, { loading, error }] = useMutation(mutation, {
    onCompleted: () => onClose(),
    variables: {
      input: {
        scheduleID,
        start: value.start, // actual truncation will be handled by backend
        end: value.end,
      },
    },
  })

  const start = DateTime.max(
    DateTime.fromISO(value.start),
    DateTime.utc(),
  ).toISO()

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will clear all temporary schedule data from ${fmt(
        start,
      )} to ${fmt(value.end)}.`}
      loading={loading}
      errors={error ? [error] : []}
      onClose={onClose}
      onSubmit={() => deleteTempSchedule()}
    />
  )
}
