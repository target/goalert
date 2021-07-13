import React, { useState } from 'react'
import { gql } from '@apollo/client'
import p from 'prop-types'
import FormDialog from '../dialogs/FormDialog'
import ScheduleForm from './ScheduleForm'
import { Mutation } from '@apollo/client/react/components'
import { nonFieldErrors, fieldErrors } from '../util/errutil'
import Query from '../util/Query'

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

export default function ScheduleEditDialog(props) {
  const [value, setValue] = useState(null)

  const renderForm = (data, commit, status) => {
    return (
      <FormDialog
        onClose={props.onClose}
        title='Edit Schedule'
        errors={nonFieldErrors(status.error)}
        onSubmit={() =>
          commit({
            variables: {
              input: {
                id: props.scheduleID,
                ...value,
              },
            },
          })
        }
        form={
          <ScheduleForm
            disabled={status.loading}
            errors={fieldErrors(status.error)}
            value={
              value || {
                name: data.name,
                description: data.description,
                timeZone: data.timeZone,
              }
            }
            onChange={(value) => setValue(value)}
          />
        }
      />
    )
  }

  const renderMutation = (data) => {
    return (
      <Mutation mutation={mutation} onCompleted={props.onClose}>
        {(...args) => renderForm(data, ...args)}
      </Mutation>
    )
  }

  return (
    <Query
      query={query}
      variables={{ id: props.scheduleID }}
      render={({ data }) => renderMutation(data.schedule)}
    />
  )
}

ScheduleEditDialog.propTypes = {
  scheduleID: p.string.isRequired,
  onClose: p.func,
}
