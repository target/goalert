import React, { useState } from 'react'
import { gql } from '@apollo/client'
import p from 'prop-types'
import { Mutation } from '@apollo/client/react/components'
import FormDialog from '../dialogs/FormDialog'
import ScheduleOverrideForm from './ScheduleOverrideForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import Query from '../util/Query'

const query = gql`
  query ($id: ID!) {
    userOverride(id: $id) {
      id
      start
      end
      target {
        id
      }
      addUser {
        id
      }
      removeUser {
        id
      }
    }
  }
`
const mutation = gql`
  mutation ($input: UpdateUserOverrideInput!) {
    updateUserOverride(input: $input)
  }
`
export default function ScheduleOverrideEditDialog(props) {
  const { onClose, overrideID } = props
  const [value, setValue] = useState(null)

  function getValue(data) {
    if (value) return value
    const newValue = {
      start: data.start,
      end: data.end,
    }

    newValue.addUserID = data.addUser ? data.addUser.id : ''
    newValue.removeUserID = data.removeUser ? data.removeUser.id : ''

    return newValue
  }

  function renderDialog(data, commit, status) {
    return (
      <FormDialog
        onClose={onClose}
        title='Edit Schedule Override'
        errors={nonFieldErrors(status.error)}
        onSubmit={() => {
          if (value === null) {
            onClose()
            return
          }
          commit({
            variables: {
              input: {
                ...value,
                id: overrideID,
              },
            },
          })
        }}
        form={
          <ScheduleOverrideForm
            add={Boolean(data.addUser)}
            remove={Boolean(data.removeUser)}
            scheduleID={data.target.id}
            disabled={status.loading}
            errors={fieldErrors(status.error)}
            value={getValue(data)}
            onChange={(value) => setValue(value)}
          />
        }
      />
    )
  }

  function renderMutation(data) {
    return (
      <Mutation mutation={mutation} onCompleted={onClose}>
        {(commit, status) => renderDialog(data, commit, status)}
      </Mutation>
    )
  }

  return (
    <Query
      query={query}
      variables={{ id: overrideID }}
      noPoll
      fetchPolicy='network-only'
      render={({ data }) => renderMutation(data.userOverride)}
    />
  )
}

ScheduleOverrideEditDialog.propTypes = {
  overrideID: p.string.isRequired,
  onClose: p.func,
}
