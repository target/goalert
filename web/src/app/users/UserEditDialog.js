import React, { useState } from 'react'
import { gql, useQuery, useMutation } from '@apollo/client'

import p from 'prop-types'

import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import UserForm from './UserForm'
import _ from 'lodash'

const query = gql`
  query user($id: ID!) {
    user(id: $id) {
      name  
      role
    }
  }
`
const mutation = gql`
  mutation updateUser($input: UpdateUserInput!) {
    updateUser(input: $input)
  }
`

export default function UserEditDialog({ userID, onClose, isAdmin }) {
  const [value, setValue] = useState(null)
  const { data, ...dataStatus } = useQuery(query, {
    variables: { id: userID },
  })
  const [save, saveStatus] = useMutation(mutation, {
    variables: { input: { ...value, id: userID } },
    onCompleted: onClose,
  })

   const defaults = {
    // default value is the user name & role
     ..._.chain(data).get('user').pick(['name', 'role']).value(),
     isAdmin: isAdmin,
  }

  const fieldErrs = fieldErrors(saveStatus.error)

  return (
    <FormDialog
      title='Edit User'
      loading={saveStatus.loading || (!data && dataStatus.loading)}
      errors={nonFieldErrors(saveStatus.error).concat(
        nonFieldErrors(dataStatus.error),
      )}
      onClose={onClose}
      onSubmit={() => save()}
      form={
        <UserForm
          errors={fieldErrs}
          disabled={Boolean(
            saveStatus.loading ||
              (!data && dataStatus.loading) ||
              dataStatus.error,
          )}
          value={value || defaults}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
UserEditDialog.propTypes = {
  userID: p.string.isRequired,
  onClose: p.func,
  isAdmin: p.bool,
}