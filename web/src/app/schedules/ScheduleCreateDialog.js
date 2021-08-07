import React, { useState } from 'react'
import { gql } from '@apollo/client'
import FormDialog from '../dialogs/FormDialog'
import ScheduleForm from './ScheduleForm'
import { Mutation } from '@apollo/client/react/components'
import { nonFieldErrors, fieldErrors } from '../util/errutil'
import { Redirect } from 'react-router'

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

export default function ScheduleCreateDialog(props) {
  const [value, setValue] = useState({
    name: '',
    description: '',
    timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
    favorite: true,
  })

  function renderForm(commit, status) {
    if (status.data && status.data.createSchedule) {
      return (
        <Redirect push to={`/schedules/${status.data.createSchedule.id}`} />
      )
    }

    return (
      <FormDialog
        onClose={props.onClose}
        title='Create New Schedule'
        errors={nonFieldErrors(status.error)}
        onSubmit={() =>
          commit({
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
            disabled={status.loading}
            errors={fieldErrors(status.error)}
            value={value}
            onChange={(value) => setValue(value)}
          />
        }
      />
    )
  }
  return <Mutation mutation={mutation}>{renderForm}</Mutation>
}
