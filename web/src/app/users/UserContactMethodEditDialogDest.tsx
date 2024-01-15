import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserContactMethodFormDest from './UserContactMethodFormDest'
import { pick } from 'lodash'
import { useQuery } from 'urql'
import { ContactMethodType, StatusUpdateState } from '../../schema'

const query = gql`
  query ($id: ID!) {
    userContactMethod(id: $id) {
      id
      name
      type
      value
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
  type: ContactMethodType
  value: string
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
  const [{ data, fetching }] = useQuery({
    query,
    variables: { id: contactMethodID },
  })
  const [commit, status] = useMutation(mutation)
  const { error } = status

  const defaultValue = {
    name: data.userContactMethod.name,
    type: data.userContactMethod.type,
    value: data.userContactMethod.value,
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
        <UserContactMethodFormDest
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
