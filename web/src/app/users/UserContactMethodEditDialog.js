import React, { useState } from 'react'
import { gql, useMutation, useQuery } from 'urql'
import p from 'prop-types'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import { pick } from 'lodash'

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

export default function UserContactMethodEditDialog({
  onClose,
  contactMethodID,
}) {
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
  const [value, setValue] = useState(null)

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
        return commit({
          variables: {
            input: {
              ...updates,
              id: contactMethodID,
            },
          },
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

UserContactMethodEditDialog.propTypes = {
  contactMethodID: p.string.isRequired,
  onClose: p.func,
}
