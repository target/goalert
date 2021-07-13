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
          return commit({
            variables: {
              // only pass 'name'
              input: {
                ...pick(value, 'name'),
                id: contactMethodID,
              },
            },
          })
        }}
        form={
          <UserContactMethodForm
            errors={fieldErrs}
            disabled={loading}
            edit={true}
            value={value || defaultValue}
            onChange={(value) => setValue(value)}
          />
        }
      />
    )
  }

  function renderMutation({ name, type, value }) {
    return (
      <Mutation mutation={mutation} onCompleted={onClose}>
        {(commit, status) =>
          renderDialog(commit, status, { name, type, value })
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
