import React, { useState } from 'react'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodFormDest'
import { gql, useMutation, useQuery } from 'urql'
import { DestinationInput, UserContactMethod } from '../../schema'

const query = gql`
  query UserContactMethod($id: ID!) {
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
  mutation UpdateUserContactMethod($input: UpdateUserContactMethodInput!) {
    updateUserContactMethod(input: $input)
  }
`

type Value = {
  name: string
  dest: DestinationInput
  statusUpdates: boolean
}

export default function UserContactMethodEditDialogDest({
  onClose,
  contactMethodID,
  disablePortal,
}: {
  onClose: () => void
  contactMethodID: string
  disablePortal?: boolean
}): React.ReactNode {
  const [{ data }] = useQuery<{
    userContactMethod: UserContactMethod
  }>({
    query,
    variables: { id: contactMethodID },
  })
  if (!data) throw new Error('no data') // shouldn't happen since we're using suspense
  const [value, setValue] = useState<Value>({
    name: data.userContactMethod.name,
    dest: data.userContactMethod.dest,
    statusUpdates:
      data.userContactMethod.statusUpdates === 'ENABLED' ||
      data.userContactMethod.statusUpdates === 'ENABLED_FORCED',
  })
  const [status, commit] = useMutation(mutation)
  const { error, fetching } = status
  if (!data) throw new Error('no data') // shouldn't happen since we're using suspense

  const fieldErrs = fieldErrors(error)

  console.log(onClose)
  console.log(contactMethodID)

  return (
    <FormDialog
      title='Edit Contact Method'
      loading={fetching}
      disablePortal={disablePortal}
      errors={nonFieldErrors(error)}
      onClose={onClose}
      onSubmit={() => {
        commit(
          {
            input: {
              name: value.name,
              enableStatusUpdates: value.statusUpdates,
              id: contactMethodID,
            },
          },
          { additionalTypenames: ['UserContactMethod'] },
        ).then((result) => {
          if (result.error) return
          onClose()
        })
      }}
      form={
        <UserContactMethodForm
          errors={fieldErrs}
          disabled={fetching}
          edit
          value={value}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
