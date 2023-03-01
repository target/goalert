import React, { useState } from 'react'
import { gql } from '@apollo/client'
import p from 'prop-types'
import { Mutation } from '@apollo/client/react/components'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import UserContactMethodForm from './UserContactMethodForm'
import Query from '../util/Query'
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
  const [value, setValue] = useState(null)

  function renderDialog(commit, status, defaultValue) {
    const { loading, error } = status
    const fieldErrs = fieldErrors(error)

    return (
      <FormDialog
        title='Edit Contact Method'
        loading={loading}
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
            disabled={loading}
            edit
            value={value || defaultValue}
            onChange={(value) => setValue(value)}
          />
        }
      />
    )
  }

  function renderMutation({ name, type, value, statusUpdates }) {
    return (
      <Mutation mutation={mutation} onCompleted={onClose}>
        {(commit, status) =>
          renderDialog(commit, status, { name, type, value, statusUpdates })
        }
      </Mutation>
    )
  }

  return (
    <Query
      query={query}
      variables={{ id: contactMethodID }}
      render={({ data }) => renderMutation(data.userContactMethod)}
      noPoll
    />
  )
}

UserContactMethodEditDialog.propTypes = {
  contactMethodID: p.string.isRequired,
  onClose: p.func,
}
