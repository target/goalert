import React, { useState } from 'react'
import { gql } from '@apollo/client'

import p from 'prop-types'
import { Mutation } from '@apollo/client/react/components'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import UserNotificationRuleForm from './UserNotificationRuleForm'

const createMutation = gql`
  mutation ($input: CreateUserNotificationRuleInput!) {
    createUserNotificationRule(input: $input) {
      id
    }
  }
`

export default function UserNotificationRuleCreateDialog({ onClose, userID }) {
  const [value, setValue] = useState({ contactMethodID: '', delayMinutes: 0 })

  function renderDialog(commit, status) {
    const { loading, error } = status
    const fieldErrs = fieldErrors(error)

    return (
      <FormDialog
        title='Create New Notification Rule'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={onClose}
        onSubmit={() => {
          return commit({
            variables: {
              input: { ...value, userID: userID },
            },
          })
        }}
        form={
          <UserNotificationRuleForm
            userID={userID}
            errors={fieldErrs}
            disabled={loading}
            value={value}
            onChange={(value) => setValue(value)}
          />
        }
      />
    )
  }

  return (
    <Mutation mutation={createMutation} onCompleted={onClose}>
      {(commit, status) => renderDialog(commit, status)}
    </Mutation>
  )
}

UserNotificationRuleCreateDialog.propTypes = {
  userID: p.string.isRequired,
  onClose: p.func,
}
