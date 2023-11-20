import React, { useState } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import FormDialog from '../dialogs/FormDialog'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import UserContactMethodForm from './UserContactMethodForm'
import { pick } from 'lodash'
import { ContactMethodType, StatusUpdateState } from '../../schema'

type Value = {
  name: string
  type: ContactMethodType
  value: string
  statusUpdates?: StatusUpdateState
}

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
function UserContactMethodEditDialog(props: {
  contactMethodID: string
  onClose: () => void
}): JSX.Element {
  const { contactMethodID, onClose } = props

  const [{ data, fetching }] = useQuery({
    query,
    variables: { id: contactMethodID },
  })

  const defaultValue = {
    name: data.userContactMethod.name,
    type: data.userContactMethod.type,
    value: data.userContactMethod.value,
    statusUpdates: data.userContactMethod.statusUpdates,
  }

  const [mutationStatus, commit] = useMutation(mutation)
  const { error } = mutationStatus
  const [value, setValue] = useState<Value | null>(null)

  const fieldErrs = fieldErrors(error)

  return (
    <FormDialog
      title='Edit Contact Method'
      loading={fetching}
      errors={nonFieldErrors(error)}
      onClose={onClose}
      onSubmit={() => {
        const updates = pick(value, 'name', 'statusUpdates') || {}
        // the form uses the 'statusUpdates' enum but the mutation simply
        // needs to know if the status updates should be enabled or not via
        // the 'enableStatusUpdates' boolean
        if ('statusUpdates' in updates) {
          delete Object.assign(updates, {
            enableStatusUpdates: updates.statusUpdates === 'ENABLED',
          }).statusUpdates
        }
        return commit({
          input: {
            ...updates,
            id: contactMethodID,
          },
        }).then((res) => {
          if (res.error) return
          props.onClose()
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

export default UserContactMethodEditDialog
