import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodFormDest'
import { pick } from 'lodash'
import { useQuery } from 'urql'
import {
  DestinationInput,
  StatusUpdateState,
  UserContactMethod,
} from '../../schema'

const query = gql`
  query ($id: ID!) {
    userContactMethod(id: $id) {
      id
      name
      dest {
        type
        values {
          fieldID
          value
        }
      }
      statusUpdates
    }
  }
`

const mutation = gql`
  mutation ($input: UpdateUserContactMethodInput!) {
    updateUserContactMethod(input: $input)
  }
`

type Value = {
  name: string
  dest: DestinationInput
  statusUpdates?: StatusUpdateState
}

export default function UserContactMethodEditDialogDest({
  onClose,
  contactMethodID,
}: {
  onClose: () => void
  contactMethodID: string
}): React.ReactNode {
  const [value, setValue] = useState<Value | null>(null)
  const [{ data, fetching }] = useQuery<{
    userContactMethod: UserContactMethod
  }>({
    query,
    variables: { id: contactMethodID },
  })
  const [commit, status] = useMutation(mutation)
  const { error } = status
  if (!data) throw new Error('no data') // shouldn't happen since we're using suspense

  const defaultValue = {
    name: data.userContactMethod.name,
    dest: data.userContactMethod.dest,
    statusUpdates: data.userContactMethod.statusUpdates,
  }

  const fieldErrs = fieldErrors(error)

  return (
    <FormDialog
      title='Edit Contact Method'
      loading={fetching}
      errors={nonFieldErrors(error)}
      onClose={onClose}
      onSubmit={() => {
        const updates = pick(value, 'name', 'statusUpdates')
        // the form uses the 'statusUpdates' enum but the mutation simply
        // needs to know if the status updates should be enabled or not via
        // the 'enableStatusUpdates' boolean
        if ('statusUpdates' in updates) {
          delete Object.assign(updates, {
            enableStatusUpdates: updates.statusUpdates === 'ENABLED',
          }).statusUpdates
        }
        commit({
          variables: {
            input: {
              ...updates,
              id: contactMethodID,
            },
          },
        }).then((result) => {
          if (result.errors) return
          onClose()
        })
      }}
      form={
        <UserContactMethodForm
          errors={fieldErrs}
          disabled={fetching}
          edit
          value={value || defaultValue}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
