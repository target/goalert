import React, { useState } from 'react'
import { useMutation } from '@apollo/client'
import { gql } from 'urql'
import FormDialog from '../dialogs/FormDialog'
import ScheduleForm, { Value } from './ScheduleForm'
import { nonFieldErrors, fieldErrors } from '../util/errutil'
import { Redirect } from 'wouter'

const mutation = gql`
  mutation ($input: CreateScheduleInput!) {
    createSchedule(input: $input) {
      id
      name
      description
      timeZone
    }
  }
`

export default function ScheduleCreateDialog(props: {
  onClose: () => void
}): React.ReactNode {
  const [value, setValue] = useState<Value>({
    name: '',
    description: '',
    timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    favorite: true,
  })

  const [createSchedule, { loading, data, error }] = useMutation(mutation)

  if (data && data.createSchedule) {
    return <Redirect to={`/schedules/${data.createSchedule.id}`} />
  }

  return (
    <FormDialog
      onClose={props.onClose}
      title='Create New Schedule'
      errors={nonFieldErrors(error)}
      onSubmit={() =>
        createSchedule({
          variables: {
            input: {
              ...value,
              targets: [
                {
                  target: { type: 'user', id: '__current_user' },
                  rules: [{}],
                },
              ],
            },
          },
        })
      }
      form={
        <ScheduleForm
          disabled={loading}
          errors={fieldErrors(error)}
          value={value}
          onChange={(value: Value) => setValue(value)}
        />
      }
    />
  )
}
