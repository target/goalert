import React, { useState } from 'react'
import { gql, useQuery, useMutation } from 'urql'
import FormDialog from '../dialogs/FormDialog'
import ScheduleForm, { Value } from './ScheduleForm'
import { GenericError } from '../error-pages'
import Spinner from '../loading/components/Spinner'

const query = gql`
  query ($id: ID!) {
    schedule(id: $id) {
      id
      name
      description
      timeZone
    }
  }
`

const mutation = gql`
  mutation ($input: UpdateScheduleInput!) {
    updateSchedule(input: $input)
  }
`

export default function ScheduleEditDialog(props: {
  scheduleID: string
  onClose: () => void
}): JSX.Element {
  const [value, setValue] = useState<Value | null>(null)

  const [{ data, error, fetching }] = useQuery({
    query,
    variables: {
      id: props.scheduleID,
    },
  })

  const [, editSchedule] = useMutation(mutation)

  if (error) {
    return <GenericError error={error.message} />
  }

  if (fetching && !data) {
    return <Spinner />
  }

  return (
    <FormDialog
      onClose={props.onClose}
      title='Edit Schedule'
      onSubmit={() =>
        editSchedule(
          {
            input: {
              id: props.scheduleID,
              ...value,
            },
          },
          { additionalTypenames: ['Schedule'] },
        ).then(() => props.onClose())
      }
      form={
        <ScheduleForm
          value={
            value || {
              name: data.schedule.name,
              description: data.schedule.description,
              timeZone: data.schedule.timeZone,
            }
          }
          onChange={(value: Value) => setValue(value)}
        />
      }
    />
  )
}
