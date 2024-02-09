import React, { useState } from 'react'
import { gql, useMutation } from '@apollo/client'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import { pick } from 'lodash'
import { useQuery } from 'urql'
import { ContactMethodType, StatusUpdateState } from '../../schema'
import UserContactMethodEditDialogDest from './UserContactMethodEditDialogDest'
import { useExpFlag } from '../util/useExpFlag'

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

type UserContactMethodEditDialogProps = {
  onClose: () => void
  contactMethodID: string
}

function UserContactMethodEditDialog({
  onClose,
  contactMethodID,
}: UserContactMethodEditDialogProps): React.ReactNode {
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

export default function UserContactMethodEditDialogSwitch(
  props: UserContactMethodEditDialogProps,
): React.ReactNode {
  const isDestTypesSet = useExpFlag('dest-types')

  if (isDestTypesSet) {
    return <UserContactMethodEditDialogDest {...props} />
  }

  return <UserContactMethodEditDialog {...props} />
}
