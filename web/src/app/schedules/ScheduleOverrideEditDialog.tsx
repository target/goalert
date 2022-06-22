import React, { useState } from 'react'
import { gql, useQuery } from 'urql'
import { useMutation } from '@apollo/client'
import FormDialog from '../dialogs/FormDialog'
import ScheduleOverrideForm from './ScheduleOverrideForm'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import { UserOverride } from '../../schema'

interface Value {
  start: string
  end: string
  addUserID: string
  removeUserID: string
}

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
export default function ScheduleOverrideEditDialog(props: {
  overrideID: string
  onClose: () => void
}): JSX.Element {
  const [value, setValue] = useState<Value | null>(null)

  const [{ data, fetching, error }] = useQuery({
    query: query,
    variables: { id: props.overrideID },
    requestPolicy: 'network-only',
  })

  const [updateOverride, updateOverrideStatus] = useMutation(mutation, {
    onCompleted: props.onClose,
  })

  if (error) {
    return <GenericError error={error.message} />
  }

  if (fetching && !data) {
    return <Spinner />
  }

  function getValue(userOverride: UserOverride): Value {
    if (value) return value
    const newValue: Value = {
      start: userOverride.start,
      end: userOverride.end,
      addUserID: userOverride.addUser ? userOverride.addUser.id : '',
      removeUserID: userOverride.removeUser ? userOverride.removeUser.id : '',
    }

    return newValue
  }

  return (
    <FormDialog
      onClose={props.onClose}
      title='Edit Schedule Override'
      errors={nonFieldErrors(error)}
      onSubmit={() => {
        if (value === null) {
          props.onClose()
          return
        }

        updateOverride({
          variables: {
            input: {
              ...value,
              id: props.overrideID,
            },
          },
        })
      }}
      form={
        <ScheduleOverrideForm
          add={Boolean(data.userOverride.addUser)}
          remove={Boolean(data.userOverride.removeUser)}
          scheduleID={data.userOverride.target.id}
          disabled={updateOverrideStatus.loading}
          errors={fieldErrors(error)}
          value={getValue(data.userOverride)}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
